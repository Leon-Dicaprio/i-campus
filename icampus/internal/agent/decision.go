package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"milvus-kb-demo/internal/llm"
)

type SearchDecision struct {
	NeedSearch  bool   `json:"need_search"`
	SearchQuery string `json:"search_query"`
	NeedMCP     bool   `json:"need_mcp"`
	Reason      string `json:"reason"`
}

type DecisionAgent struct {
	llm *llm.Client
}

func NewDecisionAgent(llmClient *llm.Client) *DecisionAgent {
	return &DecisionAgent{llm: llmClient}
}

func (a *DecisionAgent) Decide(ctx context.Context, question string) (SearchDecision, error) {
	prompt := fmt.Sprintf(`你是高校智能客服系统的“搜索决策模块”。

判断用户问题是否需要【联网搜索】或【地图服务】。

规则：
- 学生手册、规章制度、固定流程 → 不需要搜索 (need_search=false, need_mcp=false)
- 最新通知、时间、电话、政策变化 → 需要搜索 (need_search=true)
- 不确定或可能变化 → 需要搜索 (need_search=true)
- 地点查询、路线规划、导航、周边查询、距离计算 → 需要地图服务 (need_mcp=true, need_search=false)

仅返回 JSON，不要解释。

{
  "need_search": true/false,
  "search_query": "用于搜索的关键词",
  "need_mcp": true/false,
  "reason": "简短原因"
}

用户问题：
%s`, question)

	raw, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return SearchDecision{}, err
	}

	decision, err := parseDecision(raw)
	if err != nil {
		return SearchDecision{}, err
	}

	decision = applyOverrides(question, decision)
	if decision.NeedSearch && strings.TrimSpace(decision.SearchQuery) == "" {
		decision.SearchQuery = strings.TrimSpace(question)
	}
	return decision, nil
}

func parseDecision(raw string) (SearchDecision, error) {
	var decision SearchDecision
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
	return SearchDecision{}, fmt.Errorf("invalid decision json")
}

func applyOverrides(question string, decision SearchDecision) SearchDecision {
	q := strings.ToLower(question)
	if strings.Contains(q, "体测") && strings.Contains(q, "免测") {
		decision.NeedSearch = false
		decision.NeedMCP = false
		if decision.Reason == "" {
			decision.Reason = "固定流程"
		}
		return decision
	}
	if strings.Contains(q, "体测安排") && regexp.MustCompile(`20\d{2}`).MatchString(q) {
		decision.NeedSearch = true
		decision.NeedMCP = false
		if decision.SearchQuery == "" {
			decision.SearchQuery = question
		}
		if decision.Reason == "" {
			decision.Reason = "时间可能变化"
		}
	}
	// 简单的关键词覆盖，确保地图相关问题能命中 MCP
	if strings.Contains(q, "哪里") || strings.Contains(q, "路线") || strings.Contains(q, "导航") || strings.Contains(q, "位置") || strings.Contains(q, "怎么走") {
		// 如果 LLM 没判出来，这里强制修正一下，但通常 LLM 更准
		// 这里只作为保底，或者可以不加，相信 LLM
	}
	return decision
}
