package store_test

import (
	"testing"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/store"
)

func makeRecord(id string, vec []float32) store.Record {
	return store.Record{Chunk: chunker.Chunk{ID: id}, Embedding: vec}
}

func TestTopKZeroK(t *testing.T) {
	records := []store.Record{makeRecord("a", []float32{1, 0})}
	if got := store.TopK(records, []float32{1, 0}, 0); got != nil {
		t.Errorf("want nil for k=0, got %v", got)
	}
}

func TestTopKEmptyRecords(t *testing.T) {
	if got := store.TopK(nil, []float32{1, 0}, 5); got != nil {
		t.Errorf("want nil for empty records, got %v", got)
	}
}

func TestTopKFewerThanK(t *testing.T) {
	records := []store.Record{
		makeRecord("a", []float32{1, 0}),
		makeRecord("b", []float32{0, 1}),
	}
	got := store.TopK(records, []float32{1, 0}, 5)
	if len(got) != 2 {
		t.Errorf("want 2 results, got %d", len(got))
	}
}

func TestTopKReturnsKNearest(t *testing.T) {
	records := []store.Record{
		makeRecord("best", []float32{1, 0}),   // cosine=1 (exact match)
		makeRecord("mid", []float32{0, 1}),    // cosine=0 (orthogonal)
		makeRecord("worst", []float32{-1, 0}), // cosine=-1 (opposite)
	}
	got := store.TopK(records, []float32{1, 0}, 2)
	if len(got) != 2 {
		t.Fatalf("want 2 results, got %d", len(got))
	}
	if got[0].Chunk.ID != "best" {
		t.Errorf("want top result=best, got %q", got[0].Chunk.ID)
	}
}

func TestTopKOrderedByScoreDescending(t *testing.T) {
	records := []store.Record{
		makeRecord("low", []float32{0, 1}),
		makeRecord("high", []float32{1, 0}),
	}
	got := store.TopK(records, []float32{1, 0}, 2)
	if len(got) < 2 {
		t.Fatalf("want 2 results, got %d", len(got))
	}
	if got[0].Chunk.ID != "high" {
		t.Errorf("want highest-score first, got %q", got[0].Chunk.ID)
	}
}

func TestTopKEvictsLowScorer(t *testing.T) {
	// k=2: heap fills with low+mid, then high arrives and evicts low.
	// Records are ordered so the highest-scoring one comes last, forcing eviction
	// (heap.Pop followed by heap.Push) once the heap is full.
	records := []store.Record{
		makeRecord("low", []float32{0, 1}),  // cosine ≈ 0 vs query [1,0]
		makeRecord("mid", []float32{1, 1}),  // cosine ≈ 0.707 vs query [1,0]
		makeRecord("high", []float32{1, 0}), // cosine = 1.0 vs query [1,0]
	}
	got := store.TopK(records, []float32{1, 0}, 2)
	if len(got) != 2 {
		t.Fatalf("want 2 results, got %d", len(got))
	}
	// high and mid should be kept; low should have been evicted.
	ids := make(map[string]bool)
	for _, r := range got {
		ids[r.Chunk.ID] = true
	}
	if !ids["high"] {
		t.Error("want high in results")
	}
	if !ids["mid"] {
		t.Error("want mid in results")
	}
	if ids["low"] {
		t.Error("want low evicted")
	}
}

func TestTopKExactlyK(t *testing.T) {
	records := []store.Record{
		makeRecord("a", []float32{1, 0}),
		makeRecord("b", []float32{0, 1}),
		makeRecord("c", []float32{-1, 0}),
	}
	got := store.TopK(records, []float32{1, 0}, 3)
	if len(got) != 3 {
		t.Errorf("want 3 results, got %d", len(got))
	}
}
