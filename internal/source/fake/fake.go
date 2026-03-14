// Package fake provides test doubles for source interfaces.
package fake

import (
	"context"

	"github.com/jazicorn/hatch/internal/source"
)

// Fetcher is a test double for source.Fetcher.
// It returns a configurable slice of Documents (or an error).
type Fetcher struct {
	Docs []source.Document
	Err  error
}

// Fetch returns the configured Docs, or Err if set.
func (f *Fetcher) Fetch(_ context.Context) ([]source.Document, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Docs, nil
}
