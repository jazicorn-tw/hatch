package gemini

// Internal test — package gemini — covers the Complete() error path using a
// pre-cancelled context, which causes the gRPC call to fail immediately.

import (
	"context"
	"testing"
)

func TestCompleteWithCancelledContext(t *testing.T) {
	llm, err := New(Config{APIKey: "fake-api-key"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer llm.Close() //nolint:errcheck

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the call

	_, err = llm.Complete(ctx, "hello")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestNewDefaultModelValue(t *testing.T) {
	// Verify the default model constant is applied when Model is empty.
	llm, err := New(Config{APIKey: "fake-api-key", Model: ""})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer llm.Close() //nolint:errcheck
	if llm.cfg.Model != defaultModel {
		t.Errorf("want cfg.Model=%q, got %q", defaultModel, llm.cfg.Model)
	}
}
