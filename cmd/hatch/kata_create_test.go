package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

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
