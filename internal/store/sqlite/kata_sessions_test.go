package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/jazicorn/hatch/internal/kata"
)

func TestSaveKataSession(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	sess := &kata.KataSession{
		ID:        "ks1",
		Topic:     "go",
		KataID:    "kata-hello-world",
		Language:  kata.Go,
		Passed:    true,
		Attempts:  1,
		StartedAt: time.Now(),
		EndedAt:   time.Now(),
	}

	if err := s.SaveKataSession(ctx, sess); err != nil {
		t.Fatalf("SaveKataSession: %v", err)
	}
}

func TestSaveKataSessionReplaces(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	sess := &kata.KataSession{
		ID:        "ks1",
		Topic:     "go",
		KataID:    "kata-hello",
		Language:  kata.Go,
		StartedAt: time.Now(),
	}

	if err := s.SaveKataSession(ctx, sess); err != nil {
		t.Fatalf("first SaveKataSession: %v", err)
	}
	if err := s.SaveKataSession(ctx, sess); err != nil {
		t.Fatalf("second SaveKataSession (replace): %v", err)
	}
}

func TestSaveKataSessionFailed(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	sess := &kata.KataSession{
		ID:        "ks2",
		Topic:     "python",
		KataID:    "kata-fib",
		Language:  kata.Python,
		Passed:    false,
		Attempts:  3,
		StartedAt: time.Now(),
	}

	if err := s.SaveKataSession(ctx, sess); err != nil {
		t.Fatalf("SaveKataSession failed attempt: %v", err)
	}
}

func TestSaveKataSessionCancelledContext(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sess := &kata.KataSession{ID: "ks-err", Topic: "go", Language: kata.Go}
	err := s.SaveKataSession(ctx, sess)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestSaveKataSessionZeroEndedAt(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	// EndedAt is zero — ended_at should be NULL
	sess := &kata.KataSession{
		ID:        "ks3",
		Topic:     "go",
		KataID:    "kata-hello",
		Language:  kata.Go,
		StartedAt: time.Now(),
	}

	if err := s.SaveKataSession(ctx, sess); err != nil {
		t.Fatalf("SaveKataSession with zero EndedAt: %v", err)
	}
}
