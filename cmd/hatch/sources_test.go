package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// command structure
// ---------------------------------------------------------------------------

func TestNewSourcesCmdUse(t *testing.T) {
	cmd := newSourcesCmd()
	if cmd.Use != "sources" {
		t.Errorf("expected Use=sources, got %s", cmd.Use)
	}
}

func TestNewSourcesCmdSubcommands(t *testing.T) {
	cmd := newSourcesCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Use] = true
	}
	for _, want := range []string{"list", "remove"} {
		if !names[want] {
			t.Errorf("expected subcommand %q registered under sources", want)
		}
	}
}

func TestNewSourcesListCmdUse(t *testing.T) {
	cmd := newSourcesListCmd()
	if cmd.Use != "list" {
		t.Errorf("expected Use=list, got %s", cmd.Use)
	}
}

func TestNewSourcesRemoveCmdUse(t *testing.T) {
	cmd := newSourcesRemoveCmd()
	if cmd.Use != "remove" {
		t.Errorf("expected Use=remove, got %s", cmd.Use)
	}
}

// ---------------------------------------------------------------------------
// newSourcesListCmd RunE — no sources configured
// ---------------------------------------------------------------------------

func TestSourcesListNoSources(t *testing.T) {
	// Use a temp HOME with an empty config so cfg.Sources is empty.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	configYAML := "llm_provider: anthropic\nembed_provider: ollama\nssh_port: 2222\nhttp_port: 8080\nsources: []\n"
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte(configYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := newSourcesListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Errorf("sources list with no sources: %v", err)
	}
}

// ---------------------------------------------------------------------------
// runSourcesRemove — source not found
// ---------------------------------------------------------------------------

func TestRunSourcesRemoveNotFound(t *testing.T) {
	// Default config has no sources, so any name lookup returns "not found".
	err := runSourcesRemove(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error when source is not found")
	}
}

func TestRunSourcesRemoveEmptyName(t *testing.T) {
	err := runSourcesRemove(context.Background(), "")
	if err == nil {
		t.Error("expected error when source name is empty and no sources configured")
	}
}

func TestSourcesListWithSources(t *testing.T) {
	// Use a temp HOME with a config that has sources, so the tabwriter path is covered.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	configYAML := "llm_provider: anthropic\nembed_provider: ollama\nssh_port: 2222\nhttp_port: 8080\nsources:\n  - name: docs\n    path: ./docs\n    type: filesystem\n"
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte(configYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := newSourcesListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Errorf("sources list with sources: %v", err)
	}
}

func TestRunSourcesRemoveSuccess(t *testing.T) {
	// Use a fresh HOME dir with a config that has a source, so the full
	// runSourcesRemove path (find source → open DB → delete → rewrite config) runs.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// Use empty db_path so runSourcesRemove expands to HOME/.hatch/hatch.db.
	configYAML := "llm_provider: anthropic\nembed_provider: ollama\nssh_port: 2222\nhttp_port: 8080\ndb_path: \"\"\nsources:\n  - name: docs\n    path: ./docs\n    type: filesystem\n"
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte(configYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	err := runSourcesRemove(context.Background(), "docs")
	if err != nil {
		t.Errorf("runSourcesRemove: %v", err)
	}
}

func TestRunSourcesRemoveStoreOpenError(t *testing.T) {
	// Use a HOME where the config has a source but the DBPath points somewhere
	// that can't be opened (non-existent deep directory).
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	configYAML := `llm_provider: anthropic
embed_provider: ollama
ssh_port: 2222
http_port: 8080
db_path: /nonexistent/deeply/nested/path/hatch.db
sources:
  - name: docs
    path: ./docs
    type: filesystem
`
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte(configYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	err := runSourcesRemove(context.Background(), "docs")
	if err == nil {
		t.Error("expected error when DB path is non-existent")
	}
}

// ---------------------------------------------------------------------------
// newSourcesListCmd — config load error path
// ---------------------------------------------------------------------------

func TestSourcesListConfigLoadError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte("key: [unclosed"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := newSourcesListCmd()
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Error("expected error for malformed config")
	}
}

// ---------------------------------------------------------------------------
// newSourcesRemoveCmd — RunE closure
// ---------------------------------------------------------------------------

func TestNewSourcesRemoveCmdRunE(t *testing.T) {
	// Calling RunE directly exercises the closure body; source not found → error.
	cmd := newSourcesRemoveCmd()
	_ = cmd.Flags().Set("name", "nonexistent")
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Error("expected error when source not found via RunE")
	}
}

// ---------------------------------------------------------------------------
// runSourcesRemove — config load error, delete error, write config error
// ---------------------------------------------------------------------------

func TestRunSourcesRemoveConfigLoadError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte("key: [unclosed"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := runSourcesRemove(context.Background(), "docs")
	if err == nil {
		t.Error("expected error for malformed config")
	}
}

func TestRunSourcesRemoveDeleteError(t *testing.T) {
	// Cancelled context causes DeleteBySource to fail.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("HATCH_DB_PATH", filepath.Join(tmp, "test.db"))
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	configYAML := "llm_provider: anthropic\nembed_provider: ollama\nssh_port: 2222\nhttp_port: 8080\nsources:\n  - name: docs\n    path: ./docs\n    type: filesystem\n"
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte(configYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := runSourcesRemove(ctx, "docs")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestRunSourcesRemoveWriteConfigError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping as root — file permissions don't restrict root")
	}
	// config.yaml is read-only: ReadInConfig succeeds, WriteConfig fails.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("HATCH_DB_PATH", filepath.Join(tmp, "test.db"))
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	configYAML := "llm_provider: anthropic\nembed_provider: ollama\nssh_port: 2222\nhttp_port: 8080\nsources:\n  - name: docs\n    path: ./docs\n    type: filesystem\n"
	configPath := filepath.Join(hatchDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(configPath, 0o600) }) //nolint:errcheck
	err := runSourcesRemove(context.Background(), "docs")
	if err == nil {
		t.Error("expected error when config file is read-only")
	}
}
