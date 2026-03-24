package searchclient

import (
	"context"

	"milvus-kb-demo/internal/search"
	"milvus-kb-demo/internal/webretrieval/model"
)

type WebSearchClient interface {
	Search(ctx context.Context, q model.SearchQuery) ([]model.SearchResult, error)
}

type SearcherAdapter struct {
	Searcher search.Searcher
	Source   string
}

func NewSearcherAdapter(s search.Searcher, source string) *SearcherAdapter {
	return &SearcherAdapter{
		Searcher: s,
		Source:   source,
	}
}

func (a *SearcherAdapter) Search(ctx context.Context, q model.SearchQuery) ([]model.SearchResult, error) {
	results, err := a.Searcher.Search(ctx, q.Text)
	if err != nil {
		return nil, err
	}
	out := make([]model.SearchResult, 0, len(results))
	for i, r := range results {
		out = append(out, model.SearchResult{
			URL:     r.URL,
			Title:   r.Title,
			Snippet: r.Snippet,
			Source:  a.Source,
			Rank:    i,
		})
	}
	return out, nil
}

