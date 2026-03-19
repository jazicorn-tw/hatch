package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jazicorn/hatch/internal/quiz"
)

// marshalSessionOptions is injectable so tests can simulate json.Marshal failures.
var marshalSessionOptions = json.Marshal

// SaveSession persists a completed Session and all its questions to SQLite.
func (s *Store) SaveSession(ctx context.Context, sess *quiz.Session) error {
	correct, total := sess.Score()

	var endedAt *time.Time
	if !sess.EndedAt.IsZero() {
		endedAt = &sess.EndedAt
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: save session begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx,
		`INSERT OR REPLACE INTO quiz_sessions (id, topic, score, total, started_at, ended_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.Topic, correct, total, sess.StartedAt, endedAt,
	); err != nil {
		return fmt.Errorf("sqlite: insert session: %w", err)
	}

	for i, q := range sess.Questions {
		optsJSON, err := marshalSessionOptions(q.Options)
		if err != nil {
			return fmt.Errorf("sqlite: marshal options for question %s: %w", q.ID, err)
		}

		var userAnswer *int
		if i < len(sess.Answers) && sess.Answers[i] >= 0 {
			v := sess.Answers[i]
			userAnswer = &v
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT OR REPLACE INTO quiz_questions
			 (id, session_id, question_text, options, correct_index, explanation, user_answer)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			q.ID, sess.ID, q.Text, string(optsJSON),
			q.CorrectIndex, q.Explanation, userAnswer,
		); err != nil {
			return fmt.Errorf("sqlite: insert question %s: %w", q.ID, err)
		}
	}

	return tx.Commit()
}
