package sqlite

import (
	"context"
	"math"
	"path/filepath"
	"testing"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/store"
)

const (
	testDB      = "test.db"
	errOpen     = "Open: %v"
	srcToDelete = "to-delete"

	// vecDim must match the dimension declared in 002_vec.sql (float[768]).
	vecDim = 768
)

// unitVec returns a vecDim-dimensional unit vector with 1.0 at position pos.
func unitVec(pos int) []float32 {
	v := make([]float32, vecDim)
	v[pos] = 1.0
	return v
}

// openTestStore opens a fresh Store in a temporary directory and registers
// cleanup via t.Cleanup. Tests that exercise the Open function itself should
// call Open directly instead.
func openTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), testDB)
	s, err := Open(path)
	if err != nil {
		t.Fatalf(errOpen, err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), testDB)
	s, err := Open(path)
	if err != nil {
		t.Fatalf(errOpen, err)
	}
	defer s.Close()
}

func TestOpenRunsMigrations(t *testing.T) {
	path := filepath.Join(t.TempDir(), testDB)
	s, err := Open(path)
	if err != nil {
		t.Fatalf(errOpen, err)
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
	s := openTestStore(t)
	ctx := context.Background()
	records := []store.Record{
		{Chunk: chunker.Chunk{ID: "a", Source: "test", Text: "hello"}, Embedding: unitVec(0)},
		{Chunk: chunker.Chunk{ID: "b", Source: "test", Text: "world"}, Embedding: unitVec(1)},
		{Chunk: chunker.Chunk{ID: "c", Source: "test", Text: "other"}, Embedding: unitVec(2)},
	}
	if err := s.Add(ctx, records); err != nil {
		t.Fatalf("Add: %v", err)
	}

	results, err := s.Search(ctx, unitVec(0), 1)
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
	s := openTestStore(t)

	results, err := s.Search(context.Background(), unitVec(0), 5)
	if err != nil {
		t.Fatalf("Search on empty store: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestAddReplaces(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	orig := []store.Record{
		{Chunk: chunker.Chunk{ID: "a", Source: "src", Text: "original"}, Embedding: unitVec(0)},
	}
	if err := s.Add(ctx, orig); err != nil {
		t.Fatalf("Add original: %v", err)
	}

	updated := []store.Record{
		{Chunk: chunker.Chunk{ID: "a", Source: "src", Text: "updated"}, Embedding: unitVec(0)},
	}
	if err := s.Add(ctx, updated); err != nil {
		t.Fatalf("Add updated: %v", err)
	}

	results, err := s.Search(ctx, unitVec(0), 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if results[0].Chunk.Text != "updated" {
		t.Errorf("expected updated text, got %s", results[0].Chunk.Text)
	}
}

func TestUpsert(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	records := []store.Record{
		{Chunk: chunker.Chunk{ID: "u1", Source: "s", Text: "first"}, Embedding: unitVec(0)},
	}
	if err := s.Upsert(ctx, records); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	// Upsert same ID with updated text.
	records[0].Chunk.Text = "second"
	if err := s.Upsert(ctx, records); err != nil {
		t.Fatalf("Upsert (update): %v", err)
	}

	results, err := s.Search(ctx, unitVec(0), 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 || results[0].Chunk.Text != "second" {
		t.Errorf("want text=second, got %v", results)
	}
}

func TestDeleteBySource(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	records := []store.Record{
		{Chunk: chunker.Chunk{ID: "d1", Source: srcToDelete, Text: "bye"}, Embedding: unitVec(0)},
		{Chunk: chunker.Chunk{ID: "d2", Source: "keep", Text: "stay"}, Embedding: unitVec(1)},
	}
	if err := s.Upsert(ctx, records); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	if err := s.DeleteBySource(ctx, srcToDelete); err != nil {
		t.Fatalf("DeleteBySource: %v", err)
	}

	// "keep" source should still be searchable.
	results, err := s.Search(ctx, unitVec(1), 5)
	if err != nil {
		t.Fatalf("Search after delete: %v", err)
	}
	for _, r := range results {
		if r.Chunk.Source == srcToDelete {
			t.Errorf("chunk from deleted source still present: %+v", r.Chunk)
		}
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

func TestUpsertWithNoEmbedding(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	// A record with a nil embedding should be stored without error.
	records := []store.Record{
		{Chunk: chunker.Chunk{ID: "no-emb", Source: "test", Text: "hello"}},
	}
	if err := s.Upsert(ctx, records); err != nil {
		t.Fatalf("Upsert with nil embedding: %v", err)
	}
}

func TestEncodeVecEmpty(t *testing.T) {
	b := encodeVec(nil)
	if len(b) != 0 {
		t.Errorf("expected empty bytes for nil vec, got %d bytes", len(b))
	}
}

func TestDecodeVecEmpty(t *testing.T) {
	v := decodeVec([]byte{})
	if len(v) != 0 {
		t.Errorf("expected empty slice for empty bytes, got %d", len(v))
	}
}

func TestOpenPingError(t *testing.T) {
	// Path inside a non-existent directory causes Ping to fail.
	_, err := Open("/tmp/nonexistent_hatch_test_xyz_12345/sub/test.db")
	if err == nil {
		t.Error("expected error when parent directory does not exist")
	}
}

func TestUpsertCancelledContext(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := s.Upsert(ctx, []store.Record{
		{Chunk: chunker.Chunk{ID: "x", Source: "s", Text: "t"}, Embedding: unitVec(0)},
	})
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestDeleteBySourceCancelledContext(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := s.DeleteBySource(ctx, "src")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestSearchCancelledContext(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.Search(ctx, unitVec(0), 1)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestPrepareUpsertStmtsError(t *testing.T) {
	s := openTestStore(t)
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback() //nolint:errcheck
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = prepareUpsertStmts(ctx, tx)
	if err == nil {
		t.Error("expected error with cancelled context for PrepareContext")
	}
}

func TestExecUpsertRecordChunkError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	stmts, err := prepareUpsertStmts(ctx, tx)
	if err != nil {
		tx.Rollback() //nolint:errcheck
		t.Fatal(err)
	}
	// Rollback makes ExecContext on the prepared statement fail.
	tx.Rollback() //nolint:errcheck
	err = execUpsertRecord(ctx, stmts, store.Record{
		Chunk:     chunker.Chunk{ID: "x", Source: "s", Text: "t"},
		Embedding: unitVec(0),
	})
	stmts.close()
	if err == nil {
		t.Error("expected error after tx rollback")
	}
}
