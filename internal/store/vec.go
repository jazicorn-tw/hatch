package store

import (
	"fmt"
	"math"
)

// Cosine returns the cosine similarity of two float32 vectors in [-1, 1].
// Returns 0 if either vector is the zero vector.
// Panics if len(a) != len(b) — a dimension mismatch is always a programming error.
func Cosine(a, b []float32) float64 {
	if len(a) != len(b) {
		panic(fmt.Sprintf("store: Cosine: dimension mismatch %d vs %d", len(a), len(b)))
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
