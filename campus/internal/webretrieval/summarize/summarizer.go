package summarize

import (
	"context"
	"fmt"
	"strings"

	"milvus-kb-demo/internal/llm"
	"milvus-kb-demo/internal/webretrieval/model"
)

type Summarizer interface {
	Summarize(ctx context.Context, c *model.CleanContent, q model.SearchQuery, maxTokens int) (*model.Summary, error)
}

type LLMSummarizer struct {
	Client *llm.Client
}

func NewLLMSummarizer(client *llm.Client) *LLMSummarizer {
	return &LLMSummarizer{
		Client: client,
	}
}

func (s *LLMSummarizer) Summarize(ctx context.Context, c *model.CleanContent, q model.SearchQuery, maxTokens int) (*model.Summary, error) {
	// Increase context window for reading, but keep output concise
	// maxTokens parameter here is usually "max tokens to generate", not input limit.
	// But we need to feed enough input.
	// Let's assume input limit is generous (e.g. 12k chars for ~3-4k tokens).
	inputLimit := 12000
	text := c.Body
	runes := []rune(text)
	if len(runes) > inputLimit {
		text = string(runes[:inputLimit])
	}

	prompt := fmt.Sprintf(`你是东莞理工学院智能信息提取助手。

任务：仔细阅读网页正文，提取与学生问题高度相关的信息。

要求：
1. **全面提取**：不要遗漏关键细节（如具体时间、地点、电话、步骤、条件）。
2. **结构化**：如果内容包含流程或多项条件，请用列表形式整理。
3. **原文引用**：尽量保留原文的专业表述，不要过度概括导致歧义。
4. **无关省略**：与问题无关的段落（如页眉页脚、广告）直接忽略。
5. **无中生有**：如果网页内容完全不相关，直接回复“无相关信息”。

学生问题：%s

网页标题：%s

网页正文：
%s`, q.Text, c.Title, text)

	resp, err := s.Client.Call(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm summarize failed: %w", err)
	}
	resp = strings.TrimSpace(resp)

	// If LLM says "No relevant info", we might want to handle it, but for now just return it.

	return &model.Summary{
		URL:         c.URL,
		Title:       c.Title,
		SummaryText: resp,
		Tokens:      len(resp), // Estimate
		Metadata:    c.Metadata,
	}, nil
}
