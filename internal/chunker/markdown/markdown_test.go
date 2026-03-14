package markdown_test

import (
	"strings"
	"testing"

	"github.com/jazicorn/hatch/internal/chunker/markdown"
	"github.com/jazicorn/hatch/internal/source"
)

func doc(content string) source.Document {
	return source.Document{ID: "test.md", Source: "src", Content: content}
}

func TestChunkSingleH1(t *testing.T) {
	c := markdown.New()
	chunks, err := c.Chunk(doc("# Hello\n\nSome body text."))
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("want 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Metadata["level"] != "1" {
		t.Errorf("want level=1, got %q", chunks[0].Metadata["level"])
	}
	if chunks[0].Metadata["heading"] != "Hello" {
		t.Errorf("want heading=Hello, got %q", chunks[0].Metadata["heading"])
	}
	if !strings.Contains(chunks[0].Text, "# Hello") {
		t.Errorf("chunk text should contain the heading line")
	}
}

func TestChunkMultipleHeadings(t *testing.T) {
	c := markdown.New()
	content := "# Intro\nIntro body.\n\n## Section A\nA body.\n\n## Section B\nB body."
	chunks, err := c.Chunk(doc(content))
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) != 3 {
		t.Fatalf("want 3 chunks, got %d", len(chunks))
	}
	if chunks[0].Metadata["heading"] != "Intro" {
		t.Errorf("want heading=Intro, got %q", chunks[0].Metadata["heading"])
	}
	if chunks[1].Metadata["level"] != "2" {
		t.Errorf("want level=2, got %q", chunks[1].Metadata["level"])
	}
}

func TestChunkNoHeadings(t *testing.T) {
	c := markdown.New()
	chunks, err := c.Chunk(doc("Just plain text\nwith no headings."))
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("want 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Metadata["heading"] != "" {
		t.Errorf("want no heading metadata for headingless doc")
	}
}

func TestChunkEmptyDocument(t *testing.T) {
	c := markdown.New()
	chunks, err := c.Chunk(doc(""))
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) != 0 {
		t.Errorf("want 0 chunks for empty doc, got %d", len(chunks))
	}
}

func TestChunkIDFormat(t *testing.T) {
	c := markdown.New()
	chunks, err := c.Chunk(doc("# A\nbody\n# B\nbody"))
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	for _, ch := range chunks {
		if !strings.HasPrefix(ch.ID, "test.md#") {
			t.Errorf("chunk ID should start with docID#, got %q", ch.ID)
		}
	}
}

func TestChunkSourcePropagated(t *testing.T) {
	c := markdown.New()
	d := source.Document{ID: "x.md", Source: "mysrc", Content: "# H\nbody"}
	chunks, err := c.Chunk(d)
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("want at least 1 chunk")
	}
	for _, ch := range chunks {
		if ch.Source != "mysrc" {
			t.Errorf("want Source=mysrc, got %q", ch.Source)
		}
	}
}
