package extract

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func FetchAndExtract(url string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", errors.New("empty url")
	}

	client := &http.Client{
		Timeout: 12 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; WebRetriever/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("unexpected status code")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	doc.Find("header, footer, nav, script, style").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	selectors := []string{
		"article",
		".content",
		"#content",
		".article-body",
	}

	var text string
	for _, sel := range selectors {
		selection := doc.Find(sel)
		if selection.Length() == 0 {
			continue
		}
		t := strings.TrimSpace(selection.Text())
		if t != "" {
			text = t
			break
		}
	}

	if text == "" {
		body := doc.Find("body")
		if body.Length() > 0 {
			text = strings.TrimSpace(body.Text())
		}
	}

	if text == "" {
		return "", errors.New("no content extracted")
	}

	fields := strings.Fields(text)
	text = strings.Join(fields, " ")

	runes := []rune(text)
	if len(runes) > 5000 {
		text = string(runes[:5000])
	}

	return text, nil
}

