package fake_test

import (
	"context"
	"testing"

	"github.com/jazicorn/hatch/internal/embedder/fake"
)

func TestEmbed(t *testing.T) {
	e := &fake.Embedder{Dim: 8}
	vecs, err := e.Embed(context.Background(), []string{"hello", "world"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(vecs) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vecs))
	}
	for i, v := range vecs {
		if len(v) != 8 {
			t.Errorf("vec %d: expected dim 8, got %d", i, len(v))
		}
	}
}

func TestEmbedDefaultDim(t *testing.T) {
	e := &fake.Embedder{}
	vecs, err := e.Embed(context.Background(), []string{"x"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(vecs[0]) != 4 {
		t.Errorf("expected default dim 4, got %d", len(vecs[0]))
	}
}
