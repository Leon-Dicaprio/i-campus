package extract

import (
	"context"
	"strings"

	"milvus-kb-demo/internal/webretrieval/model"
)

type MarkdownExtractor struct{}

func NewMarkdownExtractor() *MarkdownExtractor {
	return &MarkdownExtractor{}
}

func (e *MarkdownExtractor) Extract(ctx context.Context, page *model.Page) (*model.Content, error) {
	// Jina returns markdown in RawHTML
	text := string(page.RawHTML)

	// Simple title extraction
	var title string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			break
		}
		if strings.HasPrefix(line, "Title: ") {
			title = strings.TrimPrefix(line, "Title: ")
			break
		}
	}
	if title == "" {
		// Use URL as title if not found
		title = page.URL
	}

	return &model.Content{
		URL:       page.URL,
		Title:     title,
		PlainText: text, // Keep markdown structure as plain text
		Metadata:  map[string]string{"format": "markdown"},
	}, nil
}
