package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jazicorn/hatch/internal/config"
)

// ---------------------------------------------------------------------------
// setupDeps
// ---------------------------------------------------------------------------

func TestSetupDepsConfigLoadError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	orig, set := os.LookupEnv("HATCH_DB_PATH")
	os.Unsetenv("HATCH_DB_PATH")
	if set {
		defer os.Setenv("HATCH_DB_PATH", orig)
	} else {
		defer os.Unsetenv("HATCH_DB_PATH")
	}
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte("key: [unclosed"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := setupDeps()
	if err == nil {
		t.Error("expected error for malformed config")
	}
}

func TestSetupDepsEmbedderError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("GEMINI_API_KEY", "")
	orig, set := os.LookupEnv("HATCH_DB_PATH")
	os.Unsetenv("HATCH_DB_PATH")
	if set {
		defer os.Setenv("HATCH_DB_PATH", orig)
	} else {
		defer os.Unsetenv("HATCH_DB_PATH")
	}
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := "llm_provider: anthropic\nanthropic_api_key: test-key\nembed_provider: gemini\ngemini_api_key: \"\"\n"
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := setupDeps()
	if err == nil {
		t.Error("expected error when gemini embed key missing")
	}
}

func TestSetupDepsLLMError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("GEMINI_API_KEY", "")
	orig, set := os.LookupEnv("HATCH_DB_PATH")
	os.Unsetenv("HATCH_DB_PATH")
	if set {
		defer os.Setenv("HATCH_DB_PATH", orig)
	} else {
		defer os.Unsetenv("HATCH_DB_PATH")
	}
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := "llm_provider: gemini\ngemini_api_key: \"\"\nembed_provider: ollama\n"
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := setupDeps()
	if err == nil {
		t.Error("expected error when gemini llm key missing")
	}
}

func TestSetupDepsResolveDBPathError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	orig, set := os.LookupEnv("HATCH_DB_PATH")
	os.Unsetenv("HATCH_DB_PATH")
	if set {
		defer os.Setenv("HATCH_DB_PATH", orig)
	} else {
		defer os.Unsetenv("HATCH_DB_PATH")
	}
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := "llm_provider: anthropic\nanthropic_api_key: test-key\nembed_provider: ollama\ndb_path: /dev/null/sub/hatch.db\n"
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := setupDeps()
	if err == nil {
		t.Error("expected error for bad db path")
	}
}

func TestSetupDepsOpenStoreError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	orig, set := os.LookupEnv("HATCH_DB_PATH")
	os.Unsetenv("HATCH_DB_PATH")
	if set {
		defer os.Setenv("HATCH_DB_PATH", orig)
	} else {
		defer os.Unsetenv("HATCH_DB_PATH")
	}
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// Create a directory where the db file would go so sqlite.Open fails.
	dbDir := filepath.Join(hatchDir, "hatch.db")
	if err := os.MkdirAll(dbDir, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := "llm_provider: anthropic\nanthropic_api_key: test-key\nembed_provider: ollama\ndb_path: \"\"\n"
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := setupDeps()
	if err == nil {
		t.Error("expected error when db path is a directory")
	}
}

func TestSetupDepsSuccess(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	orig, set := os.LookupEnv("HATCH_DB_PATH")
	os.Unsetenv("HATCH_DB_PATH")
	if set {
		defer os.Setenv("HATCH_DB_PATH", orig)
	} else {
		defer os.Unsetenv("HATCH_DB_PATH")
	}
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := "llm_provider: anthropic\nanthropic_api_key: test-key\nembed_provider: ollama\ndb_path: \"\"\n"
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	d, err := setupDeps()
	if err != nil {
		t.Fatalf("setupDeps: %v", err)
	}
	d.store.Close()
}

// ---------------------------------------------------------------------------
// newLLMCompleter
// ---------------------------------------------------------------------------

func TestNewLLMCompleterAnthropicWithKey(t *testing.T) {
	cfg := &config.Config{LLMProvider: "anthropic", AnthropicAPIKey: "test-key"}
	c, err := newLLMCompleter(cfg)
	if err != nil {
		t.Fatalf("newLLMCompleter anthropic: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil completer")
	}
}

func TestNewLLMCompleterAnthropicNoKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	cfg := &config.Config{LLMProvider: "anthropic", AnthropicAPIKey: ""}
	_, err := newLLMCompleter(cfg)
	if err == nil {
		t.Error("expected error when no anthropic key provided")
	}
}

func TestNewLLMCompleterGeminiWithKey(t *testing.T) {
	cfg := &config.Config{LLMProvider: "gemini", GeminiAPIKey: "test-key"}
	c, err := newLLMCompleter(cfg)
	if err != nil {
		t.Fatalf("newLLMCompleter gemini: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil completer")
	}
}

func TestNewLLMCompleterGeminiNoKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	cfg := &config.Config{LLMProvider: "gemini", GeminiAPIKey: ""}
	_, err := newLLMCompleter(cfg)
	if err == nil {
		t.Error("expected error when no gemini key provided")
	}
}

func TestNewLLMCompleterDefaultIsAnthropic(t *testing.T) {
	// Empty LLMProvider falls through to anthropic default.
	cfg := &config.Config{LLMProvider: "", AnthropicAPIKey: "test-key"}
	c, err := newLLMCompleter(cfg)
	if err != nil {
		t.Fatalf("newLLMCompleter default: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil completer")
	}
}

// ---------------------------------------------------------------------------
// setupStore
// ---------------------------------------------------------------------------

func TestSetupStore(t *testing.T) {
	t.Setenv("HATCH_DB_PATH", filepath.Join(t.TempDir(), "test.db"))
	st, err := setupStore()
	if err != nil {
		t.Fatalf("setupStore: %v", err)
	}
	defer st.Close()
}

func TestSetupStoreConfigLoadError(t *testing.T) {
	// Malformed config.yaml causes config.Load() to fail.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte("key: [unclosed"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := setupStore()
	if err == nil {
		t.Error("expected error when config is malformed")
	}
}

func TestSetupStoreResolveDBPathError(t *testing.T) {
	// Valid config but db_path points to a non-creatable directory.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// Unset HATCH_DB_PATH so the config file's db_path is used.
	orig, set := os.LookupEnv("HATCH_DB_PATH")
	os.Unsetenv("HATCH_DB_PATH")
	if set {
		defer os.Setenv("HATCH_DB_PATH", orig)
	} else {
		defer os.Unsetenv("HATCH_DB_PATH")
	}
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// /dev/null is a character device — MkdirAll on a subpath fails.
	configYAML := "llm_provider: anthropic\nembed_provider: ollama\nssh_port: 2222\nhttp_port: 8080\ndb_path: /dev/null/sub/hatch.db\n"
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte(configYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := setupStore()
	if err == nil {
		t.Error("expected error when db directory cannot be created")
	}
}
