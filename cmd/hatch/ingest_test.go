package main

import (
	"path/filepath"
	"testing"

	"github.com/jazicorn/hatch/internal/config"
	"github.com/jazicorn/hatch/internal/source"
)

// ---------------------------------------------------------------------------
// findSource
// ---------------------------------------------------------------------------

func TestFindSourceFound(t *testing.T) {
	cfg := &config.Config{
		Sources: []config.SourceConfig{
			{Name: "docs", Path: "./docs", Type: "filesystem"},
			{Name: "src", Path: "./src", Type: "filesystem"},
		},
	}
	s, err := findSource(cfg, "docs")
	if err != nil {
		t.Fatalf("findSource: %v", err)
	}
	if s.Name != "docs" {
		t.Errorf("expected name docs, got %s", s.Name)
	}
}

func TestFindSourceNotFound(t *testing.T) {
	cfg := &config.Config{
		Sources: []config.SourceConfig{
			{Name: "docs", Path: "./docs", Type: "filesystem"},
		},
	}
	_, err := findSource(cfg, "missing")
	if err == nil {
		t.Error("expected error for missing source")
	}
}

func TestFindSourceEmpty(t *testing.T) {
	cfg := &config.Config{}
	_, err := findSource(cfg, "anything")
	if err == nil {
		t.Error("expected error when no sources configured")
	}
}

// ---------------------------------------------------------------------------
// resolveDBPath
// ---------------------------------------------------------------------------

func TestResolveDBPathAbsolute(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "test.db")
	got, err := resolveDBPath(path)
	if err != nil {
		t.Fatalf("resolveDBPath: %v", err)
	}
	if got != path {
		t.Errorf("want %s, got %s", path, got)
	}
}

func TestResolveDBPathTilde(t *testing.T) {
	got, err := resolveDBPath("~/.hatch/test.db")
	if err != nil {
		t.Fatalf("resolveDBPath with tilde: %v", err)
	}
	if got == "" || got == "~/.hatch/test.db" {
		t.Errorf("expected expanded path, got %s", got)
	}
}

func TestResolveDBPathEmpty(t *testing.T) {
	got, err := resolveDBPath("")
	if err != nil {
		t.Fatalf("resolveDBPath empty: %v", err)
	}
	if got == "" {
		t.Error("expected default path, got empty string")
	}
}

// ---------------------------------------------------------------------------
// resolvePath
// ---------------------------------------------------------------------------

func TestResolvePathAbsolute(t *testing.T) {
	abs := "/tmp/mydir"
	got, err := resolvePath(abs)
	if err != nil {
		t.Fatalf("resolvePath: %v", err)
	}
	if got != abs {
		t.Errorf("want %s, got %s", abs, got)
	}
}

func TestResolvePathRelative(t *testing.T) {
	got, err := resolvePath("docs")
	if err != nil {
		t.Fatalf("resolvePath relative: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("expected absolute path, got %s", got)
	}
}

// ---------------------------------------------------------------------------
// dispatchChunker.Chunk
// ---------------------------------------------------------------------------

func TestDispatchChunkerMarkdown(t *testing.T) {
	d := newDispatchChunker()
	doc := source.Document{ID: "README.md", Source: "src", Content: "# Hello\nworld"}
	chunks, err := d.Chunk(doc)
	if err != nil {
		t.Fatalf("Chunk .md: %v", err)
	}
	if len(chunks) == 0 {
		t.Error("expected at least one chunk for markdown doc")
	}
}

func TestDispatchChunkerCode(t *testing.T) {
	d := newDispatchChunker()
	doc := source.Document{ID: "main.go", Source: "src", Content: "package main\n\nfunc main() {}\n"}
	chunks, err := d.Chunk(doc)
	if err != nil {
		t.Fatalf("Chunk .go: %v", err)
	}
	if len(chunks) == 0 {
		t.Error("expected at least one chunk for Go file")
	}
}

func TestDispatchChunkerMDX(t *testing.T) {
	d := newDispatchChunker()
	doc := source.Document{ID: "page.mdx", Source: "src", Content: "# Page\nContent here."}
	chunks, err := d.Chunk(doc)
	if err != nil {
		t.Fatalf("Chunk .mdx: %v", err)
	}
	if len(chunks) == 0 {
		t.Error("expected at least one chunk for MDX file")
	}
}
