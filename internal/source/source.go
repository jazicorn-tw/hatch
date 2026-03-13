package source

import "context"

// Document is a raw piece of content retrieved from a Source.
type Document struct {
	ID      string
	Source  string
	Content string
}

// Fetcher fetches documents from an external or local origin.
type Fetcher interface {
	Fetch(ctx context.Context) ([]Document, error)
}
