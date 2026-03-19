// Package anthropic provides an llm.Completer backed by the Anthropic Messages API.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultModel     = "claude-sonnet-4-6"
	defaultMaxTokens = 2048
	apiURL           = "https://api.anthropic.com/v1/messages"
	apiVersion       = "2023-06-01"
)

// apiEndpoint and jsonMarshal are package-level vars so tests can override
// them to trigger error branches that are otherwise unreachable.
var (
	apiEndpoint = apiURL
	jsonMarshal = json.Marshal
)

// Config holds Anthropic API parameters.
type Config struct {
	// APIKey is the Anthropic API key. Required.
	APIKey string
	// Model is the model name. Defaults to claude-sonnet-4-6.
	Model string
	// MaxTokens is the maximum number of tokens in the response. Defaults to 2048.
	MaxTokens int
}

// LLM calls the Anthropic Messages API.
type LLM struct {
	cfg    Config
	client *http.Client
}

// New returns a configured Anthropic LLM.
// Returns an error if APIKey is empty.
func New(cfg Config) (*LLM, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("anthropic: api key is required")
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = defaultMaxTokens
	}
	return &LLM{cfg: cfg, client: &http.Client{}}, nil
}

type apiRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	Messages  []apiMessage `json:"messages"`
}

type apiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete implements llm.Completer.
// It sends prompt as a single user message and returns the assistant's text reply.
func (l *LLM) Complete(ctx context.Context, prompt string) (string, error) {
	payload := apiRequest{
		Model:     l.cfg.Model,
		MaxTokens: l.cfg.MaxTokens,
		Messages:  []apiMessage{{Role: "user", Content: prompt}},
	}
	data, err := jsonMarshal(payload)
	if err != nil {
		return "", fmt.Errorf("anthropic: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiEndpoint, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("anthropic: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", l.cfg.APIKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic: request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("anthropic: read response: %w", err)
	}

	var body apiResponse
	if err := json.Unmarshal(raw, &body); err != nil {
		return "", fmt.Errorf("anthropic: parse response: %w", err)
	}
	if body.Error != nil {
		return "", fmt.Errorf("anthropic: api error: %s", body.Error.Message)
	}
	if len(body.Content) == 0 {
		return "", fmt.Errorf("anthropic: empty response")
	}
	return body.Content[0].Text, nil
}
