// Package gemini provides an Embedder that calls the Google Generative AI Embeddings API.
package gemini

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	defaultModel     = "text-embedding-004"
	defaultBatchSize = 100
)

// Config holds Google Generative AI API parameters for the embedder.
type Config struct {
	// APIKey is the Google API key. Required.
	APIKey string
	// Model is the embedding model name. Defaults to text-embedding-004.
	Model string
	// BatchSize is the maximum number of texts sent per API call. Defaults to 100.
	BatchSize int
}

// Embedder calls the Google Generative AI Embeddings API in batches.
type Embedder struct {
	cfg    Config
	client *genai.Client
}

// New returns a configured Gemini Embedder.
// Returns an error if APIKey is empty.
func New(cfg Config) (*Embedder, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini embedder: api key is required")
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultBatchSize
	}
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("gemini embedder: create client: %w", err)
	}
	return &Embedder{cfg: cfg, client: client}, nil
}

// Embed implements embedder.Embedder.
// It batches texts into groups of cfg.BatchSize and calls the API for each batch.
// Results are returned in the same order as the input texts.
func (e *Embedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	result := make([][]float32, 0, len(texts))
	em := e.client.EmbeddingModel(e.cfg.Model)

	for batchNum, start := 0, 0; start < len(texts); start += e.cfg.BatchSize {
		end := start + e.cfg.BatchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[start:end]

		b := em.NewBatch()
		for _, text := range batch {
			b.AddContent(genai.Text(text))
		}

		resp, err := em.BatchEmbedContents(ctx, b)
		if err != nil {
			return nil, fmt.Errorf("gemini embed batch %d: %w", batchNum, err)
		}

		for _, emb := range resp.Embeddings {
			result = append(result, emb.Values)
		}
		batchNum++
	}

	return result, nil
}

// Close releases the underlying client connection.
func (e *Embedder) Close() error {
	return e.client.Close()
}
