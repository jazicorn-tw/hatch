// Package markdown provides a heading-based Chunker for Markdown documents.
package markdown

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/source"
)

// Chunker splits a Markdown document at H1/H2/H3 heading boundaries.
// Each heading and its subsequent body (up to the next heading of equal
// or higher level) becomes one Chunk.
type Chunker struct{}

// New returns a markdown Chunker.
func New() *Chunker { return &Chunker{} }

// Chunk implements chunker.Chunker.
// If the document has no headings, the entire content is returned as one Chunk.
func (c *Chunker) Chunk(doc source.Document) ([]chunker.Chunk, error) {
	type section struct {
		heading string
		level   int
		lines   []string
	}

	var sections []section
	var current *section

	scanner := bufio.NewScanner(strings.NewReader(doc.Content))
	for scanner.Scan() {
		line := scanner.Text()
		level, heading, isHeading := parseHeading(line)
		if isHeading {
			if current != nil {
				sections = append(sections, *current)
			}
			current = &section{heading: heading, level: level, lines: []string{line}}
		} else {
			if current == nil {
				// Content before the first heading: treat as a pre-amble section.
				current = &section{heading: "", level: 0, lines: []string{}}
			}
			current.lines = append(current.lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("markdown: scan: %w", err)
	}
	if current != nil {
		sections = append(sections, *current)
	}

	// If there are no sections at all, return a single chunk of the entire content.
	if len(sections) == 0 {
		if doc.Content == "" {
			return nil, nil
		}
		return []chunker.Chunk{{
			ID:       fmt.Sprintf("%s#0", doc.ID),
			Source:   doc.Source,
			Text:     doc.Content,
			Metadata: map[string]string{},
		}}, nil
	}

	chunks := make([]chunker.Chunk, 0, len(sections))
	for i, sec := range sections {
		text := strings.Join(sec.lines, "\n")
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		meta := map[string]string{}
		if sec.heading != "" {
			meta["heading"] = sec.heading
			meta["level"] = fmt.Sprintf("%d", sec.level)
		}
		chunks = append(chunks, chunker.Chunk{
			ID:       fmt.Sprintf("%s#%d", doc.ID, i),
			Source:   doc.Source,
			Text:     text,
			Metadata: meta,
		})
	}
	return chunks, nil
}

// parseHeading detects an ATX heading (^#{1,3} text).
// Returns (level, headingText, true) on match, or (0, "", false) otherwise.
func parseHeading(line string) (int, string, bool) {
	for level := 1; level <= 3; level++ {
		prefix := strings.Repeat("#", level) + " "
		if strings.HasPrefix(line, prefix) {
			return level, strings.TrimSpace(line[len(prefix):]), true
		}
	}
	// Also accept a heading with no trailing space (e.g. "###title") only if
	// the line is entirely hashes — a pathological edge case we intentionally skip.
	return 0, "", false
}
