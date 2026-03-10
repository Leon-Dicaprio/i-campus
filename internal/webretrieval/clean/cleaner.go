package clean

import (
	"context"
	"strings"

	"milvus-kb-demo/internal/webretrieval/model"
)

type ContentCleaner interface {
	Clean(ctx context.Context, c *model.Content) (*model.CleanContent, error)
}

type DefaultCleaner struct{}

func NewDefaultCleaner() *DefaultCleaner {
	return &DefaultCleaner{}
}

func (c *DefaultCleaner) Clean(ctx context.Context, content *model.Content) (*model.CleanContent, error) {
	body := strings.TrimSpace(content.PlainText)
	body = strings.Join(strings.Fields(body), " ")
	return &model.CleanContent{
		URL:      content.URL,
		Title:    content.Title,
		Body:     body,
		Metadata: content.Metadata,
	}, nil
}

