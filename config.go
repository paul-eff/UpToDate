package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Config holds the application configuration
type Config struct {
	URL           string        `json:"url"`
	SearchConfig  SearchConfig  `json:"search"`
	Notifications Notifications `json:"notifications"`
	Interval      int           `json:"interval"`
	FetchMethod   string        `json:"fetch_method"` // "browser" or "http"
}

// SearchConfig defines what to search for and how
type SearchConfig struct {
	Type     string `json:"type"` // "string", "regex", "compound"
	Pattern  string `json:"pattern"`
	XPath    string `json:"xpath"`
	NotifyOn string `json:"notify_on"` // "found" or "not_found"
}

// CompoundPattern represents a parsed compound search pattern with AND/OR operations
type CompoundPattern struct {
	Operator string           // "AND" or "OR"
	Patterns []PatternElement // Individual patterns or nested compounds
}

// PatternElement represents either a simple pattern or a nested compound pattern
type PatternElement struct {
	Type     string // "string", "regex", "compound"
	Pattern  string
	Compound *CompoundPattern // For nested compound patterns
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

// ParseCompoundPattern parses a compound pattern string into a CompoundPattern struct
func ParseCompoundPattern(pattern string) (*CompoundPattern, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil, fmt.Errorf("empty pattern")
	}

	tokens, err := tokenizePattern(pattern)
	if err != nil {
		return nil, fmt.Errorf("tokenization error: %w", err)
	}

	return parseTokens(tokens)
}

// Token represents a parsed token from the pattern string
type Token struct {
	Type  string
	Value string
}

// tokenizePattern breaks down a pattern string into tokens, handling quoted strings
func tokenizePattern(pattern string) ([]Token, error) {
	var tokens []Token
	i := 0

	for i < len(pattern) {
		// Skip whitespace
		if pattern[i] == ' ' || pattern[i] == '\t' {
			i++
			continue
		}

		// Handle parentheses with placeholder
		if pattern[i] == '(' {
			tokens = append(tokens, Token{Type: "LPAREN", Value: "("})
			i++
			continue
		}
		if pattern[i] == ')' {
			tokens = append(tokens, Token{Type: "RPAREN", Value: ")"})
			i++
			continue
		}

		// Look for operators (AND, OR) - but only when not inside quotes
		if (i+4 <= len(pattern)) && (pattern[i:i+4] == " AND") && (i+4 >= len(pattern) || pattern[i+4] == ' ') {
			tokens = append(tokens, Token{Type: "AND", Value: "AND"})
			i += 4
			continue
		}
		if i+3 <= len(pattern) && pattern[i:i+3] == " OR" && (i+3 >= len(pattern) || pattern[i+3] == ' ') {
			tokens = append(tokens, Token{Type: "OR", Value: "OR"})
			i += 3
			continue
		}
		if i+4 <= len(pattern) && pattern[i:i+4] == "AND " {
			tokens = append(tokens, Token{Type: "AND", Value: "AND"})
			i += 4
			continue
		}
		if i+3 <= len(pattern) && pattern[i:i+3] == "OR " {
			tokens = append(tokens, Token{Type: "OR", Value: "OR"})
			i += 3
			continue
		}

		// Handle patterns that might contain quoted strings
		start := i
		inQuotes := false
		escaped := false

		for i < len(pattern) {
			char := pattern[i]

			if escaped {
				escaped = false
				i++
				continue
			}

			if char == '\\' {
				escaped = true
				i++
				continue
			}

			if char == '"' {
				inQuotes = !inQuotes
				i++
				continue
			}

			// If we're not in quotes, check for breaking conditions
			if !inQuotes {
				if char == '(' || char == ')' {
					break
				}

				// Check for operators
				if i+4 <= len(pattern) && (pattern[i:i+4] == " AND" || pattern[i:i+4] == "AND ") {
					break
				}
				if i+3 <= len(pattern) && (pattern[i:i+3] == " OR" || pattern[i:i+3] == "OR ") {
					break
				}
			}

			i++
		}

		value := strings.TrimSpace(pattern[start:i])
		if value != "" {
			tokens = append(tokens, Token{Type: "PATTERN", Value: value})
		}
	}

	return tokens, nil
}

// parseTokens parses a slice of tokens into a CompoundPattern
func parseTokens(tokens []Token) (*CompoundPattern, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens to parse")
	}

	compound, _, err := parseOrExpression(tokens, 0)
	return compound, err
}

// parseOrExpression handles OR operations (lowest precedence)
func parseOrExpression(tokens []Token, pos int) (*CompoundPattern, int, error) {
	left, newPos, err := parseAndExpression(tokens, pos)
	if err != nil {
		return nil, newPos, err
	}

	var elements []PatternElement
	elements = append(elements, PatternElement{Type: "compound", Compound: left})

	for newPos < len(tokens) && tokens[newPos].Type == "OR" {
		right, nextPos, err := parseAndExpression(tokens, newPos+1)
		if err != nil {
			return nil, nextPos, err
		}
		elements = append(elements, PatternElement{Type: "compound", Compound: right})
		newPos = nextPos
	}

	if len(elements) == 1 {
		return left, newPos, nil
	}

	return &CompoundPattern{Operator: "OR", Patterns: elements}, newPos, nil
}

// parseAndExpression handles AND operations (higher precedence)
func parseAndExpression(tokens []Token, pos int) (*CompoundPattern, int, error) {
	left, newPos, err := parsePrimary(tokens, pos)
	if err != nil {
		return nil, newPos, err
	}

	var elements []PatternElement
	elements = append(elements, PatternElement{Type: "compound", Compound: left})

	for newPos < len(tokens) && tokens[newPos].Type == "AND" {
		right, nextPos, err := parsePrimary(tokens, newPos+1)
		if err != nil {
			return nil, nextPos, err
		}
		elements = append(elements, PatternElement{Type: "compound", Compound: right})
		newPos = nextPos
	}

	if len(elements) == 1 {
		return left, newPos, nil
	}

	return &CompoundPattern{Operator: "AND", Patterns: elements}, newPos, nil
}

// parsePrimary handles parentheses and individual patterns
func parsePrimary(tokens []Token, pos int) (*CompoundPattern, int, error) {
	if pos >= len(tokens) {
		return nil, pos, fmt.Errorf("unexpected end of input")
	}

	token := tokens[pos]

	switch token.Type {
	case "LPAREN":
		// Parse expression inside parentheses
		compound, newPos, err := parseOrExpression(tokens, pos+1)
		if err != nil {
			return nil, pos, err
		}

		// Check for closing parenthesis
		if newPos >= len(tokens) || tokens[newPos].Type != "RPAREN" {
			return nil, pos, fmt.Errorf("expected closing parenthesis")
		}

		return compound, newPos + 1, nil

	case "PATTERN":
		element, err := parsePatternElement(token.Value)
		if err != nil {
			return nil, pos, err
		}

		compound := &CompoundPattern{
			Operator: "AND",
			Patterns: []PatternElement{element},
		}

		return compound, pos + 1, nil

	default:
		return nil, pos, fmt.Errorf("unexpected token type: %s", token.Type)
	}
}

// parsePatternElement parses a single pattern element with optional type prefix
func parsePatternElement(pattern string) (PatternElement, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return PatternElement{}, fmt.Errorf("empty pattern element")
	}

	// Check for type prefix (but be careful about colons in the pattern itself)
	colonIndex := strings.Index(pattern, ":")
	if colonIndex > 0 && colonIndex < len(pattern)-1 {
		// Check if this looks like a type prefix
		possibleType := strings.ToLower(strings.TrimSpace(pattern[:colonIndex]))
		if possibleType == "string" || possibleType == "regex" {
			patternValue := strings.TrimSpace(pattern[colonIndex+1:])
			if patternValue == "" {
				return PatternElement{}, fmt.Errorf("empty pattern value")
			}

			// Handle quoted patterns
			if len(patternValue) >= 2 && patternValue[0] == '"' && patternValue[len(patternValue)-1] == '"' {
				// Remove quotes and handle escape sequences
				unquoted, err := unquoteString(patternValue)
				if err != nil {
					return PatternElement{}, fmt.Errorf("error parsing quoted pattern: %w", err)
				}
				patternValue = unquoted
			}

			return PatternElement{Type: possibleType, Pattern: patternValue}, nil
		}
	}

	// Default to string type
	return PatternElement{Type: "string", Pattern: pattern}, nil
}

// unquoteString removes quotes and handles escape sequences
func unquoteString(s string) (string, error) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return s, nil
	}

	var result strings.Builder
	i := 1 // Skip opening quote

	for i < len(s)-1 { // Skip closing quote
		if s[i] == '\\' && i+1 < len(s)-1 {
			// Handle escape sequences
			switch s[i+1] {
			case '"':
				result.WriteByte('"')
			case '\\':
				result.WriteByte('\\')
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			default:
				result.WriteByte(s[i+1])
			}
			i += 2
		} else {
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String(), nil
}

// EvaluateCompoundPattern evaluates a compound pattern against content
func EvaluateCompoundPattern(compound *CompoundPattern, content string) (bool, []string, error) {
	if compound == nil {
		return false, nil, fmt.Errorf("nil compound pattern")
	}

	var allMatches []string
	results := make([]bool, len(compound.Patterns))

	// Evaluate each pattern element
	for i, element := range compound.Patterns {
		found, matches, err := evaluatePatternElement(element, content)
		if err != nil {
			return false, nil, fmt.Errorf("error evaluating pattern element %d: %w", i+1, err)
		}
		results[i] = found
		allMatches = append(allMatches, matches...)
	}

	// Apply the compound operator
	var finalResult bool
	switch strings.ToUpper(compound.Operator) {
	case "AND":
		finalResult = true
		for _, result := range results {
			if !result {
				finalResult = false
				break
			}
		}
	case "OR":
		finalResult = false
		for _, result := range results {
			if result {
				finalResult = true
				break
			}
		}
	default:
		return false, nil, fmt.Errorf("unsupported compound operator: %s", compound.Operator)
	}

	return finalResult, allMatches, nil
}

// evaluatePatternElement evaluates a single pattern element
func evaluatePatternElement(element PatternElement, content string) (bool, []string, error) {
	switch element.Type {
	case "string":
		found := strings.Contains(content, element.Pattern)
		var matches []string
		if found {
			matches = []string{element.Pattern}
		}
		return found, matches, nil

	case "regex":
		re, err := regexp.Compile(element.Pattern)
		if err != nil {
			return false, nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		matches := re.FindAllString(content, -1)
		return len(matches) > 0, matches, nil

	case "compound":
		if element.Compound == nil {
			return false, nil, fmt.Errorf("nil nested compound pattern")
		}
		return EvaluateCompoundPattern(element.Compound, content)

	default:
		return false, nil, fmt.Errorf("unsupported pattern element type: %s", element.Type)
	}
}
