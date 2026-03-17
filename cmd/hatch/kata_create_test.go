package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// newKataCreateCmd — RunE closure
// ---------------------------------------------------------------------------

func TestNewKataCreateCmdRunE(t *testing.T) {
	// Calling RunE directly exercises the closure body; it fails because
	// the file flag is empty → os.ReadFile("") returns an error.
	cmd := newKataCreateCmd()
	_ = cmd.Flags().Set("topic", "go")
	_ = cmd.Flags().Set("file", "/nonexistent/kata.json")
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Error("expected error from RunE")
	}
}

// ---------------------------------------------------------------------------
// runKataCreate
// ---------------------------------------------------------------------------

func TestRunKataCreateSuccess(t *testing.T) {
	tempDBEnv(t)

	kata := map[string]any{
		"title":        "Hello World",
		"description":  "Write a hello world program",
		"language":     "go",
		"starter_code": "package main\n",
		"tests":        "func TestHello(t *testing.T) {}",
	}
	file := writeTempJSON(t, kata)

	if err := runKataCreate(context.Background(), "go", file); err != nil {
		t.Fatalf("runKataCreate: %v", err)
	}
}

func TestRunKataCreateFileNotFound(t *testing.T) {
	tempDBEnv(t)
	err := runKataCreate(context.Background(), "go", "/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestRunKataCreateInvalidJSON(t *testing.T) {
	tempDBEnv(t)
	f := filepath.Join(t.TempDir(), "bad.json")
	_ = os.WriteFile(f, []byte("not json"), 0o600)
	err := runKataCreate(context.Background(), "go", f)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestRunKataCreateMissingTitle(t *testing.T) {
	tempDBEnv(t)
	kata := map[string]any{
		"title": "", "description": "d", "language": "go",
		"starter_code": "x", "tests": "t",
	}
	file := writeTempJSON(t, kata)
	err := runKataCreate(context.Background(), "go", file)
	if err == nil {
		t.Error("expected error for missing title")
	}
}

func TestRunKataCreateMissingTests(t *testing.T) {
	tempDBEnv(t)
	kata := map[string]any{
		"title": "T", "description": "d", "language": "go",
		"starter_code": "x", "tests": "",
	}
	file := writeTempJSON(t, kata)
	err := runKataCreate(context.Background(), "go", file)
	if err == nil {
		t.Error("expected error for missing tests")
	}
}

func TestRunKataCreateUnsupportedLanguage(t *testing.T) {
	tempDBEnv(t)
	kata := map[string]any{
		"title": "T", "description": "d", "language": "cobol",
		"starter_code": "x", "tests": "t",
	}
	file := writeTempJSON(t, kata)
	err := runKataCreate(context.Background(), "go", file)
	if err == nil {
		t.Error("expected error for unsupported language")
	}
}

func TestRunKataCreateAllLanguages(t *testing.T) {
	for _, lang := range []string{"go", "python", "javascript", "java"} {
		t.Run(lang, func(t *testing.T) {
			tempDBEnv(t)
			kata := map[string]any{
				"title": "T", "description": "d", "language": lang,
				"starter_code": "x", "tests": "t",
			}
			file := writeTempJSON(t, kata)
			if err := runKataCreate(context.Background(), lang, file); err != nil {
				t.Fatalf("runKataCreate %s: %v", lang, err)
			}
		})
	}
}

func TestRunKataCreateSetupStoreError(t *testing.T) {
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
	kata := map[string]any{
		"title": "T", "description": "d", "language": "go",
		"starter_code": "x", "tests": "t",
	}
	file := writeTempJSON(t, kata)
	err := runKataCreate(context.Background(), "go", file)
	if err == nil {
		t.Error("expected error from setupStore when config is malformed")
	}
}

func TestRunKataCreateSaveKataError(t *testing.T) {
	// Cancelled context causes st.SaveKata to fail.
	tempDBEnv(t)
	kata := map[string]any{
		"title": "T", "description": "d", "language": "go",
		"starter_code": "x", "tests": "t",
	}
	file := writeTempJSON(t, kata)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := runKataCreate(ctx, "go", file)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}
