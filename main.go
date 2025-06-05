package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var configFile string
	var runOnce bool
	
	flag.StringVar(&configFile, "config", "config.json", "Path to configuration file")
	flag.BoolVar(&runOnce, "once", false, "Run once and exit (don't monitor continuously)")
	flag.Parse()

	// Load configuration
	config, err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create client based on configuration
	var client Client
	fetchMethod := config.FetchMethod
	if fetchMethod == "" {
		fetchMethod = "http" // Default to HTTP for better compatibility
	}
	
	switch fetchMethod {
	case "browser":
		client = NewBrowser()
	case "http":
		client = NewHTTP()
	default:
		log.Fatalf("Invalid fetch_method: %s (must be 'browser' or 'http')", fetchMethod)
	}
	defer client.Close()
	
	notificationService := NewNotificationService(config)

	log.Printf("Starting UpToDate monitoring for: %s", config.URL)
	log.Printf("Fetch method: %s", fetchMethod)
	log.Printf("Search type: %s, pattern: %s", config.SearchConfig.Type, config.SearchConfig.Pattern)
	log.Printf("Notify on: %s", config.SearchConfig.NotifyOn)

	if runOnce {
		// Run once and exit
		runFetch(client, notificationService, config)
		return
	}

	// Set up signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Create ticker for periodic monitoring
	interval := time.Duration(config.Interval) * time.Second
	if interval == 0 {
		interval = 300 * time.Second // Default to 5 minutes
	}
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Monitoring every %v", interval)

	// Run initial fetch
	runFetch(client, notificationService, config)

	// Monitor loop
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

// runFetch performs a single fetch operation
func runFetch(client Client, notificationService *NotificationService, config *Config) {
	log.Printf("Fetching %s...", config.URL)
	
	result := client.Fetch(config)
	
	if result.Error != nil {
		log.Printf("Fetch error: %v", result.Error)
	} else {
		status := "not found"
		if result.Found {
			status = "found"
			if len(result.Matches) > 0 {
				log.Printf("Pattern '%s' %s. Matches found:", config.SearchConfig.Pattern, status)
				for i, match := range result.Matches {
					log.Printf("  [%d] %s", i+1, match)
				}
			} else {
				log.Printf("Pattern '%s' %s", config.SearchConfig.Pattern, status)
			}
		} else {
			log.Printf("Pattern '%s' %s", config.SearchConfig.Pattern, status)
		}
	}

	// Send notifications if needed
	if err := notificationService.SendNotification(result); err != nil {
		log.Printf("Notification error: %v", err)
	}
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if config.URL == "" {
		return fmt.Errorf("URL is required")
	}

	if config.SearchConfig.Pattern == "" {
		return fmt.Errorf("search pattern is required")
	}

	if config.SearchConfig.Type == "" {
		config.SearchConfig.Type = "string" // Default to string search
	}

	if config.SearchConfig.NotifyOn == "" {
		config.SearchConfig.NotifyOn = "found" // Default to notify when found
	}

	// Validate that at least one notification method is configured
	notifications := config.Notifications
	if notifications.Email == nil && notifications.Discord == nil && notifications.Slack == nil {
		return fmt.Errorf("at least one notification method must be configured")
	}

	// Validate email config if present
	if notifications.Email != nil {
		email := notifications.Email
		if email.SMTPHost == "" || email.Username == "" || email.Password == "" || 
		   email.From == "" || email.To == "" {
			return fmt.Errorf("email configuration is incomplete")
		}
		if email.SMTPPort == 0 {
			email.SMTPPort = 587 // Default SMTP port
		}
	}

	// Validate Discord config if present
	if notifications.Discord != nil && notifications.Discord.WebhookURL == "" {
		return fmt.Errorf("discord webhook URL is required")
	}

	// Validate Slack config if present
	if notifications.Slack != nil && notifications.Slack.WebhookURL == "" {
		return fmt.Errorf("slack webhook URL is required")
	}

	return nil
}