package clean

import (
	"context"
	"strings"

	"milvus-kb-demo/internal/webretrieval/model"
)

type MarkdownCleaner struct{}

func NewMarkdownCleaner() *MarkdownCleaner {
	return &MarkdownCleaner{}
}

func (c *MarkdownCleaner) Clean(ctx context.Context, content *model.Content) (*model.CleanContent, error) {
	// For Markdown, we want to preserve structure (newlines, lists, etc.)
	// Do not use strings.Fields() as it collapses all whitespace including newlines.
	
	body := strings.TrimSpace(content.PlainText)
	
	// Optional: Remove excessive newlines (more than 2)
	for strings.Contains(body, "\n\n\n") {
		body = strings.ReplaceAll(body, "\n\n\n", "\n\n")
	}

	return &model.CleanContent{
		URL:      content.URL,
		Title:    content.Title,
		Body:     body,
		Metadata: content.Metadata,
	}, nil
}
