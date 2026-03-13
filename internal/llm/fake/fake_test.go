package fake_test

import (
	"context"
	"testing"

	"github.com/jazicorn/hatch/internal/llm/fake"
)

func TestCompleteDefault(t *testing.T) {
	l := &fake.LLM{}
	got, err := l.Complete(context.Background(), "anything")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if got != "fake response" {
		t.Errorf("expected 'fake response', got %q", got)
	}
}

func TestCompleteCustom(t *testing.T) {
	l := &fake.LLM{Response: "custom answer"}
	got, err := l.Complete(context.Background(), "anything")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if got != "custom answer" {
		t.Errorf("expected 'custom answer', got %q", got)
	}
}
