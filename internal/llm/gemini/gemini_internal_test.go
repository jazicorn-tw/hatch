package gemini

// Internal tests (package gemini) covering error paths via fake generator injection.

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// fakeGenerator is a test double for the generator interface.
type fakeGenerator struct {
	resp *genai.GenerateContentResponse
	err  error
}

func (f *fakeGenerator) GenerateContent(_ context.Context, _ ...genai.Part) (*genai.GenerateContentResponse, error) {
	return f.resp, f.err
}

// ---------------------------------------------------------------------------
// New() — client error path
// ---------------------------------------------------------------------------

func TestNewClientError(t *testing.T) {
	orig := genaiNewClient
	genaiNewClient = func(ctx context.Context, opts ...option.ClientOption) (*genai.Client, error) {
		return nil, fmt.Errorf("forced client error")
	}
	defer func() { genaiNewClient = orig }()

	_, err := New(Config{APIKey: "key"})
	if err == nil {
		t.Error("expected error from genaiNewClient")
	}
}

// ---------------------------------------------------------------------------
// Complete() — various response shapes
// ---------------------------------------------------------------------------

func TestCompleteGenerateError(t *testing.T) {
	l := &LLM{model: &fakeGenerator{resp: nil, err: fmt.Errorf("generate failed")}}
	_, err := l.Complete(context.Background(), "test")
	if err == nil {
		t.Error("expected error when GenerateContent returns error")
	}
}

func TestCompleteEmptyCandidates(t *testing.T) {
	l := &LLM{model: &fakeGenerator{resp: &genai.GenerateContentResponse{Candidates: nil}}}
	_, err := l.Complete(context.Background(), "test")
	if err == nil {
		t.Error("expected error for empty candidates")
	}
}

func TestCompleteNilContent(t *testing.T) {
	l := &LLM{model: &fakeGenerator{resp: &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{{Content: nil}},
	}}}
	_, err := l.Complete(context.Background(), "test")
	if err == nil {
		t.Error("expected error when candidate Content is nil")
	}
}

func TestCompleteEmptyParts(t *testing.T) {
	l := &LLM{model: &fakeGenerator{resp: &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{{Content: &genai.Content{Parts: []genai.Part{}}}},
	}}}
	_, err := l.Complete(context.Background(), "test")
	if err == nil {
		t.Error("expected error when Parts is empty")
	}
}

func TestCompleteNonTextPart(t *testing.T) {
	l := &LLM{model: &fakeGenerator{resp: &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{{Content: &genai.Content{
			Parts: []genai.Part{genai.Blob{MIMEType: "image/png", Data: []byte{}}},
		}}},
	}}}
	_, err := l.Complete(context.Background(), "test")
	if err == nil {
		t.Error("expected error when Part is not genai.Text")
	}
}

func TestCompleteSuccess(t *testing.T) {
	l := &LLM{model: &fakeGenerator{resp: &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{{Content: &genai.Content{
			Parts: []genai.Part{genai.Text("hello")},
		}}},
	}}}
	got, err := l.Complete(context.Background(), "test")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if got != "hello" {
		t.Errorf("want %q, got %q", "hello", got)
	}
}
