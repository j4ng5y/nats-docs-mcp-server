package index

import (
	"fmt"
	"sync"
	"time"
)

// Manager coordinates multiple documentation indices (NATS, Synadia, and GitHub)
type Manager struct {
	natsIndex   *DocumentationIndex
	syadiaIndex  *DocumentationIndex
	githubIndex *DocumentationIndex
	mu          sync.RWMutex
}

// NewManager creates an index manager with separate indices for NATS, syncp, and GitHub documentation
func NewManager() *Manager {
	return &Manager{
		natsIndex:   NewDocumentationIndex(),
		syadiaIndex:  NewDocumentationIndex(),
		githubIndex: NewDocumentationIndex(),
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

// IndexSynadia adds or updates documents in the syncp index
func (m *Manager) IndexSynadia(docs []*Document) error {
	if len(docs) == 0 {
		return fmt.Errorf("cannot index empty document list")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, doc := range docs {
		if doc == nil {
			return fmt.Errorf("cannot index nil document")
		}
		if err := m.syadiaIndex.Index(doc); err != nil {
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

// GetSynadiaIndex returns the syncp documentation index
// The returned index should not be modified directly; use IndexSynadia instead
func (m *Manager) GetSynadiaIndex() *DocumentationIndex {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.syadiaIndex
}

// IndexGitHub adds or updates documents in the GitHub index
func (m *Manager) IndexGitHub(docs []*Document) error {
	if len(docs) == 0 {
		return fmt.Errorf("cannot index empty document list")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, doc := range docs {
		if doc == nil {
			return fmt.Errorf("cannot index nil document")
		}
		if err := m.githubIndex.Index(doc); err != nil {
			return fmt.Errorf("failed to index GitHub document %s: %w", doc.ID, err)
		}
	}

	return nil
}

// GetGitHubIndex returns the GitHub documentation index
// The returned index should not be modified directly; use IndexGitHub instead
func (m *Manager) GetGitHubIndex() *DocumentationIndex {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.githubIndex
}

// IndexStats holds statistics for all documentation indices
type IndexStats struct {
	NATSDocCount   int           // Number of documents in NATS index
	SynadiaDocCount  int           // Number of documents in syncp index
	GitHubDocCount int           // Number of documents in GitHub index
	TotalDocCount  int           // Total documents across all indices
	IndexTime      time.Duration // Time taken to build indices
}

// Stats returns statistics about all indices
func (m *Manager) Stats() IndexStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	natsCount := m.natsIndex.Count()
	syncpCount := m.syadiaIndex.Count()
	githubCount := m.githubIndex.Count()

	return IndexStats{
		NATSDocCount:   natsCount,
		SynadiaDocCount:  syncpCount,
		GitHubDocCount: githubCount,
		TotalDocCount:  natsCount + syncpCount + githubCount,
		IndexTime:      0, // Caller should track timing separately if needed
	}
}

// Reset clears all indices (useful for testing and refresh)
func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.natsIndex = NewDocumentationIndex()
	m.syadiaIndex = NewDocumentationIndex()
	m.githubIndex = NewDocumentationIndex()
}
