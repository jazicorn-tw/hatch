package memory_test

import (
	"context"
	"testing"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/store"
	"github.com/jazicorn/hatch/internal/store/memory"
)

func TestAddAndSearch(t *testing.T) {
	s := memory.New()
	ctx := context.Background()

	records := []store.Record{
		{Chunk: chunker.Chunk{ID: "a", Source: "test", Text: "hello"}, Embedding: []float32{1, 0, 0, 0}},
		{Chunk: chunker.Chunk{ID: "b", Source: "test", Text: "world"}, Embedding: []float32{0, 1, 0, 0}},
		{Chunk: chunker.Chunk{ID: "c", Source: "test", Text: "other"}, Embedding: []float32{0, 0, 1, 0}},
	}
	if err := s.Add(ctx, records); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Query closest to record "a".
	results, err := s.Search(ctx, []float32{1, 0, 0, 0}, 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Chunk.ID != "a" {
		t.Errorf("expected chunk a, got %s", results[0].Chunk.ID)
	}
}

func TestSearchEmpty(t *testing.T) {
	s := memory.New()
	results, err := s.Search(context.Background(), []float32{1, 0}, 5)
	if err != nil {
		t.Fatalf("Search on empty store: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
