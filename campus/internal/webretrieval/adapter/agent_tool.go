package adapter

import (
	"context"

	"milvus-kb-demo/internal/webretrieval/model"
	"milvus-kb-demo/internal/webretrieval/pipeline"
)

type WebSearchTool struct {
	Retriever pipeline.WebRetriever
}

type WebSearchRequest struct {
	Question    string
	SiteFilters []string
}

type WebSearchResponse struct {
	Documents []model.RankedDocument
	Citations []model.Citation
}

func NewWebSearchTool(r pipeline.WebRetriever) *WebSearchTool {
	return &WebSearchTool{
		Retriever: r,
	}
}

func (t *WebSearchTool) Invoke(ctx context.Context, req WebSearchRequest) (*WebSearchResponse, error) {
	q := model.SearchQuery{
		Text:        req.Question,
		SiteFilters: req.SiteFilters,
		TopK:        10,
		Lang:        "zh",
	}
	opt := model.RetrievalOptions{
		MaxSearchResults: 10,
		MaxDocsToFetch:   5,
		MaxTokens:        1024,
		UseCache:         true,
	}
	res, err := t.Retriever.Retrieve(ctx, q, opt)
	if err != nil {
		return nil, err
	}
	return &WebSearchResponse{
		Documents: res.Documents,
		Citations: res.Citations,
	}, nil
}

