package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// Browser handles web operations using embedded browser
type Browser struct {
	browser *rod.Browser
}

// Client interface for different fetch methods
type Client interface {
	Fetch(config *Config) *Result
	Close()
}

// Result holds the result of a fetch operation
type Result struct {
	Found     bool
	Content   string
	Error     error
	Matches   []string // Regex matches found in content
}

// NewBrowser creates a new browser instance
func NewBrowser() *Browser {
	// Launch headless browser with embedded Chromium
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
func (b *Browser) Fetch(config *Config) *Result {
	return b.fetch(config)
}

// fetch performs the web fetching and search operation
func (b *Browser) fetch(config *Config) *Result {
	var content string
	var err error

	// Create a new page with timeout
	page := b.browser.Timeout(30 * time.Second).MustPage()
	defer page.Close()

	// Navigate to the URL
	err = page.Navigate(config.URL)
	if err != nil {
		return &Result{
			Found:   false,
			Content: "",
			Error:   fmt.Errorf("failed to navigate to page: %w", err),
		}
	}

	// Wait for page to load
	page.MustWaitLoad()

	if config.SearchConfig.XPath != "" {
		// Search within specific XPath
		elements, err := page.ElementsX(config.SearchConfig.XPath)
		if err != nil {
			return &Result{
				Found:   false,
				Content: "",
				Error:   fmt.Errorf("failed to find XPath elements: %w", err),
			}
		}

		if len(elements) > 0 {
			content = elements[0].MustText()
		}
	} else {
		// Search entire page content
		content = page.MustElement("body").MustText()
	}

	// Perform the search
	found, matches, err := b.performSearch(content, &config.SearchConfig)
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

// performSearch executes the search based on configuration
func (b *Browser) performSearch(content string, searchConfig *SearchConfig) (bool, []string, error) {
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

// shouldNotify determines if a notification should be sent based on the result
func (b *Browser) shouldNotify(result *Result, notifyOn string) bool {
	if result.Error != nil {
		// Always notify on errors
		return true
	}

	switch strings.ToLower(notifyOn) {
	case "found":
		return result.Found
	case "not_found":
		return !result.Found
	default:
		log.Printf("Unknown notify_on value: %s, defaulting to 'found'", notifyOn)
		return result.Found
	}
}
