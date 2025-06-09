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
}

// SearchConfig defines what to search for and how
type SearchConfig struct {
	Type     string `json:"type"` // "string", "regex", "compound"
	Pattern  string `json:"pattern"`
	XPath    string `json:"xpath"`
	NotifyOn string `json:"notify_on"` // "found" or "not_found"
}

// CompoundPattern represents parsed compound search pattern with AND/OR operations
type CompoundPattern struct {
	Operator string           // "AND" or "OR"
	Patterns []PatternElement // Individual patterns or nested compounds
}

// PatternElement represents either a single pattern or nested compound pattern
type PatternElement struct {
	Type     string // "string", "regex", "compound"
	Pattern  string
	Compound *CompoundPattern // For nested compound patterns
}

// Notifications holds configuration for notification channels
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
	// Read file contents and unmarshal JSON into config struct
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
// Uses tokenization followed by recursive parsing to handle nested expressions
func ParseCompoundPattern(pattern string) (*CompoundPattern, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil, fmt.Errorf("empty pattern")
	}

	// First tokenize the pattern string, then parse tokens into expression tree
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

// tokenizePattern converts pattern string into tokens with quote handling
// Processes character-by-character to identify operators, parentheses and quoted strings
func tokenizePattern(pattern string) ([]Token, error) {
	var tokens []Token
	i := 0

	// Iterate through each character to build token list
	for i < len(pattern) {
		if pattern[i] == ' ' || pattern[i] == '\t' {
			i++
			continue
		}

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

		// Handle patterns that may contain quoted strings with spaces
		start := i
		inQuotes := false
		quoteChar := byte(0) // Track which quote character we're using

		for i < len(pattern) {
			char := pattern[i]

			// Track quote state to properly handle quoted content
			if (char == '"' || char == '\'') && quoteChar == 0 {
				inQuotes = true
				quoteChar = char
				i++
				continue
			} else if char == quoteChar && inQuotes {
				inQuotes = false
				quoteChar = 0
				i++
				continue
			}

			// Outside quotes, stop parsing when encountering operators or parentheses
			if !inQuotes {
				if char == '(' || char == ')' {
					break
				}

				// Check upcoming characters for AND/OR operators
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

// parseOrExpression handles OR operations with lowest precedence
// OR operators bind less tightly than AND operators
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

// parseAndExpression handles AND operations with higher precedence
// AND operators are evaluated before OR operators
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
// Processes leaf nodes and parenthesized sub-expressions
func parsePrimary(tokens []Token, pos int) (*CompoundPattern, int, error) {
	if pos >= len(tokens) {
		return nil, pos, fmt.Errorf("unexpected end of input")
	}

	token := tokens[pos]

	switch token.Type {
	case "LPAREN":
		// Parse expression inside parentheses recursively
		compound, newPos, err := parseOrExpression(tokens, pos+1)
		if err != nil {
			return nil, pos, err
		}

		// Verify closing parenthesis exists
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

// parsePatternElement parses pattern element with optional type prefix
// Recognizes type:pattern syntax and handles quoted patterns
func parsePatternElement(pattern string) (PatternElement, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return PatternElement{}, fmt.Errorf("empty pattern element")
	}

	// Search for colon to identify type prefix while avoiding pattern content
	colonIndex := strings.Index(pattern, ":")
	if colonIndex > 0 && colonIndex < len(pattern)-1 {
		possibleType := strings.ToLower(strings.TrimSpace(pattern[:colonIndex]))
		if possibleType == "string" || possibleType == "regex" {
			patternValue := strings.TrimSpace(pattern[colonIndex+1:])
			if patternValue == "" {
				return PatternElement{}, fmt.Errorf("empty pattern value")
			}

			// Process quoted patterns by removing surrounding quotes
			if len(patternValue) >= 2 {
				firstChar := patternValue[0]
				lastChar := patternValue[len(patternValue)-1]
				if (firstChar == '"' && lastChar == '"') || (firstChar == '\'' && lastChar == '\'') {
					// Remove quote characters from pattern value
					patternValue = patternValue[1 : len(patternValue)-1]
				}
			}

			return PatternElement{Type: possibleType, Pattern: patternValue}, nil
		}
	}

	return PatternElement{Type: "string", Pattern: pattern}, nil
}

// EvaluateCompoundPattern evaluates a compound pattern against content
// Applies the parsed boolean expression to web page content
func EvaluateCompoundPattern(compound *CompoundPattern, content string) (bool, []string, error) {
	if compound == nil {
		return false, nil, fmt.Errorf("nil compound pattern")
	}

	// Store matches from all sub-patterns for result reporting
	var allMatches []string
	results := make([]bool, len(compound.Patterns))

	// Test each pattern element against content and gather results
	for i, element := range compound.Patterns {
		found, matches, err := evaluatePatternElement(element, content)
		if err != nil {
			return false, nil, fmt.Errorf("error evaluating pattern element %d: %w", i+1, err)
		}
		results[i] = found
		allMatches = append(allMatches, matches...)
	}

	// Use AND/OR operator to combine individual pattern results
	var finalResult bool
	switch strings.ToUpper(compound.Operator) {
	case "AND":
		// Return true only if every pattern matches
		finalResult = true
		for _, result := range results {
			if !result {
				finalResult = false
				break
			}
		}
	case "OR":
		// Return true if any pattern matches
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

// evaluatePatternElement evaluates a single pattern element against content
func evaluatePatternElement(element PatternElement, content string) (bool, []string, error) {
	switch element.Type {
	case "string":
		found := strings.Contains(content, element.Pattern)
		matches := []string{}
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
