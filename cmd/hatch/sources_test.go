package main

import (
	"context"
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
	// Default config has no sources — should print a hint and return nil.
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
