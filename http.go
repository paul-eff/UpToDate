package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// HTTP handles HTTP-only page fetching
type HTTP struct {
	client *http.Client
}

// Close implements the Client interface
func (h *HTTP) Close() {
	//Not needed
}

// NewHTTP creates a new HTTP client
func NewHTTP() *HTTP {
	return &HTTP{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Fetch performs HTTP request and content extraction
func (h *HTTP) Fetch(config *Config) *Result {
	req, err := http.NewRequest("GET", config.URL, nil)
	if err != nil {
		return &Result{
			Found:   false,
			Content: "",
			Error:   fmt.Errorf("failed to create request: %w", err),
		}
	}

	// Set realistic user agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := h.client.Do(req)
	if err != nil {
		return &Result{
			Found:   false,
			Content: "",
			Error:   fmt.Errorf("failed to fetch URL: %w", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &Result{
			Found:   false,
			Content: "",
			Error:   fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Result{
			Found:   false,
			Content: "",
			Error:   fmt.Errorf("failed to read response body: %w", err),
		}
	}

	content := string(body)

	// TODO: No idea what I was thinking, but this should be implemented the same as browser is.
	// If XPath is specified, try to extract content from HTML
	if config.SearchConfig.XPath != "" {
		extracted, err := h.extractByXPath(content, config.SearchConfig.XPath)
		if err != nil {
			return &Result{
				Found:   false,
				Content: content,
				Error:   fmt.Errorf("XPath extraction failed: %w", err),
			}
		}
		content = extracted
	} else {
		// Extract text content from HTML
		content = h.extractTextFromHTML(content)
	}

	// Perform the search
	found, matches, err := h.performSearch(content, &config.SearchConfig)
	if err != nil {
		return &Result{
			Found:   false,
			Content: content,
			Error:   fmt.Errorf("search failed: %w", err),
			Matches: nil,
		}
	}

	return &Result{
		Found:   found,
		Content: content,
		Error:   nil,
		Matches: matches,
	}
}

// extractTextFromHTML extracts text content from HTML
func (h *HTTP) extractTextFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// If HTML parsing fails, return original content
		return htmlContent
	}

	var textContent strings.Builder
	h.extractText(doc, &textContent)
	return textContent.String()
}

// extractText recursively extracts text from HTML nodes
func (h *HTTP) extractText(n *html.Node, textContent *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			textContent.WriteString(text + " ")
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		h.extractText(c, textContent)
	}
}

// extractByXPath attempts basic XPath-like extraction (simplified)
func (h *HTTP) extractByXPath(htmlContent, xpath string) (string, error) {

	// Convert simple XPath patterns to element extraction
	if strings.Contains(xpath, "//") {
		// Simple tag extraction like //div, //span, etc.
		tagPattern := regexp.MustCompile(`//(\w+)`)
		matches := tagPattern.FindStringSubmatch(xpath)
		if len(matches) > 1 {
			tagName := matches[1]
			return h.extractByTag(htmlContent, tagName), nil
		}
	}

	// For class-based selectors like //div[@class='price']
	classPattern := regexp.MustCompile(`//(\w+)\[@class=['"]([^'"]+)['"]`)
	matches := classPattern.FindStringSubmatch(xpath)
	if len(matches) > 2 {
		tagName := matches[1]
		className := matches[2]
		return h.extractByTagAndClass(htmlContent, tagName, className), nil
	}

	// Fallback: return original content
	return htmlContent, nil
}

// extractByTag extracts content from all elements with specified tag
func (h *HTTP) extractByTag(htmlContent, tagName string) string {
	pattern := fmt.Sprintf(`<%s[^>]*>(.*?)</%s>`, tagName, tagName)
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var result strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			// Remove HTML tags from content
			cleanContent := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(match[1], "")
			result.WriteString(strings.TrimSpace(cleanContent) + " ")
		}
	}

	return result.String()
}

// extractByTagAndClass extracts content from elements with specified tag and class
func (h *HTTP) extractByTagAndClass(htmlContent, tagName, className string) string {
	pattern := fmt.Sprintf(`<%s[^>]*class=['"][^'"]*%s[^'"]*['"][^>]*>(.*?)</%s>`, tagName, className, tagName)
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var result strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			// Remove HTML tags from content
			cleanContent := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(match[1], "")
			result.WriteString(strings.TrimSpace(cleanContent) + " ")
		}
	}

	return result.String()
}

// performSearch executes the search based on configuration
func (h *HTTP) performSearch(content string, searchConfig *SearchConfig) (bool, []string, error) {
	switch strings.ToLower(searchConfig.Type) {
	case "string":
		found := strings.Contains(content, searchConfig.Pattern)
		var matches []string
		if found {
			matches = []string{searchConfig.Pattern}
		}
		return found, matches, nil
	case "regex":
		re, err := regexp.Compile(searchConfig.Pattern)
		if err != nil {
			return false, nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		matches := re.FindAllString(content, -1)
		return len(matches) > 0, matches, nil
	case "compound":
		compound, err := ParseCompoundPattern(searchConfig.Pattern)
		if err != nil {
			return false, nil, fmt.Errorf("invalid compound pattern: %w", err)
		}
		return EvaluateCompoundPattern(compound, content)
	default:
		return false, nil, fmt.Errorf("unsupported search type: %s", searchConfig.Type)
	}
}
