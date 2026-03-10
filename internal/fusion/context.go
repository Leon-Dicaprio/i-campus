package fusion

import (
	"fmt"
	"strings"

	"milvus-kb-demo/internal/search"
	"milvus-kb-demo/internal/webretrieval/model"
)

func BuildContext(ragDocs []string, webResults []search.Result) string {
	var b strings.Builder
	b.WriteString("【校内固定资料】\n")
	if len(ragDocs) == 0 {
		b.WriteString("- 暂无\n\n")
	} else {
		for _, doc := range ragDocs {
			b.WriteString("- 学生手册：")
			b.WriteString(strings.TrimSpace(doc))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("【官网实时信息】\n")
	if len(webResults) == 0 {
		b.WriteString("- 暂无\n")
		return b.String()
	}
	for _, r := range webResults {
		b.WriteString("- 《")
		b.WriteString(strings.TrimSpace(r.Title))
		b.WriteString("》\n  摘要：")
		b.WriteString(strings.TrimSpace(r.Snippet))
		b.WriteString("\n  来源：")
		b.WriteString(strings.TrimSpace(r.URL))
		b.WriteString("\n")
	}
	return b.String()
}

func BuildEnhancedContext(ragDocs []string, webResults []model.RankedDocument) string {
	var b strings.Builder
	idx := 1

	// Add RAG docs
	for _, doc := range ragDocs {
		b.WriteString(strings.TrimSpace(fmt.Sprintf("[%d] 标题: 校内知识库, 内容: %s", idx, strings.TrimSpace(doc))))
		b.WriteString("\n")
		idx++
	}

	// Add Web results
	for _, r := range webResults {
		// Clean up newlines in summary to keep it on one line if possible, or just trim
		summary := strings.ReplaceAll(strings.TrimSpace(r.SummaryText), "\n", " ")
		b.WriteString(strings.TrimSpace(fmt.Sprintf("[%d] 标题: %s, 内容: %s, 来源: %s", idx, strings.TrimSpace(r.Title), summary, strings.TrimSpace(r.URL))))
		b.WriteString("\n")
		idx++
	}

	if idx == 1 {
		return "暂无参考资料"
	}

	return b.String()
}
