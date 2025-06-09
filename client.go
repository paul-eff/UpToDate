package main

// Client interface for different fetch methods
// Defines contract for web content fetching and resource cleanup
type Client interface {
	Fetch(config *Config) *Result
	Close()
}

// Result holds the result of a fetch operation
// Stores whether patterns matched, extracted content, and any errors
type Result struct {
	Found   bool
	Content string
	Error   error
	Matches []string // Regex matches found in content
}
