package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/ledongthuc/pdf"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"milvus-kb-demo/internal/agent"
	"milvus-kb-demo/internal/auth"
	"milvus-kb-demo/internal/history"
	"milvus-kb-demo/internal/llm"
	"milvus-kb-demo/internal/mcpserver"
	"milvus-kb-demo/internal/rag"
	"milvus-kb-demo/internal/search"
	"milvus-kb-demo/internal/server"
	"milvus-kb-demo/internal/service"
)

// --- 全局配置 ---
var (
	ZillizEndpoint string
	ZillizApiKey   string
	ZhipuApiKey    string
	DeepSeekApiKey string
	CollectionName string
	PdfPath        string
	MCPConfigPath  string
)

const (
	ChunkSize    = 500
	ChunkOverlap = 50
)

// --- 初始化 ---
func initConfig() {
	_ = godotenv.Load() // 加载 .env，忽略错误（允许直接用系统环境变量）

	ZillizEndpoint = os.Getenv("ZILLIZ_ENDPOINT")
	ZillizApiKey = os.Getenv("ZILLIZ_API_KEY")
	ZhipuApiKey = os.Getenv("ZHIPU_API_KEY")
	DeepSeekApiKey = os.Getenv("DEEPSEEK_API_KEY")
	CollectionName = os.Getenv("COLLECTION_NAME")
	PdfPath = os.Getenv("PDF_PATH")
	MCPConfigPath = "internal/mcpserver/mcp_config.json"

	if CollectionName == "" {
		CollectionName = "student_handbook_kb"
	}
	if PdfPath == "" {
		PdfPath = "./handbook.pdf"
	}

	if ZillizEndpoint == "" || ZhipuApiKey == "" {
		log.Fatal("错误: 请确保 .env 文件已正确配置 ZILLIZ_ENDPOINT 和 ZHIPU_API_KEY")
	}
}

// --- 核心工具函数 ---

// 1. PDF 解析
func ParsePDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var totalText strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, _ := p.GetPlainText(nil)
		totalText.WriteString(text)
	}
	return totalText.String(), nil
}

// 2. 切片
func SplitText(text string, chunkSize int, overlap int) []string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.Join(strings.Fields(text), " ")
	contentRunes := []rune(text)
	var chunks []string

	for i := 0; i < len(contentRunes); i += (chunkSize - overlap) {
		end := i + chunkSize
		if end > len(contentRunes) {
			end = len(contentRunes)
		}
		chunk := string(contentRunes[i:end])
		if len(chunk) > 10 {
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}

// --- 业务流程 ---

// 流程 A: 数据入库 (Ingestion)
func runIngest(ctx context.Context, c client.Client, llmClient *llm.Client) {
	fmt.Printf(">>> [模式: 上传] 正在读取 PDF: %s ...\n", PdfPath)
	rawText, err := ParsePDF(PdfPath)
	if err != nil {
		log.Fatalf("读取PDF失败: %v", err)
	}

	chunks := SplitText(rawText, ChunkSize, ChunkOverlap)
	fmt.Printf(">>> 提取文本 %d 字，切分为 %d 段\n", len([]rune(rawText)), len(chunks))

	// 重建集合
	has, _ := c.HasCollection(ctx, CollectionName)
	if has {
		fmt.Println(">>> 删除旧集合...")
		c.DropCollection(ctx, CollectionName)
	}

	schema := entity.NewSchema().WithName(CollectionName).WithAutoID(false).
		WithField(entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true)).
		WithField(entity.NewField().WithName("vector").WithDataType(entity.FieldTypeFloatVector).WithDim(rag.Dim)).
		WithField(entity.NewField().WithName("content").WithDataType(entity.FieldTypeVarChar).WithMaxLength(65535))

	err = c.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		log.Fatal("创建集合失败:", err)
	}

	fmt.Println(">>> 开始向量化并入库...")
	var ids []int64
	var vectors [][]float32
	var contents []string

	for i, chunk := range chunks {
		if i%5 == 0 {
			fmt.Printf("处理进度: %d/%d\n", i, len(chunks))
		}
		vec, embedErr := llmClient.Embed(ctx, chunk)
		if embedErr != nil {
			continue
		}

		ids = append(ids, time.Now().UnixNano()+int64(i))
		vectors = append(vectors, vec)
		contents = append(contents, chunk)
		time.Sleep(100 * time.Millisecond) // 限流保护
	}

	_, err = c.Insert(ctx, CollectionName, "",
		entity.NewColumnInt64("id", ids),
		entity.NewColumnFloatVector("vector", rag.Dim, vectors),
		entity.NewColumnVarChar("content", contents))
	if err != nil {
		log.Fatal("插入数据失败:", err)
	}

	// 创建索引
	idx := entity.NewGenericIndex("auto_idx", entity.IndexType("AUTOINDEX"), map[string]string{"metric_type": "L2"})
	c.CreateIndex(ctx, CollectionName, "vector", idx, false)
	fmt.Println(">>> ✅ 入库完成！现在可以重新运行程序进入问答模式。")
}

// 流程 B: 问答交互 (Retrieval & Generation)
func runChat(ctx context.Context, c client.Client, svc *service.Service) {
	// 确保集合已加载
	err := c.LoadCollection(ctx, CollectionName, false)
	if err != nil {
		log.Fatalf("无法加载集合 (你是否执行过 -mode upload ?): %v", err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n🤖 知识库助手已启动！(输入 'exit' 退出)")
	fmt.Println("------------------------------------------------")

	for {
		fmt.Print("👉 请输入问题: ")
		question, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		question = strings.TrimSpace(question)
		if question == "exit" {
			break
		}
		if question == "" {
			continue
		}

		fmt.Println("⏳ 正在思考...")
		answer, err := svc.Answer(ctx, question)
		if err != nil {
			fmt.Println(" 生成失败:", err)
			continue
		}

		fmt.Printf("🤖 回答:\n%s\n", answer)
		fmt.Println("------------------------------------------------")
	}
}

// 流程 C: Web Server
func runWeb(ctx context.Context, c client.Client, svc *service.Service, userManager *auth.UserManager) {
	// 确保集合已加载
	err := c.LoadCollection(ctx, CollectionName, false)
	if err != nil {
		log.Printf("Warning: Failed to load collection: %v", err)
	}

	historyManager, err := history.NewHistoryManager("chat_history.json")
	if err != nil {
		log.Fatalf("Failed to initialize history manager: %v", err)
	}

	srv := server.NewServer(userManager, historyManager, svc, "./static")
	fmt.Println(">>> 🌐 启动 Web 服务器: http://localhost:8080")
	if err := srv.Run(":8080"); err != nil {
		log.Fatalf("Web Server Error: %v", err)
	}
}

// --- Main 入口 ---
func main() {
	initConfig()

	// 定义命令行参数
	mode := flag.String("mode", "chat", "运行模式: 'upload' (上传PDF), 'chat' (对话), 或 'web' (Web服务)")
	flag.Parse()

	ctx := context.Background()
	c, err := client.NewClient(ctx, client.Config{Address: ZillizEndpoint, APIKey: ZillizApiKey})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer c.Close()

	llmClient := llm.NewClient(ZhipuApiKey)

	if *mode == "upload" {
		runIngest(ctx, c, llmClient)
	} else {
		// 初始化 MCP Agent (Shared by Chat and Web mode)
		var mcpAgent *mcpserver.MCPAgent
		if DeepSeekApiKey != "" {
			var err error
			// 尝试初始化 MCP，如果失败不影响主流程，只是地图功能不可用
			mcpAgent, err = mcpserver.NewMCPAgent(MCPConfigPath, DeepSeekApiKey)
			if err != nil {
				fmt.Printf("⚠️ 警告: MCP Agent 初始化失败: %v (地图功能将不可用)\n", err)
			} else {
				fmt.Println(">>> ✅ MCP Agent (DeepSeek+Gaode) 已就绪")
			}
		} else {
			fmt.Println(">>> ℹ️ 未配置 DEEPSEEK_API_KEY，地图功能已禁用")
		}

		decisionAgent := agent.NewDecisionAgent(llmClient)
		retriever := rag.NewRetriever(c, CollectionName, llmClient)
		searcher := search.NewBingSearcher()

		svc := service.NewService(decisionAgent, retriever, searcher, llmClient, mcpAgent)

		if *mode == "web" {
			// Initialize Auth
			userManager, err := auth.NewUserManager("users.json")
			if err != nil {
				log.Fatalf("Failed to initialize user manager: %v", err)
			}
			runWeb(ctx, c, svc, userManager)
		} else {
			runChat(ctx, c, svc)
		}
	}
}
