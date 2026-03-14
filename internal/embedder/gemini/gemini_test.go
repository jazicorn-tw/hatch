package gemini_test

import (
	"testing"

	gemiembed "github.com/jazicorn/hatch/internal/embedder/gemini"
)

func TestNewRequiresAPIKey(t *testing.T) {
	_, err := gemiembed.New(gemiembed.Config{})
	if err == nil {
		t.Error("want error when APIKey is empty")
	}
}

func TestNewDefaultModel(t *testing.T) {
	e, err := gemiembed.New(gemiembed.Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if e == nil {
		t.Fatal("want non-nil Embedder")
	}
	_ = e.Close()
}

func TestNewDefaultBatchSize(t *testing.T) {
	// New should succeed with only an API key; defaults fill the rest.
	e, err := gemiembed.New(gemiembed.Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_ = e.Close()
}
