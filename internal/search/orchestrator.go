package search

import (
	"fmt"
	"sort"

	"github.com/j4ng5y/nats-docs-mcp-server/internal/classifier"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/index"
)

// SearchResult represents a single search result with source metadata
type SearchResult struct {
	Title       string  // Document title
	URL         string  // Document URL
	Snippet     string  // Relevant excerpt from the document
	Score       float64 // TF-IDF relevance score
	Source      string  // "NATS" or "Syncp" indicating documentation source
	DocumentID  string  // Internal document identifier
}

// Orchestrator coordinates searches across multiple documentation sources
type Orchestrator struct {
	natsIndex  *index.DocumentationIndex
	syncpIndex *index.DocumentationIndex
	classifier classifier.Classifier
}

// NewOrchestrator creates a search orchestrator with dual indices and classifier
func NewOrchestrator(
	natsIdx *index.DocumentationIndex,
	syncpIdx *index.DocumentationIndex,
	clf classifier.Classifier,
) *Orchestrator {
	return &Orchestrator{
		natsIndex:  natsIdx,
		syncpIndex: syncpIdx,
		classifier: clf,
	}
}

// Search performs a classified search across appropriate documentation sources
// The query is classified to determine which index/indices to search, then
// results are merged and sorted by relevance score.
func (o *Orchestrator) Search(query string, maxResults int) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if maxResults <= 0 {
		maxResults = 10 // Default limit
	}

	// Classify the query to determine which sources to search
	source := o.classifier.Classify(query)

	// Route to appropriate index based on classification
	switch source {
	case classifier.SourceNATS:
		return o.searchNATSIndex(query, maxResults)
	case classifier.SourceSyncp:
		return o.searchSyncpIndex(query, maxResults)
	case classifier.SourceBoth:
		return o.searchBothIndices(query, maxResults)
	default:
		return nil, fmt.Errorf("unknown documentation source: %v", source)
	}
}

// SearchSource performs a search against a specific documentation source
func (o *Orchestrator) SearchSource(
	query string,
	source classifier.DocumentationSource,
	maxResults int,
) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if maxResults <= 0 {
		maxResults = 10 // Default limit
	}

	switch source {
	case classifier.SourceNATS:
		return o.searchNATSIndex(query, maxResults)
	case classifier.SourceSyncp:
		return o.searchSyncpIndex(query, maxResults)
	case classifier.SourceBoth:
		return o.searchBothIndices(query, maxResults)
	default:
		return nil, fmt.Errorf("unknown documentation source: %v", source)
	}
}

// searchNATSIndex performs a search on the NATS index only
func (o *Orchestrator) searchNATSIndex(query string, maxResults int) ([]SearchResult, error) {
	natsResults, err := o.natsIndex.Search(query, maxResults)
	if err != nil {
		return nil, fmt.Errorf("failed to search NATS index: %w", err)
	}

	// Convert index results to orchestrator results with source metadata
	results := make([]SearchResult, len(natsResults))
	for i, indexResult := range natsResults {
		results[i] = SearchResult{
			Title:       indexResult.Title,
			URL:         indexResult.URL,
			Snippet:     indexResult.Summary,
			Score:       indexResult.Relevance,
			Source:      "NATS",
			DocumentID:  indexResult.DocumentID,
		}
	}

	return results, nil
}

// searchSyncpIndex performs a search on the syncp index only
func (o *Orchestrator) searchSyncpIndex(query string, maxResults int) ([]SearchResult, error) {
	syncpResults, err := o.syncpIndex.Search(query, maxResults)
	if err != nil {
		return nil, fmt.Errorf("failed to search syncp index: %w", err)
	}

	// Convert index results to orchestrator results with source metadata
	results := make([]SearchResult, len(syncpResults))
	for i, indexResult := range syncpResults {
		results[i] = SearchResult{
			Title:       indexResult.Title,
			URL:         indexResult.URL,
			Snippet:     indexResult.Summary,
			Score:       indexResult.Relevance,
			Source:      "Syncp",
			DocumentID:  indexResult.DocumentID,
		}
	}

	return results, nil
}

// searchBothIndices performs searches on both indices and merges results
func (o *Orchestrator) searchBothIndices(query string, maxResults int) ([]SearchResult, error) {
	// Search both indices (without applying limit to individual searches)
	natsResults, natsErr := o.natsIndex.Search(query, maxResults*2)
	syncpResults, syncpErr := o.syncpIndex.Search(query, maxResults*2)

	// Log errors but continue with available results
	if natsErr != nil && syncpErr != nil {
		return nil, fmt.Errorf("failed to search both indices: NATS: %v, Syncp: %v", natsErr, syncpErr)
	}

	// Merge results
	allResults := make([]SearchResult, 0, len(natsResults)+len(syncpResults))

	// Add NATS results with source metadata
	for _, indexResult := range natsResults {
		allResults = append(allResults, SearchResult{
			Title:       indexResult.Title,
			URL:         indexResult.URL,
			Snippet:     indexResult.Summary,
			Score:       indexResult.Relevance,
			Source:      "NATS",
			DocumentID:  indexResult.DocumentID,
		})
	}

	// Add Syncp results with source metadata
	for _, indexResult := range syncpResults {
		allResults = append(allResults, SearchResult{
			Title:       indexResult.Title,
			URL:         indexResult.URL,
			Snippet:     indexResult.Summary,
			Score:       indexResult.Relevance,
			Source:      "Syncp",
			DocumentID:  indexResult.DocumentID,
		})
	}

	// Sort all results by relevance score (descending)
	sort.Slice(allResults, func(i, j int) bool {
		// Higher score comes first
		if allResults[i].Score != allResults[j].Score {
			return allResults[i].Score > allResults[j].Score
		}
		// If scores are equal, maintain stable ordering
		return i < j
	})

	// Apply result limit
	if len(allResults) > maxResults {
		allResults = allResults[:maxResults]
	}

	return allResults, nil
}
