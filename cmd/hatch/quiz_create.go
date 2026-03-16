package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/jazicorn/hatch/internal/quiz"
)

// rawImportQuestion is the JSON file format for Sr-provided quiz questions.
type rawImportQuestion struct {
	Text         string   `json:"text"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correct_index"`
	Explanation  string   `json:"explanation"`
}

func newQuizCreateCmd() *cobra.Command {
	var topic string
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Import Sr-provided quiz questions from a JSON file",
		Long: `Import Sr-authored multiple-choice questions from a JSON file.

File format — JSON array of question objects:

  [
    {
      "text": "What does defer do in Go?",
      "options": ["Delays to function exit", "Panics immediately", "Returns early", "Loops"],
      "correct_index": 0,
      "explanation": "defer schedules a function to run when the surrounding function returns"
    }
  ]

Each question must have exactly 4 options and a correct_index in [0, 3].
Imported questions are stored by topic and available for use in quiz sessions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuizCreate(cmd.Context(), topic, file)
		},
	}
	cmd.Flags().StringVar(&topic, "topic", "", "Topic to tag these questions with (required)")
	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file containing quiz questions (required)")
	_ = cmd.MarkFlagRequired("topic")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func runQuizCreate(ctx context.Context, topic, file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("quiz create: read file: %w", err)
	}

	var raws []rawImportQuestion
	if err := json.Unmarshal(data, &raws); err != nil {
		return fmt.Errorf("quiz create: parse JSON: %w", err)
	}
	if len(raws) == 0 {
		return fmt.Errorf("quiz create: file contains no questions")
	}

	questions := make([]quiz.Question, 0, len(raws))
	for i, r := range raws {
		if r.Text == "" {
			return fmt.Errorf("quiz create: question %d: text is required", i+1)
		}
		if len(r.Options) != 4 {
			return fmt.Errorf("quiz create: question %d: expected 4 options, got %d", i+1, len(r.Options))
		}
		if r.CorrectIndex < 0 || r.CorrectIndex > 3 {
			return fmt.Errorf("quiz create: question %d: correct_index %d out of range [0, 3]", i+1, r.CorrectIndex)
		}
		var opts [4]string
		copy(opts[:], r.Options)
		questions = append(questions, quiz.Question{
			ID:           uuid.New().String(),
			Text:         r.Text,
			Options:      opts,
			CorrectIndex: r.CorrectIndex,
			Explanation:  r.Explanation,
		})
	}

	st, err := setupStore()
	if err != nil {
		return fmt.Errorf("quiz create: %w", err)
	}
	defer st.Close()

	if err := st.SaveQuestionBank(ctx, topic, questions); err != nil {
		return fmt.Errorf("quiz create: save: %w", err)
	}

	fmt.Printf("Imported %d question(s) for topic %q.\n", len(questions), topic)
	return nil
}
