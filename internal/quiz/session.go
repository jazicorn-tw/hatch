package quiz

import "time"

// Session tracks a single quiz attempt: the questions asked, the answers given,
// and the final score.
type Session struct {
	ID        string
	Topic     string
	Questions []Question
	// Answers holds the user's chosen option index for each question (0-3).
	// -1 indicates an unanswered question.
	Answers   []int
	StartedAt time.Time
	EndedAt   time.Time
}

// NewSession creates a Session with the given ID and topic, recording the start time.
func NewSession(id, topic string) *Session {
	return &Session{
		ID:        id,
		Topic:     topic,
		StartedAt: time.Now(),
	}
}

// Score returns the number of correct answers and the total number of questions.
func (s *Session) Score() (correct, total int) {
	total = len(s.Questions)
	for i, q := range s.Questions {
		if i < len(s.Answers) && s.Answers[i] == q.CorrectIndex {
			correct++
		}
	}
	return
}

// Finish marks the session as complete.
func (s *Session) Finish() {
	s.EndedAt = time.Now()
}
