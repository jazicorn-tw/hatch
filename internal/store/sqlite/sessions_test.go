package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/jazicorn/hatch/internal/quiz"
)

func TestSaveSession(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	sess := quiz.NewSession("sess1", "go")
	sess.Questions = []quiz.Question{
		{
			ID:           "q1",
			Text:         "What is Go?",
			Options:      [4]string{"A lang", "A game", "A city", "A food"},
			CorrectIndex: 0,
			Explanation:  "Go is a language",
		},
	}
	sess.Answers = []int{0}
	sess.Finish()

	if err := s.SaveSession(ctx, sess); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
}

func TestSaveSessionReplaces(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	sess := quiz.NewSession("sess1", "go")
	sess.Questions = []quiz.Question{
		{ID: "q1", Text: "Q?", Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 0},
	}
	sess.Answers = []int{0}

	if err := s.SaveSession(ctx, sess); err != nil {
		t.Fatalf("first SaveSession: %v", err)
	}
	// Save same ID again — INSERT OR REPLACE should succeed
	if err := s.SaveSession(ctx, sess); err != nil {
		t.Fatalf("second SaveSession (replace): %v", err)
	}
}

func TestSaveSessionZeroEndedAt(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	// EndedAt is zero — ended_at column should be NULL
	sess := quiz.NewSession("sess2", "python")
	if err := s.SaveSession(ctx, sess); err != nil {
		t.Fatalf("SaveSession with zero EndedAt: %v", err)
	}
}

func TestSaveSessionUnansweredQuestions(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	sess := quiz.NewSession("sess3", "go")
	sess.Questions = []quiz.Question{
		{ID: "q1", Text: "Q?", Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 1},
	}
	// Answers slice is shorter than Questions — user_answer should be NULL
	sess.Answers = []int{}

	if err := s.SaveSession(ctx, sess); err != nil {
		t.Fatalf("SaveSession with no answers: %v", err)
	}
}

func TestSaveSessionCancelledContext(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sess := quiz.NewSession("sess-err", "go")
	err := s.SaveSession(ctx, sess)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestSaveSessionNegativeAnswer(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	sess := quiz.NewSession("sess4", "go")
	sess.Questions = []quiz.Question{
		{ID: "q1", Text: "Q?", Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 0},
	}
	// -1 means unanswered
	sess.Answers = []int{-1}
	sess.EndedAt = time.Now()

	if err := s.SaveSession(ctx, sess); err != nil {
		t.Fatalf("SaveSession with -1 answer: %v", err)
	}
}
