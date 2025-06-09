package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// Browser handles web operations using embedded browser
// Wraps go-rod browser instance for headless web content fetching
type Browser struct {
	browser *rod.Browser
}

// NewBrowser creates a new browser instance
func NewBrowser() *Browser {
	// Start headless Chromium browser and connect to control interface
	l := launcher.New().Headless(true)
	url := l.MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()

	return &Browser{
		browser: browser,
	}
}

// Close releases the browser resources
func (b *Browser) Close() {
	if b.browser != nil {
		b.browser.MustClose()
	}
}

// Fetch implements the Client interface for browser-based fetching
// Creates page, navigates to URL, extracts content, and searches for patterns
func (b *Browser) Fetch(config *Config) *Result {
	var content string
	var err error

	// Open new browser tab with 30 second timeout
	page := b.browser.Timeout(30 * time.Second).MustPage()
	defer page.Close()

	// Load the specified URL in the browser
	if err = page.Navigate(config.URL); err != nil {
		return &Result{
			Error: fmt.Errorf("failed to navigate to page: %w", err),
		}
	}

	// Wait for page to finish loading including JavaScript execution
	page.MustWaitLoad()

	// Extract text content using XPath selector or entire page body
	if config.SearchConfig.XPath != "" {
		// Find elements matching the XPath expression
		elements, err := page.ElementsX(config.SearchConfig.XPath)
		if err != nil {
			return &Result{
				Error: fmt.Errorf("failed to find XPath elements: %w", err),
			}
		}

		if len(elements) > 0 {
			content = elements[0].MustText()
		}
	} else {
		// Get all text content from the page body element
		content = page.MustElement("body").MustText()
	}

	// Search the extracted text using configured pattern type
	found, matches, err := b.performSearch(content, &config.SearchConfig)
	if err != nil {
		return &Result{
			Content: content,
			Error:   fmt.Errorf("search failed: %w", err),
		}
	}

	return &Result{
		Found:   found,
		Content: content,
		Error:   nil,
		Matches: matches,
	}
}

// performSearch executes search based on configuration
// Handles string, regex, and compound pattern matching
func (b *Browser) performSearch(content string, searchConfig *SearchConfig) (bool, []string, error) {
	switch strings.ToLower(searchConfig.Type) {
	case "string":
		// Check if pattern text appears anywhere in content
		found := strings.Contains(content, searchConfig.Pattern)
		matches := []string{}
		if found {
			matches = []string{searchConfig.Pattern}
		}
		return found, matches, nil
	case "regex":
		// Compile regex and find all matches in content
		re, err := regexp.Compile(searchConfig.Pattern)
		if err != nil {
			return false, nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		matches := re.FindAllString(content, -1)
		return len(matches) > 0, matches, nil
	case "compound":
		// Parse and evaluate boolean pattern expressions
		compound, err := ParseCompoundPattern(searchConfig.Pattern)
		if err != nil {
			return false, nil, fmt.Errorf("invalid compound pattern: %w", err)
		}
		return EvaluateCompoundPattern(compound, content)
	default:
		return false, nil, fmt.Errorf("unsupported search type: %s", searchConfig.Type)
	}
}
