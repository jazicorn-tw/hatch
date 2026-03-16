package quiz

import "testing"

func TestCheckCorrect(t *testing.T) {
	e := NewEvaluator()
	q := Question{Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 2}
	if !e.Check(q, 2) {
		t.Error("expected Check to return true for correct answer")
	}
}

func TestCheckIncorrect(t *testing.T) {
	e := NewEvaluator()
	q := Question{Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 2}
	if e.Check(q, 0) {
		t.Error("expected Check to return false for wrong answer")
	}
}

func TestCheckBoundary(t *testing.T) {
	e := NewEvaluator()
	q := Question{CorrectIndex: 0}
	if !e.Check(q, 0) {
		t.Error("expected true for index 0")
	}
	q.CorrectIndex = 3
	if !e.Check(q, 3) {
		t.Error("expected true for index 3")
	}
}
