package memory

import (
	"context"
	"sync"

	"github.com/jazicorn/hatch/internal/store"
)

// Store is a non-persistent, in-memory implementation of store.Store.
// Intended for unit tests only.
type Store struct {
	mu      sync.RWMutex
	records []store.Record
}

// New returns an empty in-memory Store.
func New() *Store { return &Store{} }

// Add appends records to the in-memory store.
func (s *Store) Add(_ context.Context, records []store.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, records...)
	return nil
}

// Search returns the k nearest records by cosine similarity.
func (s *Store) Search(_ context.Context, vec []float32, k int) ([]store.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return store.TopK(s.records, vec, k), nil
}

// Close is a no-op for the in-memory store.
func (s *Store) Close() error { return nil }
