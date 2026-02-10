package service

import (
	"context"
	"fmt"
	"strings"

	"milvus-kb-demo/internal/agent"
	"milvus-kb-demo/internal/fusion"
	"milvus-kb-demo/internal/llm"
	"milvus-kb-demo/internal/mcpserver"
	"milvus-kb-demo/internal/rag"
	"milvus-kb-demo/internal/search"
)

type Service struct {
	DecisionAgent *agent.DecisionAgent
	Retriever     *rag.Retriever
	Searcher      search.Searcher
	LLM           *llm.Client
	MCPAgent      *mcpserver.MCPAgent
}

type StreamEvent struct {
	Type    string `json:"type"` // "status", "token", "error"
	Content string `json:"content"`
}

func NewService(decisionAgent *agent.DecisionAgent, retriever *rag.Retriever, searcher search.Searcher, llmClient *llm.Client, mcpAgent *mcpserver.MCPAgent) *Service {
	return &Service{
		DecisionAgent: decisionAgent,
		Retriever:     retriever,
		Searcher:      searcher,
		LLM:           llmClient,
		MCPAgent:      mcpAgent,
	}
}

func (s *Service) Answer(ctx context.Context, question string) (string, error) {
	decision, err := s.DecisionAgent.Decide(ctx, question)
	if err != nil {
		decision = agent.SearchDecision{
			NeedSearch:  true,
			SearchQuery: question,
		}
	}

	// 优先处理地图相关请求
	if decision.NeedMCP {
		fmt.Printf("  [决策] 需要地图服务 (原因: %s)\n", decision.Reason)
		if s.MCPAgent != nil {
			fmt.Println("  [MCP] 转交 MCPAgent 处理...")
			return s.MCPAgent.Process(ctx, question)
		} else {
			fmt.Println("  [警告] MCP Agent 未启用，降级为普通搜索")
			decision.NeedSearch = true
			if decision.SearchQuery == "" {
				decision.SearchQuery = question
			}
		}
	}

	if decision.NeedSearch {
		fmt.Printf("  [决策] 需要联网搜索 (原因: %s)\n", decision.Reason)
	} else {
		fmt.Printf("  [决策] 不需要联网搜索 (原因: %s)\n", decision.Reason)
	}

	ragDocs, err := s.Retriever.Search(ctx, question, 5)
	if err != nil {
		ragDocs = nil
	}

	var webResults []search.Result
	if decision.NeedSearch {
		query := strings.TrimSpace(decision.SearchQuery)
		if query == "" {
			query = question
		}
		fmt.Printf("  [搜索] 正在搜索: %s ...\n", query)
		var err error
		webResults, err = s.Searcher.Search(ctx, query)
		if err != nil {
			fmt.Printf("  [搜索] ❌ 搜索出错: %v\n", err)
		} else {
			fmt.Printf("  [搜索] 找到 %d 条相关网页\n", len(webResults))
			for i, res := range webResults {
				fmt.Printf("    %d. %s (%s...)\n", i+1, res.Title, limitStr(res.Snippet, 30))
			}
		}
	}

	ctxText := fusion.BuildContext(ragDocs, webResults)
	prompt := buildFinalPrompt(ctxText, question)
	return s.LLM.Call(ctx, prompt)
}

func (s *Service) AnswerStream(ctx context.Context, question string) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent)

	go func() {
		defer close(ch)

		ch <- StreamEvent{Type: "status", Content: "Analyzing question..."}
		decision, err := s.DecisionAgent.Decide(ctx, question)
		if err != nil {
			decision = agent.SearchDecision{
				NeedSearch:  true,
				SearchQuery: question,
			}
		}

		if decision.NeedMCP {
			ch <- StreamEvent{Type: "status", Content: "Calling Map Service..."}
			if s.MCPAgent != nil {
				res, err := s.MCPAgent.Process(ctx, question)
				if err != nil {
					ch <- StreamEvent{Type: "error", Content: err.Error()}
				} else {
					ch <- StreamEvent{Type: "token", Content: res}
				}
				return
			}
			decision.NeedSearch = true
			if decision.SearchQuery == "" {
				decision.SearchQuery = question
			}
		}

		ch <- StreamEvent{Type: "status", Content: "Retrieving knowledge..."}
		ragDocs, err := s.Retriever.Search(ctx, question, 5)
		if err != nil {
			ragDocs = nil
		}

		var webResults []search.Result
		if decision.NeedSearch {
			ch <- StreamEvent{Type: "status", Content: "Searching web..."}
			query := strings.TrimSpace(decision.SearchQuery)
			if query == "" {
				query = question
			}

			var err error
			webResults, err = s.Searcher.Search(ctx, query)
			if err != nil {
				// ignore error
			}
		}

		ch <- StreamEvent{Type: "status", Content: "Generating answer..."}
		ctxText := fusion.BuildContext(ragDocs, webResults)
		prompt := buildFinalPrompt(ctxText, question)

		textChan, errChan := s.LLM.CallStream(ctx, prompt)

		for {
			select {
			case text, ok := <-textChan:
				if !ok {
					textChan = nil
				} else {
					ch <- StreamEvent{Type: "token", Content: text}
				}
			case err, ok := <-errChan:
				if ok && err != nil {
					ch <- StreamEvent{Type: "error", Content: err.Error()}
				}
				errChan = nil
			}
			if textChan == nil && errChan == nil {
				break
			}
		}
	}()

	return ch, nil
}

func buildFinalPrompt(contextText, question string) string {
	return fmt.Sprintf(`你是东莞理工学院官方智能客服。

请【基于已知信息】回答学生问题。
如果已知信息中包含相关通知或时间，请直接总结并列出。
如果已知信息不足或没有具体细节，再说明“建议查看官网”。

已知信息：
%s

学生问题：%s`, contextText, question)
}

func limitStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}
