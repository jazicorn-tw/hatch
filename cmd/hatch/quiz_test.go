package main

import (
	"bufio"
	"strings"
	"testing"
)

// promptAnswer reads from a bufio.Reader until it gets a valid 1-based answer,
// then returns it as a 0-based index.

func TestPromptAnswerValidFirstTry(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("2\n"))
	got := promptAnswer(r, 4)
	if got != 1 {
		t.Errorf("expected 1 (0-based), got %d", got)
	}
}

func TestPromptAnswerLowerBound(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("1\n"))
	got := promptAnswer(r, 4)
	if got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestPromptAnswerUpperBound(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("4\n"))
	got := promptAnswer(r, 4)
	if got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
}

func TestPromptAnswerInvalidThenValid(t *testing.T) {
	// "abc" is not a number — should retry; "3" is valid
	r := bufio.NewReader(strings.NewReader("abc\n3\n"))
	got := promptAnswer(r, 4)
	if got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
}

func TestPromptAnswerOutOfRangeThenValid(t *testing.T) {
	// "5" is out of range for 4 options; "1" is valid
	r := bufio.NewReader(strings.NewReader("5\n1\n"))
	got := promptAnswer(r, 4)
	if got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestPromptAnswerZeroThenValid(t *testing.T) {
	// "0" is below the valid range; "2" is valid
	r := bufio.NewReader(strings.NewReader("0\n2\n"))
	got := promptAnswer(r, 4)
	if got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}

func TestPromptAnswerWhitespaceTrimmed(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("  3  \n"))
	got := promptAnswer(r, 4)
	if got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
}
