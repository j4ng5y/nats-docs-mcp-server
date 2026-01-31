package classifier

import (
	"strings"
	"unicode"
)

// DocumentationSource represents which documentation set(s) to search
type DocumentationSource int

const (
	// SourceNATS indicates the query is specific to NATS documentation
	SourceNATS DocumentationSource = iota
	// SourceSyncp indicates the query is specific to Syncp documentation
	SourceSyncp
	// SourceBoth indicates the query should search both NATS and Syncp documentation
	SourceBoth
)

// String returns the string representation of the DocumentationSource
func (ds DocumentationSource) String() string {
	switch ds {
	case SourceNATS:
		return "NATS"
	case SourceSyncp:
		return "Syncp"
	case SourceBoth:
		return "Both"
	default:
		return "Unknown"
	}
}

// Classifier determines which documentation source(s) are relevant for a query
type Classifier interface {
	// Classify analyzes a query and returns the appropriate documentation source(s)
	Classify(query string) DocumentationSource
}

// KeywordClassifier implements classification using keyword matching
type KeywordClassifier struct {
	syncpKeywords map[string]bool
	natsKeywords  map[string]bool
}

// NewKeywordClassifier creates a classifier with the given keyword lists
func NewKeywordClassifier(syncpKeywords, natsKeywords []string) *KeywordClassifier {
	kc := &KeywordClassifier{
		syncpKeywords: make(map[string]bool),
		natsKeywords:  make(map[string]bool),
	}

	// Normalize and store syncp keywords
	for _, kw := range syncpKeywords {
		kc.syncpKeywords[strings.ToLower(kw)] = true
	}

	// Normalize and store NATS keywords
	for _, kw := range natsKeywords {
		kc.natsKeywords[strings.ToLower(kw)] = true
	}

	return kc
}

// Classify implements the Classifier interface
// Classification algorithm:
// 1. Normalize query to lowercase
// 2. Check if keywords appear in the query (substring or word matching)
// 3. Count matches against syncp keyword list
// 4. Count matches against NATS keyword list
// 5. Apply classification rules:
//    - If syncp matches > 0 AND nats matches == 0 → SourceSyncp
//    - If nats matches > 0 AND syncp matches == 0 → SourceNATS
//    - If both have matches OR both have zero matches → SourceBoth
func (kc *KeywordClassifier) Classify(query string) DocumentationSource {
	if query == "" {
		return SourceBoth
	}

	// Normalize to lowercase
	normalizedQuery := strings.ToLower(query)

	// Count matches by checking if keywords appear in the query
	syncpMatches := 0
	natsMatches := 0

	// Check syncp keywords
	for kw := range kc.syncpKeywords {
		if matchesKeywordInQuery(normalizedQuery, kw) {
			syncpMatches++
		}
	}

	// Check NATS keywords
	for kw := range kc.natsKeywords {
		if matchesKeywordInQuery(normalizedQuery, kw) {
			natsMatches++
		}
	}

	// If no keywords matched at all, default to SourceBoth
	if syncpMatches == 0 && natsMatches == 0 {
		return SourceBoth
	}

	// Apply classification rules
	if syncpMatches > 0 && natsMatches == 0 {
		return SourceSyncp
	}
	if natsMatches > 0 && syncpMatches == 0 {
		return SourceNATS
	}
	// Both have matches OR both have zero matches
	return SourceBoth
}

// matchesKeywordInQuery checks if a keyword appears in the query with word boundaries.
// This handles:
// - Single words: "jetstream", "syncp" (must be surrounded by non-alphanumeric)
// - Hyphenated words: "control-plane", "nats-server" (hyphens are word separators)
// - Multi-word phrases: "object store", "core nats" (spaces separate words)
// - Words with dots: "nats.go", "nats.js" (dots are word separators)
func matchesKeywordInQuery(query, keyword string) bool {
	// Find all occurrences of the keyword in the query
	start := 0
	for {
		idx := strings.Index(query[start:], keyword)
		if idx == -1 {
			return false
		}

		// Adjust index to absolute position
		idx = start + idx

		// Check word boundary before keyword
		if idx > 0 {
			prevChar := query[idx-1]
			if !isBoundaryChar(prevChar) {
				start = idx + 1
				continue
			}
		}

		// Check word boundary after keyword
		endIdx := idx + len(keyword)
		if endIdx < len(query) {
			nextChar := query[endIdx]
			if !isBoundaryChar(nextChar) {
				start = idx + 1
				continue
			}
		}

		// Keyword found with proper word boundaries
		return true
	}
}

// isBoundaryChar checks if a character is a word boundary (space, punctuation, etc)
func isBoundaryChar(r byte) bool {
	return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
}

// tokenizeQuery splits a query string into words, removing punctuation and special characters
func tokenizeQuery(query string) []string {
	var words []string
	var currentWord strings.Builder

	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			currentWord.WriteRune(r)
		} else {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
	}

	// Don't forget the last word
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// DefaultSyncpKeywords returns the default list of Syncp-specific keywords
func DefaultSyncpKeywords() []string {
	return []string{
		"control-plane",
		"syncp",
		"syn-cp",
		"synadia platform",
		"control plane",
		"platform",
		"multi-tenant",
		"personal access token",
		"system owner",
		"account owner",
		"application owner",
	}
}

// DefaultNATSKeywords returns the default list of NATS-specific keywords
func DefaultNATSKeywords() []string {
	return []string{
		"jetstream",
		"stream",
		"consumer",
		"subject",
		"publish",
		"subscribe",
		"request",
		"reply",
		"core nats",
		"kv",
		"object store",
		"nats-server",
		"nats.go",
		"nats.js",
	}
}
