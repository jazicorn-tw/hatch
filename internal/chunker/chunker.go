package chunker

import "github.com/jazicorn/hatch/internal/source"

// Chunk is a sub-section of a Document, ready for embedding.
type Chunk struct {
	ID       string
	Source   string
	Text     string
	Metadata map[string]string
}

// Chunker splits a Document into indexable Chunks.
type Chunker interface {
	Chunk(doc source.Document) ([]Chunk, error)
}
