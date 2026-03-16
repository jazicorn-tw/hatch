package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/jazicorn/hatch/internal/config"
	"github.com/jazicorn/hatch/internal/kata"
	"github.com/jazicorn/hatch/internal/kata/sandbox"
	"github.com/jazicorn/hatch/internal/store/sqlite"
)

func newKataCmd() *cobra.Command {
	var topic string

	cmd := &cobra.Command{
		Use:   "kata",
		Short: "Solve an AI-generated code kata",
		Long:  "Generates a code kata from ingested content and opens an in-terminal editor for your solution.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKata(cmd.Context(), topic)
		},
	}
	cmd.Flags().StringVar(&topic, "topic", "", "Topic to focus on (optional — omit to generate a general kata)")
	return cmd
}

func runKata(ctx context.Context, topic string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("kata: load config: %w", err)
	}

	emb, err := newEmbedder(cfg)
	if err != nil {
		return fmt.Errorf("kata: create embedder: %w", err)
	}

	completer, err := newLLMCompleter(cfg)
	if err != nil {
		return fmt.Errorf("kata: create llm: %w", err)
	}

	dbPath, err := resolveDBPath(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("kata: resolve db path: %w", err)
	}
	st, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("kata: open store: %w", err)
	}
	defer st.Close()

	label := topic
	if label == "" {
		label = "general"
	}

	fmt.Fprintf(os.Stderr, "Generating kata on %q…\n", label)
	gen := kata.NewGenerator(emb, st, completer, kata.GeneratorConfig{TopK: 10})
	k, err := gen.Generate(ctx, label)
	if err != nil {
		return fmt.Errorf("kata: generate: %w", err)
	}

	sess := &kata.KataSession{
		ID:        uuid.New().String(),
		Topic:     label,
		KataID:    k.ID,
		Language:  k.Language,
		StartedAt: time.Now(),
	}

	fmt.Printf("\n=== Kata: %s [%s] ===\n\n", k.Title, k.Language)
	fmt.Println(k.Description)
	fmt.Println()

	attempt := 0
	for {
		attempt++
		sess.Attempts = attempt

		// Open the in-TUI editor pre-filled with starter code (or previous attempt).
		solution, submitted := runEditor(k.StarterCode)
		if !submitted {
			fmt.Println("Kata cancelled.")
			break
		}

		fmt.Fprintln(os.Stderr, "Running tests…")
		result, err := sandbox.Run(ctx, *k, solution, sandbox.Config{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "sandbox error: %v\n", err)
		}

		fmt.Println()
		fmt.Println(strings.TrimSpace(result.Output))
		fmt.Printf("\nTime: %v\n", result.Duration.Round(time.Millisecond))

		if result.Passed {
			fmt.Printf("\n✓ All tests passed! (attempt %d)\n", attempt)
			sess.Passed = true
			// Pre-fill subsequent re-runs with the passing solution.
			k.StarterCode = solution
			break
		}

		fmt.Printf("\n✗ Tests failed. (attempt %d)\n", attempt)
		fmt.Print("Try again? [y/N]: ")
		var yn string
		fmt.Scanln(&yn) //nolint:errcheck
		if !strings.EqualFold(strings.TrimSpace(yn), "y") {
			break
		}
		// Pre-fill editor with the last attempt.
		k.StarterCode = solution
	}

	sess.EndedAt = time.Now()
	if err := st.SaveKataSession(ctx, sess); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save kata session: %v\n", err)
	}
	return nil
}

// editorModel is a minimal Bubble Tea program wrapping a textarea.
type editorModel struct {
	ta        textarea.Model
	submitted bool
	quit      bool
}

func initialEditorModel(content string) editorModel {
	ta := textarea.New()
	ta.SetValue(content)
	ta.Focus()
	ta.ShowLineNumbers = true
	ta.CharLimit = 0 // unlimited
	ta.SetWidth(100)
	ta.SetHeight(30)
	return editorModel{ta: ta}
}

func (m editorModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m editorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlS:
			m.submitted = true
			return m, tea.Quit
		case tea.KeyEsc, tea.KeyCtrlC:
			m.quit = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	return m, cmd
}

func (m editorModel) View() string {
	return fmt.Sprintf(
		"  Edit your solution. Ctrl+S to submit • Esc/Ctrl+C to cancel\n\n%s\n",
		m.ta.View(),
	)
}

// runEditor opens the Bubble Tea editor and returns (solution, submitted).
func runEditor(starter string) (string, bool) {
	m := initialEditorModel(starter)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return starter, false
	}
	final, ok := result.(editorModel)
	if !ok || final.quit {
		return starter, false
	}
	return final.ta.Value(), final.submitted
}
