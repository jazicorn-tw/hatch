package sqlite

// Tests covering error paths in kata_bank.go.

import (
	"context"
	"testing"

	"github.com/jazicorn/hatch/internal/kata"
)

// TestListKatasScanError covers the rows.Scan error (line 44-46) by inserting
// a row with a NULL value in a non-nullable column via a raw SQL bypass.
func TestListKatasScanError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	// Insert directly with NULL language to cause Scan to fail when scanning
	// into a string (NULL → string is unsupported in database/sql).
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO kata_bank (id, topic, title, description, starter_code, tests, language)
		VALUES ('k-null', 'go', 'T', 'D', '', '', NULL)
	`)
	if err != nil {
		// If the schema prevents NULL, skip this test.
		t.Skipf("cannot insert NULL language: %v", err)
	}

	_, err = s.ListKatas(ctx, "go")
	if err == nil {
		t.Error("expected error when Scan encounters NULL for string column")
	}
}

// TestListKatasRowsErrError covers the rows.Err() path (line 50-52) via a
// cancelled context during ListKatas after inserting a kata.
func TestListKatasRowsErrError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	k := &kata.Kata{ID: "k-rows", Topic: "go", Title: "T", Language: kata.Go}
	if err := s.SaveKata(ctx, k); err != nil {
		t.Fatal(err)
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	_, err := s.ListKatas(cancelCtx, "go")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}
