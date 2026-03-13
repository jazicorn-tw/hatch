package store_test

import (
	"math"
	"testing"

	"github.com/jazicorn/hatch/internal/store"
)

func TestCosineOrthogonal(t *testing.T) {
	if got := store.Cosine([]float32{1, 0}, []float32{0, 1}); got != 0 {
		t.Errorf("orthogonal vectors: want 0, got %v", got)
	}
}

func TestCosineIdentical(t *testing.T) {
	if got := store.Cosine([]float32{1, 1}, []float32{1, 1}); math.Abs(got-1.0) > 1e-6 {
		t.Errorf("identical vectors: want 1.0, got %v", got)
	}
}

func TestCosineZeroVector(t *testing.T) {
	if got := store.Cosine([]float32{0, 0}, []float32{1, 0}); got != 0 {
		t.Errorf("zero vector: want 0, got %v", got)
	}
}

func TestCosineDimensionMismatchPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on dimension mismatch, got none")
		}
	}()
	store.Cosine([]float32{1, 0}, []float32{1, 0, 0})
}
