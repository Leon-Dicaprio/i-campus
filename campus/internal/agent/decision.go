package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

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

func (a *DecisionAgent) Decide(ctx context.Context, question string) (DecisionResult, error) {
	currentTime := time.Now().Format("2006年1月2日 星期一")
	prompt := fmt.Sprintf(`你是高校智能客服系统的“任务分配与决策模块”。

你的任务是判断用户的问题是否需要执行以下任务：
1. **本地知识库检索 (Local RAG)**: 针对《学生手册》及校园规章制度。
2. **联网搜索 (Web Search)**: 针对最新动态、通知、时间、电话或政策变化。
3. **地图服务 (Map Service)**: 针对地点查询、路线规划、导航、周边查询、距离计算。

当前系统时间是：%s

### 1. 本地知识库检索 (need_local_rag) 判断逻辑：
当用户提问涉及以下《学生手册》核心领域时，**必须**设置为 true：
- **学术与教务**：入学注册、转专业流程、绩点计算、学位授予条件、考勤旷课规定、辅修申请、通识课修读、学籍异动（休学/退学）。
- **奖惩与申诉**：国家奖助学金申请、勤工助学、学生违纪处分标准（作弊、打架等）、学术不端处理、学生申诉程序。
- **校园生活与安全**：学生宿舍管理（作息、违章电器清单）、校内交通（机动车/电动车管理）、消防安全、网络使用规范。
- **专项政策**：大学生参军入伍的激励措施与学业帮扶、学生伤害事故的处理程序。
- **学校概况**：校训、办学理念、学校章程等。

具体场景示例：
- 询问具体的“怎么做”：如“如何申请转专业？”、“怎么拿奖学金？”。
- 询问具体的“标准/要求”：如“学位申请需要多少绩点？”、“旷课多少节会被处分？”。
- 询问具体的“权利/义务”：如“宿舍能用吹风机吗？”、“对处分不服怎么办？”。
- 询问校园内的“管理规则”：如“电动车能进校园吗？”、“校园网怎么实名？”。

如果提问属于通用知识（如“什么是人工智能？”）或与东莞理工学院（莞工）校规无关，则设置为 false。

### 2. 联网搜索 (need_search) 判断逻辑：
- 用户说明要联网搜索
- 涉及最新通知、具体时间（如“今年放假时间”）、电话、时效性强的政策（如“2025年最新补考通知”）。
- 不确定或可能随时变化的信息。

### 3. 地图服务 (need_mcp) 判断逻辑：
- 涉及具体地点、位置、路线导航、周边设施。

### 关键词改写要求：
1. 将年份缩写补全，如“25年”改为“2025年”。
2. 将口语化词汇改为正式书面语。
3. 如果问题隐含特定学校，必须加上“东莞理工学院”。

仅返回 JSON，不要解释。

{
  "need_local_rag": true/false, 
  "need_search": true/false,
  "search_queries": ["用于搜索的关键词1", "关键词2"], // 重点！让大模型拆分并改写搜索词
  "need_mcp": true/false,
  "mcp_tool_name": "amap_search", // 如果 need_mcp 为 true，填 "amap_search"，否则为空
  "reason": "简短原因"
}

用户问题：
%s`, currentTime, question)

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

func parseDecision(raw string) (DecisionResult, error) {
	decision := DecisionResult{
		NeedLocalRAG: false, // Default to false
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

	// 如果包含年份，通常需要联网搜索最新通知，但可能也需要 RAG
	yearRegex := regexp.MustCompile(`(20)?2[4-6]年?`)
	if yearRegex.MatchString(q) {
		decision.NeedSearch = true
		if len(decision.SearchQueries) == 0 {
			decision.SearchQueries = []string{question}
		}
		if decision.Reason == "" {
			decision.Reason = "需要查询特定年份的最新通知"
		}
		// 不要直接 return，让后面的逻辑继续判断是否需要 RAG 或 MCP
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
	}

	// 基础免测流程通常是固定的，需要 RAG
	if strings.Contains(q, "体测") && strings.Contains(q, "免测") {
		decision.NeedLocalRAG = true
		if decision.Reason == "" {
			decision.Reason = "体测免测固定流程 (若需最新年份通知请说明年份)"
		}
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
