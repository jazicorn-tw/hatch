package gemini

// Internal test — package gemini — to cover the empty-input early return in
// Embed() and the BatchSize default branch in New(), which are unreachable
// from the external test package.

import (
	"context"
	"testing"
)

func TestEmbedEmptyNil(t *testing.T) {
	e, err := New(Config{APIKey: "fake-key"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer e.Close() //nolint:errcheck

	got, err := e.Embed(context.Background(), nil)
	if err != nil || got != nil {
		t.Errorf("nil input: got err=%v result=%v", err, got)
	}
}

func TestEmbedEmptySlice(t *testing.T) {
	e, err := New(Config{APIKey: "fake-key"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer e.Close() //nolint:errcheck

	got, err := e.Embed(context.Background(), []string{})
	if err != nil || got != nil {
		t.Errorf("empty slice: got err=%v result=%v", err, got)
	}
}

func TestNewCustomBatchSize(t *testing.T) {
	e, err := New(Config{APIKey: "fake-key", BatchSize: 50})
	if err != nil {
		t.Fatalf("New with BatchSize=50: %v", err)
	}
	defer e.Close() //nolint:errcheck
	if e.cfg.BatchSize != 50 {
		t.Errorf("want BatchSize=50, got %d", e.cfg.BatchSize)
	}
}

func TestNewDefaultBatchSizeApplied(t *testing.T) {
	// BatchSize=0 should be replaced with defaultBatchSize.
	e, err := New(Config{APIKey: "fake-key", BatchSize: 0})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer e.Close() //nolint:errcheck
	if e.cfg.BatchSize != defaultBatchSize {
		t.Errorf("want BatchSize=%d, got %d", defaultBatchSize, e.cfg.BatchSize)
	}
}

func TestEmbedCancelledContext(t *testing.T) {
	// A pre-cancelled context causes BatchEmbedContents to fail immediately,
	// covering the error return path in the Embed loop.
	e, err := New(Config{APIKey: "fake-key"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer e.Close() //nolint:errcheck

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the call

	_, err = e.Embed(ctx, []string{"hello"})
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}
