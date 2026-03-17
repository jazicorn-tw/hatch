package sqlite

import (
	"context"
	"testing"

	"github.com/jazicorn/hatch/internal/quiz"
)

func TestSaveAndListQuestionBank(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	questions := []quiz.Question{
		{
			ID:           "q1",
			Text:         "What is 1+1?",
			Options:      [4]string{"1", "2", "3", "4"},
			CorrectIndex: 1,
			Explanation:  "basic addition",
		},
		{
			ID:           "q2",
			Text:         "What is 2+2?",
			Options:      [4]string{"2", "3", "4", "5"},
			CorrectIndex: 2,
			Explanation:  "basic addition again",
		},
	}

	if err := s.SaveQuestionBank(ctx, "math", questions); err != nil {
		t.Fatalf("SaveQuestionBank: %v", err)
	}

	got, err := s.ListQuestionBank(ctx, "math")
	if err != nil {
		t.Fatalf("ListQuestionBank: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(got))
	}
	if got[0].ID != "q1" {
		t.Errorf("expected id q1, got %s", got[0].ID)
	}
	if got[0].Text != "What is 1+1?" {
		t.Errorf("expected text 'What is 1+1?', got %s", got[0].Text)
	}
	if got[0].Options != questions[0].Options {
		t.Errorf("options mismatch: want %v, got %v", questions[0].Options, got[0].Options)
	}
	if got[0].CorrectIndex != 1 {
		t.Errorf("expected CorrectIndex 1, got %d", got[0].CorrectIndex)
	}
	if got[0].Explanation != "basic addition" {
		t.Errorf("expected explanation 'basic addition', got %s", got[0].Explanation)
	}
}

func TestListQuestionBankEmpty(t *testing.T) {
	s := openTestStore(t)

	got, err := s.ListQuestionBank(context.Background(), "math")
	if err != nil {
		t.Fatalf("ListQuestionBank on empty store: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 questions, got %d", len(got))
	}
}

func TestSaveQuestionBankReplaces(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	orig := []quiz.Question{
		{
			ID:           "q1",
			Text:         "original text",
			Options:      [4]string{"a", "b", "c", "d"},
			CorrectIndex: 0,
			Explanation:  "original",
		},
	}
	if err := s.SaveQuestionBank(ctx, "go", orig); err != nil {
		t.Fatalf("SaveQuestionBank original: %v", err)
	}

	updated := []quiz.Question{
		{
			ID:           "q1",
			Text:         "updated text",
			Options:      [4]string{"w", "x", "y", "z"},
			CorrectIndex: 3,
			Explanation:  "updated",
		},
	}
	if err := s.SaveQuestionBank(ctx, "go", updated); err != nil {
		t.Fatalf("SaveQuestionBank updated: %v", err)
	}

	got, err := s.ListQuestionBank(ctx, "go")
	if err != nil {
		t.Fatalf("ListQuestionBank: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 question, got %d", len(got))
	}
	if got[0].Text != "updated text" {
		t.Errorf("expected updated text, got %s", got[0].Text)
	}
}

func TestSaveQuestionBankCancelledContext(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := s.SaveQuestionBank(ctx, "go", []quiz.Question{
		{ID: "q1", Text: "Q?", Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 0},
	})
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestListQuestionBankCancelledContext(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.ListQuestionBank(ctx, "go")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestListQuestionBankFiltersByTopic(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	goQ := []quiz.Question{
		{ID: "g1", Text: "go question", Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 0},
	}
	pyQ := []quiz.Question{
		{ID: "p1", Text: "python question", Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 1},
	}
	if err := s.SaveQuestionBank(ctx, "go", goQ); err != nil {
		t.Fatalf("SaveQuestionBank go: %v", err)
	}
	if err := s.SaveQuestionBank(ctx, "python", pyQ); err != nil {
		t.Fatalf("SaveQuestionBank python: %v", err)
	}

	got, err := s.ListQuestionBank(ctx, "go")
	if err != nil {
		t.Fatalf("ListQuestionBank go: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 question for topic go, got %d", len(got))
	}
	if got[0].ID != "g1" {
		t.Errorf("expected id g1, got %s", got[0].ID)
	}
}
