package anthropic

// Internal test — package anthropic — gives direct access to LLM, apiResponse,
// defaultModel, and defaultMaxTokens so we can inject a custom HTTP transport
// without changing the production source.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// redirectTransport rewrites every outbound request to the given target host,
// preserving path/query. This lets us point the hardcoded apiURL at a local
// httptest server without modifying production code.
type redirectTransport struct {
	target *url.URL
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r2 := req.Clone(req.Context())
	r2.URL.Scheme = t.target.Scheme
	r2.URL.Host = t.target.Host
	return http.DefaultTransport.RoundTrip(r2)
}

// testLLM returns an LLM whose HTTP client forwards all requests to srv.
func testLLM(t *testing.T, srv *httptest.Server) *LLM {
	t.Helper()
	u, _ := url.Parse(srv.URL)
	return &LLM{
		cfg: Config{
			APIKey:    "test-key",
			Model:     defaultModel,
			MaxTokens: defaultMaxTokens,
		},
		client: &http.Client{Transport: &redirectTransport{target: u}},
	}
}

func TestCompleteSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := apiResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: "hello world"}},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	llm := testLLM(t, srv)
	got, err := llm.Complete(context.Background(), "say hello")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if got != "hello world" {
		t.Errorf("want %q, got %q", "hello world", got)
	}
}

func TestCompleteAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := apiResponse{
			Error: &struct {
				Message string `json:"message"`
			}{Message: "invalid api key"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	llm := testLLM(t, srv)
	_, err := llm.Complete(context.Background(), "prompt")
	if err == nil {
		t.Error("expected error from API error response")
	}
}

func TestCompleteEmptyContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No content field → empty response
		resp := apiResponse{}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	llm := testLLM(t, srv)
	_, err := llm.Complete(context.Background(), "prompt")
	if err == nil {
		t.Error("expected error for empty response content")
	}
}

func TestCompleteInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	llm := testLLM(t, srv)
	_, err := llm.Complete(context.Background(), "prompt")
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestCompleteRequestError(t *testing.T) {
	// Point at a port that refuses connections.
	u, _ := url.Parse("http://127.0.0.1:1")
	llm := &LLM{
		cfg:    Config{APIKey: "key", Model: defaultModel, MaxTokens: defaultMaxTokens},
		client: &http.Client{Transport: &redirectTransport{target: u}},
	}
	_, err := llm.Complete(context.Background(), "prompt")
	if err == nil {
		t.Error("expected error when server is unreachable")
	}
}

// errorReader always returns an error from Read, simulating a broken connection.
type errorReader struct{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("simulated read error")
}

// errorBodyTransport returns a well-formed HTTP response whose body always
// errors on Read, to cover the io.ReadAll error path in Complete.
type errorBodyTransport struct{}

func (t *errorBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(&errorReader{}),
		Request:    req,
	}, nil
}

func TestCompleteBodyReadError(t *testing.T) {
	llm := &LLM{
		cfg:    Config{APIKey: "key", Model: defaultModel, MaxTokens: defaultMaxTokens},
		client: &http.Client{Transport: &errorBodyTransport{}},
	}
	_, err := llm.Complete(context.Background(), "prompt")
	if err == nil {
		t.Error("expected error when body read fails")
	}
}

func TestCompleteMarshalError(t *testing.T) {
	orig := jsonMarshal
	jsonMarshal = func(_ any) ([]byte, error) { return nil, fmt.Errorf("forced marshal error") }
	defer func() { jsonMarshal = orig }()

	l := &LLM{cfg: Config{APIKey: "key", Model: defaultModel, MaxTokens: defaultMaxTokens}, client: http.DefaultClient}
	_, err := l.Complete(context.Background(), "prompt")
	if err == nil {
		t.Error("expected error when json.Marshal fails")
	}
}

func TestCompleteCreateRequestError(t *testing.T) {
	orig := apiEndpoint
	apiEndpoint = "%" // malformed URL — http.NewRequestWithContext will fail
	defer func() { apiEndpoint = orig }()

	l := &LLM{cfg: Config{APIKey: "key", Model: defaultModel, MaxTokens: defaultMaxTokens}, client: http.DefaultClient}
	_, err := l.Complete(context.Background(), "prompt")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
