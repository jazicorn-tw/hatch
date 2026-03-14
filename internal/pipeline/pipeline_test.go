package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jazicorn/hatch/internal/chunker/markdown"
	fakeembed "github.com/jazicorn/hatch/internal/embedder/fake"
	"github.com/jazicorn/hatch/internal/pipeline"
	"github.com/jazicorn/hatch/internal/source"
	fakesource "github.com/jazicorn/hatch/internal/source/fake"
	fakestore "github.com/jazicorn/hatch/internal/store/fake"
)

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
