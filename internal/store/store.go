package store

import (
	"context"

	"github.com/jazicorn/hatch/internal/chunker"
)

// Record is a Chunk stored alongside its embedding vector.
type Record struct {
	Chunk     chunker.Chunk
	Embedding []float32
}

// Store persists and retrieves embedded chunks.
type Store interface {
	// Add indexes one or more records.
	Add(ctx context.Context, records []Record) error
	// Search returns the k nearest records to the given vector.
	Search(ctx context.Context, vec []float32, k int) ([]Record, error)
	// Close releases any held resources.
	Close() error
}
