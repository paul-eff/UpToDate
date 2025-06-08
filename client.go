package main

// Client interface for different fetch methods
type Client interface {
	Fetch(config *Config) *Result
	Close()
}

// Result holds the result of a fetch operation
type Result struct {
	Found   bool
	Content string
	Error   error
	Matches []string // Regex matches found in content
}
