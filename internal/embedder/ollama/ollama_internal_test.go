package ollama

// Internal test — package ollama — to cover the io.ReadAll error path in
// Embed(), which requires direct access to the unexported client field.

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
)

// errorBodyReader always returns an error on Read.
type errorBodyReader struct{}

func (r *errorBodyReader) Read(_ []byte) (int, error) { return 0, fmt.Errorf("read error") }

// errorBodyTransport returns a 200 response whose body always errors on Read.
type errorBodyTransport struct{}

func (t *errorBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(&errorBodyReader{}),
		Request:    req,
	}, nil
}

func TestEmbedBodyReadError(t *testing.T) {
	e := &Embedder{
		cfg:    Config{Host: "http://localhost", Model: "test"},
		client: &http.Client{Transport: &errorBodyTransport{}},
	}
	_, err := e.Embed(context.Background(), []string{"hello"})
	if err == nil {
		t.Error("expected error when body read fails")
	}
}
