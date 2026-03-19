package fs

// Internal tests (package fs) covering error paths that are unreachable
// from the external test package.

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jazicorn/hatch/internal/source"
)

// errorOnReadFile implements io.ReadSeekCloser but errors on first Read.
type errorOnReadFile struct{ pos int }

func (f *errorOnReadFile) Read(_ []byte) (int, error)         { return 0, fmt.Errorf("read error") }
func (f *errorOnReadFile) Seek(_ int64, _ int) (int64, error) { return 0, nil }
func (f *errorOnReadFile) Close() error                       { return nil }

// errorOnSeekFile reads the sniff buffer successfully but fails on Seek.
type errorOnSeekFile struct {
	data []byte
	pos  int
}

func (f *errorOnSeekFile) Read(p []byte) (int, error) {
	if f.pos >= len(f.data) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.pos:])
	f.pos += n
	return n, nil
}

func (f *errorOnSeekFile) Seek(_ int64, _ int) (int64, error) {
	return 0, fmt.Errorf("seek not supported")
}

func (f *errorOnSeekFile) Close() error { return nil }

// errorOnReadAllFile reads the sniff buffer and seeks successfully but fails on ReadAll.
type errorOnReadAllFile struct {
	data  []byte
	pos   int
	reads int
}

func (f *errorOnReadAllFile) Read(p []byte) (int, error) {
	if f.reads > 0 {
		return 0, fmt.Errorf("read error after seek")
	}
	if f.pos >= len(f.data) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.pos:])
	f.pos += n
	return n, nil
}

func (f *errorOnReadAllFile) Seek(_ int64, _ int) (int64, error) {
	f.pos = 0
	f.reads++
	return 0, nil
}

func (f *errorOnReadAllFile) Close() error { return nil }

func TestReadTextFileReadSniffError(t *testing.T) {
	orig := osOpenFile
	osOpenFile = func(_ string) (io.ReadSeekCloser, error) {
		return &errorOnReadFile{}, nil
	}
	defer func() { osOpenFile = orig }()

	_, err := readTextFile("/fake/path")
	if err == nil {
		t.Error("expected error when sniff Read fails")
	}
}

func TestReadTextFileSeekError(t *testing.T) {
	orig := osOpenFile
	osOpenFile = func(_ string) (io.ReadSeekCloser, error) {
		return &errorOnSeekFile{data: []byte("text content")}, nil
	}
	defer func() { osOpenFile = orig }()

	_, err := readTextFile("/fake/path")
	if err == nil {
		t.Error("expected error when Seek fails")
	}
}

func TestReadTextFileReadAllError(t *testing.T) {
	orig := osOpenFile
	osOpenFile = func(_ string) (io.ReadSeekCloser, error) {
		return &errorOnReadAllFile{data: []byte("text content")}, nil
	}
	defer func() { osOpenFile = orig }()

	_, err := readTextFile("/fake/path")
	if err == nil {
		t.Error("expected error when ReadAll fails")
	}
}

func TestWalkEntryRelError(t *testing.T) {
	// Override filepathRel to force an error.
	orig := filepathRel
	filepathRel = func(_, _ string) (string, error) {
		return "", fmt.Errorf("forced rel error")
	}
	defer func() { filepathRel = orig }()

	root := t.TempDir()
	// Write a real file so WalkDir enters the callback with a file entry.
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	f := &Fetcher{cfg: Config{Root: root, SourceName: "s"}}
	var docs []source.Document
	ctx := context.Background()

	// Manually call walkEntry simulating a file entry.
	info, err := os.Lstat(filepath.Join(root, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	entry := &fakeEntry{name: "a.txt", info: info}
	err = f.walkEntry(ctx, &docs, filepath.Join(root, "a.txt"), entry, nil)
	if err == nil {
		t.Error("expected error when filepathRel fails")
	}
}

// fakeEntry implements os.DirEntry for testing.
type fakeEntry struct {
	name string
	info os.FileInfo
}

func (e *fakeEntry) Name() string               { return e.name }
func (e *fakeEntry) IsDir() bool                { return e.info.IsDir() }
func (e *fakeEntry) Type() os.FileMode          { return e.info.Mode().Type() }
func (e *fakeEntry) Info() (os.FileInfo, error) { return e.info, nil }
