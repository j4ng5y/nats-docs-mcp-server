package index

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// ============================================================================
// Unit Tests
// ============================================================================

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.natsIndex == nil {
		t.Fatal("NATS index is nil")
	}

	if manager.syncpIndex == nil {
		t.Fatal("syncp index is nil")
	}

	if manager.GetNATSIndex().Count() != 0 {
		t.Error("new NATS index should be empty")
	}

	if manager.GetSyncpIndex().Count() != 0 {
		t.Error("new syncp index should be empty")
	}
}

func TestIndexNATS(t *testing.T) {
	manager := NewManager()

	docs := []*Document{
		{
			ID:      "nats-001",
			Title:   "NATS Overview",
			URL:     "https://docs.nats.io/nats-overview",
			Content: "NATS is a messaging system",
		},
		{
			ID:      "nats-002",
			Title:   "JetStream",
			URL:     "https://docs.nats.io/jetstream",
			Content: "JetStream is a persistent message streaming platform",
		},
	}

	err := manager.IndexNATS(docs)
	if err != nil {
		t.Fatalf("failed to index NATS documents: %v", err)
	}

	if manager.GetNATSIndex().Count() != 2 {
		t.Errorf("expected 2 documents in NATS index, got %d", manager.GetNATSIndex().Count())
	}

	// Syncp index should still be empty
	if manager.GetSyncpIndex().Count() != 0 {
		t.Error("syncp index should still be empty")
	}
}

func TestIndexSyncp(t *testing.T) {
	manager := NewManager()

	docs := []*Document{
		{
			ID:      "syncp-001",
			Title:   "Control Plane",
			URL:     "https://docs.synadia.com/control-plane",
			Content: "Synadia Control Plane is a management platform",
		},
	}

	err := manager.IndexSyncp(docs)
	if err != nil {
		t.Fatalf("failed to index syncp documents: %v", err)
	}

	if manager.GetSyncpIndex().Count() != 1 {
		t.Errorf("expected 1 document in syncp index, got %d", manager.GetSyncpIndex().Count())
	}

	// NATS index should still be empty
	if manager.GetNATSIndex().Count() != 0 {
		t.Error("NATS index should still be empty")
	}
}

func TestIndexBothSources(t *testing.T) {
	manager := NewManager()

	natsDocs := []*Document{
		{
			ID:      "nats-001",
			Title:   "NATS Concepts",
			URL:     "https://docs.nats.io/concepts",
			Content: "NATS messaging concepts",
		},
	}

	syncpDocs := []*Document{
		{
			ID:      "syncp-001",
			Title:   "Getting Started",
			URL:     "https://docs.synadia.com/getting-started",
			Content: "Getting started with Synadia",
		},
	}

	if err := manager.IndexNATS(natsDocs); err != nil {
		t.Fatalf("failed to index NATS: %v", err)
	}

	if err := manager.IndexSyncp(syncpDocs); err != nil {
		t.Fatalf("failed to index syncp: %v", err)
	}

	if manager.GetNATSIndex().Count() != 1 {
		t.Errorf("expected 1 document in NATS index, got %d", manager.GetNATSIndex().Count())
	}

	if manager.GetSyncpIndex().Count() != 1 {
		t.Errorf("expected 1 document in syncp index, got %d", manager.GetSyncpIndex().Count())
	}
}

func TestIndexNATS_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		docs        []*Document
		shouldError bool
	}{
		{
			name:        "empty document list",
			docs:        []*Document{},
			shouldError: true,
		},
		{
			name:        "nil document in list",
			docs:        []*Document{nil},
			shouldError: true,
		},
		{
			name: "document with empty ID",
			docs: []*Document{
				{
					ID:      "",
					Title:   "Test",
					Content: "Test content",
				},
			},
			shouldError: true,
		},
		{
			name: "valid documents",
			docs: []*Document{
				{
					ID:      "valid-001",
					Title:   "Test",
					Content: "Test content",
				},
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager() // Fresh manager for each test
			err := manager.IndexNATS(tt.docs)
			if (err != nil) != tt.shouldError {
				t.Errorf("expected error: %v, got error: %v", tt.shouldError, err != nil)
			}
		})
	}
}

func TestStats(t *testing.T) {
	manager := NewManager()

	// Initial stats should show empty indices
	stats := manager.Stats()
	if stats.NATSDocCount != 0 || stats.SyncpDocCount != 0 {
		t.Error("initial stats should show 0 documents")
	}

	// Index NATS documents
	natsDocs := []*Document{
		{ID: "nats-001", Title: "NATS 1", Content: "content1"},
		{ID: "nats-002", Title: "NATS 2", Content: "content2"},
	}
	manager.IndexNATS(natsDocs)

	// Index syncp documents
	syncpDocs := []*Document{
		{ID: "syncp-001", Title: "Syncp 1", Content: "content1"},
	}
	manager.IndexSyncp(syncpDocs)

	stats = manager.Stats()
	if stats.NATSDocCount != 2 {
		t.Errorf("expected 2 NATS docs, got %d", stats.NATSDocCount)
	}

	if stats.SyncpDocCount != 1 {
		t.Errorf("expected 1 syncp doc, got %d", stats.SyncpDocCount)
	}

	if stats.TotalDocCount != 3 {
		t.Errorf("expected 3 total docs, got %d", stats.TotalDocCount)
	}
}

func TestReset(t *testing.T) {
	manager := NewManager()

	// Index some documents
	natsDocs := []*Document{
		{ID: "nats-001", Title: "NATS", Content: "content"},
	}
	syncpDocs := []*Document{
		{ID: "syncp-001", Title: "Syncp", Content: "content"},
	}

	manager.IndexNATS(natsDocs)
	manager.IndexSyncp(syncpDocs)

	// Verify documents are indexed
	if manager.GetNATSIndex().Count() != 1 || manager.GetSyncpIndex().Count() != 1 {
		t.Fatal("documents not properly indexed")
	}

	// Reset
	manager.Reset()

	// Verify both indices are empty
	if manager.GetNATSIndex().Count() != 0 {
		t.Error("NATS index should be empty after reset")
	}

	if manager.GetSyncpIndex().Count() != 0 {
		t.Error("syncp index should be empty after reset")
	}
}

func TestGetNATSIndex_ReturnsSameInstance(t *testing.T) {
	manager := NewManager()

	index1 := manager.GetNATSIndex()
	index2 := manager.GetNATSIndex()

	// Should return the same instance
	if index1 != index2 {
		t.Error("GetNATSIndex should return the same instance")
	}
}

func TestGetSyncpIndex_ReturnsSameInstance(t *testing.T) {
	manager := NewManager()

	index1 := manager.GetSyncpIndex()
	index2 := manager.GetSyncpIndex()

	// Should return the same instance
	if index1 != index2 {
		t.Error("GetSyncpIndex should return the same instance")
	}
}

// ============================================================================
// Property-Based Tests
// ============================================================================

// Feature: syncp-documentation-support, Property 1: Index Independence
// VALIDATES: Requirements 2.1
func TestProperty_IndexIndependence(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property(
		"indexing NATS documents does not affect syncp index",
		prop.ForAll(
			func(natDocCount uint32) bool {
				manager := NewManager()

				// Index some NATS documents
				count := 1 + (int(natDocCount) % 10) // 1-10 documents
				natsDocs := make([]*Document, count)
				for i := 0; i < count; i++ {
					natsDocs[i] = &Document{
						ID:      fmt.Sprintf("nats-%d", i),
						Title:   "NATS Doc",
						Content: "content",
					}
				}

				manager.IndexNATS(natsDocs)
				natsCount := manager.GetNATSIndex().Count()
				syncpCountBefore := manager.GetSyncpIndex().Count()

				// Verify syncp index is still empty
				syncpCountAfter := manager.GetSyncpIndex().Count()

				return natsCount == count && syncpCountBefore == 0 && syncpCountAfter == 0
			},
			gen.UInt32(),
		),
	)

	properties.Property(
		"indexing syncp documents does not affect NATS index",
		prop.ForAll(
			func(syncpDocCount uint32) bool {
				manager := NewManager()

				// Index some syncp documents
				count := 1 + (int(syncpDocCount) % 10) // 1-10 documents
				syncpDocs := make([]*Document, count)
				for i := 0; i < count; i++ {
					syncpDocs[i] = &Document{
						ID:      fmt.Sprintf("syncp-%d", i),
						Title:   "Syncp Doc",
						Content: "content",
					}
				}

				manager.IndexSyncp(syncpDocs)
				syncpCount := manager.GetSyncpIndex().Count()
				natsCountBefore := manager.GetNATSIndex().Count()

				// Verify NATS index is still empty
				natsCountAfter := manager.GetNATSIndex().Count()

				return syncpCount == count && natsCountBefore == 0 && natsCountAfter == 0
			},
			gen.UInt32(),
		),
	)

	properties.Property(
		"modifications to NATS index do not affect syncp index after both are populated",
		prop.ForAll(
			func(natsCount uint32, syncpCount uint32) bool {
				manager := NewManager()

				// Index documents in both
				natsNum := 1 + (int(natsCount) % 10)
				syncpNum := 1 + (int(syncpCount) % 10)

				natsDocs := make([]*Document, natsNum)
				for i := 0; i < natsNum; i++ {
					natsDocs[i] = &Document{
						ID:      fmt.Sprintf("nats-%d", i),
						Title:   "NATS",
						Content: "content",
					}
				}

				syncpDocs := make([]*Document, syncpNum)
				for i := 0; i < syncpNum; i++ {
					syncpDocs[i] = &Document{
						ID:      fmt.Sprintf("syncp-%d", i),
						Title:   "Syncp",
						Content: "content",
					}
				}

				manager.IndexNATS(natsDocs)
				manager.IndexSyncp(syncpDocs)

				syncpCountBefore := manager.GetSyncpIndex().Count()

				// Index more NATS documents
				moreNatsDocs := make([]*Document, 3)
				for i := 0; i < 3; i++ {
					moreNatsDocs[i] = &Document{
						ID:      fmt.Sprintf("nats-extra-%d", i),
						Title:   "NATS Extra",
						Content: "extra",
					}
				}
				manager.IndexNATS(moreNatsDocs)

				syncpCountAfter := manager.GetSyncpIndex().Count()

				return syncpCountBefore == syncpCountAfter && syncpCountAfter == syncpNum
			},
			gen.UInt32(),
			gen.UInt32(),
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: syncp-documentation-support, Property 2: Parser Consistency (placeholder)
// VALIDATES: Requirements 2.2
// Note: This is a placeholder test. Full parser consistency testing will be
// done when we implement the parser extension component.
func TestProperty_ParserConsistency_Placeholder(t *testing.T) {
	// Placeholder for future parser consistency tests
	// Parser consistency will be validated when both NATS and syncp
	// documents are parsed with the same logic
}
