package fetch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func fetchAndExtractHTML(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// Add User-Agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	title := extractTitle(doc)
	text := extractBodyText(doc)

	// Format as simple Markdown for compatibility with MarkdownExtractor
	var sb strings.Builder
	if title != "" {
		sb.WriteString("# ")
		sb.WriteString(title)
		sb.WriteString("\n\n")
	}
	sb.WriteString(text)

	return []byte(sb.String()), nil
}

func extractTitle(doc *html.Node) string {
	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			// Extract all text inside title tag
			var sb strings.Builder
			var extractText func(*html.Node)
			extractText = func(c *html.Node) {
				if c.Type == html.TextNode {
					sb.WriteString(c.Data)
				}
				for child := c.FirstChild; child != nil; child = child.NextSibling {
					extractText(child)
				}
			}
			extractText(n)
			title = sb.String()
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
		// If no body tag, try root
		body = doc
	}
	var b strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "script" || n.Data == "style" || n.Data == "noscript" || n.Data == "head" || n.Data == "iframe" {
				return
			}
		}
		if n.Type == html.TextNode {
			data := strings.TrimSpace(n.Data)
			if data != "" {
				b.WriteString(data)
				b.WriteString("\n") // Use newline to separate blocks
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(body)
	return b.String()
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
