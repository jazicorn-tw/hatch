package main

import (
	"fmt"
	"os"

	"github.com/jazicorn/hatch/internal/config"
	"github.com/jazicorn/hatch/internal/embedder"
	"github.com/jazicorn/hatch/internal/llm"
	anthropicllm "github.com/jazicorn/hatch/internal/llm/anthropic"
	gemillm "github.com/jazicorn/hatch/internal/llm/gemini"
	"github.com/jazicorn/hatch/internal/store/sqlite"
)

// deps holds the shared dependencies for quiz and kata commands.
type deps struct {
	cfg   *config.Config
	emb   embedder.Embedder
	llm   llm.Completer
	store *sqlite.Store
}

// setupDeps loads config, constructs the embedder, LLM, and opens the store.
// The caller is responsible for calling deps.store.Close().
func setupDeps() (*deps, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	emb, err := newEmbedder(cfg)
	if err != nil {
		return nil, fmt.Errorf("create embedder: %w", err)
	}
	completer, err := newLLMCompleter(cfg)
	if err != nil {
		return nil, fmt.Errorf("create llm: %w", err)
	}
	dbPath, err := resolveDBPath(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("resolve db path: %w", err)
	}
	st, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	return &deps{cfg: cfg, emb: emb, llm: completer, store: st}, nil
}

// newLLMCompleter constructs an LLM Completer from config.
// newEmbedder is defined in ingest.go and shared across commands.
func newLLMCompleter(cfg *config.Config) (llm.Completer, error) {
	switch cfg.LLMProvider {
	case "gemini":
		apiKey := cfg.GeminiAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
		return gemillm.New(gemillm.Config{APIKey: apiKey})
	default: // "anthropic" and unset
		apiKey := cfg.AnthropicAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
		return anthropicllm.New(anthropicllm.Config{APIKey: apiKey})
	}
}
