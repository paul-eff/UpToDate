# UpToDate

A simple web monitoring tool that checks websites for specific content and sends notifications when patterns are found or not found.

## Features

- **Dual Fetch Modes**: HTTP client (no dependencies) or headless browser (with embedded Chromium)
- **Flexible Search**: Search for exact strings or regex patterns
- **XPath Support**: Target specific page elements
- **Multiple Notifications**: Email (SMTP), Discord webhooks, and Slack webhooks
- **Configurable Monitoring**: Set custom intervals for continuous monitoring
- **Error Notification**: Automatic notifications when errors occur

## Installation

Use the provided `uptodate` binary provided in this repository or:

1. Install dependencies:
```bash
go mod tidy
```

2. Build the application:
```bash
go build -o uptodate
```

## Configuration

Create a `config.json` file based on the [provided examples](./examples):

### HTTP Mode Example (no Chromium required)
```json
{
  "url": "https://example.com",
  "fetch_method": "http",
  "search": {
    "type": "string",
    "pattern": "Sale",
    "notify_on": "found"
  },
  "notifications": {
    "email": {
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "your-email@gmail.com", 
      "password": "your-app-password",
      "from": "your-email@gmail.com",
      "to": "recipient@example.com"
    }
  },
  "interval": 300
}
```

### Browser Mode Example (will install Chromium to cache)
```json
{
  "url": "https://example.com/product",
  "fetch_method": "browser",
  "search": {
    "type": "regex",
    "pattern": "^[1-9]\\d*$",
    "xpath": "//div[@class='units_available']",
    "notify_on": "found"
  },
  "notifications": {
    "discord": {
      "webhook_url": "https://discord.com/api/webhooks/YOUR_WEBHOOK_URL"
    }
  },
  "interval": 600
}
```

### Configuration Options

- **url**: The webpage to monitor
- **fetch_method**: `"http"` (default) or `"browser"` (will install Chromium to cache)
- **search.type**: `"string"` or `"regex"`
- **search.pattern**: Text or regex pattern to search for
- **search.xpath**: Optional XPath to limit search scope (full XPath in browser mode, basic in HTTP mode)
- **search.notify_on**: `"found"` or `"not_found"`
- **interval**: Check interval in seconds (default: 300)
- **notifications**: At least one notification method required

### Fetch Methods

#### HTTP Mode (`"fetch_method": "http"`)
- **Pros**: No external dependencies, typically faster
- **Cons**: No JavaScript execution, basic XPath support only
- **Best for**: Static content, simple HTML pages, APIs

#### Browser Mode (`"fetch_method": "browser"`)
- **Pros**: Full JavaScript support, complete XPath support, handles dynamic content
- **Cons**: Slower, more resource intensive, more error prone
- **Best for**: Anything available via the browser

### Notification Setup

#### Email (SMTP)
- Use app passwords for Gmail/Outlook
- Common SMTP ports: 587 (TLS) or 465 (SSL)

#### Discord Webhook
1. Go to Server Settings → Integrations → Webhooks
2. Create webhook and copy URL

#### Slack Webhook  
1. Create Slack app at [api.slack.com](api.slack.com)
2. Add Incoming Webhooks feature
3. Copy webhook URL

## Usage

### Continuous Monitoring
```bash
./uptodate -config config.json
```

### Run Once
```bash
./uptodate -config config.json -once
```

### Command Line Options
- `-config`: Path to configuration file (default: `config.json`)
- `-once`: Run once and exit instead of continuous monitoring

## Docker

Use the `Dockerfile` and `docker-compose.yaml` to let UpToDate run in a containerized environment.

In your root directory
```bash
docker compose up -d
```
