package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"milvus-kb-demo/internal/webretrieval/model"
)

type JinaFetcher struct {
	Client *http.Client
}

func NewJinaFetcher(timeout time.Duration) *JinaFetcher {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &JinaFetcher{
		Client: &http.Client{Timeout: timeout},
	}
}

func (f *JinaFetcher) Fetch(ctx context.Context, url string) (*model.Page, error) {
	jinaURL := fmt.Sprintf("https://r.jina.ai/%s", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jinaURL, nil)
	if err != nil {
		return f.fallbackFetch(ctx, url, err)
	}

	resp, err := f.Client.Do(req)
	if err != nil {
		return f.fallbackFetch(ctx, url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return f.fallbackFetch(ctx, url, fmt.Errorf("status %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return f.fallbackFetch(ctx, url, err)
	}

	// Simple check if Jina returned empty or error message
	if len(body) < 50 || strings.Contains(string(body), "Jina Reader") && strings.Contains(string(body), "Error") {
		// Try fallback
		return f.fallbackFetch(ctx, url, fmt.Errorf("jina content suspicious"))
	}

	return &model.Page{
		URL:         url,
		RawHTML:     body, // Jina returns Markdown
		RetrievedAt: time.Now(),
	}, nil
}

func (f *JinaFetcher) fallbackFetch(ctx context.Context, url string, reason error) (*model.Page, error) {
	fmt.Printf("[JinaFetcher] Fallback for %s due to: %v\n", url, reason)

	// Create a new client for fallback if needed, or use f.Client
	// Note: fetchAndExtractHTML uses f.Client which has timeout set.
	content, err := fetchAndExtractHTML(ctx, f.Client, url)
	if err != nil {
		return nil, fmt.Errorf("both jina and fallback failed: %v (fallback err: %v)", reason, err)
	}

	return &model.Page{
		URL:         url,
		RawHTML:     content, // Extracted text formatted as Markdown
		RetrievedAt: time.Now(),
	}, nil
}
