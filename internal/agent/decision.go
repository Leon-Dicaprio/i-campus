package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"milvus-kb-demo/internal/history"
	"milvus-kb-demo/internal/llm"
)

type DecisionResult struct {
	NeedLocalRAG  bool     `json:"need_local_rag"`
	NeedSearch    bool     `json:"need_search"`
	SearchQueries []string `json:"search_queries"`
	NeedMCP       bool     `json:"need_mcp"`
	MCPToolName   string   `json:"mcp_tool_name"`
	Reason        string   `json:"reason"`
}

type DecisionAgent struct {
	llm *llm.Client
}

func NewDecisionAgent(llmClient *llm.Client) *DecisionAgent {
	return &DecisionAgent{llm: llmClient}
}

func (a *DecisionAgent) Decide(ctx context.Context, messages []history.Message) (DecisionResult, error) {
	if len(messages) == 0 {
		return DecisionResult{}, fmt.Errorf("no messages provided")
	}
	question := lastUserContent(messages)
	historyText := formatHistory(messages)
	currentTime := time.Now().Format("2006年1月2日 星期一")
	prompt := fmt.Sprintf(`你是高校智能客服系统的“搜索决策模块”。

判断用户问题是否需要【联网搜索】或【地图服务】。
当前系统时间是：%s

对话历史：
%s

规则：
- 学生手册、规章制度、固定流程 → 不需要搜索 (need_search=false, need_mcp=false)
- 最新通知、时间、电话、政策变化 → 需要搜索 (need_search=true)
- 不确定或可能变化 → 需要搜索 (need_search=true)
- 地点查询、路线规划、导航、周边查询、距离计算 → 需要地图服务 (need_mcp=true, need_search=false)

关键词改写要求：
1. 将年份缩写补全，如“25年”改为“2025年”。
2. 将口语化词汇改为正式书面语。
3. 如果问题隐含特定学校，必须加上“东莞理工学院”。

仅返回 JSON，不要解释。

{
  "need_local_rag": true/false, // 建议加这个：如果明确是问天气/外网新闻，没必要查 Milvus，节省时间
  "need_search": true/false,
  "search_queries": ["用于搜索的关键词1", "关键词2"], // 重点！让大模型拆分并改写搜索词
  "need_mcp": true/false,
  "mcp_tool_name": "工具名称", // 如果需要地图服务，填写 amap_search
  "reason": "简短原因"
}

用户问题：
%s`, currentTime, historyText, question)

	raw, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return DecisionResult{}, err
	}

	decision, err := parseDecision(raw)
	if err != nil {
		return DecisionResult{}, err
	}

	decision = applyOverrides(question, decision)
	if decision.NeedSearch && len(decision.SearchQueries) == 0 {
		decision.SearchQueries = []string{strings.TrimSpace(question)}
	}
	return decision, nil
}

func formatHistory(messages []history.Message) string {
	var builder strings.Builder
	for _, msg := range messages {
		role := msg.Role
		if role == "user" {
			role = "用户"
		} else if role == "assistant" {
			role = "助手"
		}
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("%s：%s\n", role, content))
	}
	return strings.TrimSpace(builder.String())
}

func lastUserContent(messages []history.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		content := strings.TrimSpace(messages[i].Content)
		if content == "" {
			continue
		}
		if messages[i].Role == "user" {
			return content
		}
	}
	return strings.TrimSpace(messages[len(messages)-1].Content)
}

func parseDecision(raw string) (DecisionResult, error) {
	decision := DecisionResult{
		NeedLocalRAG: true, // Default to true
	}
	if err := json.Unmarshal([]byte(raw), &decision); err == nil {
		return decision, nil
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(raw[start:end+1]), &decision); err == nil {
			return decision, nil
		}
	}
	return DecisionResult{}, fmt.Errorf("invalid decision json")
}

func applyOverrides(question string, decision DecisionResult) DecisionResult {
	q := strings.ToLower(question)

	// 如果包含年份，通常需要联网搜索最新通知
	yearRegex := regexp.MustCompile(`(20)?2[4-6]年?`)
	if yearRegex.MatchString(q) {
		decision.NeedSearch = true
		if len(decision.SearchQueries) == 0 {
			decision.SearchQueries = []string{question}
		}
		if decision.Reason == "" {
			decision.Reason = "需要查询特定年份的最新通知"
		}
		return decision
	}

	// 体测安排通常需要联网搜索
	if strings.Contains(q, "体测安排") {
		decision.NeedSearch = true
		if len(decision.SearchQueries) == 0 {
			decision.SearchQueries = []string{question}
		}
		if decision.Reason == "" {
			decision.Reason = "体测安排具有时效性"
		}
		return decision
	}

	// 基础免测流程通常是固定的
	if strings.Contains(q, "体测") && strings.Contains(q, "免测") {
		decision.NeedSearch = false
		decision.NeedMCP = false
		if decision.Reason == "" {
			decision.Reason = "体测免测固定流程 (若需最新年份通知请说明年份)"
		}
		return decision
	}

	// 地图保底逻辑
	if strings.Contains(q, "哪里") || strings.Contains(q, "路线") || strings.Contains(q, "导航") || strings.Contains(q, "位置") || strings.Contains(q, "怎么走") {
		decision.NeedMCP = true
		if decision.MCPToolName == "" {
			decision.MCPToolName = "amap_search"
		}
	}

	return decision
}
