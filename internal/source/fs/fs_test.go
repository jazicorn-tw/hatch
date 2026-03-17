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

	docs := mustFetch(t, root, "test")
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

	ids := docIDs(mustFetch(t, root, "test"))
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

	docs := mustFetch(t, root, "s")
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

	ids := docIDs(mustFetch(t, root, "s"))
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

	ids := docIDs(mustFetch(t, root, "s"))
	if slices.Contains(ids, ".hidden/secret.md") {
		t.Error(".hidden/secret.md should be skipped")
	}
	if !slices.Contains(ids, "visible.md") {
		t.Error("visible.md should be included")
	}
}

func TestNewNonExistentRoot(t *testing.T) {
	_, err := fssource.New(fssource.Config{Root: "/nonexistent/path/xyz_test", SourceName: "s"})
	if err == nil {
		t.Error("want error for non-existent root")
	}
}

func TestNewRootIsFile(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "notadir.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := fssource.New(fssource.Config{Root: filePath, SourceName: "s"})
	if err == nil {
		t.Error("want error when root is a file, not a directory")
	}
}

func TestFetchGitignoreDir(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".gitignore", "vendor/\n")
	writeFile(t, root, "main.go", "package main")
	writeFile(t, root, "vendor/lib.go", "package vendor")

	ids := docIDs(mustFetch(t, root, "s"))
	if slices.Contains(ids, "vendor/lib.go") {
		t.Error("vendor/lib.go should be gitignored")
	}
	if !slices.Contains(ids, "main.go") {
		t.Error("main.go should be included")
	}
}

func TestFetchWalkError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping as root — chmod restrictions don't apply")
	}
	root := t.TempDir()
	subDir := filepath.Join(root, "restricted")
	if err := os.MkdirAll(subDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// Remove all permissions so WalkDir gets an error entering the directory.
	if err := os.Chmod(subDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(subDir, 0o700) }) //nolint:errcheck

	f, err := fssource.New(fssource.Config{Root: root, SourceName: "s"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = f.Fetch(context.Background())
	if err == nil {
		t.Error("want error for inaccessible directory")
	}
}

func TestFetchSkipsUnreadableFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping as root — chmod restrictions don't apply")
	}
	root := t.TempDir()
	writeFile(t, root, "visible.md", "hello")
	writeFile(t, root, "noperm.md", "secret")
	noPermPath := filepath.Join(root, "noperm.md")
	if err := os.Chmod(noPermPath, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(noPermPath, 0o600) }) //nolint:errcheck

	ids := docIDs(mustFetch(t, root, "s"))
	if slices.Contains(ids, "noperm.md") {
		t.Error("unreadable file should be skipped")
	}
	if !slices.Contains(ids, "visible.md") {
		t.Error("visible.md should be included")
	}
}

func TestFetchSkipsNonRegularFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "real.md", "hello")
	linkPath := filepath.Join(root, "link.md")
	if err := os.Symlink(filepath.Join(root, "real.md"), linkPath); err != nil {
		t.Skip("symlinks not supported:", err)
	}

	ids := docIDs(mustFetch(t, root, "s"))
	if slices.Contains(ids, "link.md") {
		t.Error("symlink link.md should be skipped as non-regular file")
	}
	if !slices.Contains(ids, "real.md") {
		t.Error("real.md should be included")
	}
}

// helpers

func mustFetch(t *testing.T, root, sourceName string) []source.Document {
	t.Helper()
	f, err := fssource.New(fssource.Config{Root: root, SourceName: sourceName})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	docs, err := f.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	return docs
}

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
