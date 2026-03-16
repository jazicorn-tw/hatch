package kata

import "time"

// Language identifies the programming language for a kata.
type Language string

const (
	Go         Language = "go"
	Python     Language = "python"
	JavaScript Language = "javascript"
	Java       Language = "java"
)

// Kata is a code challenge with a description, starter code, and test cases.
type Kata struct {
	ID          string
	Title       string
	Description string
	StarterCode string   // pre-filled in the editor
	Tests       string   // test file run against the user's solution
	Language    Language // programming language for this kata
	Topic       string
	Source      string // ingested source this kata was derived from
}

// KataSession tracks a single kata attempt.
type KataSession struct {
	ID        string
	Topic     string
	KataID    string
	Language  Language
	Passed    bool
	Attempts  int
	StartedAt time.Time
	EndedAt   time.Time
}
