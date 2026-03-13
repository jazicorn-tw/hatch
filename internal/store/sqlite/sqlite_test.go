package sqlite

import (
	"context"
	"math"
	"path/filepath"
	"testing"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/store"
)

func TestOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()
}

func TestOpenRunsMigrations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	// Migrations are idempotent — opening the same DB twice should not error.
	s2, err := Open(path)
	if err != nil {
		t.Fatalf("Open (second time): %v", err)
	}
	defer s2.Close()
}

func TestAddAndSearch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	records := []store.Record{
		{Chunk: chunker.Chunk{ID: "a", Source: "test", Text: "hello"}, Embedding: []float32{1, 0, 0, 0}},
		{Chunk: chunker.Chunk{ID: "b", Source: "test", Text: "world"}, Embedding: []float32{0, 1, 0, 0}},
		{Chunk: chunker.Chunk{ID: "c", Source: "test", Text: "other"}, Embedding: []float32{0, 0, 1, 0}},
	}
	if err := s.Add(ctx, records); err != nil {
		t.Fatalf("Add: %v", err)
	}

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
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	results, err := s.Search(context.Background(), []float32{1, 0}, 5)
	if err != nil {
		t.Fatalf("Search on empty store: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestAddReplaces(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	orig := []store.Record{
		{Chunk: chunker.Chunk{ID: "a", Source: "src", Text: "original"}, Embedding: []float32{1, 0}},
	}
	if err := s.Add(ctx, orig); err != nil {
		t.Fatalf("Add original: %v", err)
	}

	updated := []store.Record{
		{Chunk: chunker.Chunk{ID: "a", Source: "src", Text: "updated"}, Embedding: []float32{1, 0}},
	}
	if err := s.Add(ctx, updated); err != nil {
		t.Fatalf("Add updated: %v", err)
	}

	results, err := s.Search(ctx, []float32{1, 0}, 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if results[0].Chunk.Text != "updated" {
		t.Errorf("expected updated text, got %s", results[0].Chunk.Text)
	}
}

func TestEncodeDecodeVec(t *testing.T) {
	in := []float32{1.5, -0.5, 0, math.MaxFloat32}
	out := decodeVec(encodeVec(in))
	if len(out) != len(in) {
		t.Fatalf("length mismatch: want %d, got %d", len(in), len(out))
	}
	for i := range in {
		if out[i] != in[i] {
			t.Errorf("[%d]: want %v, got %v", i, in[i], out[i])
		}
	}
}

func TestDecodeVecOddBytes(t *testing.T) {
	if got := decodeVec([]byte{1, 2, 3}); got != nil {
		t.Errorf("expected nil for odd-length input, got %v", got)
	}
}
