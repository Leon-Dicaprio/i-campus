package search

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type BingSearcher struct {
	client *http.Client
}

func NewBingSearcher() *BingSearcher {
	return &BingSearcher{
		client: &http.Client{Timeout: 12 * time.Second},
	}
}

func (s *BingSearcher) Search(ctx context.Context, query string) ([]Result, error) {
	searchQuery := fmt.Sprintf("site:dgut.edu.cn %s", query)
	// 使用 cn.bing.com 提高国内稳定性
	endpoint := "https://cn.bing.com/search?q=" + url.QueryEscape(searchQuery)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	// 模拟真实浏览器 UA 和 Cookie
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cookie", "SRCHHPGUSR=CW=1920&CH=1080; _EDGE_S=F=1") // 简单 Cookie 规避

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search status: %d", resp.StatusCode)
	}

	// 读取 HTML 内容（为了调试，如果不成功可以打印出来）
	// 注意：html.Parse 会消耗 Body，所以如果要调试需要先读出来
	// 这里为了性能，我们先直接 Parse。如果 Parse 结果为 0，我们在未来可以考虑 dump Body。

	// 使用 TeeReader 可以在解析的同时保留一部分内容用于调试（如果需要）
	// 但为了简单，我们先直接解析
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	// 尝试解析结果
	results := parseResultsRobust(doc)

	// 如果没找到结果，尝试提取 Title 看看是不是被拦截了
	if len(results) == 0 {
		title := extractPageTitle(doc)
		// 打印一些页面中的链接用于调试
		var debugLinks []string
		debugCount := 0
		var walkDebug func(*html.Node)
		walkDebug = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "a" {
				href := getAttr(n, "href")
				if href != "" && debugCount < 5 {
					debugLinks = append(debugLinks, href)
					debugCount++
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walkDebug(c)
			}
		}
		walkDebug(doc)

		return nil, fmt.Errorf("no results found (Page Title: %s, Links found: %d, Sample: %v)", title, debugCount, debugLinks)
	}

	return results, nil
}

// extractPageTitle 提取页面标题用于调试
func extractPageTitle(n *html.Node) string {
	var title string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "title" {
			title = textContent(node)
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if title != "" {
				return
			}
			walk(c)
		}
	}
	walk(n)
	return title
}

// parseResultsRobust 使用启发式规则提取搜索结果，不依赖特定 class
func parseResultsRobust(n *html.Node) []Result {
	var results []Result
	seenURLs := make(map[string]bool)

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		// 寻找所有链接
		if node.Type == html.ElementNode && node.Data == "a" {
			href := getAttr(node, "href")
			// 必须是有效链接且属于目标域名
			if href != "" && isAllowedDomain(href) && !seenURLs[href] {

				// 检查链接文本（标题）是否足够长，过滤掉“首页”、“登录”等短链
				title := strings.TrimSpace(textContent(node))
				// 过滤掉时间筛选器等无效链接
				if isGarbageTitle(title) {
					// 这里的 return 是跳过当前节点及其子节点的遍历，相当于 continue 效果
					return
				}

				if len([]rune(title)) > 4 { // 至少5个字

					// 提取摘要（尝试找父级元素下的 p 或 div）
					snippet := findNearbySnippet(node)

					results = append(results, Result{
						Title:   title,
						URL:     href,
						Snippet: snippet,
					})
					seenURLs[href] = true
				}
			}
		}

		if len(results) >= 3 {
			return
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
			if len(results) >= 3 {
				return
			}
		}
	}
	walk(n)
	return results
}

func isGarbageTitle(title string) bool {
	garbage := []string{"24 小时内", "24小时内", "过去24小时", "一周内", "一个月内", "一年内", "所有时间", "按时间排序", "按相关性排序"}
	for _, g := range garbage {
		if strings.Contains(title, g) {
			return true
		}
	}
	return false
}

// findNearbySnippet 寻找链接附近的文本作为摘要
func findNearbySnippet(aNode *html.Node) string {
	// 策略：向上找父节点，直到找到一个包含大量文本的容器（例如 li 或 div）
	// 然后提取其中的所有文本，排除掉标题本身

	container := aNode.Parent
	// 向上找最多 3 层
	for i := 0; i < 3; i++ {
		if container == nil || container.Type != html.ElementNode {
			break
		}
		// 如果容器包含大量文本，可能就是摘要容器
		text := extractTextExclude(container, aNode)
		if len([]rune(text)) > 20 {
			return text
		}
		container = container.Parent
	}

	return ""
}

// extractTextExclude 提取节点文本，但排除某个特定子节点（通常是标题）
func extractTextExclude(root *html.Node, exclude *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == exclude {
			return
		}
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return strings.Join(strings.Fields(b.String()), " ")
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

func isAllowedDomain(link string) bool {
	// 如果链接直接包含 dgut.edu.cn，直接通过
	if strings.Contains(link, "dgut.edu.cn") {
		return true
	}

	parsed, err := url.Parse(link)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Host)
	// 允许 dgut.edu.cn 的所有子域名
	return strings.HasSuffix(host, "dgut.edu.cn")
}
