package sqlite

import (
	"context"
	"fmt"

	"github.com/jazicorn/hatch/internal/kata"
)

// SaveKata inserts a Sr-authored kata into the kata_bank.
// An existing row with the same ID is replaced.
func (s *Store) SaveKata(ctx context.Context, k *kata.Kata) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO kata_bank
		 (id, topic, title, description, starter_code, tests, language)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		k.ID, k.Topic, k.Title, k.Description, k.StarterCode, k.Tests, string(k.Language),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save kata: %w", err)
	}
	return nil
}

// ListKatas returns all Sr-authored katas for the given topic,
// ordered by insertion time.
func (s *Store) ListKatas(ctx context.Context, topic string) ([]*kata.Kata, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, topic, title, description, starter_code, tests, language
		 FROM kata_bank
		 WHERE topic = ?
		 ORDER BY created_at`,
		topic,
	)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list katas: %w", err)
	}
	defer rows.Close()

	var katas []*kata.Kata
	for rows.Next() {
		k := &kata.Kata{}
		var lang string
		if err := rows.Scan(&k.ID, &k.Topic, &k.Title, &k.Description, &k.StarterCode, &k.Tests, &lang); err != nil {
			return nil, fmt.Errorf("sqlite: scan kata: %w", err)
		}
		k.Language = kata.Language(lang)
		katas = append(katas, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: kata bank rows: %w", err)
	}
	return katas, nil
}
