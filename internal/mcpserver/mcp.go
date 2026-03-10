package mcpserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// ==========================================
// Part 1: 数据结构定义
// ==========================================

type Config struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

type ServerConfig struct {
	URL string `json:"url"`
}

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      int         `json:"id,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
	ID      int             `json:"id"`
}

type MCPTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ==========================================
// Part 2: 远程 SSE MCP Client
// ==========================================

type MCPHTTPClient struct {
	SSEUrl      string
	PostURL     string
	Client      *http.Client
	MsgID       int
	PendingReqs sync.Map
	IsReady     chan bool
}

func NewMCPHTTPClient(sseURL string) *MCPHTTPClient {
	return &MCPHTTPClient{
		SSEUrl:  sseURL,
		Client:  &http.Client{Timeout: 60 * time.Second}, // 给网络请求多一点时间
		IsReady: make(chan bool),
	}
}

// resolvePostURL 处理相对路径 URL
func (c *MCPHTTPClient) resolvePostURL(endpoint string) string {
	// 如果 Server 返回的是完整 URL (http开头)，直接使用
	if strings.HasPrefix(endpoint, "http") {
		return endpoint
	}

	// 如果是相对路径，需要基于 SSE URL 进行拼接
	// 解析原始 SSE URL
	u, err := url.Parse(c.SSEUrl)
	if err != nil {
		return endpoint // 降级处理
	}

	// 拼接路径
	// 注意：这里是一个简化的 URL 拼接，真实场景可能需要更严谨的处理
	// 假设 SSEUrl 是 https://mcp.amap.com/mcp?key=xxx
	// endpoint 是 /messages?key=xxx 或者只是 /messages

	// 简单策略：通常 MCP 实现会返回完整的 URL 或者基于域名的路径
	// 如果 endpoint 是 /api/message
	newURL := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, endpoint)

	// 如果原始 URL 有 query 参数 (比如 key=xxx)，通常 POST 接口也需要带上
	// 这里视 Server 实现而定。如果 endpoint 里没带参数，把原始参数补上
	if u.RawQuery != "" && !strings.Contains(endpoint, "?") {
		newURL = newURL + "?" + u.RawQuery
	}

	return newURL
}

func (c *MCPHTTPClient) Start() {
	go func() {
		fmt.Printf("正在连接: %s\n", c.SSEUrl)

		// 1. 构造 Cursor 同款的握手包
		initPayload := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "initialize",
			"params": map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"roots":    map[string]interface{}{"listChanged": true},
					"sampling": map[string]interface{}{},
				},
				"clientInfo": map[string]string{
					"name":    "Cursor",
					"version": "0.44.11", // 伪装成真实的 Cursor 版本
				},
			},
			"id": 0,
		}
		jsonBytes, _ := json.Marshal(initPayload)

		// 2. 发起 POST 请求
		req, _ := http.NewRequest("POST", c.SSEUrl, bytes.NewBuffer(jsonBytes))

		// 3. 设置 Headers
		req.Header.Set("Accept", "text/event-stream, application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
		req.Header.Set("Origin", "https://mcp.amap.com")

		resp, err := c.Client.Do(req)
		if err != nil {
			log.Printf("网络错误: %v", err)
			return
		}
		defer resp.Body.Close()

		// 4. 打印调试信息
		contentType := resp.Header.Get("Content-Type")
		fmt.Printf("✅ 服务器响应: %d | Type: %s\n", resp.StatusCode, contentType)

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("❌ 握手失败: %s", string(body))
			return
		}

		// === 核心修改：如果是 JSON，直接读取，不要当成 SSE 解析 ===
		if strings.Contains(contentType, "application/json") {
			fmt.Println("⚡ 检测到直接 JSON 响应 (非流式)...")
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("📦 收到握手响应: %s\n", string(body))

			// 既然握手成功了，通常下一步就是拿工具列表
			// 在这种非标准 SSE 模式下，POST 地址通常就是原地址
			c.PostURL = c.SSEUrl
			fmt.Printf("🔗 指令地址已锁定: %s\n", c.PostURL)

			// 标记就绪
			select {
			case <-c.IsReady:
			default:
				close(c.IsReady)
			}
			return // 任务完成，退出协程 (不需要长连接监听了)
		}

		// === 如果是 SSE 流，则进入循环读取 ===
		fmt.Println("🌊 检测到 SSE 流，开始监听...")
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					log.Printf("🔴 流断开: %v", err)
				} else {
					log.Println("🔴 流结束 (EOF)")
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// 解析 endpoint 事件
			if strings.HasPrefix(line, "event: endpoint") {
				// 读下一行 data
				dataLine, _ := reader.ReadString('\n')
				url := strings.TrimSpace(strings.TrimPrefix(dataLine, "data: "))
				c.PostURL = c.resolvePostURL(url)
				fmt.Printf("🔗 [SSE握手] 指令地址: %s\n", c.PostURL)
				select {
				case <-c.IsReady:
				default:
					close(c.IsReady)
				}
			}

			// 解析普通消息
			if strings.HasPrefix(line, "data:") {
				// ... 这里可以加数据解析逻辑 ...
			}
		}
	}()
}

func (c *MCPHTTPClient) Call(method string, params interface{}) (json.RawMessage, error) {
	// 等待握手完成
	select {
	case <-c.IsReady:
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("connection timeout: handshake failed")
	}

	c.MsgID++
	reqData := JSONRPCRequest{
		JSONRPC: "2.0", Method: method, Params: params, ID: c.MsgID,
	}

	// 1. 构造 Body
	payload, _ := json.Marshal(reqData)

	// 2. 创建请求
	req, err := http.NewRequest("POST", c.PostURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	// === 核心修复：Accept 必须完全满足服务器的变态要求 ===
	// 哪怕我们并不打算读流，也必须告诉服务器我们支持流，否则它不干
	req.Header.Set("Accept", "text/event-stream, application/json")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://mcp.amap.com")
	req.Header.Set("Referer", "https://mcp.amap.com/")
	req.Header.Set("Cache-Control", "no-cache")

	// 3. 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// 4. 读取结果
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		// 有时候服务器虽然 Accept 要求流，但返回的其实是 JSON
		// 如果这里报错，说明可能真的返回了流数据，但根据之前的经验，它大概率是 JSON
		return nil, fmt.Errorf("invalid json: %v | raw: %s", err, string(body))
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("MCP Error: %v", rpcResp.Error)
	}

	return rpcResp.Result, nil
}

// ==========================================
// Part 3: 管理器与主程序
// ==========================================

type ClientManager struct {
	Clients   map[string]*MCPHTTPClient
	toolDefs  map[string]MCPTool
	toolOwner map[string]*MCPHTTPClient
	mu        sync.RWMutex
}

func NewClientManager(configFile string) (*ClientManager, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	mgr := &ClientManager{
		Clients:   make(map[string]*MCPHTTPClient),
		toolDefs:  make(map[string]MCPTool),
		toolOwner: make(map[string]*MCPHTTPClient),
	}

	for name, serverCfg := range cfg.MCPServers {
		if serverCfg.URL == "" {
			continue
		}
		client := NewMCPHTTPClient(serverCfg.URL)
		client.Start()
		mgr.Clients[name] = client
	}
	return mgr, nil
}

func (m *ClientManager) ensureToolCache() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.toolDefs) > 0 {
		return nil
	}

	for _, client := range m.Clients {
		res, err := client.Call("tools/list", nil)
		if err != nil {
			log.Printf("获取工具失败: %v", err)
			continue
		}

		var listRes struct {
			Tools []MCPTool `json:"tools"`
		}
		json.Unmarshal(res, &listRes)

		for _, t := range listRes.Tools {
			m.toolDefs[t.Name] = t
			m.toolOwner[t.Name] = client
		}
	}

	if len(m.toolDefs) == 0 {
		return fmt.Errorf("no tools available")
	}

	return nil
}

func (m *ClientManager) LoadToolsByNames(names []string) ([]openai.Tool, error) {
	if err := m.ensureToolCache(); err != nil {
		return nil, err
	}

	var tools []openai.Tool
	if len(names) == 0 {
		m.mu.RLock()
		defer m.mu.RUnlock()
		for _, t := range m.toolDefs {
			tools = append(tools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.InputSchema,
				},
			})
		}
		return tools, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	missing := 0
	for _, name := range names {
		t, ok := m.toolDefs[name]
		if !ok {
			missing++
			continue
		}
		tools = append(tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}
	if len(tools) == 0 && missing > 0 {
		return nil, fmt.Errorf("tool not found: %s", names[0])
	}
	return tools, nil
}

func (m *ClientManager) Execute(name string, args string) (string, error) {
	if err := m.ensureToolCache(); err != nil {
		return "", err
	}

	m.mu.RLock()
	client, ok := m.toolOwner[name]
	m.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("tool not found")
	}

	var argMap map[string]interface{}
	json.Unmarshal([]byte(args), &argMap)

	res, err := client.Call("tools/call", map[string]interface{}{
		"name": name, "arguments": argMap,
	})
	if err != nil {
		return "", err
	}

	var contentRes struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	json.Unmarshal(res, &contentRes)

	var sb strings.Builder
	for _, c := range contentRes.Content {
		sb.WriteString(c.Text)
	}
	return sb.String(), nil
}

// ==========================================
// Part 4: Agent 封装
// ==========================================

type MCPAgent struct {
	Manager *ClientManager
	LLM     *openai.Client
	Tools   []openai.Tool
}

func NewMCPAgent(configFile, apiKey string) (*MCPAgent, error) {
	// 1. 加载 MCP 客户端 (从 json)
	fmt.Println("正在初始化 MCP 客户端...")
	mgr, err := NewClientManager(configFile)
	if err != nil {
		return nil, fmt.Errorf("配置加载失败: %v", err)
	}

	// 2. 启动 DeepSeek 对话
	// 注意：go-openai 库默认 base url 是 openai官网，deepseek 需修改
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.deepseek.com"
	llm := openai.NewClientWithConfig(config)

	return &MCPAgent{
		Manager: mgr,
		LLM:     llm,
	}, nil
}

func (a *MCPAgent) ensureTools(names []string) error {
	tools, err := a.Manager.LoadToolsByNames(names)
	if err != nil {
		if len(names) > 0 {
			tools, err = a.Manager.LoadToolsByNames(nil)
			if err == nil {
				a.Tools = tools
				return nil
			}
		}
		return err
	}
	a.Tools = tools
	return nil
}

func (a *MCPAgent) Process(ctx context.Context, query string) (string, error) {
	if err := a.ensureTools(nil); err != nil {
		return "", err
	}
	return a.processWithTools(ctx, query)
}

func (a *MCPAgent) ProcessWithTools(ctx context.Context, query string, toolNames []string) (string, error) {
	if err := a.ensureTools(toolNames); err != nil {
		return "", err
	}
	return a.processWithTools(ctx, query)
}

func (a *MCPAgent) processWithTools(ctx context.Context, query string) (string, error) {
	systemPrompt := `你是一个专注于“东莞理工学院（松山湖校区）”的智能校园出行助手。
你的核心职责是利用地图工具回答用户关于校园及周边的位置、导航和生活服务问题。

【重要上下文】
1. 默认位置：东莞理工学院松山湖校区（地址：广东省东莞市松山湖大学路1号）。
2. 当用户问“附近”、“周边”或“去xxx怎么走”且未指定起点时，默认起点/中心点均为“东莞理工学院松山湖校区”。
3. 搜索关键词策略：
   - 必须保留用户提到的特定约束（如“西餐”、“粤菜”、“星巴克”等）。不要将其简化为泛分类（如“餐厅”）。
   - 如果用户的问题包含具体品类，优先使用该品类作为 keywords 调用周边搜索工具。
4. 必须积极使用工具获取信息，不要直接拒绝对话或要求用户提供已知的位置信息。`

	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
		{Role: openai.ChatMessageRoleUser, Content: query},
	}

	// 最大轮数限制，防止死循环
	maxTurns := 20
	for i := 0; i < maxTurns; i++ {
		// fmt.Printf("  [MCP] Turn %d/%d\n", i+1, maxTurns)
		resp, err := a.LLM.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model: "deepseek-chat", Messages: msgs, Tools: a.Tools,
		})
		if err != nil {
			return "", fmt.Errorf("LLM 请求错误: %v", err)
		}

		msg := resp.Choices[0].Message
		msgs = append(msgs, msg)

		// 如果没有工具调用，说明本次回复结束
		if len(msg.ToolCalls) == 0 {
			return msg.Content, nil
		}

		// 处理工具调用
		for _, tc := range msg.ToolCalls {
			fmt.Printf("  [MCP] 调用 %s 参数: %s\n", tc.Function.Name, tc.Function.Arguments)

			// 执行远程工具
			resStr, err := a.Manager.Execute(tc.Function.Name, tc.Function.Arguments)
			if err != nil {
				resStr = fmt.Sprintf("Error: %v", err)
			}

			// 限制输出长度，防止 Token 爆炸
			if len(resStr) > 2000 {
				resStr = resStr[:2000] + "..."
			}

			fmt.Printf("  [MCP] <- 结果: %s...\n", limitString(resStr, 50))

			msgs = append(msgs, openai.ChatCompletionMessage{
				Role: openai.ChatMessageRoleTool, Content: resStr, ToolCallID: tc.ID,
			})
		}
		// 继续循环，把工具结果发回给 LLM
	}

	return "", fmt.Errorf("exceeded max turns")
}

func (a *MCPAgent) ProcessStream(ctx context.Context, query string) (<-chan string, <-chan error) {
	if err := a.ensureTools(nil); err != nil {
		textChan := make(chan string)
		errChan := make(chan error, 1)
		close(textChan)
		errChan <- err
		close(errChan)
		return textChan, errChan
	}
	return a.processStreamWithTools(ctx, query)
}

func (a *MCPAgent) ProcessStreamWithTools(ctx context.Context, query string, toolNames []string) (<-chan string, <-chan error) {
	if err := a.ensureTools(toolNames); err != nil {
		textChan := make(chan string)
		errChan := make(chan error, 1)
		close(textChan)
		errChan <- err
		close(errChan)
		return textChan, errChan
	}
	return a.processStreamWithTools(ctx, query)
}

func (a *MCPAgent) processStreamWithTools(ctx context.Context, query string) (<-chan string, <-chan error) {
	textChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(textChan)
		defer close(errChan)

		systemPrompt := `你是一个专注于“东莞理工学院（松山湖校区）”的智能校园出行助手。
你的核心职责是利用地图工具回答用户关于校园及周边的位置、导航和生活服务问题。

【重要上下文】
1. 默认位置：东莞理工学院松山湖校区（地址：广东省东莞市松山湖大学路1号）。
2. 当用户问“附近”、“周边”或“去xxx怎么走”且未指定起点时，默认起点/中心点均为“东莞理工学院松山湖校区”。
3. 搜索关键词策略：
   - 必须保留用户提到的特定约束（如“西餐”、“粤菜”、“星巴克”等）。不要将其简化为泛分类（如“餐厅”）。
   - 如果用户的问题包含具体品类，优先使用该品类作为 keywords 调用周边搜索工具。
4. 必须积极使用工具获取信息，不要直接拒绝对话或要求用户提供已知的位置信息。`

		msgs := []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: query},
		}

		maxTurns := 20
		for i := 0; i < maxTurns; i++ {
			stream, err := a.LLM.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
				Model: "deepseek-chat", Messages: msgs, Tools: a.Tools, Stream: true,
			})
			if err != nil {
				errChan <- fmt.Errorf("LLM 流请求错误: %v", err)
				return
			}

			var fullContent strings.Builder
			var toolCalls []openai.ToolCall

			for {
				resp, err := stream.Recv()
				if err != nil {
					if err.Error() == "EOF" {
						break
					}
					errChan <- fmt.Errorf("流读取错误: %v", err)
					stream.Close()
					return
				}

				if len(resp.Choices) == 0 {
					continue
				}

				delta := resp.Choices[0].Delta

				// 处理内容
				if delta.Content != "" {
					fullContent.WriteString(delta.Content)
					textChan <- delta.Content
				}

				// 处理工具调用 (分片累加)
				if len(delta.ToolCalls) > 0 {
					for _, tc := range delta.ToolCalls {
						idx := 0
						if tc.Index != nil {
							idx = *tc.Index
						}

						// 扩容
						for len(toolCalls) <= idx {
							toolCalls = append(toolCalls, openai.ToolCall{})
						}

						if tc.ID != "" {
							toolCalls[idx].ID = tc.ID
						}
						if tc.Type != "" {
							toolCalls[idx].Type = tc.Type
						}
						if tc.Function.Name != "" {
							toolCalls[idx].Function.Name += tc.Function.Name
						}
						if tc.Function.Arguments != "" {
							toolCalls[idx].Function.Arguments += tc.Function.Arguments
						}
					}
				}
			}
			stream.Close()

			// 将 LLM 的回复加入历史
			assistantMsg := openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: fullContent.String(),
			}
			if len(toolCalls) > 0 {
				assistantMsg.ToolCalls = toolCalls
			}
			msgs = append(msgs, assistantMsg)

			// 如果没有工具调用，说明本次回复结束
			if len(toolCalls) == 0 {
				return
			}

			// 执行工具
			for _, tc := range toolCalls {
				fmt.Printf("  [MCP] 调用 %s 参数: %s\n", tc.Function.Name, tc.Function.Arguments)
				resStr, err := a.Manager.Execute(tc.Function.Name, tc.Function.Arguments)
				if err != nil {
					resStr = fmt.Sprintf("Error: %v", err)
				}
				if len(resStr) > 2000 {
					resStr = resStr[:2000] + "..."
				}
				fmt.Printf("  [MCP] <- 结果: %s...\n", limitString(resStr, 50))

				msgs = append(msgs, openai.ChatCompletionMessage{
					Role: openai.ChatMessageRoleTool, Content: resStr, ToolCallID: tc.ID,
				})
			}
		}
		errChan <- fmt.Errorf("exceeded max turns")
	}()

	return textChan, errChan
}

func limitString(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
