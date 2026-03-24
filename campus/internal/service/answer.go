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
	"milvus-kb-demo/internal/webretrieval/model"
	"milvus-kb-demo/internal/webretrieval/pipeline"
)

type Service struct {
	DecisionAgent *agent.DecisionAgent
	Retriever     *rag.Retriever
	WebRetriever  pipeline.WebRetriever
	LLM           *llm.Client
	MCPAgent      *mcpserver.MCPAgent
}

type StreamEvent struct {
	Type    string `json:"type"` // "status", "token", "error"
	Content string `json:"content"`
}

func NewService(decisionAgent *agent.DecisionAgent, retriever *rag.Retriever, webRetriever pipeline.WebRetriever, llmClient *llm.Client, mcpAgent *mcpserver.MCPAgent) *Service {
	return &Service{
		DecisionAgent: decisionAgent,
		Retriever:     retriever,
		WebRetriever:  webRetriever,
		LLM:           llmClient,
		MCPAgent:      mcpAgent,
	}
}

func (s *Service) Answer(ctx context.Context, question string) (string, error) {
	decision, err := s.DecisionAgent.Decide(ctx, question)
	if err != nil {
		decision = agent.DecisionResult{
			NeedSearch:    true,
			SearchQueries: []string{question},
		}
	}

	// 优先处理地图相关请求
	if decision.NeedMCP {
		fmt.Printf("  [决策] 需要地图服务 (原因: %s)\n", decision.Reason)
		if s.MCPAgent != nil {
			fmt.Println("  [MCP] 转交 MCPAgent 处理...")
			toolNames := []string{}
			if decision.MCPToolName != "" {
				toolNames = append(toolNames, decision.MCPToolName)
			}
			return s.MCPAgent.ProcessWithTools(ctx, question, toolNames)
		} else {
			fmt.Println("  [警告] MCP Agent 未启用，降级为普通搜索")
			decision.NeedSearch = true
			if len(decision.SearchQueries) == 0 {
				decision.SearchQueries = []string{question}
			}
		}
	}

	if decision.NeedSearch {
		fmt.Printf("  [决策] 需要联网搜索 (原因: %s)\n", decision.Reason)
	} else {
		fmt.Printf("  [决策] 不需要联网搜索 (原因: %s)\n", decision.Reason)
	}

	var ragDocs []string
	// 如果明确不需要 Local RAG (例如天气/外网新闻)，则跳过
	if decision.NeedLocalRAG {
		ragDocs, err = s.Retriever.Search(ctx, question, 5)
		if err != nil {
			ragDocs = nil
		}
	} else {
		fmt.Printf("  [决策] 不需要查询本地知识库 (NeedLocalRAG=false)\n")
	}

	var webResults []model.RankedDocument
	if decision.NeedSearch {
		if s.WebRetriever == nil {
			fmt.Println("  [警告] 联网搜索已禁用，无法执行搜索")
		} else {
			queries := decision.SearchQueries
			if len(queries) == 0 {
				queries = []string{question}
			}

			var validQueries []string
			for _, q := range queries {
				if strings.TrimSpace(q) != "" {
					validQueries = append(validQueries, strings.TrimSpace(q))
				}
			}

			if len(validQueries) > 0 {
				fmt.Printf("  [搜索] 正在通过高级流水线并行搜索: %v ...\n", validQueries)
				opt := model.RetrievalOptions{
					MaxSearchResults: 5,
					MaxDocsToFetch:   3,
					UseCache:         true,
				}
				res, err := s.WebRetriever.BatchRetrieve(ctx, validQueries, opt)
				if err != nil {
					fmt.Printf("  [搜索] ❌ 搜索出错: %v\n", err)
				} else {
					webResults = res.Documents
					fmt.Printf("  [搜索] 找到 %d 条相关网页内容\n", len(webResults))
					for i, doc := range webResults {
						fmt.Printf("    %d. %s (%s...)\n", i+1, doc.Title, limitStr(doc.SummaryText, 30))
					}
				}
			}
		}
	}

	ctxText := fusion.BuildEnhancedContext(ragDocs, webResults)
	prompt := buildFinalPrompt(ctxText, question)
	return s.LLM.Call(ctx, prompt)
}

func (s *Service) GenerateTitle(ctx context.Context, question string) (string, error) {
	prompt := fmt.Sprintf("请根据以下用户的问题，生成一个极简的会话标题（不超过10个字）。只需返回标题内容，不要有任何解释或标点符号。\n\n问题：%s", question)
	return s.LLM.Call(ctx, prompt)
}

func (s *Service) AnswerStream(ctx context.Context, question string) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent)

	go func() {
		defer close(ch)

		ch <- StreamEvent{Type: "status", Content: "Analyzing question..."}
		decision, err := s.DecisionAgent.Decide(ctx, question)
		if err != nil {
			decision = agent.DecisionResult{
				NeedSearch:    true,
				SearchQueries: []string{question},
			}
		}

		if decision.NeedMCP {
			ch <- StreamEvent{Type: "status", Content: "Calling Map Service..."}
			if s.MCPAgent != nil {
				toolNames := []string{}
				if decision.MCPToolName != "" {
					toolNames = append(toolNames, decision.MCPToolName)
				}
				textChan, errChan := s.MCPAgent.ProcessStreamWithTools(ctx, question, toolNames)
				for {
					select {
					case text, ok := <-textChan:
						if !ok {
							textChan = nil
						} else {
							ch <- StreamEvent{Type: "token", Content: text}
						}
					case e, ok := <-errChan:
						if ok && e != nil {
							ch <- StreamEvent{Type: "error", Content: e.Error()}
						}
						errChan = nil
					}
					if textChan == nil && errChan == nil {
						break
					}
				}
				return
			}
			decision.NeedSearch = true
			if len(decision.SearchQueries) == 0 {
				decision.SearchQueries = []string{question}
			}
		}

		var ragDocs []string
		if decision.NeedLocalRAG {
			ch <- StreamEvent{Type: "status", Content: "Retrieving knowledge..."}
			var errSearch error
			ragDocs, errSearch = s.Retriever.Search(ctx, question, 5)
			if errSearch != nil {
				ragDocs = nil
			}
		} else {
			ch <- StreamEvent{Type: "status", Content: "Skipping knowledge retrieval..."}
		}

		var webResults []model.RankedDocument
		if decision.NeedSearch {
			if s.WebRetriever != nil {
				ch <- StreamEvent{Type: "status", Content: "Searching web..."}
				queries := decision.SearchQueries
				if len(queries) == 0 {
					queries = []string{question}
				}

				var validQueries []string
				for _, q := range queries {
					if strings.TrimSpace(q) != "" {
						validQueries = append(validQueries, strings.TrimSpace(q))
					}
				}

				if len(validQueries) > 0 {
					opt := model.RetrievalOptions{
						MaxSearchResults: 5,
						MaxDocsToFetch:   3,
						UseCache:         true,
						OnProgress: func(msg string) {
							ch <- StreamEvent{Type: "status", Content: msg}
						},
					}
					res, err := s.WebRetriever.BatchRetrieve(ctx, validQueries, opt)
					if err == nil {
						webResults = res.Documents
					}
				}
			} else {
				ch <- StreamEvent{Type: "status", Content: "Web search disabled, skipping..."}
			}
		}

		ch <- StreamEvent{Type: "status", Content: "Generating answer..."}
		ctxText := fusion.BuildEnhancedContext(ragDocs, webResults)
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
			case e, ok := <-errChan:
				if ok && e != nil {
					ch <- StreamEvent{Type: "error", Content: e.Error()}
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
	return fmt.Sprintf(`你是一个严谨的研究助手。请严格且仅根据我提供的参考资料回答问题。

参考资料：
%s

回答要求：
1. **真实性**：如果参考资料中没有相关信息，请直接回答“根据搜索结果未找到相关信息”，绝不允许使用你的内部知识编造。
2. **来源引用**：你的每一句陈述，必须在句末使用角标标明信息来源，格式为 [1] 或 [1][2]。
3. **手册引用**：如果答案来自《学生手册》相关资料，**务必**在回答中引用具体的章节或政策名称（例如：“根据《学生社区管理规定》，...”）。
4. **语气**：保持客观中立、专业且富有帮助的语气。

学生问题：%s`, contextText, question)
}

func limitStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}
