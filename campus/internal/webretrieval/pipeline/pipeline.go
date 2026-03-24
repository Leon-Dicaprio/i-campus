package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"milvus-kb-demo/internal/webretrieval/cache"
	"milvus-kb-demo/internal/webretrieval/cite"
	"milvus-kb-demo/internal/webretrieval/clean"
	"milvus-kb-demo/internal/webretrieval/extract"
	"milvus-kb-demo/internal/webretrieval/fetch"
	"milvus-kb-demo/internal/webretrieval/model"
	"milvus-kb-demo/internal/webretrieval/rerank"
	"milvus-kb-demo/internal/webretrieval/searchclient"
	"milvus-kb-demo/internal/webretrieval/summarize"

	"golang.org/x/sync/errgroup"
)

type WebRetriever interface {
	Retrieve(ctx context.Context, q model.SearchQuery, opt model.RetrievalOptions) (*model.RetrievalResult, error)
	BatchRetrieve(ctx context.Context, queries []string, opt model.RetrievalOptions) (*model.RetrievalResult, error)
}

type webRetriever struct {
	searchClient searchclient.WebSearchClient
	fetcher      fetch.PageFetcher
	extractor    extract.ContentExtractor
	cleaner      clean.ContentCleaner
	summarizer   summarize.Summarizer
	reranker     rerank.Reranker
	citer        cite.CitationGenerator
	cache        cache.Cache
	defaultTTL   time.Duration
}

func NewWebRetriever(
	searchClient searchclient.WebSearchClient,
	fetcher fetch.PageFetcher,
	extractor extract.ContentExtractor,
	cleaner clean.ContentCleaner,
	summarizer summarize.Summarizer,
	reranker rerank.Reranker,
	citer cite.CitationGenerator,
	c cache.Cache,
	defaultTTL time.Duration,
) WebRetriever {
	if defaultTTL <= 0 {
		defaultTTL = 10 * time.Minute
	}
	return &webRetriever{
		searchClient: searchClient,
		fetcher:      fetcher,
		extractor:    extractor,
		cleaner:      cleaner,
		summarizer:   summarizer,
		reranker:     reranker,
		citer:        citer,
		cache:        c,
		defaultTTL:   defaultTTL,
	}
}

func (w *webRetriever) Retrieve(ctx context.Context, q model.SearchQuery, opt model.RetrievalOptions) (*model.RetrievalResult, error) {
	return w.BatchRetrieve(ctx, []string{q.Text}, opt)
}

func (w *webRetriever) BatchRetrieve(ctx context.Context, queries []string, opt model.RetrievalOptions) (*model.RetrievalResult, error) {
	if len(queries) == 0 {
		return nil, fmt.Errorf("no queries provided")
	}

	if opt.MaxSearchResults <= 0 {
		opt.MaxSearchResults = 10
	}
	if opt.MaxDocsToFetch <= 0 {
		opt.MaxDocsToFetch = 3 // Default Top 3 per query as requested
	}
	if opt.MaxTokens <= 0 {
		opt.MaxTokens = 1024
	}

	cacheKey := ""
	if opt.UseCache && w.cache != nil {
		cacheKey = buildBatchCacheKey(queries, opt)
		if cacheKey != "" {
			if v, ok, _ := w.cache.Get(ctx, cacheKey); ok {
				var res model.RetrievalResult
				if err := json.Unmarshal(v, &res); err == nil {
					res.UsedCache = true
					return &res, nil
				}
			}
		}
	}

	// 1. Concurrent Search
	var searchResults []model.SearchResult
	var searchMutex sync.Mutex
	searchGroup, searchCtx := errgroup.WithContext(ctx)

	for _, queryText := range queries {
		qStr := queryText
		searchGroup.Go(func() error {
			// Construct a SearchQuery object for each text
			sq := model.SearchQuery{
				Text:        qStr,
				SiteFilters: nil, // We could propagate from opt if needed, but BatchRetrieve signature doesn't have it easily
				TopK:        opt.MaxSearchResults,
				Lang:        "zh", // Default or propagate
			}
			res, err := w.searchClient.Search(searchCtx, sq)
			if err != nil {
				fmt.Printf("[Pipeline] Search error for '%s': %v\n", qStr, err)
				return nil
			}

			// Filter Top K per query immediately
			limit := opt.MaxDocsToFetch
			if len(res) > limit {
				res = res[:limit]
			}
			fmt.Printf("[Pipeline] Search '%s' found %d results (taking top %d)\n", qStr, len(res), limit)

			searchMutex.Lock()
			searchResults = append(searchResults, res...)
			searchMutex.Unlock()
			return nil
		})
	}

	if err := searchGroup.Wait(); err != nil {
		fmt.Printf("[Pipeline] Search group error: %v\n", err)
		return nil, err
	}

	// Deduplicate URLs
	uniqueResults := make([]model.SearchResult, 0, len(searchResults))
	seenURLs := make(map[string]bool)
	for _, r := range searchResults {
		if !seenURLs[r.URL] {
			seenURLs[r.URL] = true
			uniqueResults = append(uniqueResults, r)
		}
	}
	searchResults = uniqueResults
	fmt.Printf("[Pipeline] Total unique URLs to fetch: %d\n", len(searchResults))

	if opt.OnProgress != nil && len(searchResults) > 0 {
		opt.OnProgress(fmt.Sprintf("Reading %d reference materials...", len(searchResults)))
	}

	// 2. Concurrent Fetch & Process
	summaries := make([]model.Summary, 0, len(searchResults))
	var summaryMutex sync.Mutex
	fetchGroup, _ := errgroup.WithContext(ctx)
	// We don't use fetchCtx for cancellation to avoid one failure stopping others,
	// but we should respect parent ctx.

	// Limit concurrency if needed, e.g. 10 concurrent fetches
	sem := make(chan struct{}, 10)

	for _, r := range searchResults {
		result := r
		fetchGroup.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			// Timeout for individual fetch to prevent hanging
			fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Second) // Increase timeout to 10s for robustness
			defer cancel()

			page, fetchErr := w.fetcher.Fetch(fetchCtx, result.URL)
			if fetchErr != nil {
				fmt.Printf("[Pipeline] Fetch error for %s: %v\n", result.URL, fetchErr)
				return nil
			}

			// Processing (Extract -> Clean -> Summarize)
			content, extractErr := w.extractor.Extract(ctx, page)
			if extractErr != nil {
				fmt.Printf("[Pipeline] Extract error for %s: %v\n", result.URL, extractErr)
				return nil
			}
			cleanContent, cleanErr := w.cleaner.Clean(ctx, content)
			if cleanErr != nil {
				fmt.Printf("[Pipeline] Clean error for %s: %v\n", result.URL, cleanErr)
				return nil
			}

			// Check if content is empty
			if len(strings.TrimSpace(cleanContent.Body)) < 50 {
				fmt.Printf("[Pipeline] Content too short for %s, skipping summarization\n", result.URL)
				return nil
			}

			combinedQuery := model.SearchQuery{Text: strings.Join(queries, " ")}

			summary, summarizeErr := w.summarizer.Summarize(ctx, cleanContent, combinedQuery, opt.MaxTokens)
			if summarizeErr != nil {
				fmt.Printf("[Pipeline] Summarize error for %s: %v\n", result.URL, summarizeErr)
				return nil
			}

			summaryMutex.Lock()
			summaries = append(summaries, *summary)
			summaryMutex.Unlock()
			return nil
		})
	}

	if err := fetchGroup.Wait(); err != nil {
		fmt.Printf("[Pipeline] Fetch group error: %v\n", err)
		return nil, err
	}
	fmt.Printf("[Pipeline] Successfully summarized %d documents\n", len(summaries))

	// 3. Rerank (using combined query)
	combinedQuery := model.SearchQuery{Text: strings.Join(queries, " ")}
	ranked, err := w.reranker.Rerank(ctx, combinedQuery, summaries)
	if err != nil {
		return nil, err
	}

	// 4. Cite
	citations, err := w.citer.Generate(ctx, ranked, 5)
	if err != nil {
		return nil, err
	}

	res := &model.RetrievalResult{
		Query:     combinedQuery,
		Documents: ranked,
		Citations: citations,
		UsedCache: false,
		DebugInfo: map[string]any{
			"search_result_count": len(searchResults),
			"summary_count":       len(summaries),
		},
	}

	if opt.UseCache && w.cache != nil && cacheKey != "" {
		if b, err := json.Marshal(res); err == nil {
			_ = w.cache.Set(ctx, cacheKey, b, w.defaultTTL)
		}
	}

	return res, nil
}

func buildBatchCacheKey(queries []string, opt model.RetrievalOptions) string {
	b, err := json.Marshal(struct {
		Q   []string
		Opt model.RetrievalOptions
	}{Q: queries, Opt: opt})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return "webretrieval_batch:" + hex.EncodeToString(sum[:])
}
