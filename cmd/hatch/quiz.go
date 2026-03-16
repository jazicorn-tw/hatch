package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/jazicorn/hatch/internal/quiz"
)

func newQuizCmd() *cobra.Command {
	var topic string
	var count int

	cmd := &cobra.Command{
		Use:   "quiz",
		Short: "Take an AI-generated quiz on a topic",
		Long:  "Generates multiple-choice questions from ingested content and runs an interactive quiz session.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuiz(cmd.Context(), topic, count)
		},
	}
	cmd.Flags().StringVar(&topic, "topic", "", "Topic to focus on (optional — omit to quiz across all ingested content)")
	cmd.Flags().IntVar(&count, "count", 5, "Number of questions to generate")
	cmd.AddCommand(newQuizCreateCmd())
	return cmd
}

func runQuiz(ctx context.Context, topic string, count int) error {
	d, err := setupDeps()
	if err != nil {
		return fmt.Errorf("quiz: %w", err)
	}
	defer d.store.Close()

	gen := quiz.NewGenerator(d.emb, d.store, d.llm, quiz.GeneratorConfig{TopK: 10})

	label := topic
	if label == "" {
		label = "general"
	}
	fmt.Fprintf(os.Stderr, "Generating %d question(s) on %q…\n", count, label)
	questions, err := gen.Generate(ctx, label, count)
	if err != nil {
		return fmt.Errorf("quiz: generate: %w", err)
	}
	if len(questions) == 0 {
		return fmt.Errorf("quiz: no questions generated — make sure content has been ingested")
	}

	sess := quiz.NewSession(uuid.New().String(), label)
	sess.Questions = questions

	eval := quiz.NewEvaluator()
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\n=== Quiz: %s (%d questions) ===\n\n", label, len(questions))

	for i, q := range questions {
		fmt.Printf("Q%d. %s\n", i+1, q.Text)
		for j, opt := range q.Options {
			fmt.Printf("  %d. %s\n", j+1, opt)
		}

		answer := promptAnswer(reader, len(q.Options))
		sess.Answers = append(sess.Answers, answer)

		if eval.Check(q, answer) {
			fmt.Printf("✓ Correct!\n")
		} else {
			fmt.Printf("✗ Incorrect. Correct answer: %d. %s\n", q.CorrectIndex+1, q.Options[q.CorrectIndex])
		}
		fmt.Printf("  %s\n\n", q.Explanation)
	}

	sess.Finish()
	correct, total := sess.Score()
	fmt.Printf("=== Score: %d / %d (%.0f%%) ===\n", correct, total, float64(correct)/float64(total)*100)

	if err := d.store.SaveSession(ctx, sess); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save session: %v\n", err)
	}
	return nil
}

// promptAnswer reads a valid answer (1–optionCount) from stdin, retrying on bad input.
func promptAnswer(r *bufio.Reader, optionCount int) int {
	for {
		fmt.Printf("Your answer (1-%d): ", optionCount)
		line, _ := r.ReadString('\n')
		line = strings.TrimSpace(line)
		n, err := strconv.Atoi(line)
		if err == nil && n >= 1 && n <= optionCount {
			return n - 1 // convert to 0-based
		}
		fmt.Printf("Please enter a number between 1 and %d.\n", optionCount)
	}
}
