package fake

import "context"

// Embedder is a test double that returns zero vectors of a fixed dimension.
type Embedder struct {
	Dim int // vector dimension; defaults to 4 if zero
}

// Embed returns a slice of zero vectors, one per input text.
func (e *Embedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	dim := e.Dim
	if dim == 0 {
		dim = 4
	}
	out := make([][]float32, len(texts))
	for i := range out {
		out[i] = make([]float32, dim)
	}
	return out, nil
}
