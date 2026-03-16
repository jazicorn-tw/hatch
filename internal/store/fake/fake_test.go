package fake_test

import (
	"context"
	"testing"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/store"
	"github.com/jazicorn/hatch/internal/store/fake"
)

func TestNew(t *testing.T) {
	s := fake.New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
}

func TestUpsertRecordsCall(t *testing.T) {
	s := fake.New()
	ctx := context.Background()

	records := []store.Record{
		{Chunk: chunker.Chunk{ID: "a", Source: "src", Text: "hello"}, Embedding: []float32{1, 0}},
	}

	if err := s.Upsert(ctx, records); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	if len(s.UpsertCalls) != 1 {
		t.Fatalf("expected 1 UpsertCall, got %d", len(s.UpsertCalls))
	}
	if len(s.UpsertCalls[0]) != 1 || s.UpsertCalls[0][0].Chunk.ID != "a" {
		t.Errorf("unexpected UpsertCalls content: %v", s.UpsertCalls)
	}
}

func TestUpsertMakesRecordsSearchable(t *testing.T) {
	s := fake.New()
	ctx := context.Background()

	records := []store.Record{
		{Chunk: chunker.Chunk{ID: "a", Source: "src", Text: "hello"}, Embedding: []float32{1, 0}},
	}

	if err := s.Upsert(ctx, records); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	results, err := s.Search(ctx, []float32{1, 0}, 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 || results[0].Chunk.ID != "a" {
		t.Errorf("expected record a in search results, got %v", results)
	}
}

func TestDeleteBySourceRecordsCall(t *testing.T) {
	s := fake.New()
	ctx := context.Background()

	if err := s.DeleteBySource(ctx, "my-source"); err != nil {
		t.Fatalf("DeleteBySource: %v", err)
	}

	if len(s.DeleteBySourceCalls) != 1 || s.DeleteBySourceCalls[0] != "my-source" {
		t.Errorf("unexpected DeleteBySourceCalls: %v", s.DeleteBySourceCalls)
	}
}
