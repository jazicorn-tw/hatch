package sqlite

// Tests covering error paths in question_bank.go.

import (
	"context"
	"fmt"
	"testing"

	"github.com/jazicorn/hatch/internal/quiz"
)

// TestSaveQuestionBankInsertError covers the ExecContext error for question
// INSERT (line 30-32) using a trigger that fails on question_bank inserts.
func TestSaveQuestionBankInsertError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	_, err := s.db.ExecContext(ctx, `
		CREATE TRIGGER fail_question_bank_insert
		BEFORE INSERT ON question_bank
		BEGIN
			SELECT RAISE(FAIL, 'forced failure');
		END
	`)
	if err != nil {
		t.Fatal(err)
	}

	err = s.SaveQuestionBank(ctx, "go", []quiz.Question{
		{ID: "q1", Text: "Q?", Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 0},
	})
	if err == nil {
		t.Error("expected error when question INSERT fails due to trigger")
	}
}

// TestListQuestionBankScanError covers the rows.Scan error (line 56-58)
// by inserting a row with NULL in a non-nullable column.
func TestListQuestionBankScanError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO question_bank (id, topic, question_text, options, correct_index, explanation)
		VALUES ('q-null', 'go', NULL, '["a","b","c","d"]', 0, '')
	`)
	if err != nil {
		t.Skipf("cannot insert NULL question_text: %v", err)
	}

	_, err = s.ListQuestionBank(ctx, "go")
	if err == nil {
		t.Error("expected error when Scan encounters NULL for string column")
	}
}

// TestListQuestionBankBadJSON covers the json.Unmarshal error (line 60-62)
// by inserting a row with invalid JSON in the options column.
func TestListQuestionBankBadJSON(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO question_bank (id, topic, question_text, options, correct_index, explanation)
		VALUES ('q-bad-json', 'go', 'Q?', 'not-json', 0, '')
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.ListQuestionBank(ctx, "go")
	if err == nil {
		t.Error("expected error when options column contains invalid JSON")
	}
}

// TestListQuestionBankWrongOptionsCount covers the wrong options count error
// (line 63-65) by inserting a row with fewer than 4 options.
func TestListQuestionBankWrongOptionsCount(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO question_bank (id, topic, question_text, options, correct_index, explanation)
		VALUES ('q-3opts', 'go', 'Q?', '["a","b","c"]', 0, '')
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.ListQuestionBank(ctx, "go")
	if err == nil {
		t.Error("expected error when options array has 3 elements instead of 4")
	}
}

// TestListQuestionBankRowsErrError covers the rows.Err() path (line 69-71)
// via a cancelled context.
func TestListQuestionBankRowsErrError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	if err := s.SaveQuestionBank(ctx, "go", []quiz.Question{
		{ID: "qrows", Text: "Q?", Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 0},
	}); err != nil {
		t.Fatal(err)
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	_, err := s.ListQuestionBank(cancelCtx, "go")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

// TestSaveQuestionBankMarshalError covers the json.Marshal error path (line 22-24)
// by overriding marshalQuestionOptions.
func TestSaveQuestionBankMarshalError(t *testing.T) {
	orig := marshalQuestionOptions
	marshalQuestionOptions = func(_ any) ([]byte, error) {
		return nil, fmt.Errorf("forced marshal error")
	}
	defer func() { marshalQuestionOptions = orig }()

	s := openTestStore(t)
	ctx := context.Background()

	err := s.SaveQuestionBank(ctx, "go", []quiz.Question{
		{ID: "q1", Text: "Q?", Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 0},
	})
	if err == nil {
		t.Error("expected error when json.Marshal fails for question options")
	}
}
