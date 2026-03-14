package store

import "context"

// VecStore extends Store with idempotent upsert and source-scoped delete.
// The SQLite implementation satisfies this interface; memory.Store satisfies
// the base Store interface only, which is sufficient for unit tests that do
// not exercise the full ingestion pipeline.
type VecStore interface {
	Store
	// Upsert inserts or replaces records, keyed by Chunk.ID.
	Upsert(ctx context.Context, records []Record) error
	// DeleteBySource removes all records whose Chunk.Source matches source.
	DeleteBySource(ctx context.Context, source string) error
}
