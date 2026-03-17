package code_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jazicorn/hatch/internal/chunker/code"
	"github.com/jazicorn/hatch/internal/source"
)

func makeDoc(lines int) source.Document {
	rows := make([]string, lines)
	for i := range rows {
		rows[i] = fmt.Sprintf("line%d", i+1)
	}
	return source.Document{ID: "main.go", Source: "src", Content: strings.Join(rows, "\n")}
}

func TestChunkWindowsCorrect(t *testing.T) {
	// 100 lines, window=50, overlap=10 → step=40
	// chunks: [0-49], [40-89], [80-99] → 3 chunks
	c, err := code.New(code.Config{WindowSize: 50, Overlap: 10})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	chunks, err := c.Chunk(makeDoc(100))
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) != 3 {
		t.Errorf("want 3 chunks, got %d", len(chunks))
	}
}

func TestChunkMetadataLines(t *testing.T) {
	c, err := code.New(code.Config{WindowSize: 10, Overlap: 2})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	chunks, err := c.Chunk(makeDoc(10))
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("want at least 1 chunk")
	}
	if chunks[0].Metadata["lines"] != "1-10" {
		t.Errorf("want lines=1-10, got %q", chunks[0].Metadata["lines"])
	}
}

func TestChunkOverlapContent(t *testing.T) {
	// window=5, overlap=2 → step=3; two adjacent chunks share 2 lines
	c, err := code.New(code.Config{WindowSize: 5, Overlap: 2})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	chunks, err := c.Chunk(makeDoc(8))
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("want at least 2 chunks, got %d", len(chunks))
	}
	lines0 := strings.Split(chunks[0].Text, "\n")
	lines1 := strings.Split(chunks[1].Text, "\n")
	// Last 2 lines of chunk 0 == first 2 lines of chunk 1.
	tail0 := lines0[len(lines0)-2:]
	head1 := lines1[:2]
	for i := range tail0 {
		if tail0[i] != head1[i] {
			t.Errorf("overlap mismatch at pos %d: %q vs %q", i, tail0[i], head1[i])
		}
	}
}

func TestChunkSingleWindow(t *testing.T) {
	c, err := code.New(code.Config{WindowSize: 50, Overlap: 5})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	chunks, err := c.Chunk(makeDoc(20))
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) != 1 {
		t.Errorf("want 1 chunk for small file, got %d", len(chunks))
	}
}

func TestChunkInvalidOverlap(t *testing.T) {
	_, err := code.New(code.Config{WindowSize: 10, Overlap: 10})
	if err == nil {
		t.Error("want error when overlap >= window size")
	}
}

func TestChunkEmptyDocument(t *testing.T) {
	c, err := code.New(code.DefaultConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	chunks, err := c.Chunk(source.Document{ID: "empty.go", Source: "s", Content: ""})
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) != 0 {
		t.Errorf("want 0 chunks for empty doc, got %d", len(chunks))
	}
}

func TestNewZeroWindowSize(t *testing.T) {
	// Overlap must be less than WindowSize to bypass the overlap check and
	// reach the WindowSize <= 0 guard. Use Overlap=-1, WindowSize=0.
	_, err := code.New(code.Config{WindowSize: 0, Overlap: -1})
	if err == nil {
		t.Error("want error when WindowSize is zero")
	}
}

func TestChunkTrailingNewline(t *testing.T) {
	// Content ending with "\n" produces a trailing empty string after Split;
	// the chunker should remove it and produce the same result as without it.
	c, err := code.New(code.Config{WindowSize: 3, Overlap: 0})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	withoutNewline := source.Document{ID: "f.go", Source: "s", Content: "a\nb\nc"}
	withNewline := source.Document{ID: "f.go", Source: "s", Content: "a\nb\nc\n"}

	chunksA, err := c.Chunk(withoutNewline)
	if err != nil {
		t.Fatalf("Chunk (without newline): %v", err)
	}
	chunksB, err := c.Chunk(withNewline)
	if err != nil {
		t.Fatalf("Chunk (with newline): %v", err)
	}
	if len(chunksA) != len(chunksB) {
		t.Errorf("want same chunk count with/without trailing newline: %d vs %d",
			len(chunksA), len(chunksB))
	}
}
