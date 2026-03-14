package fs_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/jazicorn/hatch/internal/source"
	fssource "github.com/jazicorn/hatch/internal/source/fs"
)

func TestFetchReturnsFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "# Hello")
	writeFile(t, root, "b.md", "# World")
	writeFile(t, root, "sub/c.md", "# Sub")

	f, err := fssource.New(fssource.Config{Root: root, SourceName: "test"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	docs, err := f.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(docs) != 3 {
		t.Errorf("want 3 docs, got %d", len(docs))
	}
	for _, d := range docs {
		if d.Source != "test" {
			t.Errorf("want Source=test, got %q", d.Source)
		}
	}
}

func TestFetchRespectsGitignore(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".gitignore", "*.log\n")
	writeFile(t, root, "notes.md", "hello")
	writeFile(t, root, "debug.log", "ignored")

	f, err := fssource.New(fssource.Config{Root: root, SourceName: "test"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	docs, err := f.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	ids := docIDs(docs)
	if slices.Contains(ids, "debug.log") {
		t.Error("debug.log should be gitignored")
	}
	if !slices.Contains(ids, "notes.md") {
		t.Error("notes.md should be included")
	}
}

func TestFetchRelativeIDs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "sub/file.md", "content")

	f, err := fssource.New(fssource.Config{Root: root, SourceName: "s"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	docs, err := f.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("want 1 doc, got %d", len(docs))
	}
	if docs[0].ID != "sub/file.md" {
		t.Errorf("want ID=sub/file.md, got %q", docs[0].ID)
	}
}

func TestFetchSkipsBinary(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "text.md", "hello")
	// Write a binary file with a null byte.
	if err := os.WriteFile(filepath.Join(root, "bin.dat"), []byte{0x01, 0x00, 0x02}, 0o600); err != nil {
		t.Fatal(err)
	}

	f, err := fssource.New(fssource.Config{Root: root, SourceName: "s"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	docs, err := f.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	ids := docIDs(docs)
	if slices.Contains(ids, "bin.dat") {
		t.Error("binary file should be skipped")
	}
	if !slices.Contains(ids, "text.md") {
		t.Error("text.md should be included")
	}
}

func TestFetchContextCancel(t *testing.T) {
	root := t.TempDir()
	for i := range 10 {
		writeFile(t, root, filepath.Join("sub", string(rune('a'+i))+".md"), "x")
	}

	f, err := fssource.New(fssource.Config{Root: root, SourceName: "s"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	_, err = f.Fetch(ctx)
	if err == nil {
		t.Error("want error from cancelled context, got nil")
	}
}

func TestFetchSkipsHiddenDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "visible.md", "hello")
	writeFile(t, root, ".hidden/secret.md", "hidden")

	f, err := fssource.New(fssource.Config{Root: root, SourceName: "s"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	docs, err := f.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	ids := docIDs(docs)
	if slices.Contains(ids, ".hidden/secret.md") {
		t.Error(".hidden/secret.md should be skipped")
	}
	if !slices.Contains(ids, "visible.md") {
		t.Error("visible.md should be included")
	}
}

// helpers

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func docIDs(docs []source.Document) []string {
	ids := make([]string, len(docs))
	for i, d := range docs {
		ids[i] = d.ID
	}
	return ids
}
