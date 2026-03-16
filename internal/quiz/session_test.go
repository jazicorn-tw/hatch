package quiz

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	before := time.Now()
	s := NewSession("s1", "go")
	after := time.Now()

	if s.ID != "s1" {
		t.Errorf("expected ID s1, got %s", s.ID)
	}
	if s.Topic != "go" {
		t.Errorf("expected topic go, got %s", s.Topic)
	}
	if s.StartedAt.Before(before) || s.StartedAt.After(after) {
		t.Errorf("StartedAt out of range: %v", s.StartedAt)
	}
	if !s.EndedAt.IsZero() {
		t.Error("EndedAt should be zero before Finish")
	}
}

func TestScoreAllCorrect(t *testing.T) {
	s := NewSession("s1", "go")
	s.Questions = []Question{
		{CorrectIndex: 0},
		{CorrectIndex: 1},
	}
	s.Answers = []int{0, 1}

	correct, total := s.Score()
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if correct != 2 {
		t.Errorf("expected correct 2, got %d", correct)
	}
}

func TestScorePartiallyCorrect(t *testing.T) {
	s := NewSession("s1", "go")
	s.Questions = []Question{
		{CorrectIndex: 0},
		{CorrectIndex: 1},
		{CorrectIndex: 2},
	}
	s.Answers = []int{0, 3, 2}

	correct, total := s.Score()
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if correct != 2 {
		t.Errorf("expected correct 2, got %d", correct)
	}
}

func TestScoreUnanswered(t *testing.T) {
	s := NewSession("s1", "go")
	s.Questions = []Question{
		{CorrectIndex: 0},
		{CorrectIndex: 1},
	}
	// No answers provided
	correct, total := s.Score()
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if correct != 0 {
		t.Errorf("expected 0 correct for no answers, got %d", correct)
	}
}

func TestScoreEmpty(t *testing.T) {
	s := NewSession("s1", "go")
	correct, total := s.Score()
	if total != 0 || correct != 0 {
		t.Errorf("expected 0/0, got %d/%d", correct, total)
	}
}

func TestFinish(t *testing.T) {
	s := NewSession("s1", "go")
	before := time.Now()
	s.Finish()
	after := time.Now()

	if s.EndedAt.Before(before) || s.EndedAt.After(after) {
		t.Errorf("EndedAt out of range: %v", s.EndedAt)
	}
}
