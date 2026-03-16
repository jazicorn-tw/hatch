package config_test

import (
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
