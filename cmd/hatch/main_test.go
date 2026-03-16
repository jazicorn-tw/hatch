package main

import (
	"testing"
)

func TestNewRootCmdUse(t *testing.T) {
	cmd := newRootCmd()
	if cmd.Use != "hatch" {
		t.Errorf("expected Use=hatch, got %s", cmd.Use)
	}
}

func TestNewRootCmdSubcommands(t *testing.T) {
	cmd := newRootCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Use] = true
	}
	for _, want := range []string{"config", "ingest", "sources", "quiz", "kata"} {
		if !names[want] {
			t.Errorf("expected subcommand %q to be registered", want)
		}
	}
}

func TestNewConfigCmdUse(t *testing.T) {
	cmd := newConfigCmd()
	if cmd.Use != "config" {
		t.Errorf("expected Use=config, got %s", cmd.Use)
	}
}

func TestNewConfigCmdHasInitSubcommand(t *testing.T) {
	cmd := newConfigCmd()
	for _, sub := range cmd.Commands() {
		if sub.Use == "init" {
			return
		}
	}
	t.Error("expected 'init' subcommand under config")
}

func TestNewConfigInitCmdUse(t *testing.T) {
	cmd := newConfigInitCmd()
	if cmd.Use != "init" {
		t.Errorf("expected Use=init, got %s", cmd.Use)
	}
}

func TestConfigInitCmdRunE(t *testing.T) {
	cmd := newConfigInitCmd()
	// RunE calls config.Init() which writes or skips an existing config — always non-error.
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Errorf("config init RunE: %v", err)
	}
}
