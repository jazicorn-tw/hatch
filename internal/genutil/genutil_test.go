package genutil_test

import (
	"testing"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/genutil"
	"github.com/jazicorn/hatch/internal/store"
)

// ---------------------------------------------------------------------------
// ChunkFile
// ---------------------------------------------------------------------------

func TestChunkFileWithFragment(t *testing.T) {
	if got := genutil.ChunkFile("foo.go#5"); got != "foo.go" {
		t.Errorf("want foo.go, got %s", got)
	}
}

func TestChunkFileWithoutFragment(t *testing.T) {
	if got := genutil.ChunkFile("foo.go"); got != "foo.go" {
		t.Errorf("want foo.go, got %s", got)
	}
}

func TestChunkFileLastFragment(t *testing.T) {
	if got := genutil.ChunkFile("a#b#3"); got != "a#b" {
		t.Errorf("want a#b, got %s", got)
	}
}

func TestChunkFileEmpty(t *testing.T) {
	if got := genutil.ChunkFile(""); got != "" {
		t.Errorf("want empty, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// StripMarkdownFence
// ---------------------------------------------------------------------------

func TestStripMarkdownFenceNone(t *testing.T) {
	in := `[{"text":"hi"}]`
	if got := genutil.StripMarkdownFence(in); got != in {
		t.Errorf("want %q, got %q", in, got)
	}
}

func TestStripMarkdownFenceJSON(t *testing.T) {
	in := "```json\n[{\"text\":\"hi\"}]\n```"
	want := `[{"text":"hi"}]`
	if got := genutil.StripMarkdownFence(in); got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestStripMarkdownFencePlain(t *testing.T) {
	in := "```\nhello\n```"
	want := "hello"
	if got := genutil.StripMarkdownFence(in); got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestStripMarkdownFenceWhitespace(t *testing.T) {
	in := "  ```json\n[]\n```  "
	want := "[]"
	if got := genutil.StripMarkdownFence(in); got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

// ---------------------------------------------------------------------------
// DiversifyBySource
// ---------------------------------------------------------------------------

func records(ids ...string) []store.Record {
	out := make([]store.Record, len(ids))
	for i, id := range ids {
		out[i] = store.Record{Chunk: chunker.Chunk{ID: id, Source: id}}
	}
	return out
}

func TestDiversifyBySourceEmpty(t *testing.T) {
	got := genutil.DiversifyBySource(nil, 5)
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestDiversifyBySourceLimit(t *testing.T) {
	// 5 records from different files, limit 3
	recs := records("a#0", "b#0", "c#0", "d#0", "e#0")
	got := genutil.DiversifyBySource(recs, 3)
	if len(got) != 3 {
		t.Errorf("expected 3, got %d", len(got))
	}
}

func TestDiversifyBySourceDeduplicatesFile(t *testing.T) {
	// 3 records from only 2 files
	recs := records("a#0", "a#1", "b#0")
	got := genutil.DiversifyBySource(recs, 5)
	if len(got) != 2 {
		t.Errorf("expected 2 (one per file), got %d", len(got))
	}
}

func TestDiversifyBySourceFirstWins(t *testing.T) {
	// Two chunks from same file; first one should be chosen
	recs := records("a#0", "a#1")
	got := genutil.DiversifyBySource(recs, 5)
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if got[0].Chunk.ID != "a#0" {
		t.Errorf("expected a#0 to win, got %s", got[0].Chunk.ID)
	}
}
