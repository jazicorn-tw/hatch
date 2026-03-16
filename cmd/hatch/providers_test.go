package main

import (
	"path/filepath"
	"testing"

	"github.com/jazicorn/hatch/internal/config"
)

// ---------------------------------------------------------------------------
// newLLMCompleter
// ---------------------------------------------------------------------------

func TestNewLLMCompleterAnthropicWithKey(t *testing.T) {
	cfg := &config.Config{LLMProvider: "anthropic", AnthropicAPIKey: "test-key"}
	c, err := newLLMCompleter(cfg)
	if err != nil {
		t.Fatalf("newLLMCompleter anthropic: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil completer")
	}
}

func TestNewLLMCompleterAnthropicNoKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	cfg := &config.Config{LLMProvider: "anthropic", AnthropicAPIKey: ""}
	_, err := newLLMCompleter(cfg)
	if err == nil {
		t.Error("expected error when no anthropic key provided")
	}
}

func TestNewLLMCompleterGeminiWithKey(t *testing.T) {
	cfg := &config.Config{LLMProvider: "gemini", GeminiAPIKey: "test-key"}
	c, err := newLLMCompleter(cfg)
	if err != nil {
		t.Fatalf("newLLMCompleter gemini: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil completer")
	}
}

func TestNewLLMCompleterGeminiNoKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	cfg := &config.Config{LLMProvider: "gemini", GeminiAPIKey: ""}
	_, err := newLLMCompleter(cfg)
	if err == nil {
		t.Error("expected error when no gemini key provided")
	}
}

func TestNewLLMCompleterDefaultIsAnthropic(t *testing.T) {
	// Empty LLMProvider falls through to anthropic default.
	cfg := &config.Config{LLMProvider: "", AnthropicAPIKey: "test-key"}
	c, err := newLLMCompleter(cfg)
	if err != nil {
		t.Fatalf("newLLMCompleter default: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil completer")
	}
}

// ---------------------------------------------------------------------------
// setupStore
// ---------------------------------------------------------------------------

func TestSetupStore(t *testing.T) {
	t.Setenv("HATCH_DB_PATH", filepath.Join(t.TempDir(), "test.db"))
	st, err := setupStore()
	if err != nil {
		t.Fatalf("setupStore: %v", err)
	}
	defer st.Close()
}
