//go:build property
// +build property

package index

import (
	"fmt"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty6_IndexedDocumentsAreSearchable validates Property 6:
// For any document that has been indexed, searching for terms from its title
// or content should return that document in the results.
//
// **Validates: Requirements 4.2**
// Feature: nats-docs-mcp-server, Property 6: Indexed Documents Are Searchable
func TestProperty6_IndexedDocumentsAreSearchable(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100 // Minimum 100 iterations as specified
	properties := gopter.NewProperties(parameters)

	properties.Property("indexed documents are searchable by title terms", prop.ForAll(
		func(docID string, title string, content string) bool {
			// Skip if title is empty or only whitespace
			if strings.TrimSpace(title) == "" {
				return true
			}

			// Create and index a document
			idx := NewDocumentationIndex()
			doc := &Document{
				ID:      docID,
				Title:   title,
				URL:     fmt.Sprintf("https://docs.nats.io/%s", docID),
				Content: content,
			}

			err := idx.Index(doc)
			if err != nil {
				t.Logf("Failed to index document: %v", err)
				return false
			}

			// Extract a term from the title
			titleTerms := tokenize(title)
			if len(titleTerms) == 0 {
				return true // No searchable terms
			}

			// Search for the first term from the title
			searchTerm := titleTerms[0]
			results, err := idx.Search(searchTerm, 10)
			if err != nil {
				t.Logf("Search failed: %v", err)
				return false
			}

			// The document should be in the results
			found := false
			for _, result := range results {
				if result.DocumentID == docID {
					found = true
					break
				}
			}

			if !found {
				t.Logf("Document '%s' with title '%s' not found when searching for term '%s'",
					docID, title, searchTerm)
			}

			return found
		},
		genDocID(),
		genTitle(),
		genContent(),
	))

	properties.Property("indexed documents are searchable by content terms", prop.ForAll(
		func(docID string, title string, content string) bool {
			// Skip if content is empty or only whitespace
			if strings.TrimSpace(content) == "" {
				return true
			}

			// Create and index a document
			idx := NewDocumentationIndex()
			doc := &Document{
				ID:      docID,
				Title:   title,
				URL:     fmt.Sprintf("https://docs.nats.io/%s", docID),
				Content: content,
			}

			err := idx.Index(doc)
			if err != nil {
				t.Logf("Failed to index document: %v", err)
				return false
			}

			// Extract a term from the content
			contentTerms := tokenize(content)
			if len(contentTerms) == 0 {
				return true // No searchable terms
			}

			// Search for the first term from the content
			searchTerm := contentTerms[0]
			results, err := idx.Search(searchTerm, 10)
			if err != nil {
				t.Logf("Search failed: %v", err)
				return false
			}

			// The document should be in the results
			found := false
			for _, result := range results {
				if result.DocumentID == docID {
					found = true
					break
				}
			}

			if !found {
				t.Logf("Document '%s' not found when searching for content term '%s'",
					docID, searchTerm)
			}

			return found
		},
		genDocID(),
		genTitle(),
		genContent(),
	))

	properties.Property("indexed documents are searchable by multi-term queries", prop.ForAll(
		func(docID string, title string, content string) bool {
			// Skip if both title and content are empty
			if strings.TrimSpace(title) == "" && strings.TrimSpace(content) == "" {
				return true
			}

			// Create and index a document
			idx := NewDocumentationIndex()
			doc := &Document{
				ID:      docID,
				Title:   title,
				URL:     fmt.Sprintf("https://docs.nats.io/%s", docID),
				Content: content,
			}

			err := idx.Index(doc)
			if err != nil {
				t.Logf("Failed to index document: %v", err)
				return false
			}

			// Extract terms from both title and content
			allTerms := append(tokenize(title), tokenize(content)...)
			if len(allTerms) < 2 {
				return true // Need at least 2 terms for multi-term query
			}

			// Create a multi-term query from first two terms
			query := allTerms[0] + " " + allTerms[1]
			results, err := idx.Search(query, 10)
			if err != nil {
				t.Logf("Search failed: %v", err)
				return false
			}

			// The document should be in the results
			found := false
			for _, result := range results {
				if result.DocumentID == docID {
					found = true
					break
				}
			}

			if !found {
				t.Logf("Document '%s' not found when searching for multi-term query '%s'",
					docID, query)
			}

			return found
		},
		genDocID(),
		genTitle(),
		genContent(),
	))

	// Run properties
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genDocID generates random document IDs
func genDocID() gopter.Gen {
	return gen.Identifier().SuchThat(func(id string) bool {
		return id != ""
	}).Map(func(id string) string {
		return strings.ToLower(id)
	})
}

// genTitle generates random document titles with meaningful words
func genTitle() gopter.Gen {
	return gen.OneConstOf(
		"NATS Messaging System",
		"JetStream Guide",
		"Security Documentation",
		"Configuration Tutorial",
		"Deployment Concepts",
		"Monitoring Performance",
		"Authentication Guide",
		"Clustering Documentation",
		"NATS Concepts",
		"JetStream Tutorial",
	)
}

// genContent generates random document content with meaningful sentences
func genContent() gopter.Gen {
	return gen.OneConstOf(
		"NATS is a high performance messaging system for cloud native applications",
		"JetStream provides persistence and streaming capabilities for NATS",
		"The system supports publish subscribe patterns and request reply",
		"Authentication can be configured using tokens or JWT credentials",
		"Clustering enables high availability and fault tolerance",
		"Monitoring tools help track system performance and health",
		"Security features include TLS encryption and authorization",
		"Configuration files use YAML format for easy management",
		"Documentation provides comprehensive guides and tutorials",
		"Tutorials help users get started with NATS quickly",
	)
}
