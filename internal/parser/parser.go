package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Precompiled regex patterns for performance
var delistPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)Binance\s+Will\s+Delist\s+(.+?)\s+on\s+\d{4}-\d{2}-\d{2}`),
	regexp.MustCompile(`(?i)Binance\s+Announced\s+the\s+First\s+Batch\s+of\s+Vote\s+to\s+Delist\s+Results\s+and\s+Will\s+Delist\s+(.+?)\s+on\s+\d{4}-\d{2}-\d{2}`),
}

// ExtractPairs extracts trading pairs (e.g., "BTCUSDT", "ETHUSDT") from Binance delisting announcement titles.
//
// Examples:
//
//	"Binance Will Delist BTCUSDT, ETHUSDT on 2025-10-31"
//	→ ["BTCUSDT", "ETHUSDT"]
func ExtractPairs(title string) []string {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil
	}

	for _, pattern := range delistPatterns {
		if matches := pattern.FindStringSubmatch(title); len(matches) > 1 {
			return cleanPairs(matches[1])
		}
	}

	fmt.Printf("[%s] ⚠️  No known delisting pattern matched: %q\n", timestamp(), title)
	return nil
}

// cleanPairs splits a string of comma-separated pairs and trims spaces.
func cleanPairs(input string) []string {
	// Handles commas, semicolons, or multiple spaces as delimiters.
	fields := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == ';'
	})

	var pairs []string
	for _, f := range fields {
		if trimmed := strings.TrimSpace(f); trimmed != "" {
			pairs = append(pairs, trimmed)
		}
	}
	return pairs
}

// timestamp returns a formatted timestamp string.
func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
