package markdown

// Internal tests (package markdown) covering error paths that are
// unreachable from the external test package.

import (
	"fmt"
	"io"
	"testing"

	"github.com/jazicorn/hatch/internal/source"
)

// errReader returns an error after the specified number of successful reads.
type errReader struct {
	data []byte
	pos  int
	fail int // fail on the Nth Read call (0-indexed)
	call int
}

func (r *errReader) Read(p []byte) (int, error) {
	if r.call >= r.fail {
		return 0, fmt.Errorf("simulated read error")
	}
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	r.call++
	return n, nil
}

func TestChunkScanSectionsError(t *testing.T) {
	// Override scanSectionsImpl to return an error.
	orig := scanSectionsImpl
	scanSectionsImpl = func(_ string) ([]section, error) {
		return nil, fmt.Errorf("forced scan error")
	}
	defer func() { scanSectionsImpl = orig }()

	c := New()
	_, err := c.Chunk(source.Document{ID: "x.md", Source: "s", Content: "# Hello"})
	if err == nil {
		t.Error("expected error when scanSections fails")
	}
}

func TestChunkNoSectionsNonEmptyContent(t *testing.T) {
	// Override scanSectionsImpl to return empty sections for non-empty content,
	// covering the "no sections, content != empty" branch.
	orig := scanSectionsImpl
	scanSectionsImpl = func(_ string) ([]section, error) {
		return []section{}, nil
	}
	defer func() { scanSectionsImpl = orig }()

	c := New()
	chunks, err := c.Chunk(source.Document{ID: "x.md", Source: "s", Content: "some content"})
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("want 1 chunk for non-empty content with no sections, got %d", len(chunks))
	}
	if chunks[0].Text != "some content" {
		t.Errorf("want Text=some content, got %q", chunks[0].Text)
	}
}

func TestScanSectionsScannerError(t *testing.T) {
	// Override newContentReader to return a reader that errors mid-scan.
	orig := newContentReader
	newContentReader = func(_ string) io.Reader {
		return &errReader{data: []byte("# Hello\n"), fail: 0}
	}
	defer func() { newContentReader = orig }()

	_, err := scanSections("# Hello\n")
	if err == nil {
		t.Error("expected error when reader fails during scan")
	}
}
