package main

import (
	"encoding/json"
	"os"
)

// Config holds the application configuration
type Config struct {
	URL          string        `json:"url"`
	SearchConfig SearchConfig  `json:"search"`
	Notifications Notifications `json:"notifications"`
	Interval     int           `json:"interval"` // in seconds
	FetchMethod  string        `json:"fetch_method"` // "browser" or "http"
}

// SearchConfig defines what to search for and how
type SearchConfig struct {
	Type     string `json:"type"`     // "string", "regex"
	Pattern  string `json:"pattern"`  // the search pattern
	XPath    string `json:"xpath"`    // optional xpath selector
	NotifyOn string `json:"notify_on"` // "found" or "not_found"
}

// Notifications configuration for various notification methods
type Notifications struct {
	Email   *EmailConfig   `json:"email,omitempty"`
	Discord *DiscordConfig `json:"discord,omitempty"`
	Slack   *SlackConfig   `json:"slack,omitempty"`
}

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
	Subject  string `json:"subject"`
}

// DiscordConfig holds Discord webhook configuration
type DiscordConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// SlackConfig holds Slack webhook configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}