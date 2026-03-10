package webretrieval

import (
	"time"

	"milvus-kb-demo/internal/llm"
	"milvus-kb-demo/internal/search"
	"milvus-kb-demo/internal/webretrieval/cache"
	"milvus-kb-demo/internal/webretrieval/cite"
	"milvus-kb-demo/internal/webretrieval/clean"
	"milvus-kb-demo/internal/webretrieval/extract"
	"milvus-kb-demo/internal/webretrieval/fetch"
	"milvus-kb-demo/internal/webretrieval/pipeline"
	"milvus-kb-demo/internal/webretrieval/rerank"
	"milvus-kb-demo/internal/webretrieval/searchclient"
	"milvus-kb-demo/internal/webretrieval/summarize"
)

func NewDefaultWebRetriever(llmClient *llm.Client, searcher search.Searcher) pipeline.WebRetriever {
	searchClient := searchclient.NewSearcherAdapter(searcher, "api")
	// Use JinaFetcher for fetching pages (Markdown)
	fetcher := fetch.NewJinaFetcher(10 * time.Second)
	// Use MarkdownExtractor for Jina content
	extractor := extract.NewMarkdownExtractor()
	// Use MarkdownCleaner to preserve structure
	cleaner := clean.NewMarkdownCleaner()
	summarizer := summarize.NewLLMSummarizer(llmClient)
	reranker := rerank.NewLexicalReranker()
	citer := cite.NewDefaultCiter()
	c := cache.NewInMemoryCache()

	return pipeline.NewWebRetriever(
		searchClient,
		fetcher,
		extractor,
		cleaner,
		summarizer,
		reranker,
		citer,
		c,
		10*time.Minute,
	)
}
