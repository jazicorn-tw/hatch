// Package openai provides an Embedder that calls the OpenAI Embeddings API.
package openai

import (
	"context"
	"fmt"

	oai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	defaultModel     = oai.EmbeddingModelTextEmbedding3Small
	defaultBatchSize = 100
)

// Config holds OpenAI API parameters for the embedder.
type Config struct {
	// APIKey is the OpenAI API key. Required.
	APIKey string
	// Model is the embedding model name. Defaults to text-embedding-3-small.
	Model string
	// BatchSize is the maximum number of texts sent per API call. Defaults to 100.
	BatchSize int
}

// Embedder calls the OpenAI Embeddings API in batches.
type Embedder struct {
	cfg    Config
	client *oai.Client
}

// New returns a configured OpenAI Embedder.
// Returns an error if APIKey is empty.
func New(cfg Config) (*Embedder, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai embedder: api key is required")
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultBatchSize
	}
	client := oai.NewClient(option.WithAPIKey(cfg.APIKey))
	return &Embedder{cfg: cfg, client: &client}, nil
}

// Embed implements embedder.Embedder.
// It batches texts into groups of cfg.BatchSize and calls the API for each batch.
// Results are returned in the same order as the input texts.
func (e *Embedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	result := make([][]float32, 0, len(texts))

	for batchNum, start := 0, 0; start < len(texts); start += e.cfg.BatchSize {
		end := start + e.cfg.BatchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[start:end]

		resp, err := e.client.Embeddings.New(ctx, oai.EmbeddingNewParams{
			Model: e.cfg.Model,
			Input: oai.EmbeddingNewParamsInputUnion{
				OfArrayOfStrings: batch,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("openai embed batch %d: %w", batchNum, err)
		}

		for _, emb := range resp.Data {
			vec := make([]float32, len(emb.Embedding))
			for i, v := range emb.Embedding {
				vec[i] = float32(v)
			}
			result = append(result, vec)
		}
		batchNum++
	}

	return result, nil
}
