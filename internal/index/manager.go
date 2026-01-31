package index

import (
	"fmt"
	"sync"
	"time"
)

// Manager coordinates multiple documentation indices (NATS and Syncp)
type Manager struct {
	natsIndex  *DocumentationIndex
	syncpIndex *DocumentationIndex
	mu         sync.RWMutex
}

// NewManager creates an index manager with separate indices for NATS and syncp documentation
func NewManager() *Manager {
	return &Manager{
		natsIndex:  NewDocumentationIndex(),
		syncpIndex: NewDocumentationIndex(),
	}
}

// IndexNATS adds or updates documents in the NATS index
func (m *Manager) IndexNATS(docs []*Document) error {
	if len(docs) == 0 {
		return fmt.Errorf("cannot index empty document list")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, doc := range docs {
		if doc == nil {
			return fmt.Errorf("cannot index nil document")
		}
		if err := m.natsIndex.Index(doc); err != nil {
			return fmt.Errorf("failed to index NATS document %s: %w", doc.ID, err)
		}
	}

	return nil
}

// IndexSyncp adds or updates documents in the syncp index
func (m *Manager) IndexSyncp(docs []*Document) error {
	if len(docs) == 0 {
		return fmt.Errorf("cannot index empty document list")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, doc := range docs {
		if doc == nil {
			return fmt.Errorf("cannot index nil document")
		}
		if err := m.syncpIndex.Index(doc); err != nil {
			return fmt.Errorf("failed to index syncp document %s: %w", doc.ID, err)
		}
	}

	return nil
}

// GetNATSIndex returns the NATS documentation index
// The returned index should not be modified directly; use IndexNATS instead
func (m *Manager) GetNATSIndex() *DocumentationIndex {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.natsIndex
}

// GetSyncpIndex returns the syncp documentation index
// The returned index should not be modified directly; use IndexSyncp instead
func (m *Manager) GetSyncpIndex() *DocumentationIndex {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.syncpIndex
}

// IndexStats holds statistics for both documentation indices
type IndexStats struct {
	NATSDocCount  int           // Number of documents in NATS index
	SyncpDocCount int           // Number of documents in syncp index
	TotalDocCount int           // Total documents across both indices
	IndexTime     time.Duration // Time taken to build indices
}

// Stats returns statistics about both indices
func (m *Manager) Stats() IndexStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	natsCount := m.natsIndex.Count()
	syncpCount := m.syncpIndex.Count()

	return IndexStats{
		NATSDocCount:  natsCount,
		SyncpDocCount: syncpCount,
		TotalDocCount: natsCount + syncpCount,
		IndexTime:     0, // Caller should track timing separately if needed
	}
}

// Reset clears all indices (useful for testing)
func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.natsIndex = NewDocumentationIndex()
	m.syncpIndex = NewDocumentationIndex()
}
