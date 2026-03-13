package embedder

import "context"

// Embedder converts text into dense vector representations.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}
