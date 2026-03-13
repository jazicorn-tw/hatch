package llm

import "context"

// Completer generates text completions from a prompt.
type Completer interface {
	Complete(ctx context.Context, prompt string) (string, error)
}
