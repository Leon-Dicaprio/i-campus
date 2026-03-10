package cite

import (
	"context"

	"milvus-kb-demo/internal/webretrieval/model"
)

type CitationGenerator interface {
	Generate(ctx context.Context, docs []model.RankedDocument, topN int) ([]model.Citation, error)
}

type DefaultCiter struct{}

func NewDefaultCiter() *DefaultCiter {
	return &DefaultCiter{}
}

func (c *DefaultCiter) Generate(ctx context.Context, docs []model.RankedDocument, topN int) ([]model.Citation, error) {
	if topN <= 0 {
		topN = 3
	}
	if len(docs) < topN {
		topN = len(docs)
	}
	out := make([]model.Citation, 0, topN)
	for i := 0; i < topN; i++ {
		d := docs[i]
		excerpt := d.SummaryText
		runes := []rune(excerpt)
		if len(runes) > 120 {
			excerpt = string(runes[:120])
		}
		out = append(out, model.Citation{
			URL:     d.URL,
			Title:   d.Title,
			Excerpt: excerpt,
			Rank:    i,
		})
	}
	return out, nil
}

