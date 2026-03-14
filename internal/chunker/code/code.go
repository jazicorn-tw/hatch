// Package code provides a fixed-size sliding-window Chunker for source code files.
package code

import (
	"fmt"
	"strings"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/source"
)

// Config controls the sliding-window parameters.
type Config struct {
	// WindowSize is the number of lines per chunk.
	WindowSize int
	// Overlap is the number of lines shared between adjacent chunks.
	Overlap int
}

// DefaultConfig returns sensible defaults for code chunking.
func DefaultConfig() Config {
	return Config{WindowSize: 50, Overlap: 10}
}

// Chunker splits a source code Document into overlapping line windows.
type Chunker struct{ cfg Config }

// New returns a Chunker with the given Config.
// Returns an error if Overlap >= WindowSize.
func New(cfg Config) (*Chunker, error) {
	if cfg.Overlap >= cfg.WindowSize {
		return nil, fmt.Errorf("code: overlap (%d) must be less than window size (%d)",
			cfg.Overlap, cfg.WindowSize)
	}
	if cfg.WindowSize <= 0 {
		return nil, fmt.Errorf("code: window size must be positive")
	}
	return &Chunker{cfg: cfg}, nil
}

// Chunk implements chunker.Chunker.
// It splits doc.Content into overlapping windows of cfg.WindowSize lines,
// stepping forward by (WindowSize - Overlap) lines per chunk.
func (c *Chunker) Chunk(doc source.Document) ([]chunker.Chunk, error) {
	if doc.Content == "" {
		return nil, nil
	}

	lines := strings.Split(doc.Content, "\n")
	// Remove a trailing empty line that results from a trailing newline.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return nil, nil
	}

	step := c.cfg.WindowSize - c.cfg.Overlap
	var chunks []chunker.Chunk
	idx := 0

	for start := 0; start < len(lines); start += step {
		end := start + c.cfg.WindowSize
		if end > len(lines) {
			end = len(lines)
		}

		window := lines[start:end]
		text := strings.Join(window, "\n")

		// 1-indexed, inclusive line range.
		startLine := start + 1
		endLine := start + len(window)

		chunks = append(chunks, chunker.Chunk{
			ID:     fmt.Sprintf("%s#%d", doc.ID, idx),
			Source: doc.Source,
			Text:   text,
			Metadata: map[string]string{
				"lines": fmt.Sprintf("%d-%d", startLine, endLine),
			},
		})
		idx++

		// If we've already included the last line, stop.
		if end == len(lines) {
			break
		}
	}

	return chunks, nil
}
