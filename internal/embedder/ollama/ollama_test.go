package ollama_test

import (
	"context"
	"encoding/json"
	"github.com/jazicorn/hatch/internal/embedder/ollama"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// New()
// ---------------------------------------------------------------------------

func TestNewDefaults(t *testing.T) {
	e, err := ollama.New(ollama.Config{})
	if err != nil {
		t.Fatalf("New with empty config: %v", err)
	}
	if e == nil {
		t.Fatal("expected non-nil Embedder")
	}
}

func TestNewCustomHost(t *testing.T) {
	e, err := ollama.New(ollama.Config{Host: "http://localhost:9999", Model: "llama3"})
	if err != nil {
		t.Fatalf("New with custom host: %v", err)
	}
	if e == nil {
		t.Fatal("expected non-nil Embedder")
	}
}

// ---------------------------------------------------------------------------
// Embed()
// ---------------------------------------------------------------------------

func TestEmbedEmpty(t *testing.T) {
	e, _ := ollama.New(ollama.Config{})
	vecs, err := e.Embed(context.Background(), nil)
	if err != nil {
		t.Fatalf("Embed nil: %v", err)
	}
	if vecs != nil {
		t.Errorf("expected nil for empty input, got %v", vecs)
	}

	vecs, err = e.Embed(context.Background(), []string{})
	if err != nil {
		t.Fatalf("Embed empty slice: %v", err)
	}
	if vecs != nil {
		t.Errorf("expected nil for empty slice, got %v", vecs)
	}
}

func TestEmbedSuccess(t *testing.T) {
	want := [][]float32{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		resp := map[string]any{"embeddings": want}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	e, err := ollama.New(ollama.Config{Host: srv.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	vecs, err := e.Embed(context.Background(), []string{"hello", "world"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(vecs) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vecs))
	}
	if vecs[0][0] != 0.1 {
		t.Errorf("expected vecs[0][0]=0.1, got %v", vecs[0][0])
	}
}

func TestEmbedAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"error": "model not found"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	e, _ := ollama.New(ollama.Config{Host: srv.URL})
	_, err := e.Embed(context.Background(), []string{"test"})
	if err == nil {
		t.Error("expected error from API error response")
	}
}

func TestEmbedInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	e, _ := ollama.New(ollama.Config{Host: srv.URL})
	_, err := e.Embed(context.Background(), []string{"test"})
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestEmbedRequestError(t *testing.T) {
	// Point to a server that immediately closes the connection
	e, _ := ollama.New(ollama.Config{Host: "http://127.0.0.1:1"})
	_, err := e.Embed(context.Background(), []string{"test"})
	if err == nil {
		t.Error("expected error when server is unreachable")
	}
}
