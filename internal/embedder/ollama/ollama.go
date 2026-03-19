// Package ollama provides an Embedder that calls a locally running Ollama instance.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultModel = "nomic-embed-text"
	defaultHost  = "http://localhost:11434"
)

// jsonMarshal is a package-level var so tests can override it to inject
// errors in the marshal path.
var jsonMarshal = json.Marshal

// Config holds Ollama API parameters for the embedder.
type Config struct {
	// Host is the base URL of the Ollama server. Defaults to http://localhost:11434.
	Host string
	// Model is the embedding model name. Defaults to nomic-embed-text.
	Model string
}

// Embedder calls the Ollama /api/embed endpoint.
type Embedder struct {
	cfg    Config
	client *http.Client
}

// New returns a configured Ollama Embedder.
func New(cfg Config) (*Embedder, error) {
	if cfg.Host == "" {
		cfg.Host = defaultHost
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	return &Embedder{cfg: cfg, client: &http.Client{}}, nil
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Error      string      `json:"error"`
}

// Embed implements embedder.Embedder.
func (e *Embedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	payload := embedRequest{Model: e.cfg.Model, Input: texts}
	data, err := jsonMarshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ollama embed: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		e.cfg.Host+"/api/embed", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("ollama embed: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embed: request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama embed: read response: %w", err)
	}

	var body embedResponse
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("ollama embed: parse response: %w", err)
	}
	if body.Error != "" {
		return nil, fmt.Errorf("ollama embed: api error: %s", body.Error)
	}
	return body.Embeddings, nil
}
