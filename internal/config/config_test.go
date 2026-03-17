package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jazicorn/hatch/internal/config"
)

const errLoad = "Load: %v"

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf(errLoad, err)
	}
	if cfg.LLMProvider != "anthropic" {
		t.Errorf("LLMProvider: want anthropic, got %s", cfg.LLMProvider)
	}
	if cfg.EmbedProvider != "ollama" {
		t.Errorf("EmbedProvider: want ollama, got %s", cfg.EmbedProvider)
	}
	if cfg.SSHPort != 2222 {
		t.Errorf("SSHPort: want 2222, got %d", cfg.SSHPort)
	}
	if cfg.HTTPPort != 8080 {
		t.Errorf("HTTPPort: want 8080, got %d", cfg.HTTPPort)
	}
	if cfg.DBPath != "~/.hatch/hatch.db" {
		t.Errorf("DBPath: want ~/.hatch/hatch.db, got %s", cfg.DBPath)
	}
}

func TestValidateDefaults(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf(errLoad, err)
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("default config should be valid, got: %v", err)
	}
}

func TestValidateUnknownLLMProvider(t *testing.T) {
	cfg := &config.Config{LLMProvider: "unknown", EmbedProvider: "ollama", SSHPort: 2222, HTTPPort: 8080}
	if cfg.Validate() == nil {
		t.Error("expected error for unknown llm_provider, got nil")
	}
}

func TestValidateUnknownEmbedProvider(t *testing.T) {
	cfg := &config.Config{LLMProvider: "anthropic", EmbedProvider: "unknown", SSHPort: 2222, HTTPPort: 8080}
	if cfg.Validate() == nil {
		t.Error("expected error for unknown embed_provider, got nil")
	}
}

func TestValidatePortOutOfRange(t *testing.T) {
	cfg := &config.Config{LLMProvider: "anthropic", EmbedProvider: "ollama", SSHPort: 0, HTTPPort: 8080}
	if cfg.Validate() == nil {
		t.Error("expected error for ssh_port=0, got nil")
	}

	cfg.SSHPort = 2222
	cfg.HTTPPort = 99999
	if cfg.Validate() == nil {
		t.Error("expected error for http_port=99999, got nil")
	}
}

func TestValidateSourceValid(t *testing.T) {
	cfg := &config.Config{
		LLMProvider:   "anthropic",
		EmbedProvider: "ollama",
		SSHPort:       2222,
		HTTPPort:      8080,
		Sources: []config.SourceConfig{
			{Name: "docs", Path: "./docs", Type: "filesystem"},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config with source, got: %v", err)
	}
}

func TestValidateSourceMissingName(t *testing.T) {
	cfg := &config.Config{
		LLMProvider:   "anthropic",
		EmbedProvider: "ollama",
		SSHPort:       2222,
		HTTPPort:      8080,
		Sources: []config.SourceConfig{
			{Path: "./docs", Type: "filesystem"},
		},
	}
	if cfg.Validate() == nil {
		t.Error("expected error for source missing name")
	}
}

func TestValidateSourceMissingPath(t *testing.T) {
	cfg := &config.Config{
		LLMProvider:   "anthropic",
		EmbedProvider: "ollama",
		SSHPort:       2222,
		HTTPPort:      8080,
		Sources: []config.SourceConfig{
			{Name: "docs", Type: "filesystem"},
		},
	}
	if cfg.Validate() == nil {
		t.Error("expected error for source missing path")
	}
}

func TestValidateSourceUnknownType(t *testing.T) {
	cfg := &config.Config{
		LLMProvider:   "anthropic",
		EmbedProvider: "ollama",
		SSHPort:       2222,
		HTTPPort:      8080,
		Sources: []config.SourceConfig{
			{Name: "docs", Path: "./docs", Type: "s3"},
		},
	}
	if cfg.Validate() == nil {
		t.Error("expected error for unknown source type")
	}
}

func TestInit(t *testing.T) {
	// Init writes a new config or reports the file already exists. Both are non-error.
	if err := config.Init(); err != nil {
		t.Errorf("Init: %v", err)
	}
}

func TestInitWritesNewConfig(t *testing.T) {
	// Point HOME to a fresh temp dir so Init() must create the file from scratch.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := config.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".hatch", "config.yaml")); err != nil {
		t.Errorf("config file not created: %v", err)
	}
}

func TestInitAlreadyExists(t *testing.T) {
	// Second Init on the same dir should be a no-op (already-exists branch).
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := config.Init(); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	if err := config.Init(); err != nil {
		t.Fatalf("second Init: %v", err)
	}
}

func TestLoadReadInConfigError(t *testing.T) {
	// A malformed config.yaml triggers the non-FileNotFoundError branch in Load.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// Write invalid YAML that causes a parse error (not a FileNotFoundError).
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte("key: [unclosed"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := config.Load()
	if err == nil {
		t.Error("expected error for malformed config file")
	}
}

func TestInitMkdirError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping as root — chmod restrictions don't apply")
	}
	// Create a file named ".hatch" so MkdirAll fails (can't create dir over a file).
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// Create a regular file at the path where .hatch dir should be.
	if err := os.WriteFile(filepath.Join(tmp, ".hatch"), []byte("blocker"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := config.Init()
	if err == nil {
		t.Error("expected error when .hatch exists as a file")
	}
}

func TestInitWriteError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping as root — chmod restrictions don't apply")
	}
	// Create a read-only .hatch dir so WriteFile fails.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o500); err != nil { // read+exec only
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(hatchDir, 0o700) }) //nolint:errcheck
	err := config.Init()
	if err == nil {
		t.Error("expected error writing to read-only directory")
	}
}

func TestLoadEnvOverride(t *testing.T) {
	t.Setenv("HATCH_LLM_PROVIDER", "openai")
	t.Setenv("HATCH_EMBED_PROVIDER", "openai")
	t.Setenv("HATCH_SSH_PORT", "3333")
	t.Setenv("HATCH_HTTP_PORT", "9090")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf(errLoad, err)
	}
	if cfg.LLMProvider != "openai" {
		t.Errorf("LLMProvider: want openai, got %s", cfg.LLMProvider)
	}
	if cfg.EmbedProvider != "openai" {
		t.Errorf("EmbedProvider: want openai, got %s", cfg.EmbedProvider)
	}
	if cfg.SSHPort != 3333 {
		t.Errorf("SSHPort: want 3333, got %d", cfg.SSHPort)
	}
	if cfg.HTTPPort != 9090 {
		t.Errorf("HTTPPort: want 9090, got %d", cfg.HTTPPort)
	}
}
