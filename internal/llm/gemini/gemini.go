// Package gemini provides an llm.Completer backed by the Google Generative AI API.
package gemini

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const defaultModel = "gemini-2.0-flash"

// Config holds Google Generative AI API parameters for the LLM.
type Config struct {
	// APIKey is the Google API key. Required.
	APIKey string
	// Model is the generative model name. Defaults to gemini-2.0-flash.
	Model string
}

// LLM calls the Google Generative AI content generation API.
type LLM struct {
	cfg    Config
	client *genai.Client
}

// New returns a configured Gemini LLM.
// Returns an error if APIKey is empty.
func New(cfg Config) (*LLM, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini llm: api key is required")
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("gemini llm: create client: %w", err)
	}
	return &LLM{cfg: cfg, client: client}, nil
}

// Complete implements llm.Completer.
// It sends prompt as a single user message and returns the model's text reply.
func (l *LLM) Complete(ctx context.Context, prompt string) (string, error) {
	model := l.client.GenerativeModel(l.cfg.Model)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini llm: generate: %w", err)
	}
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil ||
		len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini llm: empty response")
	}
	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("gemini llm: unexpected response part type")
	}
	return string(text), nil
}

// Close releases the underlying client connection.
func (l *LLM) Close() error {
	return l.client.Close()
}
