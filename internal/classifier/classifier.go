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
	// SourceSynadia indicates the query is specific to Synadia documentation
	SourceSynadia
	// SourceGitHub indicates the query is specific to GitHub documentation
	SourceGitHub
	// SourceAll indicates the query should search all documentation sources
	SourceAll
)

// String returns the string representation of the DocumentationSource
func (ds DocumentationSource) String() string {
	switch ds {
	case SourceNATS:
		return "NATS"
	case SourceSynadia:
		return "Synadia"
	case SourceGitHub:
		return "GitHub"
	case SourceAll:
		return "All"
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
	syadiaKeywords map[string]bool
	natsKeywords   map[string]bool
	githubKeywords map[string]bool
}

// NewKeywordClassifier creates a classifier with the given keyword lists
func NewKeywordClassifier(syadiaKeywords, natsKeywords []string, githubKeywords []string) *KeywordClassifier {
	kc := &KeywordClassifier{
		syadiaKeywords: make(map[string]bool),
		natsKeywords:   make(map[string]bool),
		githubKeywords: make(map[string]bool),
	}

	// Normalize and store Synadia keywords
	for _, kw := range syadiaKeywords {
		kc.syadiaKeywords[strings.ToLower(kw)] = true
	}

	// Normalize and store NATS keywords
	for _, kw := range natsKeywords {
		kc.natsKeywords[strings.ToLower(kw)] = true
	}

	// Normalize and store GitHub keywords
	for _, kw := range githubKeywords {
		kc.githubKeywords[strings.ToLower(kw)] = true
	}

	return kc
}

// Classify implements the Classifier interface
// Classification algorithm:
// 1. Normalize query to lowercase
// 2. Check if keywords appear in the query (substring or word matching)
// 3. Count matches against Synadia, NATS, and GitHub keyword lists
// 4. Apply classification rules:
//    - If only one source has matches → Return that source
//    - If multiple sources have matches OR no matches → Return SourceAll
func (kc *KeywordClassifier) Classify(query string) DocumentationSource {
	if query == "" {
		return SourceAll
	}

	// Normalize to lowercase
	normalizedQuery := strings.ToLower(query)

	// Count matches by checking if keywords appear in the query
	syadiaMatches := 0
	natsMatches := 0
	githubMatches := 0

	// Check Synadia keywords
	for kw := range kc.syadiaKeywords {
		if matchesKeywordInQuery(normalizedQuery, kw) {
			syadiaMatches++
		}
	}

	// Check NATS keywords
	for kw := range kc.natsKeywords {
		if matchesKeywordInQuery(normalizedQuery, kw) {
			natsMatches++
		}
	}

	// Check GitHub keywords
	for kw := range kc.githubKeywords {
		if matchesKeywordInQuery(normalizedQuery, kw) {
			githubMatches++
		}
	}

	// Count how many sources have matches
	matchCount := 0
	if syadiaMatches > 0 {
		matchCount++
	}
	if natsMatches > 0 {
		matchCount++
	}
	if githubMatches > 0 {
		matchCount++
	}

	// If no keywords matched at all, default to SourceAll
	if matchCount == 0 {
		return SourceAll
	}

	// If exactly one source matched, return that source
	if matchCount == 1 {
		if syadiaMatches > 0 {
			return SourceSynadia
		}
		if natsMatches > 0 {
			return SourceNATS
		}
		if githubMatches > 0 {
			return SourceGitHub
		}
	}

	// Multiple sources matched or ambiguous - search all
	return SourceAll
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

// DefaultSyadiaKeywords returns the default list of Synadia-specific keywords
func DefaultSyadiaKeywords() []string {
	return []string{
		"control-plane",
		"synadia",
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

// DefaultGitHubKeywords returns the default list of GitHub-specific keywords
func DefaultGitHubKeywords() []string {
	return []string{
		"server",
		"implementation",
		"code",
		"go",
		"golang",
		"client",
		"example",
		"tutorial",
		"source",
		"repository",
		"readme",
		"github",
		"sdk",
		"library",
	}
}
