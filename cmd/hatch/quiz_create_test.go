package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeTempJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	f := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(f, data, 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return f
}

func tempDBEnv(t *testing.T) {
	t.Helper()
	t.Setenv("HATCH_DB_PATH", filepath.Join(t.TempDir(), "test.db"))
}

// ---------------------------------------------------------------------------
// runQuizCreate
// ---------------------------------------------------------------------------

func TestRunQuizCreateSuccess(t *testing.T) {
	tempDBEnv(t)

	questions := []map[string]any{
		{
			"text":          "What is Go?",
			"options":       []string{"A language", "A game", "A city", "A food"},
			"correct_index": 0,
			"explanation":   "Go is a programming language",
		},
	}
	file := writeTempJSON(t, questions)

	if err := runQuizCreate(context.Background(), "go", file); err != nil {
		t.Fatalf("runQuizCreate: %v", err)
	}
}

func TestRunQuizCreateFileNotFound(t *testing.T) {
	tempDBEnv(t)
	err := runQuizCreate(context.Background(), "go", "/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestRunQuizCreateInvalidJSON(t *testing.T) {
	tempDBEnv(t)
	f := filepath.Join(t.TempDir(), "bad.json")
	_ = os.WriteFile(f, []byte("not json"), 0o600)
	err := runQuizCreate(context.Background(), "go", f)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestRunQuizCreateEmptyFile(t *testing.T) {
	tempDBEnv(t)
	file := writeTempJSON(t, []any{})
	err := runQuizCreate(context.Background(), "go", file)
	if err == nil {
		t.Error("expected error for empty questions array")
	}
}

func TestRunQuizCreateMissingText(t *testing.T) {
	tempDBEnv(t)
	questions := []map[string]any{
		{"text": "", "options": []string{"a", "b", "c", "d"}, "correct_index": 0},
	}
	file := writeTempJSON(t, questions)
	err := runQuizCreate(context.Background(), "go", file)
	if err == nil {
		t.Error("expected error for empty question text")
	}
}

func TestRunQuizCreateWrongOptionCount(t *testing.T) {
	tempDBEnv(t)
	questions := []map[string]any{
		{"text": "Q?", "options": []string{"a", "b", "c"}, "correct_index": 0},
	}
	file := writeTempJSON(t, questions)
	err := runQuizCreate(context.Background(), "go", file)
	if err == nil {
		t.Error("expected error for wrong option count")
	}
}

func TestRunQuizCreateIndexOutOfRange(t *testing.T) {
	tempDBEnv(t)
	questions := []map[string]any{
		{"text": "Q?", "options": []string{"a", "b", "c", "d"}, "correct_index": 5},
	}
	file := writeTempJSON(t, questions)
	err := runQuizCreate(context.Background(), "go", file)
	if err == nil {
		t.Error("expected error for out-of-range correct_index")
	}
}

// ---------------------------------------------------------------------------
// newQuizCreateCmd — RunE closure
// ---------------------------------------------------------------------------

func TestNewQuizCreateCmdRunE(t *testing.T) {
	// Calling RunE directly exercises the closure body.
	cmd := newQuizCreateCmd()
	_ = cmd.Flags().Set("topic", "go")
	_ = cmd.Flags().Set("file", "/nonexistent/questions.json")
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Error("expected error from RunE")
	}
}

// ---------------------------------------------------------------------------
// runQuizCreate — setupStore and save error paths
// ---------------------------------------------------------------------------

func TestRunQuizCreateSetupStoreError(t *testing.T) {
	// Malformed config causes setupStore() to fail.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte("key: [unclosed"), 0o600); err != nil {
		t.Fatal(err)
	}
	questions := []map[string]any{
		{"text": "Q?", "options": []string{"a", "b", "c", "d"}, "correct_index": 0, "explanation": "e"},
	}
	file := writeTempJSON(t, questions)
	err := runQuizCreate(context.Background(), "go", file)
	if err == nil {
		t.Error("expected error from setupStore when config is malformed")
	}
}

func TestRunQuizCreateSaveError(t *testing.T) {
	// Cancelled context causes st.SaveQuestionBank to fail.
	tempDBEnv(t)
	questions := []map[string]any{
		{"text": "Q?", "options": []string{"a", "b", "c", "d"}, "correct_index": 0, "explanation": "e"},
	}
	file := writeTempJSON(t, questions)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := runQuizCreate(ctx, "go", file)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}
