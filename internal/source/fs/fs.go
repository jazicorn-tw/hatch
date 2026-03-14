// Package fs provides a filesystem-backed implementation of source.Fetcher.
package fs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/monochromegane/go-gitignore"

	"github.com/jazicorn/hatch/internal/source"
)

// Config holds the constructor options for the filesystem Fetcher.
type Config struct {
	// Root is the absolute path of the directory to walk.
	Root string
	// SourceName is the value assigned to Document.Source for every document
	// emitted from this fetcher.
	SourceName string
}

// Fetcher walks a directory tree and emits one Document per text file,
// respecting any .gitignore files found in the tree.
type Fetcher struct {
	cfg    Config
	ignore gitignore.IgnoreMatcher
}

// New returns a Fetcher for the given Config.
// It loads the .gitignore in Config.Root (if present) for pattern matching.
// Returns an error if Root does not exist or is not a directory.
func New(cfg Config) (*Fetcher, error) {
	info, err := os.Stat(cfg.Root)
	if err != nil {
		return nil, fmt.Errorf("fs: stat root %s: %w", cfg.Root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("fs: root %s is not a directory", cfg.Root)
	}

	// Load .gitignore from root if present; ignore errors (file may not exist).
	matcher, _ := gitignore.NewGitIgnore(filepath.Join(cfg.Root, ".gitignore"), cfg.Root)

	return &Fetcher{cfg: cfg, ignore: matcher}, nil
}

// Fetch walks cfg.Root and returns one Document per text file.
// Binary files (null byte in first 512 bytes) are skipped silently.
// The walk is cancelled if ctx is done.
func (f *Fetcher) Fetch(ctx context.Context) ([]source.Document, error) {
	var docs []source.Document

	err := filepath.WalkDir(f.cfg.Root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Check for cancellation before processing each entry.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip the root itself.
		if path == f.cfg.Root {
			return nil
		}

		rel, err := filepath.Rel(f.cfg.Root, path)
		if err != nil {
			return fmt.Errorf("fs: rel path for %s: %w", path, err)
		}

		// Apply gitignore rules to both files and directories.
		if f.ignore != nil && f.ignore.Match(path, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden directories (e.g. .git, .github).
		if d.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.Type().IsRegular() {
			return nil
		}

		content, err := readTextFile(path)
		if err != nil {
			// Skip unreadable or binary files.
			return nil //nolint:nilerr
		}

		docs = append(docs, source.Document{
			ID:      filepath.ToSlash(rel),
			Source:  f.cfg.SourceName,
			Content: content,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fs: walk %s: %w", f.cfg.Root, err)
	}
	return docs, nil
}

// readTextFile reads a file's content as a string.
// Returns an error if the file is binary (contains a null byte in the first
// 512 bytes) or cannot be read.
func readTextFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Read a sniff buffer to detect binary content.
	sniff := make([]byte, 512)
	n, err := f.Read(sniff)
	if err != nil && err != io.EOF {
		return "", err
	}
	for _, b := range sniff[:n] {
		if b == 0 {
			return "", fmt.Errorf("binary file")
		}
	}

	// Re-read the full file.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
