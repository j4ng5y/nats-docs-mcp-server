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
	syncpKeywords := []string{"control-plane", "syncp"}
	natsKeywords := []string{"jetstream", "stream"}

	kc := NewKeywordClassifier(syncpKeywords, natsKeywords)

	if kc == nil {
		t.Fatal("NewKeywordClassifier returned nil")
	}

	if len(kc.syncpKeywords) != len(syncpKeywords) {
		t.Errorf("expected %d syncp keywords, got %d", len(syncpKeywords), len(kc.syncpKeywords))
	}

	if len(kc.natsKeywords) != len(natsKeywords) {
		t.Errorf("expected %d nats keywords, got %d", len(natsKeywords), len(kc.natsKeywords))
	}
}

func TestClassify_SyncpOnly(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected DocumentationSource
	}{
		{
			name:     "query with single syncp keyword",
			query:    "control-plane",
			expected: SourceSyncp,
		},
		{
			name:     "query with syncp keyword in sentence",
			query:    "how do I configure the control-plane",
			expected: SourceSyncp,
		},
		{
			name:     "query with multiple syncp keywords",
			query:    "control-plane syncp multi-tenant",
			expected: SourceSyncp,
		},
		{
			name:     "query with syncp keyword case insensitive",
			query:    "CONTROL-PLANE setup",
			expected: SourceSyncp,
		},
	}

	kc := NewKeywordClassifier(DefaultSyncpKeywords(), DefaultNATSKeywords())

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

	kc := NewKeywordClassifier(DefaultSyncpKeywords(), DefaultNATSKeywords())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kc.Classify(tt.query)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestClassify_Both(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected DocumentationSource
	}{
		{
			name:     "query with both syncp and nats keywords",
			query:    "jetstream and control-plane",
			expected: SourceBoth,
		},
		{
			name:     "query with no keywords",
			query:    "how do I do things",
			expected: SourceBoth,
		},
		{
			name:     "empty query",
			query:    "",
			expected: SourceBoth,
		},
		{
			name:     "query with only special characters",
			query:    "!@#$%^&*()",
			expected: SourceBoth,
		},
		{
			name:     "query with spaces only",
			query:    "   ",
			expected: SourceBoth,
		},
	}

	kc := NewKeywordClassifier(DefaultSyncpKeywords(), DefaultNATSKeywords())

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
			expected: SourceBoth,
		},
		{
			name:     "query with exact keyword",
			query:    "jetstream",
			expected: SourceNATS,
		},
		{
			name:     "query with hyphenated keyword partial match",
			query:    "control",
			expected: SourceBoth,
		},
		{
			name:     "query with hyphenated keyword full match",
			query:    "control-plane",
			expected: SourceSyncp,
		},
		{
			name:     "very long query with single keyword",
			query:    "this is a very long query with many words and at some point we mention jetstream somewhere",
			expected: SourceNATS,
		},
	}

	kc := NewKeywordClassifier(DefaultSyncpKeywords(), DefaultNATSKeywords())

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

// Feature: syncp-documentation-support, Property 3: Classification Determinism
// VALIDATES: Requirements 3.1
func TestProperty_ClassificationDeterminism(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property(
		"classification is deterministic - same query always produces same result",
		prop.ForAll(
			func(query string) bool {
				kc := NewKeywordClassifier(DefaultSyncpKeywords(), DefaultNATSKeywords())

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

// Feature: syncp-documentation-support, Property 4: Keyword-Based Classification Correctness
// VALIDATES: Requirements 3.2, 3.3, 3.4
func TestProperty_ClassificationCorrectness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Test 1: Query with only syncp keywords → SourceSyncp
	syncpKeywords := DefaultSyncpKeywords()
	natsKeywords := DefaultNATSKeywords()

	properties.Property(
		"query with only syncp keywords classifies as SourceSyncp",
		prop.ForAll(
			func() bool {
				kc := NewKeywordClassifier(syncpKeywords, natsKeywords)

				// Test each syncp keyword individually
				for _, syncpKw := range syncpKeywords {
					result := kc.Classify(syncpKw)
					if result != SourceSyncp {
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
				kc := NewKeywordClassifier(syncpKeywords, natsKeywords)

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

	// Test 3: Query with both keywords → SourceBoth
	properties.Property(
		"query with both syncp and nats keywords classifies as SourceBoth",
		prop.ForAll(
			func(syncpIdx int, natsIdx int) bool {
				kc := NewKeywordClassifier(syncpKeywords, natsKeywords)

				// Safely access array elements with modulo
				idx1 := ((syncpIdx % len(syncpKeywords)) + len(syncpKeywords)) % len(syncpKeywords)
				idx2 := ((natsIdx % len(natsKeywords)) + len(natsKeywords)) % len(natsKeywords)

				syncpKw := syncpKeywords[idx1]
				natsKw := natsKeywords[idx2]

				// Only test with non-overlapping keywords to ensure Both classification
				// Some keywords might overlap if someone configured them poorly
				query := syncpKw + " and " + natsKw

				result := kc.Classify(query)
				return result == SourceBoth
			},
			gen.Int(),
			gen.Int(),
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: syncp-documentation-support, Property 5: Classification Configuration Sensitivity
// VALIDATES: Requirements 3.5
func TestProperty_ClassificationConfigurationSensitivity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property(
		"different keyword configurations produce different classifications",
		prop.ForAll(
			func() bool {
				query := "my special keyword"

				// Config 1: keyword is syncp-specific
				kc1 := NewKeywordClassifier(
					[]string{"special"},
					[]string{},
				)
				result1 := kc1.Classify(query)

				// Config 2: keyword is nats-specific
				kc2 := NewKeywordClassifier(
					[]string{},
					[]string{"special"},
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
		{SourceSyncp, "Syncp"},
		{SourceBoth, "Both"},
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
	syncpKeywords := DefaultSyncpKeywords()
	natsKeywords := DefaultNATSKeywords()

	if len(syncpKeywords) == 0 {
		t.Error("DefaultSyncpKeywords returned empty list")
	}

	if len(natsKeywords) == 0 {
		t.Error("DefaultNATSKeywords returned empty list")
	}

	// Check for no duplicates in syncp keywords
	syncpSet := make(map[string]bool)
	for _, kw := range syncpKeywords {
		if syncpSet[kw] {
			t.Errorf("duplicate syncp keyword: %q", kw)
		}
		syncpSet[kw] = true
	}

	// Check for no duplicates in nats keywords
	natsSet := make(map[string]bool)
	for _, kw := range natsKeywords {
		if natsSet[kw] {
			t.Errorf("duplicate nats keyword: %q", kw)
		}
		natsSet[kw] = true
	}
}
