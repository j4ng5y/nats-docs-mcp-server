package classifier

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// ============================================================================
// Unit Tests
// ============================================================================

func TestNewKeywordClassifier(t *testing.T) {
	SynadiaKeywords := []string{"control-plane", "Synadia"}
	natsKeywords := []string{"jetstream", "stream"}
	githubKeywords := []string{"server", "code"}

	kc := NewKeywordClassifier(SynadiaKeywords, natsKeywords, githubKeywords)

	if kc == nil {
		t.Fatal("NewKeywordClassifier returned nil")
	}

	if len(kc.syadiaKeywords) != len(SynadiaKeywords) {
		t.Errorf("expected %d Synadia keywords, got %d", len(SynadiaKeywords), len(kc.syadiaKeywords))
	}

	if len(kc.natsKeywords) != len(natsKeywords) {
		t.Errorf("expected %d nats keywords, got %d", len(natsKeywords), len(kc.natsKeywords))
	}

	if len(kc.githubKeywords) != len(githubKeywords) {
		t.Errorf("expected %d github keywords, got %d", len(githubKeywords), len(kc.githubKeywords))
	}
}

func TestClassify_SynadiaOnly(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected DocumentationSource
	}{
		{
			name:     "query with single Synadia keyword",
			query:    "control-plane",
			expected: SourceSynadia,
		},
		{
			name:     "query with Synadia keyword in sentence",
			query:    "how do I configure the control-plane",
			expected: SourceSynadia,
		},
		{
			name:     "query with multiple Synadia keywords",
			query:    "control-plane Synadia multi-tenant",
			expected: SourceSynadia,
		},
		{
			name:     "query with Synadia keyword case insensitive",
			query:    "CONTROL-PLANE setup",
			expected: SourceSynadia,
		},
	}

	kc := NewKeywordClassifier(DefaultSyadiaKeywords(), DefaultNATSKeywords(), DefaultGitHubKeywords())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kc.Classify(tt.query)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestClassify_NATSOnly(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected DocumentationSource
	}{
		{
			name:     "query with single nats keyword",
			query:    "jetstream",
			expected: SourceNATS,
		},
		{
			name:     "query with nats keyword in sentence",
			query:    "how to use jetstream consumers",
			expected: SourceNATS,
		},
		{
			name:     "query with multiple nats keywords",
			query:    "jetstream stream consumer",
			expected: SourceNATS,
		},
		{
			name:     "query with nats keyword case insensitive",
			query:    "JETSTREAM setup",
			expected: SourceNATS,
		},
	}

	kc := NewKeywordClassifier(DefaultSyadiaKeywords(), DefaultNATSKeywords(), DefaultGitHubKeywords())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kc.Classify(tt.query)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestClassify_All(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected DocumentationSource
	}{
		{
			name:     "query with both Synadia and nats keywords",
			query:    "jetstream and control-plane",
			expected: SourceAll,
		},
		{
			name:     "query with no keywords",
			query:    "how do I do things",
			expected: SourceAll,
		},
		{
			name:     "empty query",
			query:    "",
			expected: SourceAll,
		},
		{
			name:     "query with only special characters",
			query:    "!@#$%^&*()",
			expected: SourceAll,
		},
		{
			name:     "query with spaces only",
			query:    "   ",
			expected: SourceAll,
		},
	}

	kc := NewKeywordClassifier(DefaultSyadiaKeywords(), DefaultNATSKeywords(), DefaultGitHubKeywords())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kc.Classify(tt.query)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestClassify_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected DocumentationSource
	}{
		{
			name:     "query with keyword as substring",
			query:    "jetstreaming",
			expected: SourceAll,
		},
		{
			name:     "query with exact keyword",
			query:    "jetstream",
			expected: SourceNATS,
		},
		{
			name:     "query with hyphenated keyword partial match",
			query:    "control",
			expected: SourceAll,
		},
		{
			name:     "query with hyphenated keyword full match",
			query:    "control-plane",
			expected: SourceSynadia,
		},
		{
			name:     "very long query with single keyword",
			query:    "this is a very long query with many words and at some point we mention jetstream somewhere",
			expected: SourceNATS,
		},
	}

	kc := NewKeywordClassifier(DefaultSyadiaKeywords(), DefaultNATSKeywords(), DefaultGitHubKeywords())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kc.Classify(tt.query)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTokenizeQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "simple words",
			query:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "hyphenated words kept as single token",
			query:    "control-plane",
			expected: []string{"control-plane"},
		},
		{
			name:     "with punctuation",
			query:    "hello, world!",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with numbers and dots",
			query:    "version 1.2.3",
			expected: []string{"version", "1", "2", "3"},
		},
		{
			name:     "empty string",
			query:    "",
			expected: []string{},
		},
		{
			name:     "only spaces",
			query:    "   ",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenizeQuery(tt.query)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tokens, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, v := range result {
				if i >= len(tt.expected) || v != tt.expected[i] {
					t.Errorf("token mismatch at index %d: expected %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}

// ============================================================================
// Property-Based Tests (using gopter)
// ============================================================================

// Feature: Synadia-documentation-support, Property 3: Classification Determinism
// VALIDATES: Requirements 3.1
func TestProperty_ClassificationDeterminism(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property(
		"classification is deterministic - same query always produces same result",
		prop.ForAll(
			func(query string) bool {
				kc := NewKeywordClassifier(DefaultSyadiaKeywords(), DefaultNATSKeywords(), DefaultGitHubKeywords())

				// Classify the same query multiple times
				result1 := kc.Classify(query)
				result2 := kc.Classify(query)
				result3 := kc.Classify(query)

				// All results should be identical
				return result1 == result2 && result2 == result3
			},
			gen.AnyString(),
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: Synadia-documentation-support, Property 4: Keyword-Based Classification Correctness
// VALIDATES: Requirements 3.2, 3.3, 3.4
func TestProperty_ClassificationCorrectness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Use minimal non-overlapping keyword sets to test the core property
	// without hitting edge cases from default keyword overlaps
	SynadiaKeywords := []string{"Synadia", "control-plane", "synadia"}
	natsKeywords := []string{"jetstream", "stream", "consumer"}
	githubKeywords := []string{"code", "implementation", "repository"}

	properties.Property(
		"query with only Synadia keywords classifies as SourceSynadia",
		prop.ForAll(
			func() bool {
				kc := NewKeywordClassifier(SynadiaKeywords, natsKeywords, githubKeywords)

				// Test each Synadia keyword individually
				for _, SynadiaKw := range SynadiaKeywords {
					result := kc.Classify(SynadiaKw)
					if result != SourceSynadia {
						return false
					}
				}
				return true
			},
		),
	)

	// Test 2: Query with only nats keywords → SourceNATS
	properties.Property(
		"query with only nats keywords classifies as SourceNATS",
		prop.ForAll(
			func() bool {
				kc := NewKeywordClassifier(SynadiaKeywords, natsKeywords, githubKeywords)

				// Test each nats keyword individually
				for _, natsKw := range natsKeywords {
					result := kc.Classify(natsKw)
					if result != SourceNATS {
						return false
					}
				}
				return true
			},
		),
	)

	// Test 3: Query with both keywords → SourceAll
	properties.Property(
		"query with both Synadia and nats keywords classifies as SourceAll",
		prop.ForAll(
			func(SynadiaIdx int, natsIdx int) bool {
				kc := NewKeywordClassifier(SynadiaKeywords, natsKeywords, githubKeywords)

				// Safely access array elements with modulo
				idx1 := ((SynadiaIdx % len(SynadiaKeywords)) + len(SynadiaKeywords)) % len(SynadiaKeywords)
				idx2 := ((natsIdx % len(natsKeywords)) + len(natsKeywords)) % len(natsKeywords)

				SynadiaKw := SynadiaKeywords[idx1]
				natsKw := natsKeywords[idx2]

				// Only test with non-overlapping keywords to ensure All classification
				// Some keywords might overlap if someone configured them poorly
				query := SynadiaKw + " and " + natsKw

				result := kc.Classify(query)
				return result == SourceAll
			},
			gen.Int(),
			gen.Int(),
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: Synadia-documentation-support, Property 5: Classification Configuration Sensitivity
// VALIDATES: Requirements 3.5
func TestProperty_ClassificationConfigurationSensitivity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property(
		"different keyword configurations produce different classifications",
		prop.ForAll(
			func() bool {
				query := "my special keyword"

				// Config 1: keyword is Synadia-specific
				kc1 := NewKeywordClassifier(
					[]string{"special"},
					[]string{},
					[]string{},
				)
				result1 := kc1.Classify(query)

				// Config 2: keyword is nats-specific
				kc2 := NewKeywordClassifier(
					[]string{},
					[]string{"special"},
					[]string{},
				)
				result2 := kc2.Classify(query)

				// Results should be different
				return result1 != result2
			},
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ============================================================================
// Helper Tests
// ============================================================================

func TestDocumentationSourceString(t *testing.T) {
	tests := []struct {
		source   DocumentationSource
		expected string
	}{
		{SourceNATS, "NATS"},
		{SourceSynadia, "Synadia"},
		{SourceGitHub, "GitHub"},
		{SourceAll, "All"},
		{DocumentationSource(999), "Unknown"},
	}

	for _, tt := range tests {
		result := tt.source.String()
		if result != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, result)
		}
	}
}

func TestDefaultKeywordLists(t *testing.T) {
	SynadiaKeywords := DefaultSyadiaKeywords()
	natsKeywords := DefaultNATSKeywords()
	githubKeywords := DefaultGitHubKeywords()

	if len(SynadiaKeywords) == 0 {
		t.Error("DefaultSyadiaKeywords returned empty list")
	}

	if len(natsKeywords) == 0 {
		t.Error("DefaultNATSKeywords returned empty list")
	}

	if len(githubKeywords) == 0 {
		t.Error("DefaultGitHubKeywords returned empty list")
	}

	// Check for no duplicates in Synadia keywords
	SynadiaSet := make(map[string]bool)
	for _, kw := range SynadiaKeywords {
		if SynadiaSet[kw] {
			t.Errorf("duplicate Synadia keyword: %q", kw)
		}
		SynadiaSet[kw] = true
	}

	// Check for no duplicates in nats keywords
	natsSet := make(map[string]bool)
	for _, kw := range natsKeywords {
		if natsSet[kw] {
			t.Errorf("duplicate nats keyword: %q", kw)
		}
		natsSet[kw] = true
	}

	// Check for no duplicates in github keywords
	githubSet := make(map[string]bool)
	for _, kw := range githubKeywords {
		if githubSet[kw] {
			t.Errorf("duplicate github keyword: %q", kw)
		}
		githubSet[kw] = true
	}
}
