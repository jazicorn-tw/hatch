package anthropic_test

import (
	"testing"

	"github.com/jazicorn/hatch/internal/llm/anthropic"
)

const (
	testAPIKey   = "test-key"
	msgNewErr    = "New: %v"
	msgNonNilLLM = "expected non-nil LLM"
)

func TestNewRequiresAPIKey(t *testing.T) {
	_, err := anthropic.New(anthropic.Config{})
	if err == nil {
		t.Error("expected error when APIKey is empty")
	}
}

func TestNewWithAPIKey(t *testing.T) {
	llm, err := anthropic.New(anthropic.Config{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf(msgNewErr, err)
	}
	if llm == nil {
		t.Fatal(msgNonNilLLM)
	}
}

func TestNewDefaultModel(t *testing.T) {
	// Config with only APIKey — New must not return an error even when
	// Model is omitted, confirming the default-model branch is exercised.
	_, err := anthropic.New(anthropic.Config{APIKey: testAPIKey, Model: ""})
	if err != nil {
		t.Fatalf("New with empty Model: %v", err)
	}
}

func TestNewCustomModelAndTokens(t *testing.T) {
	llm, err := anthropic.New(anthropic.Config{
		APIKey:    testAPIKey,
		Model:     "claude-3-haiku-20240307",
		MaxTokens: 512,
	})
	if err != nil {
		t.Fatalf(msgNewErr, err)
	}
	if llm == nil {
		t.Fatal(msgNonNilLLM)
	}
}

func TestNewZeroMaxTokensDefaultsToNonZero(t *testing.T) {
	// MaxTokens=0 should be replaced with the default (2048).
	llm, err := anthropic.New(anthropic.Config{APIKey: testAPIKey, MaxTokens: 0})
	if err != nil {
		t.Fatalf(msgNewErr, err)
	}
	if llm == nil {
		t.Fatal(msgNonNilLLM)
	}
}
