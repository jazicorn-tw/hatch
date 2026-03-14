// Package fake provides test doubles for store interfaces.
package fake

import (
	"context"

	"github.com/jazicorn/hatch/internal/store"
	"github.com/jazicorn/hatch/internal/store/memory"
)

// VecStore is a test double for store.VecStore.
// It embeds memory.Store for Add/Search/Close and records Upsert/DeleteBySource calls
// so tests can assert on pipeline behaviour without requiring CGO or a real database.
type VecStore struct {
	*memory.Store
	UpsertCalls         [][]store.Record
	DeleteBySourceCalls []string
}

// New returns an empty FakeVecStore.
func New() *VecStore {
	return &VecStore{Store: memory.New()}
}

// Upsert records the call and delegates to Add for in-memory visibility.
func (s *VecStore) Upsert(_ context.Context, records []store.Record) error {
	s.UpsertCalls = append(s.UpsertCalls, records)
	return s.Store.Add(context.Background(), records)
}

// DeleteBySource records the call. The in-memory store has no source-scoped delete,
// so this is a no-op beyond recording.
func (s *VecStore) DeleteBySource(_ context.Context, source string) error {
	s.DeleteBySourceCalls = append(s.DeleteBySourceCalls, source)
	return nil
}
