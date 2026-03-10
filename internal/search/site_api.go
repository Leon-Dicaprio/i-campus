package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SearchAPIClient interface {
	Search(ctx context.Context, query string, limit int) ([]Result, error)
}

type SiteLimitedSearcher struct {
	Client     SearchAPIClient
	SiteDomain string
	MaxResults int
}

func NewSiteLimitedSearcher(client SearchAPIClient, siteDomain string, maxResults int) *SiteLimitedSearcher {
	if maxResults <= 0 || maxResults > 5 {
		maxResults = 5
	}
	return &SiteLimitedSearcher{
		Client:     client,
		SiteDomain: siteDomain,
		MaxResults: maxResults,
	}
}

func (s *SiteLimitedSearcher) Search(ctx context.Context, query string) ([]Result, error) {
	base := strings.TrimSpace(query)
	if base == "" {
		return nil, nil
	}
	siteConfig := strings.TrimSpace(s.SiteDomain)
	if siteConfig == "" {
		return s.Client.Search(ctx, base, s.MaxResults)
	}

	// 支持多域名搜索，逗号分隔，例如 "dgut.edu.cn,dgut.cn"
	domains := strings.Split(siteConfig, ",")
	var siteParts []string
	for _, d := range domains {
		d = strings.TrimSpace(d)
		if d != "" {
			siteParts = append(siteParts, fmt.Sprintf("site:%s", d))
		}
	}

	var q string
	if len(siteParts) == 0 {
		q = base
	} else if len(siteParts) == 1 {
		q = fmt.Sprintf("%s %s", siteParts[0], base)
	} else {
		// 构造 (site:a.com OR site:b.com) query
		q = fmt.Sprintf("(%s) %s", strings.Join(siteParts, " OR "), base)
	}

	results, err := s.Client.Search(ctx, q, s.MaxResults)
	if err != nil {
		return nil, err
	}

	// Fallback mechanism: if insufficient results, try searching with domain as keyword (not site operator)
	// Only if we searched with site operator previously
	if len(results) < 3 && len(siteParts) > 0 {
		fmt.Printf("[Search] Site-limited results (%d) insufficient, trying fallback search...\n", len(results))

		// Use the first domain as keyword, or join them?
		// Let's just append the raw siteConfig string as keyword, hoping search engine handles "dgut.edu.cn" well.
		// Or better: if domains are ["dgut.edu.cn"], append "dgut.edu.cn".
		// But usually we want school name. Since we don't have school name, we use domain.
		fallbackQuery := fmt.Sprintf("%s %s", base, siteConfig)

		fallbackResults, err := s.Client.Search(ctx, fallbackQuery, s.MaxResults)
		if err == nil && len(fallbackResults) > 0 {
			// Merge results with domain validation
			seen := make(map[string]bool)
			for _, r := range results {
				seen[r.URL] = true
			}

			// Validate fallback results: must contain domain
			validDomains := strings.Split(siteConfig, ",")
			addedCount := 0
			for _, r := range fallbackResults {
				if seen[r.URL] {
					continue
				}

				// Strict domain check for fallback results
				isValid := false
				u, parseErr := url.Parse(r.URL)
				if parseErr == nil {
					hostname := u.Hostname()
					for _, domain := range validDomains {
						domain = strings.TrimSpace(domain)
						if domain != "" && (hostname == domain || strings.HasSuffix(hostname, "."+domain)) {
							isValid = true
							break
						}
					}
				}

				if isValid {
					results = append(results, r)
					seen[r.URL] = true
					addedCount++
				} else {
					fmt.Printf("[Search] Ignored fallback result (domain mismatch): %s\n", r.URL)
				}
			}
			fmt.Printf("[Search] Fallback found %d new valid results\n", addedCount)
		} else if err != nil {
			fmt.Printf("[Search] Fallback search failed: %v\n", err)
		}
	}

	return results, nil
}

type SerpAPISearchClient struct {
	APIKey string
	Engine string // e.g. "google"
	Client *http.Client
}

func NewSerpAPISearchClient(apiKey string) *SerpAPISearchClient {
	return &SerpAPISearchClient{
		APIKey: apiKey,
		Engine: "google",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *SerpAPISearchClient) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit <= 0 || limit > 10 {
		limit = 10
	}
	values := url.Values{}
	values.Set("q", query)
	values.Set("api_key", c.APIKey)
	values.Set("engine", c.Engine)
	values.Set("num", fmt.Sprintf("%d", limit))

	reqURL := "https://serpapi.com/search?" + values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("serpapi status: %d, body: %s", resp.StatusCode, string(body))
	}

	var data struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic_results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []Result
	for _, r := range data.OrganicResults {
		results = append(results, Result{
			Title:   r.Title,
			URL:     r.Link,
			Snippet: r.Snippet,
		})
	}
	return results, nil
}

type HTTPAPISearchClient struct {
	Endpoint string
	APIKey   string
	Client   *http.Client
}

func NewHTTPAPISearchClient(endpoint string, apiKey string) *HTTPAPISearchClient {
	return &HTTPAPISearchClient{
		Endpoint: endpoint,
		APIKey:   apiKey,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *HTTPAPISearchClient) httpClient() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

func (c *HTTPAPISearchClient) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit <= 0 || limit > 5 {
		limit = 5
	}
	values := url.Values{}
	values.Set("q", query)
	values.Set("limit", fmt.Sprintf("%d", limit))

	endpoint := strings.TrimSpace(c.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("empty search endpoint")
	}

	reqURL := endpoint
	if strings.Contains(reqURL, "?") {
		reqURL = reqURL + "&" + values.Encode()
	} else {
		reqURL = reqURL + "?" + values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search api status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Snippet string `json:"snippet"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	out := make([]Result, 0, len(apiResp.Results))
	for i, r := range apiResp.Results {
		if i >= limit {
			break
		}
		out = append(out, Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Snippet,
		})
	}
	return out, nil
}

type MockSearchAPIClient struct {
	Results []Result
}

func NewMockSearchAPIClient(results []Result) *MockSearchAPIClient {
	return &MockSearchAPIClient{
		Results: results,
	}
}

func (m *MockSearchAPIClient) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit <= 0 || limit > 5 {
		limit = 5
	}
	if len(m.Results) <= limit {
		return m.Results, nil
	}
	return m.Results[:limit], nil
}
