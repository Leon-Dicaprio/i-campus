package fetch

import (
	"context"
	"io"
	"net/http"
	"time"

	"milvus-kb-demo/internal/webretrieval/model"
)

type PageFetcher interface {
	Fetch(ctx context.Context, url string) (*model.Page, error)
}

type HTTPFetcher struct {
	Client *http.Client
}

func NewHTTPFetcher(timeout time.Duration) *HTTPFetcher {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &HTTPFetcher{
		Client: &http.Client{Timeout: timeout},
	}
}

func (f *HTTPFetcher) Fetch(ctx context.Context, url string) (*model.Page, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &model.Page{
		URL:         url,
		RawHTML:     body,
		RetrievedAt: time.Now(),
	}, nil
}

