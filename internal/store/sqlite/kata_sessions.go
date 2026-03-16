package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/jazicorn/hatch/internal/kata"
)

// SaveKataSession persists a completed KataSession to SQLite.
func (s *Store) SaveKataSession(ctx context.Context, sess *kata.KataSession) error {
	var endedAt *time.Time
	if !sess.EndedAt.IsZero() {
		endedAt = &sess.EndedAt
	}

	passed := 0
	if sess.Passed {
		passed = 1
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO kata_sessions (id, topic, kata_id, language, passed, attempts, started_at, ended_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.Topic, sess.KataID, string(sess.Language), passed, sess.Attempts, sess.StartedAt, endedAt,
	)
	if err != nil {
		return fmt.Errorf("sqlite: save kata session: %w", err)
	}
	return nil
}
