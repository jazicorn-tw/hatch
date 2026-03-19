// Package markdown provides a heading-based Chunker for Markdown documents.
package markdown

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/source"
)

// scanSectionsImpl is a var so internal tests can override scanSections to
// inject errors or control what it returns.
var scanSectionsImpl = scanSections

// Chunker splits a Markdown document at H1/H2/H3 heading boundaries.
// Each heading and its subsequent body (up to the next heading of equal
// or higher level) becomes one Chunk.
type Chunker struct{}

// New returns a markdown Chunker.
func New() *Chunker { return &Chunker{} }

type section struct {
	heading string
	level   int
	lines   []string
}

// Chunk implements chunker.Chunker.
// If the document has no headings, the entire content is returned as one Chunk.
func (c *Chunker) Chunk(doc source.Document) ([]chunker.Chunk, error) {
	sections, err := scanSectionsImpl(doc.Content)
	if err != nil {
		return nil, err
	}

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

	return buildChunks(sections, doc.ID, doc.Source), nil
}

// newContentReader creates a reader for the given content string.
// It is a var so internal tests can inject an error-returning reader.
var newContentReader = func(s string) io.Reader { return strings.NewReader(s) }

// scanSections parses content into heading-delimited sections.
func scanSections(content string) ([]section, error) {
	var sections []section
	var current *section

	scanner := bufio.NewScanner(newContentReader(content))
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
	return sections, nil
}

// buildChunks converts sections into Chunks, skipping empty sections.
func buildChunks(sections []section, docID, docSource string) []chunker.Chunk {
	chunks := make([]chunker.Chunk, 0, len(sections))
	for i, sec := range sections {
		text := strings.TrimSpace(strings.Join(sec.lines, "\n"))
		if text == "" {
			continue
		}
		meta := map[string]string{}
		if sec.heading != "" {
			meta["heading"] = sec.heading
			meta["level"] = fmt.Sprintf("%d", sec.level)
		}
		chunks = append(chunks, chunker.Chunk{
			ID:       fmt.Sprintf("%s#%d", docID, i),
			Source:   docSource,
			Text:     text,
			Metadata: meta,
		})
	}
	return chunks
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
