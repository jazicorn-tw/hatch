package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ---------------------------------------------------------------------------
// initialEditorModel
// ---------------------------------------------------------------------------

func TestInitialEditorModel(t *testing.T) {
	m := initialEditorModel("package main\n")
	if m.submitted {
		t.Error("expected submitted=false")
	}
	if m.quit {
		t.Error("expected quit=false")
	}
}

// ---------------------------------------------------------------------------
// editorModel.Init
// ---------------------------------------------------------------------------

func TestEditorModelInit(t *testing.T) {
	m := initialEditorModel("package main\n")
	cmd := m.Init()
	// Init returns textarea.Blink which is non-nil.
	if cmd == nil {
		t.Error("expected non-nil Init command")
	}
}

// ---------------------------------------------------------------------------
// editorModel.Update
// ---------------------------------------------------------------------------

func TestEditorModelUpdateCtrlS(t *testing.T) {
	m := initialEditorModel("code")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	final, ok := updated.(editorModel)
	if !ok {
		t.Fatal("expected editorModel back from Update")
	}
	if !final.submitted {
		t.Error("expected submitted=true after Ctrl+S")
	}
}

func TestEditorModelUpdateEsc(t *testing.T) {
	m := initialEditorModel("code")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	final, ok := updated.(editorModel)
	if !ok {
		t.Fatal("expected editorModel back from Update")
	}
	if !final.quit {
		t.Error("expected quit=true after Esc")
	}
}

func TestEditorModelUpdateCtrlC(t *testing.T) {
	m := initialEditorModel("code")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	final, ok := updated.(editorModel)
	if !ok {
		t.Fatal("expected editorModel back from Update")
	}
	if !final.quit {
		t.Error("expected quit=true after Ctrl+C")
	}
}

func TestEditorModelUpdateOtherKey(t *testing.T) {
	m := initialEditorModel("code")
	// A non-special key falls through to m.ta.Update.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	_, ok := updated.(editorModel)
	if !ok {
		t.Fatal("expected editorModel back from Update")
	}
}

// ---------------------------------------------------------------------------
// editorModel.View
// ---------------------------------------------------------------------------

func TestEditorModelView(t *testing.T) {
	m := initialEditorModel("package main\n")
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
	if !strings.Contains(view, "Ctrl+S") {
		t.Error("expected view to contain Ctrl+S instruction")
	}
}

// ---------------------------------------------------------------------------
// runEditor (no terminal — p.Run fails, returns starter unchanged)
// ---------------------------------------------------------------------------

func TestRunEditorNoTerminal(t *testing.T) {
	// Without a real terminal, tea.Program.Run returns an error, so runEditor
	// falls back to returning the starter with submitted=false.
	starter := "package main\n"
	solution, submitted := runEditor(starter)
	if submitted {
		// If somehow BubbleTea ran without error, that's also acceptable.
		t.Log("BubbleTea ran without error in non-terminal environment")
	} else {
		if solution != starter {
			t.Errorf("expected unchanged starter=%q, got %q", starter, solution)
		}
	}
}

// ---------------------------------------------------------------------------
// newKataCmd structure
// ---------------------------------------------------------------------------

func TestNewKataCmdUse(t *testing.T) {
	cmd := newKataCmd()
	if cmd.Use != "kata" {
		t.Errorf("expected Use=kata, got %s", cmd.Use)
	}
}

func TestNewKataCmdHasCreateSubcommand(t *testing.T) {
	cmd := newKataCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Use] = true
	}
	if !names["create"] {
		t.Error("expected 'create' subcommand registered under kata")
	}
}

func TestNewKataCmdRunE(t *testing.T) {
	// Trigger the RunE closure body by calling it with a malformed config so
	// setupDeps() returns an error immediately.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, ".hatch")
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hatchDir, "config.yaml"), []byte("key: [unclosed"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := newKataCmd()
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Error("expected error from RunE when config is malformed")
	}
}
