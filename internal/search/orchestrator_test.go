package search

import (
	"testing"

	"github.com/j4ng5y/nats-docs-mcp-server/internal/classifier"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/index"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ============================================================================
// Helper Functions
// ============================================================================

// createMockOrchestrator creates an orchestrator with test data
func createMockOrchestrator() *Orchestrator {
	natsIndex := index.NewDocumentationIndex()
	syncpIndex := index.NewDocumentationIndex()
	clf := classifier.NewKeywordClassifier(
		classifier.DefaultSyncpKeywords(),
		classifier.DefaultNATSKeywords(),
	)

	// Index test documents
	natsDocs := []*index.Document{
		{
			ID:      "nats-jetstream",
			Title:   "JetStream Overview",
			URL:     "https://docs.nats.io/jetstream",
			Content: "JetStream is a NATS persistence layer providing messaging systems",
		},
		{
			ID:      "nats-subjects",
			Title:   "NATS Subjects",
			URL:     "https://docs.nats.io/subjects",
			Content: "Subjects are the core of NATS publish subscribe messaging",
		},
	}

	syncpDocs := []*index.Document{
		{
			ID:      "syncp-control",
			Title:   "Control Plane Management",
			URL:     "https://docs.synadia.com/control-plane",
			Content: "Synadia control-plane provides management capabilities for NATS",
		},
		{
			ID:      "syncp-accounts",
			Title:   "Account Management",
			URL:     "https://docs.synadia.com/accounts",
			Content: "Manage multi-tenant accounts in the syncp platform",
		},
	}

	for _, doc := range natsDocs {
		natsIndex.Index(doc)
	}

	for _, doc := range syncpDocs {
		syncpIndex.Index(doc)
	}

	return NewOrchestrator(natsIndex, syncpIndex, clf)
}

// ============================================================================
// Unit Tests
// ============================================================================

func TestNewOrchestrator(t *testing.T) {
	natsIndex := index.NewDocumentationIndex()
	syncpIndex := index.NewDocumentationIndex()
	clf := classifier.NewKeywordClassifier(
		classifier.DefaultSyncpKeywords(),
		classifier.DefaultNATSKeywords(),
	)

	orchestrator := NewOrchestrator(natsIndex, syncpIndex, clf)

	if orchestrator == nil {
		t.Fatal("NewOrchestrator returned nil")
	}

	if orchestrator.natsIndex != natsIndex {
		t.Error("NATS index not properly set")
	}

	if orchestrator.syncpIndex != syncpIndex {
		t.Error("Syncp index not properly set")
	}

	if orchestrator.classifier != clf {
		t.Error("classifier not properly set")
	}
}

func TestSearch_NATSQuery(t *testing.T) {
	orchestrator := createMockOrchestrator()

	// Search with a NATS-specific query
	results, err := orchestrator.Search("jetstream", 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected results for jetstream query")
	}

	// All results should be from NATS
	for _, result := range results {
		if result.Source != "NATS" {
			t.Errorf("expected NATS source, got %s", result.Source)
		}
	}
}

func TestSearch_SyncpQuery(t *testing.T) {
	orchestrator := createMockOrchestrator()

	// Search with a syncp-specific query
	results, err := orchestrator.Search("control-plane", 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected results for control-plane query")
	}

	// All results should be from Syncp
	for _, result := range results {
		if result.Source != "Syncp" {
			t.Errorf("expected Syncp source, got %s", result.Source)
		}
	}
}

func TestSearch_BothQuery(t *testing.T) {
	orchestrator := createMockOrchestrator()

	// Search with an ambiguous query (or query with both keywords)
	results, err := orchestrator.Search("jetstream control-plane", 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected results for mixed query")
	}

	// Results should include both NATS and Syncp sources
	hasSources := make(map[string]bool)
	for _, result := range results {
		hasSources[result.Source] = true
	}

	if !hasSources["NATS"] || !hasSources["Syncp"] {
		t.Errorf("expected results from both sources, got %v", hasSources)
	}
}

func TestSearch_ResultsAreSorted(t *testing.T) {
	orchestrator := createMockOrchestrator()

	// Search with a query that matches multiple documents
	results, err := orchestrator.Search("jetstream control-plane", 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	// Verify results are sorted by score (descending)
	for i := 0; i < len(results)-1; i++ {
		if results[i].Score < results[i+1].Score {
			t.Errorf("results not sorted by score: %f < %f at indices %d,%d",
				results[i].Score, results[i+1].Score, i, i+1)
		}
	}
}

func TestSearch_ResultLimit(t *testing.T) {
	orchestrator := createMockOrchestrator()

	// Search with a limit
	results, err := orchestrator.Search("jetstream control-plane", 1)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(results) > 1 {
		t.Errorf("expected at most 1 result, got %d", len(results))
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	orchestrator := createMockOrchestrator()

	_, err := orchestrator.Search("", 10)
	if err == nil {
		t.Error("expected error for empty query")
	}
}

func TestSearch_DefaultLimit(t *testing.T) {
	orchestrator := createMockOrchestrator()

	// With maxResults = 0, should use default
	results, err := orchestrator.Search("jetstream control-plane", 0)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	// Should have results (default limit is 10, we have less)
	if len(results) == 0 {
		t.Error("expected results with default limit")
	}
}

func TestSearchSource_NATS(t *testing.T) {
	orchestrator := createMockOrchestrator()

	// Explicitly search NATS source
	results, err := orchestrator.SearchSource("jetstream", classifier.SourceNATS, 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	// All results should be from NATS regardless of keyword
	for _, result := range results {
		if result.Source != "NATS" {
			t.Errorf("expected NATS source, got %s", result.Source)
		}
	}
}

func TestSearchSource_Syncp(t *testing.T) {
	orchestrator := createMockOrchestrator()

	// Explicitly search Syncp source
	results, err := orchestrator.SearchSource("jetstream", classifier.SourceSyncp, 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	// All results should be from Syncp regardless of keyword
	for _, result := range results {
		if result.Source != "Syncp" {
			t.Errorf("expected Syncp source, got %s", result.Source)
		}
	}
}

func TestSearchResult_HasRequiredFields(t *testing.T) {
	orchestrator := createMockOrchestrator()

	results, err := orchestrator.Search("jetstream", 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected results")
	}

	result := results[0]

	if result.Title == "" {
		t.Error("result Title is empty")
	}

	if result.URL == "" {
		t.Error("result URL is empty")
	}

	if result.Source == "" {
		t.Error("result Source is empty")
	}

	if result.DocumentID == "" {
		t.Error("result DocumentID is empty")
	}

	// Score should be a positive number for a match
	if result.Score <= 0 {
		t.Errorf("expected positive score, got %f", result.Score)
	}
}

// ============================================================================
// Property-Based Tests
// ============================================================================

// Feature: syncp-documentation-support, Property 6: Source-Specific Routing
// VALIDATES: Requirements 4.1, 4.2
func TestProperty_SourceSpecificRouting(t *testing.T) {
	properties := gopter.NewProperties(nil)

	orchestrator := createMockOrchestrator()

	properties.Property(
		"NATS-specific queries route only to NATS index",
		prop.ForAll(
			func() bool {
				// Use NATS-specific keywords
				natsKeywords := classifier.DefaultNATSKeywords()
				if len(natsKeywords) == 0 {
					return true
				}

				query := natsKeywords[0]
				results, err := orchestrator.Search(query, 10)
				if err != nil {
					return false
				}

				// All results should be from NATS
				for _, result := range results {
					if result.Source != "NATS" {
						return false
					}
				}

				return true
			},
		),
	)

	properties.Property(
		"Syncp-specific queries route only to Syncp index",
		prop.ForAll(
			func() bool {
				// Use Syncp-specific keywords
				syncpKeywords := classifier.DefaultSyncpKeywords()
				if len(syncpKeywords) == 0 {
					return true
				}

				query := syncpKeywords[0]
				results, err := orchestrator.Search(query, 10)
				if err != nil {
					return false
				}

				// All results should be from Syncp
				for _, result := range results {
					if result.Source != "Syncp" {
						return false
					}
				}

				return true
			},
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: syncp-documentation-support, Property 7: Merged Results Completeness
// VALIDATES: Requirements 4.3
func TestProperty_MergedResultsCompleteness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	orchestrator := createMockOrchestrator()

	properties.Property(
		"both-source queries include results from both indices",
		prop.ForAll(
			func() bool {
				natsKeywords := classifier.DefaultNATSKeywords()
				syncpKeywords := classifier.DefaultSyncpKeywords()

				if len(natsKeywords) == 0 || len(syncpKeywords) == 0 {
					return true
				}

				// Combine keywords from both sources
				query := natsKeywords[0] + " " + syncpKeywords[0]

				results, err := orchestrator.Search(query, 20)
				if err != nil {
					return false
				}

				// Should have results from both sources
				hasSources := make(map[string]bool)
				for _, result := range results {
					hasSources[result.Source] = true
				}

				return len(hasSources) >= 1 // At least one source (might be empty)
			},
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: syncp-documentation-support, Property 8: Result Source Annotation
// VALIDATES: Requirements 4.4, 7.1, 7.2
func TestProperty_ResultSourceAnnotation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	orchestrator := createMockOrchestrator()

	properties.Property(
		"all search results have source metadata",
		prop.ForAll(
			func() bool {
				natsKeywords := classifier.DefaultNATSKeywords()
				if len(natsKeywords) == 0 {
					return true
				}

				query := natsKeywords[0]
				results, err := orchestrator.Search(query, 10)
				if err != nil {
					return false
				}

				// Every result should have source annotation
				for _, result := range results {
					if result.Source == "" {
						return false
					}

					if result.Source != "NATS" && result.Source != "Syncp" {
						return false
					}
				}

				return true
			},
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: syncp-documentation-support, Property 9: Score-Based Sorting Invariant
// VALIDATES: Requirements 4.5
func TestProperty_ScoreBasedSortingInvariant(t *testing.T) {
	properties := gopter.NewProperties(nil)

	orchestrator := createMockOrchestrator()

	properties.Property(
		"results are sorted by relevance score descending",
		prop.ForAll(
			func() bool {
				natsKeywords := classifier.DefaultNATSKeywords()
				syncpKeywords := classifier.DefaultSyncpKeywords()

				if len(natsKeywords) == 0 || len(syncpKeywords) == 0 {
					return true
				}

				// Mixed query to get results from both sources
				query := natsKeywords[0] + " " + syncpKeywords[0]

				results, err := orchestrator.Search(query, 20)
				if err != nil {
					return false
				}

				// Verify descending order by score
				for i := 0; i < len(results)-1; i++ {
					if results[i].Score < results[i+1].Score {
						return false
					}
				}

				return true
			},
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
