package sqlite

// Tests covering error paths in sessions.go.

import (
	"context"
	"fmt"
	"testing"

	"github.com/jazicorn/hatch/internal/quiz"
)

// TestSaveSessionInsertError covers the ExecContext error for session INSERT
// (line 31-33) by using PRAGMA query_only to prevent writes.
func TestSaveSessionInsertError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	if _, err := s.db.Exec("PRAGMA query_only = ON"); err != nil {
		t.Fatal(err)
	}

	sess := quiz.NewSession("sess-err", "go")
	err := s.SaveSession(ctx, sess)
	if err == nil {
		t.Error("expected error when session INSERT fails due to read-only DB")
	}
}

// TestSaveSessionQuestionInsertError covers the ExecContext error for question
// INSERT (line 53-55) using a BEFORE INSERT trigger on quiz_questions.
func TestSaveSessionQuestionInsertError(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	_, err := s.db.ExecContext(ctx, `
		CREATE TRIGGER fail_quiz_questions_insert
		BEFORE INSERT ON quiz_questions
		BEGIN
			SELECT RAISE(FAIL, 'forced failure');
		END
	`)
	if err != nil {
		t.Fatal(err)
	}

	sess := quiz.NewSession("sess-q-err", "go")
	sess.Questions = []quiz.Question{
		{ID: "q1", Text: "Q?", Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 0},
	}
	sess.Answers = []int{0}

	if err := s.SaveSession(ctx, sess); err == nil {
		t.Error("expected error when question INSERT fails due to trigger")
	}
}

// TestSaveSessionMarshalError covers the json.Marshal error path (line 37-39)
// by overriding marshalSessionOptions.
func TestSaveSessionMarshalError(t *testing.T) {
	orig := marshalSessionOptions
	marshalSessionOptions = func(_ any) ([]byte, error) {
		return nil, fmt.Errorf("forced marshal error")
	}
	defer func() { marshalSessionOptions = orig }()

	s := openTestStore(t)
	ctx := context.Background()

	sess := quiz.NewSession("sess-marshal", "go")
	sess.Questions = []quiz.Question{
		{ID: "q1", Text: "Q?", Options: [4]string{"a", "b", "c", "d"}, CorrectIndex: 0},
	}

	if err := s.SaveSession(ctx, sess); err == nil {
		t.Error("expected error when json.Marshal fails for session options")
	}
}
