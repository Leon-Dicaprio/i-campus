package model

import "time"

type SearchQuery struct {
	Text        string
	SiteFilters []string
	TopK        int
	Lang        string
}

type SearchResult struct {
	URL     string
	Title   string
	Snippet string
	Source  string
	Rank    int
}

type Page struct {
	URL         string
	RawHTML     []byte
	RetrievedAt time.Time
}

type Content struct {
	URL       string
	Title     string
	PlainText string
	Metadata  map[string]string
}

type CleanContent struct {
	URL      string
	Title    string
	Body     string
	Metadata map[string]string
}

type Summary struct {
	URL         string
	Title       string
	SummaryText string
	Tokens      int
	Metadata    map[string]string
}

type RankedDocument struct {
	URL         string
	Title       string
	SummaryText string
	Score       float64
	Metadata    map[string]string
}

type Citation struct {
	URL     string
	Title   string
	Excerpt string
	Rank    int
}

type RetrievalOptions struct {
	MaxSearchResults int
	MaxDocsToFetch   int
	MaxTokens        int
	UseCache         bool
	OnProgress       func(msg string)
}

type RetrievalResult struct {
	Query     SearchQuery
	Documents []RankedDocument
	Citations []Citation
	UsedCache bool
	DebugInfo map[string]any
}

