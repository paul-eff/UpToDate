# UpToDate

A simple web monitoring tool that checks websites for specific content and sends notifications when patterns are found or not found.

## Features

- **Dual Fetch Modes**: HTTP client (no dependencies) or headless browser (with embedded Chromium)
- **Flexible Search**: Search for exact strings, regex patterns, or compound patterns with AND/OR logic
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
  "interval": 300
}
```

### Compound Pattern Example (NEW!)
```json
{
  "url": "https://example.com/product",
  "fetch_method": "http",
  "search": {
    "type": "compound",
    "pattern": "(string:sale OR string:discount) AND regex:\\€[0-9]+\\.[0-9]{2}",
    "notify_on": "found"
  },
  "notifications": {
    "discord": {
      "webhook_url": "YOUR_DISCORD_WEBHOOK_URL"
    }
  },
  "interval": 300
}
```

### Configuration Options

- **url**: The webpage to monitor
- **fetch_method**: `"http"` (default) or `"browser"` (will install Chromium to cache)
- **search.type**: `"string"`, `"regex"`, or `"compound"`
- **search.pattern**: Text, regex pattern, or compound pattern to search for
- **search.xpath**: Optional XPath to limit search scope (full XPath in browser mode, basic in HTTP mode)
- **search.notify_on**: `"found"` or `"not_found"`
- **notifications**: At least one notification method required
- **interval**: Check interval in seconds (default: 300)

### Compound Pattern Syntax

Compound patterns allow you to combine multiple search criteria using AND/OR logic:

#### Basic Operators
- **AND**: Both patterns must be found
- **OR**: Either pattern must be found

#### Pattern Types
- **pattern** - Defaults to string matching
- **string:pattern** - Exact text matching
- **string:\"this AND that\"** - Exact text matching, but containing a control word
- **regex:pattern** - Regular expression matching

#### Examples
```json
// Simple AND - both must be found
"pattern1 AND pattern2"

// Simple OR - either can be found  
"pattern1 OR pattern2"

// Mixed types - combine string and regex
"string:sale AND regex:[0-9]+\\.[0-9]{2}"

// Parentheses grouping
"(urgent OR breaking) AND news"

// Complex combinations
"regex:error|failed OR (string:maintenance AND string:scheduled)"

// Price monitoring
"string:price AND regex:\\$[0-9]+\\.[0-9]{2}"

// Stock alerts
"(string:in stock OR string:available) AND regex:item-[0-9]+"

// Quoted patterns for text with AND/OR keywords, or special characters
"string:\"HOT AND COLD\" OR string:DECAPITATED"

// Patterns with colons or special characters
"string:\"Time: 12:30:45\" AND regex:[0-9]{2}:[0-9]{2}"

// Escaped quotes in patterns
"string:\"He said \\\"hello\\\" to me\" AND test"
```

#### Important Notes for Complex Patterns
- **Use quotes** when your pattern contains AND/OR keywords, or colons
- **Escape special characters** inside quoted patterns with backslash: `\.`
- **Type prefixes** work with quoted strings: `string:"text with spaces"`
- **Regex patterns** can contain any characters when quoted: `regex:"\\$[0-9]+\\.[0-9]{2}"`

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
