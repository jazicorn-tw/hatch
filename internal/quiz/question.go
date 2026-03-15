// Package quiz implements question generation, answer evaluation, and session tracking.
package quiz

// Question is a single multiple-choice question with exactly 4 options.
type Question struct {
	// ID is a unique identifier for this question within a session.
	ID string
	// Text is the question prompt.
	Text string
	// Options holds the four answer choices.
	Options [4]string
	// CorrectIndex is the 0-based index into Options of the correct answer.
	CorrectIndex int
	// Explanation is a brief justification for the correct answer.
	Explanation string
	// SourceChunks are the chunk IDs that informed this question.
	SourceChunks []string
}
