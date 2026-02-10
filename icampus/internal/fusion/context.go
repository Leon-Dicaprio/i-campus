package fusion

import (
	"strings"

	"milvus-kb-demo/internal/search"
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
