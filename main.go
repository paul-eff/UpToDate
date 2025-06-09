package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Parse command line flags for configuration file and execution mode
	var configFile string
	var runOnce bool

	flag.StringVar(&configFile, "config", "config.json", "Path to config file.")
	flag.BoolVar(&runOnce, "once", false, "Run once and exit.")
	flag.Parse()

	// Load JSON configuration from file and validate all settings
	config, err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := validateConfig(config); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create browser instance and notification service from config
	client := NewBrowser()
	defer client.Close()

	notificationService := NewNotificationService(config)

	log.Printf("Starting UpToDate monitoring for: %s", config.URL)
	log.Printf("Search type: %s, pattern: %s", config.SearchConfig.Type, config.SearchConfig.Pattern)
	log.Printf("Notify on: %s", config.SearchConfig.NotifyOn)

	// Execute single fetch when -once flag is provided
	if runOnce {
		runFetch(client, notificationService, config)
		return
	}

	// Set up signal handling for graceful shutdown and configure monitoring interval
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	interval := time.Duration(config.Interval) * time.Second
	if interval == 0 {
		interval = 300 * time.Second // Default to 5 minutes
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Monitoring every %v", interval)

	// Run first fetch immediately before starting timer
	runFetch(client, notificationService, config)

	// Wait for timer ticks or shutdown signals in infinite loop
	for {
		select {
		case <-ticker.C:
			runFetch(client, notificationService, config)
		case <-c:
			log.Println("Received shutdown signal, exiting...")
			return
		}
	}
}

// runFetch performs a single fetch operation and handles logging
func runFetch(client Client, notificationService *NotificationService, config *Config) {
	log.Printf("Fetching %s...", config.URL)

	// Call browser client to fetch page and search for patterns
	result := client.Fetch(config)

	// Output search results and any regex matches to console
	if result.Error != nil {
		log.Printf("Fetch error: %v", result.Error)
	} else {
		status := "not found"
		if result.Found {
			status = "found"
		}
		
		if result.Found && len(result.Matches) > 0 {
			log.Printf("Pattern '%s' %s. Matches found:", config.SearchConfig.Pattern, status)
			for i, match := range result.Matches {
				log.Printf("  [%d] %s", i+1, match)
			}
		} else {
			log.Printf("Pattern '%s' %s", config.SearchConfig.Pattern, status)
		}
	}

	// Send notifications if conditions are met based on search outcome
	if err := notificationService.SendNotification(result); err != nil {
		log.Printf("Notification error: %v", err)
	}
}

// validateConfig validates configuration and applies defaults
// Checks required fields and sets defaults for missing optional values
func validateConfig(config *Config) error {
	if config.URL == "" {
		return fmt.Errorf("URL is required")
	}

	if config.SearchConfig.Pattern == "" {
		return fmt.Errorf("search pattern is required")
	}

	if config.SearchConfig.Type == "" {
		config.SearchConfig.Type = "string"
	}

	// Parse compound patterns to validate syntax before monitoring starts
	if strings.ToLower(config.SearchConfig.Type) == "compound" {
		_, err := ParseCompoundPattern(config.SearchConfig.Pattern)
		if err != nil {
			return fmt.Errorf("invalid compound pattern: %w", err)
		}
	}

	if config.SearchConfig.NotifyOn == "" {
		config.SearchConfig.NotifyOn = "found"
	}

	// Ensure at least one notification method is available
	notifications := config.Notifications
	if notifications.Email == nil && notifications.Discord == nil && notifications.Slack == nil {
		return fmt.Errorf("at least one notification method must be configured")
	}

	// Check all required SMTP fields and apply default port and subject
	if notifications.Email != nil {
		email := notifications.Email
		if email.SMTPHost == "" || email.Username == "" || email.Password == "" ||
			email.From == "" || email.To == "" {
			return fmt.Errorf("email configuration is incomplete")
		}
		if email.SMTPPort == 0 {
			email.SMTPPort = 587
		}
		if email.Subject == "" {
			email.Subject = "UpToDate Alert!"
		}
	}

	if notifications.Discord != nil && notifications.Discord.WebhookURL == "" {
		return fmt.Errorf("discord webhook URL is required")
	}

	if notifications.Slack != nil && notifications.Slack.WebhookURL == "" {
		return fmt.Errorf("slack webhook URL is required")
	}

	return nil
}
