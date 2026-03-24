package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	// Go 语言的 OpenAI 客户端库
	"github.com/sashabaranov/go-openai"
)

// --- DeepSeek 模型常量定义 ---
const (
	// DeepSeek 的 API 基础 URL
	DeepSeekBaseURL = "https://api.deepseek.com/v1"
	// DeepSeek 聊天模型名称
	DeepseekChatModel = "deepseek-chat"
)

// --- 知识库定义 ---
var knowledgeBase = map[string]string{
	"KB1": "ABC 公司的主要产品是 Alpha 处理器，于 2024 年发布。",
	"KB2": "Alpha 处理器运行频率为 4.0 GHz，功耗为 85 瓦。",
	"KB3": "该公司的 CEO 是 Jane Doe，她于 2020 年加入公司。",
	"KB4": "发往北美地区的运输通常需要 5-7 个工作日。",
}

// --- 核心工具函数 ---

// retrieveContextByKeyword：使用关键词匹配模拟检索 (替代 Embedding)
// 这是一个简化的检索器，只要文档包含查询中的词，就认为相关。
func retrieveContextByKeyword(query string) string {
	var relevantDocs []string

	// 简单的分词：按空格分割
	keywords := strings.Fields(query)

	for _, docContent := range knowledgeBase {
		for _, keyword := range keywords {
			// 忽略太短的词（如 "是", "的"）
			if len(keyword) < 4 {
				continue
			}
			// 如果文档包含关键词，则加入上下文
			if strings.Contains(docContent, keyword) {
				relevantDocs = append(relevantDocs, docContent)
				break // 防止同一段文档被重复添加
			}
		}
	}

	if len(relevantDocs) == 0 {
		return ""
	}

	// 将找到的所有相关文档拼接起来
	return strings.Join(relevantDocs, "\n")
}

// generateAnswer：生成步骤 (保持不变，这是 RAG 的核心)
func generateAnswer(client *openai.Client, ctxContent, query string) (string, error) {
	systemPrompt := "你是一个有用的助手。请**严格使用**提供的 Context 来回答用户的问题。如果信息不足，请说明信息不足。"
	userPrompt := fmt.Sprintf("Context: %s\n\nQuestion: %s", ctxContent, query)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: DeepseekChatModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("DeepSeek Chat Completion API 调用失败: %w", err)
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func main() {
	// 1. 设置 DeepSeek API Key 环境变量
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 DEEPSEEK_API_KEY 环境变量。")
	}

	// 2. 初始化客户端
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = DeepSeekBaseURL
	client := openai.NewClientWithConfig(config)

	// 待查询的问题
	userQuery := "Alpha 处理器的工作频率是多少？"
	fmt.Printf("🎯 用户查询: %s\n\n", userQuery)

	// 3. RAG 流程 - 检索 (改为关键词检索)
	fmt.Println("🔍 第一步：从知识库中检索最相关的上下文 (Keyword Mode)...")

	// *** 使用新的关键词检索函数 ***
	retrievedContext := retrieveContextByKeyword(userQuery)

	if retrievedContext == "" {
		fmt.Println("--- 检索结果: 未找到足够相关的上下文。程序终止。---")
		return
	}

	fmt.Printf("--- 检索结果 (找到的上下文): %s\n\n", retrievedContext)

	// 4. RAG 流程 - 生成 (Generation)
	fmt.Println("🧠 第二步：使用 DeepSeek LLM 和上下文生成答案...")
	answer, err := generateAnswer(client, retrievedContext, userQuery)
	if err != nil {
		log.Fatalf("生成答案失败: %v", err)
	}

	fmt.Printf("✅ 最终答案: %s\n", answer)
}
