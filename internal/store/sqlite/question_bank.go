package sqlite

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jazicorn/hatch/internal/quiz"
)

// SaveQuestionBank inserts a batch of Sr-authored questions for the given topic.
// Existing rows with the same ID are replaced.
func (s *Store) SaveQuestionBank(ctx context.Context, topic string, questions []quiz.Question) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: save question bank begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	for _, q := range questions {
		optsJSON, err := json.Marshal(q.Options)
		if err != nil {
			return fmt.Errorf("sqlite: marshal options for question %s: %w", q.ID, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT OR REPLACE INTO question_bank
			 (id, topic, question_text, options, correct_index, explanation)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			q.ID, topic, q.Text, string(optsJSON), q.CorrectIndex, q.Explanation,
		); err != nil {
			return fmt.Errorf("sqlite: insert question %s: %w", q.ID, err)
		}
	}
	return tx.Commit()
}

// ListQuestionBank returns all Sr-authored questions for the given topic,
// ordered by insertion time.
func (s *Store) ListQuestionBank(ctx context.Context, topic string) ([]quiz.Question, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, question_text, options, correct_index, explanation
		 FROM question_bank
		 WHERE topic = ?
		 ORDER BY created_at`,
		topic,
	)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list question bank: %w", err)
	}
	defer rows.Close()

	var questions []quiz.Question
	for rows.Next() {
		var q quiz.Question
		var optsJSON string
		if err := rows.Scan(&q.ID, &q.Text, &optsJSON, &q.CorrectIndex, &q.Explanation); err != nil {
			return nil, fmt.Errorf("sqlite: scan question: %w", err)
		}
		var optSlice []string
		if err := json.Unmarshal([]byte(optsJSON), &optSlice); err != nil {
			return nil, fmt.Errorf("sqlite: unmarshal options for %s: %w", q.ID, err)
		}
		if len(optSlice) != 4 {
			return nil, fmt.Errorf("sqlite: question %s has %d options, expected 4", q.ID, len(optSlice))
		}
		copy(q.Options[:], optSlice)
		questions = append(questions, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: question bank rows: %w", err)
	}
	return questions, nil
}
