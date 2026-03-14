package openai_test

import (
	"testing"

	oaiembed "github.com/jazicorn/hatch/internal/embedder/openai"
)

func TestNewRequiresAPIKey(t *testing.T) {
	_, err := oaiembed.New(oaiembed.Config{})
	if err == nil {
		t.Error("want error when APIKey is empty")
	}
}

func TestNewDefaultModel(t *testing.T) {
	e, err := oaiembed.New(oaiembed.Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if e == nil {
		t.Fatal("want non-nil Embedder")
	}
}

func TestNewDefaultBatchSize(t *testing.T) {
	// New should succeed with only an API key; defaults fill the rest.
	_, err := oaiembed.New(oaiembed.Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
}
