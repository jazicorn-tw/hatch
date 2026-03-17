package openai

// Internal test — package openai — uses option.WithBaseURL to redirect the
// SDK's HTTP calls to a local httptest server without changing production code.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	oai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// testEmbedder returns an Embedder whose HTTP client is redirected to srv.
func testEmbedder(t *testing.T, srv *httptest.Server, batchSize int) *Embedder {
	t.Helper()
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	client := oai.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(srv.URL+"/"),
	)
	return &Embedder{
		cfg:    Config{APIKey: "test-key", Model: defaultModel, BatchSize: batchSize},
		client: &client,
	}
}

// openAIEmbedResponse mirrors the shape the openai-go SDK expects from the
// embeddings endpoint.
type openAIEmbedResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func TestEmbedEmpty(t *testing.T) {
	e := &Embedder{cfg: Config{APIKey: "key", Model: defaultModel, BatchSize: defaultBatchSize}}
	got, err := e.Embed(context.Background(), nil)
	if err != nil || got != nil {
		t.Errorf("empty input: got err=%v result=%v", err, got)
	}
	got, err = e.Embed(context.Background(), []string{})
	if err != nil || got != nil {
		t.Errorf("empty slice: got err=%v result=%v", err, got)
	}
}

func TestEmbedSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIEmbedResponse{
			Object: "list",
			Data: []struct {
				Object    string    `json:"object"`
				Embedding []float64 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{Object: "embedding", Embedding: []float64{0.1, 0.2, 0.3}, Index: 0},
				{Object: "embedding", Embedding: []float64{0.4, 0.5, 0.6}, Index: 1},
			},
			Model: "text-embedding-3-small",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	e := testEmbedder(t, srv, 0)
	vecs, err := e.Embed(context.Background(), []string{"hello", "world"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(vecs) != 2 {
		t.Fatalf("want 2 vectors, got %d", len(vecs))
	}
	if vecs[0][0] != float32(0.1) {
		t.Errorf("want vecs[0][0]=0.1, got %v", vecs[0][0])
	}
}

func TestEmbedBatchSplit(t *testing.T) {
	// batchSize=2 with 3 texts → 2 batches
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := openAIEmbedResponse{
			Object: "list",
			Data: []struct {
				Object    string    `json:"object"`
				Embedding []float64 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{Object: "embedding", Embedding: []float64{0.1}, Index: 0},
			},
			Model: defaultModel,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	e := testEmbedder(t, srv, 2)
	// 3 texts: first batch has 2, second has 1 → server called twice
	_, _ = e.Embed(context.Background(), []string{"a", "b", "c"})
	if callCount != 2 {
		t.Errorf("want 2 server calls for batch split, got %d", callCount)
	}
}

func TestEmbedAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"invalid key","type":"auth","code":"invalid_api_key"}}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	e := testEmbedder(t, srv, 0)
	_, err := e.Embed(context.Background(), []string{"test"})
	if err == nil {
		t.Error("expected error from API error response")
	}
}
