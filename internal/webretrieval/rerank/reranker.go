package rerank

import (
	"context"
	"strings"

	"milvus-kb-demo/internal/webretrieval/model"
)

type Reranker interface {
	Rerank(ctx context.Context, q model.SearchQuery, docs []model.Summary) ([]model.RankedDocument, error)
}

type LexicalReranker struct{}

func NewLexicalReranker() *LexicalReranker {
	return &LexicalReranker{}
}

func (r *LexicalReranker) Rerank(ctx context.Context, q model.SearchQuery, docs []model.Summary) ([]model.RankedDocument, error) {
	queryTerms := splitTerms(q.Text)
	ranked := make([]model.RankedDocument, 0, len(docs))
	for _, d := range docs {
		score := scoreDocument(queryTerms, d.SummaryText)
		ranked = append(ranked, model.RankedDocument{
			URL:         d.URL,
			Title:       d.Title,
			SummaryText: d.SummaryText,
			Score:       score,
			Metadata:    d.Metadata,
		})
	}
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[j].Score > ranked[i].Score {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}
	return ranked, nil
}

func splitTerms(text string) []string {
	text = strings.ToLower(text)
	fields := strings.Fields(text)
	return fields
}

func scoreDocument(terms []string, content string) float64 {
	if len(terms) == 0 {
		return 0
	}
	contentLower := strings.ToLower(content)
	var score float64
	for _, t := range terms {
		if t == "" {
			continue
		}
		count := strings.Count(contentLower, t)
		if count > 0 {
			score += float64(count)
		}
	}
	return score
}

