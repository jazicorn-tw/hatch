package gemini_test

import (
	"testing"

	"github.com/jazicorn/hatch/internal/llm/gemini"
)

const (
	testAPIKey   = "fake-api-key"
	msgNonNilLLM = "expected non-nil LLM"
)

// ---------------------------------------------------------------------------
// New()
// ---------------------------------------------------------------------------

func TestNewRequiresAPIKey(t *testing.T) {
	_, err := gemini.New(gemini.Config{})
	if err == nil {
		t.Error("expected error when APIKey is empty")
	}
}

func TestNewWithAPIKey(t *testing.T) {
	llm, err := gemini.New(gemini.Config{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("New with API key: %v", err)
	}
	if llm == nil {
		t.Fatal(msgNonNilLLM)
	}
	_ = llm.Close()
}

func TestNewDefaultModel(t *testing.T) {
	// New should succeed and apply the default model when Model is empty.
	llm, err := gemini.New(gemini.Config{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if llm == nil {
		t.Fatal(msgNonNilLLM)
	}
	_ = llm.Close()
}

func TestNewCustomModel(t *testing.T) {
	llm, err := gemini.New(gemini.Config{APIKey: testAPIKey, Model: "gemini-1.5-pro"})
	if err != nil {
		t.Fatalf("New with custom model: %v", err)
	}
	if llm == nil {
		t.Fatal(msgNonNilLLM)
	}
	_ = llm.Close()
}

// ---------------------------------------------------------------------------
// Close()
// ---------------------------------------------------------------------------

func TestClose(t *testing.T) {
	llm, err := gemini.New(gemini.Config{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := llm.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
