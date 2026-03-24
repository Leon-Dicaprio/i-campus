package extract

import (
	"bytes"
	"context"
	"strings"

	"golang.org/x/net/html"

	"milvus-kb-demo/internal/webretrieval/model"
)

type ContentExtractor interface {
	Extract(ctx context.Context, page *model.Page) (*model.Content, error)
}

type HTMLExtractor struct{}

func NewHTMLExtractor() *HTMLExtractor {
	return &HTMLExtractor{}
}

func (e *HTMLExtractor) Extract(ctx context.Context, page *model.Page) (*model.Content, error) {
	doc, err := html.Parse(bytes.NewReader(page.RawHTML))
	if err != nil {
		return nil, err
	}

	title := extractTitle(doc)
	body := extractBodyText(doc)

	return &model.Content{
		URL:       page.URL,
		Title:     title,
		PlainText: body,
		Metadata:  map[string]string{},
	}, nil
}

func extractTitle(doc *html.Node) string {
	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			title = textContent(n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if title != "" {
				return
			}
			f(c)
		}
	}
	f(doc)
	return strings.TrimSpace(title)
}

func extractBodyText(doc *html.Node) string {
	body := findBody(doc)
	if body == nil {
		return ""
	}
	var b strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "script" || n.Data == "style" || n.Data == "noscript" {
				return
			}
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteString(" ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(body)
	text := strings.Join(strings.Fields(b.String()), " ")
	return strings.TrimSpace(text)
}

func findBody(doc *html.Node) *html.Node {
	var body *html.Node
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			body = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if body != nil {
				return
			}
			f(c)
		}
	}
	f(doc)
	return body
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var f func(*html.Node)
	f = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
			b.WriteString(" ")
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	text := strings.Join(strings.Fields(b.String()), " ")
	return strings.TrimSpace(text)
}

