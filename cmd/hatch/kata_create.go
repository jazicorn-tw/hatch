package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/jazicorn/hatch/internal/config"
	"github.com/jazicorn/hatch/internal/kata"
	"github.com/jazicorn/hatch/internal/store/sqlite"
)

// rawImportKata is the JSON file format for Sr-provided katas.
type rawImportKata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Language    string `json:"language"`
	StarterCode string `json:"starter_code"`
	Tests       string `json:"tests"`
}

func newKataCreateCmd() *cobra.Command {
	var topic string
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Import a Sr-provided kata from a JSON file",
		Long: `Import a Sr-authored code kata from a JSON file.

File format — JSON object:

  {
    "title": "Reverse a string",
    "description": "Write a function that reverses a string in-place.",
    "language": "go",
    "starter_code": "package kata\n\nfunc Reverse(s string) string {\n\treturn \"\"\n}",
    "tests": "package kata_test\n\nimport \"testing\"\n\nfunc TestReverse(t *testing.T) { ... }"
  }

Supported languages: go, python, javascript, java
The tests field must contain a complete test file for the target language.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKataCreate(cmd.Context(), topic, file)
		},
	}
	cmd.Flags().StringVar(&topic, "topic", "", "Topic to tag this kata with (required)")
	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file containing the kata (required)")
	_ = cmd.MarkFlagRequired("topic")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func runKataCreate(ctx context.Context, topic, file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("kata create: read file: %w", err)
	}

	var raw rawImportKata
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("kata create: parse JSON: %w", err)
	}

	if raw.Title == "" {
		return fmt.Errorf("kata create: title is required")
	}
	if raw.Tests == "" {
		return fmt.Errorf("kata create: tests is required")
	}

	lang := kata.Language(raw.Language)
	switch lang {
	case kata.Go, kata.Python, kata.JavaScript, kata.Java:
	default:
		return fmt.Errorf("kata create: unsupported language %q (valid: go, python, javascript, java)", raw.Language)
	}

	k := &kata.Kata{
		ID:          uuid.New().String(),
		Topic:       topic,
		Title:       raw.Title,
		Description: raw.Description,
		StarterCode: raw.StarterCode,
		Tests:       raw.Tests,
		Language:    lang,
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("kata create: load config: %w", err)
	}
	dbPath, err := resolveDBPath(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("kata create: %w", err)
	}
	st, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("kata create: open store: %w", err)
	}
	defer st.Close()

	if err := st.SaveKata(ctx, k); err != nil {
		return fmt.Errorf("kata create: save: %w", err)
	}

	fmt.Printf("Imported kata %q (ID: %s) for topic %q.\n", k.Title, k.ID, topic)
	return nil
}
