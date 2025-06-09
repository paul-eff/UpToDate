# UpToDate 🔍

**Never miss important website changes again!**

UpToDate monitors websites and notifies you when specific content appears or disappears. Perfect for tracking price drops, stock availability, news alerts, or any content that matters to you.

## ✨ What Can It Do? What you want it to!

- **🛍️ Price Monitoring** - Get notified when prices drop or sales start
- **📦 Stock Alerts** - Know instantly when items come back in stock  
- **📰 News Tracking** - Stay updated on breaking news or specific topics
- **🚨 Service Monitoring** - Monitor API health, error rates, or uptime
- **📈 Content Changes** - Track any text, numbers, or patterns on websites

## 🚀 Key Features

- **Smart Pattern Matching** - Find exact text, use regex, or combine multiple conditions
- **Multiple Notifications** - Email, Discord, and Slack alerts
- **Real Browser Engine** - Handles JavaScript and dynamic content perfectly
- **Flexible Scheduling** - Check every minute or once a day
- **XPath Support** - Target specific page elements precisely
- **Docker Ready** - Easy deployment with Docker Compose

## 🏃‍♂️ Quick Start

### 1. Download & Install
```bash
# Download the binary or build from source
go mod tidy
go build -o uptodate
```

### 2. Create Your First Monitor
Create a `config.json` file:

```json
{
  "url": "https://example-store.com/product/123",
  "search": {
    "type": "string",
    "pattern": "In Stock",
    "notify_on": "found"
  },
  "notifications": {
    "discord": {
      "webhook_url": "https://discord.com/api/webhooks/YOUR_WEBHOOK_URL"
    }
  },
  "interval": 300
}
```

### 3. Start Monitoring
```bash
# Using the binary
./uptodate -config config.json
# Using docker compose
docker compose up -d
```

That's it! You'll get notified on Discord when "In Stock" appears on the page.

## 📋 Configuration Examples

### Price Drop Alert
*Monitors a product page and sends email alerts when the price drops to $10-$49 range.*

```json
{
  "url": "https://store.com/expensive-item",
  "search": {
    "type": "regex",
    "pattern": "\\$[1-4][0-9]\\.[0-9]{2}",
    "xpath": "//div[@class='price']",
    "notify_on": "found"
  },
  "notifications": {
    "email": {
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "your-email@gmail.com",
      "password": "your-app-password",
      "from": "your-email@gmail.com",
      "to": "alerts@example.com",
      "subject": "💰 Price Drop Alert!"
    }
  },
  "interval": 600
}
```

### News Monitoring
*Watches a news site and sends Slack alerts when breaking news articles with dates are published.*

```json
{
  "url": "https://news-site.com",
  "search": {
    "type": "compound",
    "pattern": "string:'breaking news' AND regex:[0-9]{4}-[0-9]{2}-[0-9]{2}",
    "notify_on": "found"
  },
  "notifications": {
    "slack": {
      "webhook_url": "YOUR_SLACK_WEBHOOK"
    }
  },
  "interval": 60
}
```

### Service Health Check
*Monitors an API health endpoint and sends alerts to Discord and email when the service becomes unhealthy.*

```json
{
  "url": "https://api.yourservice.com/health",
  "search": {
    "type": "string",
    "pattern": "healthy",
    "notify_on": "not_found"
  },
  "notifications": {
    "discord": {
      "webhook_url": "YOUR_DISCORD_WEBHOOK"
    },
    "email": {
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "monitoring@company.com",
      "password": "password",
      "from": "monitoring@company.com",
      "to": "devops@company.com",
      "subject": "🚨 Service Down!"
    }
  },
  "interval": 120
}
```

## 🔧 Configuration Options

### Required Settings
- **`url`** - The webpage to monitor
- **`search.pattern`** - What to look for on the page
- **`notifications`** - At least one notification method (email, discord, or slack)

### Search Options
- **`search.type`** - `"string"` (exact text), `"regex"` (pattern), or `"compound"` (multiple conditions)
- **`search.notify_on`** - `"found"` (notify when pattern is found) or `"not_found"` (notify when pattern is not found)
- **`search.xpath`** - Optional: target specific page elements (e.g., `"//div[@class='price']"`)

### Timing
- **`interval`** - How often to check in seconds (default: 300 = 5 minutes)

## 📧 Setting Up Notifications

### Email (SMTP)
Works with most common email providers (Gmail, Hotmail, ...):

```json
"email": {
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "username": "your-email@gmail.com",
  "password": "your-app-password",
  "from": "your-email@gmail.com",
  "to": "alerts@example.com",
  "subject": "UpToDate Alert"
}
```

**Gmail Setup:**
1. Enable 2-factor authentication
2. Generate an "App Password" 
3. Use the app password instead of your regular password

### Discord Webhook
1. Go to your Discord server
2. Server Settings → Integrations → Webhooks
3. Create webhook and copy the URL

```json
"discord": {
  "webhook_url": "https://discord.com/api/webhooks/YOUR_WEBHOOK_URL"
}
```

### Slack Webhook
1. Create a Slack app
2. Add "Incoming Webhooks" feature
3. Copy the webhook URL

```json
"slack": {
  "webhook_url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
}
```

## 🎯 Pattern Matching Guide

### Simple Text Search
```json
"search": {
  "type": "string",
  "pattern": "Sale"
}
```

### Regular Expressions
```json
"search": {
  "type": "regex", 
  "pattern": "\\$[0-9]+\\.[0-9]{2}"
}
```

### Complex Conditions (Compound Patterns)
Use `AND` and `OR` to combine multiple conditions:

```json
"search": {
  "type": "compound",
  "pattern": "string:'price drop' AND regex:\\$[0-9]+\\.[0-9]{2}"
}
```

**Compound Pattern Examples:**
- `"string:sale OR string:discount"` - Either word appears
- `"string:'breaking news' AND regex:[0-9]{4}"` - Both conditions must be true
- `"(string:error OR string:failed) AND regex:[0-9]{2}:[0-9]{2}"` - Use parentheses for grouping

**Important:** Use single quotes for text containing spaces or special characters: `string:'Hot Deal'`

### XPath Targeting
Target specific page elements:

```json
"search": {
  "type": "string",
  "pattern": "In Stock",
  "xpath": "//div[@class='availability']"
}
```

Common XPath examples:
- `"//div[@class='price']"` - Element with specific class
- `"//span[@id='stock-status']"` - Element with specific ID
- `"//h1"` - All H1 headings
- `"//div[contains(@class, 'product')]"` - Class contains text

## 🚀 Running UpToDate

### Command Line Options
```bash
# Continuous monitoring (default)
./uptodate -config config.json

# Run once and exit
./uptodate -config config.json -once

# Use different config file
./uptodate -config /path/to/my-config.json
```

### Docker
```bash
# Using docker-compose
docker compose up -d
```

## ❓ Troubleshooting

**"Pattern not found" but you can see it on the page**
- The content might be lazy loaded by JavaScript - Currently this is not handled, feel free to create an issue.
- Check if the text is in a specific element using XPath
- Verify the exact text (case-sensitive)

**Getting too many notifications**
- Increase the `interval` (time between checks)
- Make your pattern more specific
- Use XPath to target a specific page section

**Pattern syntax errors**
- Use single quotes for text with spaces, AND/OR or special characters: `string:'hello world'`
- Escape special regex characters: `\\$` for dollar signs
- Check parentheses are balanced in compound patterns

---

## 👨‍💻 Developer Section

### Architecture Overview

UpToDate uses a modular architecture:
- **Browser Engine**: go-rod with embedded Chromium for JavaScript support
- **Pattern Parser**: Recursive descent parser for compound patterns  
- **Notification System**: Plugin-style support for multiple channels
- **Configuration**: JSON-based with validation

### Building from Source
```bash
git clone https://github.com/paul-eff/UpToDate
cd UpToDate
go mod tidy
go build -o uptodate
```

### Project Structure
```
├── main.go              # Application entry point
├── config.go            # Configuration parsing & compound patterns  
├── browser.go           # Browser-based web fetching
├── client.go            # Client interface
├── notifications.go     # Multi-channel notification system
└── examples/            # Configuration examples
```

### Adding New Notification Channels

1. Add config struct to `config.go`:
```go
type MyServiceConfig struct {
    APIKey string `json:"api_key"`
    Channel string `json:"channel"`
}
```

2. Add to main Notifications struct:
```go
type Notifications struct {
    Email   *EmailConfig     `json:"email,omitempty"`
    Discord *DiscordConfig   `json:"discord,omitempty"`
    Slack   *SlackConfig     `json:"slack,omitempty"`
    MyService *MyServiceConfig `json:"myservice,omitempty"`
}
```

3. Implement sending in `notifications.go`:
```go
func (ns *NotificationService) sendMyService(message string) error {
    // Implementation here
}
```

### Pattern Parser Details

The compound pattern parser uses recursive descent parsing:
- **Tokenization**: Handles quoted strings and operators
- **Precedence**: AND has higher precedence than OR
- **Quotes**: Single quotes preferred (no JSON escaping needed)
- **Evaluation**: Returns boolean result and matched strings

### Configuration Validation

All configuration is validated at startup:
- Required fields checked
- Pattern syntax validated  
- Notification channels verified
- Defaults applied where appropriate

### Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## ⭐ Support

If UpToDate helps you catch that perfect deal or stay informed, please consider giving it a star!