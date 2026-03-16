package main

import (
	"os"

	"github.com/jazicorn/hatch/internal/config"
	"github.com/jazicorn/hatch/internal/llm"
	anthropicllm "github.com/jazicorn/hatch/internal/llm/anthropic"
	gemillm "github.com/jazicorn/hatch/internal/llm/gemini"
)

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
