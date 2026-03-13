package llm

import "context"

// LLM generates text completions from a prompt.
type LLM interface {
	Complete(ctx context.Context, prompt string) (string, error)
}
