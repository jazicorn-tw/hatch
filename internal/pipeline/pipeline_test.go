package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/chunker/markdown"
	fakeembed "github.com/jazicorn/hatch/internal/embedder/fake"
	"github.com/jazicorn/hatch/internal/pipeline"
	"github.com/jazicorn/hatch/internal/source"
	fakesource "github.com/jazicorn/hatch/internal/source/fake"
	"github.com/jazicorn/hatch/internal/store"
	fakestore "github.com/jazicorn/hatch/internal/store/fake"
)

// errChunker always returns an error from Chunk.
type errChunker struct{ err error }

func (c *errChunker) Chunk(_ source.Document) ([]chunker.Chunk, error) {
	return nil, c.err
}

// nilChunker always returns an empty (nil) chunk slice.
type nilChunker struct{}

func (c *nilChunker) Chunk(_ source.Document) ([]chunker.Chunk, error) {
	return nil, nil
}

// mismatchEmbedder returns one fewer vector than texts given.
type mismatchEmbedder struct{}

func (e *mismatchEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	// Return one fewer vector to trigger the mismatch check.
	out := make([][]float32, len(texts)-1)
	for i := range out {
		out[i] = []float32{0}
	}
	return out, nil
}

// errStore always returns an error from Upsert.
type errStore struct{}

func (s *errStore) Upsert(_ context.Context, _ []store.Record) error {
	return errors.New("upsert failed")
}
func (s *errStore) Add(_ context.Context, _ []store.Record) error { return nil }
func (s *errStore) Search(_ context.Context, _ []float32, _ int) ([]store.Record, error) {
	return nil, nil
}
func (s *errStore) DeleteBySource(_ context.Context, _ string) error { return nil }
func (s *errStore) Close() error                                     { return nil }

func TestRunBasic(t *testing.T) {
	src := &fakesource.Fetcher{Docs: []source.Document{
		{ID: "a.md", Source: "test", Content: "# Hello\nworld"},
	}}
	chk := markdown.New()
	emb := &fakeembed.Embedder{Dim: 4}
	st := fakestore.New()

	if err := pipeline.Run(context.Background(), src, chk, emb, st, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(st.UpsertCalls) == 0 {
		t.Error("want at least one Upsert call")
	}
}

func TestRunProgress(t *testing.T) {
	src := &fakesource.Fetcher{Docs: []source.Document{
		{ID: "a.md", Source: "s", Content: "# A\nbody"},
		{ID: "b.md", Source: "s", Content: "# B\nbody"},
	}}
	chk := markdown.New()
	emb := &fakeembed.Embedder{Dim: 4}
	st := fakestore.New()

	progressCh := make(chan pipeline.Progress, 10)
	if err := pipeline.Run(context.Background(), src, chk, emb, st, progressCh); err != nil {
		t.Fatalf("Run: %v", err)
	}
	close(progressCh)

	var updates []pipeline.Progress
	for p := range progressCh {
		updates = append(updates, p)
	}
	if len(updates) == 0 {
		t.Error("want progress updates")
	}
	last := updates[len(updates)-1]
	if last.Done != last.Total {
		t.Errorf("want Done==Total at end, got Done=%d Total=%d", last.Done, last.Total)
	}
}

func TestRunNilProgressCh(t *testing.T) {
	src := &fakesource.Fetcher{Docs: []source.Document{
		{ID: "a.md", Source: "s", Content: "# Title\nbody"},
	}}
	// Should not panic with nil progress channel.
	err := pipeline.Run(context.Background(), src, markdown.New(), &fakeembed.Embedder{Dim: 4}, fakestore.New(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRunContextCancel(t *testing.T) {
	src := &fakesource.Fetcher{Docs: []source.Document{
		{ID: "a.md", Source: "s", Content: "# A\nbody"},
	}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := pipeline.Run(ctx, src, markdown.New(), &fakeembed.Embedder{Dim: 4}, fakestore.New(), nil)
	if err == nil {
		t.Error("want error from cancelled context")
	}
}

func TestRunFetchError(t *testing.T) {
	src := &fakesource.Fetcher{Err: errors.New("fetch failed")}
	err := pipeline.Run(context.Background(), src, markdown.New(), &fakeembed.Embedder{Dim: 4}, fakestore.New(), nil)
	if err == nil {
		t.Error("want error when Fetch fails")
	}
}

func TestRunEmbedError(t *testing.T) {
	src := &fakesource.Fetcher{Docs: []source.Document{
		{ID: "a.md", Source: "s", Content: "# A\nbody"},
	}}
	emb := &fakeembed.Embedder{Dim: 4, Err: errors.New("embed failed")}
	err := pipeline.Run(context.Background(), src, markdown.New(), emb, fakestore.New(), nil)
	if err == nil {
		t.Error("want error when Embed fails")
	}
}

func TestRunChunkError(t *testing.T) {
	src := &fakesource.Fetcher{Docs: []source.Document{
		{ID: "a.md", Source: "s", Content: "# A\nbody"},
	}}
	chk := &errChunker{err: errors.New("chunk failed")}
	err := pipeline.Run(context.Background(), src, chk, &fakeembed.Embedder{Dim: 4}, fakestore.New(), nil)
	if err == nil {
		t.Error("want error when Chunk fails")
	}
}

func TestRunSkipsDocWithNoChunks(t *testing.T) {
	// nilChunker returns empty chunks; document should be silently skipped.
	src := &fakesource.Fetcher{Docs: []source.Document{
		{ID: "a.md", Source: "s", Content: "# A\nbody"},
	}}
	st := fakestore.New()
	err := pipeline.Run(context.Background(), src, &nilChunker{}, &fakeembed.Embedder{Dim: 4}, st, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(st.UpsertCalls) != 0 {
		t.Error("want no Upsert calls when all docs produce no chunks")
	}
}

func TestRunVectorMismatch(t *testing.T) {
	src := &fakesource.Fetcher{Docs: []source.Document{
		{ID: "a.md", Source: "s", Content: "# A\nbody"},
	}}
	err := pipeline.Run(context.Background(), src, markdown.New(), &mismatchEmbedder{}, fakestore.New(), nil)
	if err == nil {
		t.Error("want error when embedder returns wrong number of vectors")
	}
}

func TestRunUpsertError(t *testing.T) {
	src := &fakesource.Fetcher{Docs: []source.Document{
		{ID: "a.md", Source: "s", Content: "# A\nbody"},
	}}
	err := pipeline.Run(context.Background(), src, markdown.New(), &fakeembed.Embedder{Dim: 4}, &errStore{}, nil)
	if err == nil {
		t.Error("want error when Upsert fails")
	}
}
