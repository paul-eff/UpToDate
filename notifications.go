package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"time"
)

// NotificationService handles sending notifications
type NotificationService struct {
	config *Config
}

// NewNotificationService creates a new notification service
func NewNotificationService(config *Config) *NotificationService {
	return &NotificationService{config: config}
}

// SendNotification sends notifications based on fetch results
func (ns *NotificationService) SendNotification(result *Result) error {
	shouldNotify := ns.shouldNotify(result)
	if !shouldNotify {
		return nil
	}

	message := ns.buildMessage(result)
	reason := ns.getNotificationReason(result)

	var errors []error
	var sentChannels []string

	// Send email notification
	if ns.config.Notifications.Email != nil {
		if err := ns.sendEmail(message); err != nil {
			errors = append(errors, fmt.Errorf("email notification failed: %w", err))
		} else {
			sentChannels = append(sentChannels, "email")
		}
	}

	// Send Discord notification
	if ns.config.Notifications.Discord != nil {
		if err := ns.sendDiscord(message); err != nil {
			errors = append(errors, fmt.Errorf("discord notification failed: %w", err))
		} else {
			sentChannels = append(sentChannels, "discord")
		}
	}

	// Send Slack notification
	if ns.config.Notifications.Slack != nil {
		if err := ns.sendSlack(message); err != nil {
			errors = append(errors, fmt.Errorf("slack notification failed: %w", err))
		} else {
			sentChannels = append(sentChannels, "slack")
		}
	}

	// Log notification status
	if len(sentChannels) > 0 {
		log.Printf("Notification sent via %v - Reason: %s", sentChannels, reason)
	}

	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %v", errors)
	}

	return nil
}

// shouldNotify determines if notifications should be sent
func (ns *NotificationService) shouldNotify(result *Result) bool {
	if result.Error != nil {
		return true
	}

	notifyOn := ns.config.SearchConfig.NotifyOn
	switch notifyOn {
	case "found":
		return result.Found
	case "not_found":
		return !result.Found
	default:
		return result.Found // Default to notifying when found
	}
}

// getNotificationReason returns the reason for sending the notification
func (ns *NotificationService) getNotificationReason(result *Result) string {
	if result.Error != nil {
		return "fetch error occurred"
	}

	notifyOn := ns.config.SearchConfig.NotifyOn
	switch notifyOn {
	case "found":
		if result.Found {
			return "pattern found"
		}
	case "not_found":
		if !result.Found {
			return "pattern not found"
		}
	default:
		if result.Found {
			return "pattern found (default)"
		}
	}
	return "unknown reason"
}

// buildMessage creates a notification message
func (ns *NotificationService) buildMessage(result *Result) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	if result.Error != nil {
		return fmt.Sprintf("[%s] Error monitoring %s: %s", timestamp, ns.config.URL, result.Error.Error())
	}

	status := "NOT FOUND"
	if result.Found {
		status = "FOUND"
	}

	message := fmt.Sprintf("[%s] Pattern '%s' %s on %s",
		timestamp,
		ns.config.SearchConfig.Pattern,
		status,
		ns.config.URL)

	// Add matches if found and using regex
	if result.Found && len(result.Matches) > 0 {
		message += "\n\nMatches found:"
		for i, match := range result.Matches {
			message += fmt.Sprintf("\n  [%d] %s", i+1, match)
		}
	}

	return message
}

// sendEmail sends email notification
func (ns *NotificationService) sendEmail(message string) error {
	emailConfig := ns.config.Notifications.Email

	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.SMTPHost)

	to := []string{emailConfig.To}
	subject := emailConfig.Subject
	if subject == "" {
		subject = "UpToDate Monitoring Alert"
	}

	body := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", emailConfig.To, subject, message)

	addr := fmt.Sprintf("%s:%d", emailConfig.SMTPHost, emailConfig.SMTPPort)
	return smtp.SendMail(addr, auth, emailConfig.From, to, []byte(body))
}

// DiscordWebhook represents a Discord webhook payload
type DiscordWebhook struct {
	Content string `json:"content"`
}

// sendDiscord sends Discord webhook notification
func (ns *NotificationService) sendDiscord(message string) error {
	webhook := DiscordWebhook{Content: message}

	jsonData, err := json.Marshal(webhook)
	if err != nil {
		return err
	}

	resp, err := http.Post(ns.config.Notifications.Discord.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// SlackWebhook represents a Slack webhook payload
type SlackWebhook struct {
	Text string `json:"text"`
}

// sendSlack sends Slack webhook notification
func (ns *NotificationService) sendSlack(message string) error {
	webhook := SlackWebhook{Text: message}

	jsonData, err := json.Marshal(webhook)
	if err != nil {
		return err
	}

	resp, err := http.Post(ns.config.Notifications.Slack.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}
