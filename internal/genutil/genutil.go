// Package genutil provides helpers shared across generator packages.
package genutil

import (
	"strings"

	"github.com/jazicorn/hatch/internal/store"
)

// ChunkData is the template context for a single retrieved chunk.
// Both the quiz and kata generators embed this in their promptData structs.
type ChunkData struct {
	Text string
}

// DiversifyBySource picks at most one chunk per file from candidates,
// returning up to limit results. Candidates are assumed to be ranked by
// relevance (closest first); the first chunk seen for each file wins.
// File identity is derived from the chunk ID by stripping the trailing "#N" index.
func DiversifyBySource(candidates []store.Record, limit int) []store.Record {
	seen := make(map[string]bool, limit)
	out := make([]store.Record, 0, limit)
	for _, r := range candidates {
		file := ChunkFile(r.Chunk.ID)
		if seen[file] {
			continue
		}
		seen[file] = true
		out = append(out, r)
		if len(out) == limit {
			break
		}
	}
	return out
}

// ChunkFile returns the file path portion of a chunk ID (everything before the last "#").
func ChunkFile(id string) string {
	if i := strings.LastIndexByte(id, '#'); i >= 0 {
		return id[:i]
	}
	return id
}

// StripMarkdownFence removes ```json ... ``` or ``` ... ``` wrappers if present.
func StripMarkdownFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Remove opening fence line.
		if idx := strings.IndexByte(s, '\n'); idx >= 0 {
			s = s[idx+1:]
		}
		// Remove closing fence.
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	return s
}
