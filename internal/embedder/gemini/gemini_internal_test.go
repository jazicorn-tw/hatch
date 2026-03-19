package gemini

// Internal tests (package gemini) covering error paths and success path via
// injectable vars.

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// fakeBatcher is a test double for the batcher interface.
type fakeBatcher struct {
	resp *genai.BatchEmbedContentsResponse
	err  error
}

func (f *fakeBatcher) EmbedBatch(_ context.Context, _ []string) (*genai.BatchEmbedContentsResponse, error) {
	return f.resp, f.err
}

func TestNewClientError(t *testing.T) {
	orig := genaiNewClient
	genaiNewClient = func(_ context.Context, _ ...option.ClientOption) (*genai.Client, error) {
		return nil, fmt.Errorf("forced client error")
	}
	defer func() { genaiNewClient = orig }()

	_, err := New(Config{APIKey: "key"})
	if err == nil {
		t.Error("expected error from genaiNewClient")
	}
}

func TestEmbedSuccess(t *testing.T) {
	origModel := newEmbeddingModel
	newEmbeddingModel = func(_ *genai.Client, _ string) batcher {
		return &fakeBatcher{
			resp: &genai.BatchEmbedContentsResponse{
				Embeddings: []*genai.ContentEmbedding{
					{Values: []float32{0.1, 0.2, 0.3}},
				},
			},
		}
	}
	defer func() { newEmbeddingModel = origModel }()

	e := &Embedder{cfg: Config{APIKey: "key", Model: defaultModel, BatchSize: defaultBatchSize}}
	result, err := e.Embed(context.Background(), []string{"hello"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("want 1 embedding, got %d", len(result))
	}
	if len(result[0]) != 3 {
		t.Errorf("want 3 floats, got %d", len(result[0]))
	}
}
