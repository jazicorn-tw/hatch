package quiz

import (
	"context"
	"errors"
	"testing"

	embfake "github.com/jazicorn/hatch/internal/embedder/fake"
	llmfake "github.com/jazicorn/hatch/internal/llm/fake"
	"github.com/jazicorn/hatch/internal/store/memory"
)

// ---------------------------------------------------------------------------
// parseQuestions (internal)
// ---------------------------------------------------------------------------

func TestParseQuestionsValid(t *testing.T) {
	raw := `[{"text":"What is Go?","options":["A language","A game","A city","A food"],"correct_index":0,"explanation":"Go is a language"}]`
	questions, err := parseQuestions(raw, []string{"chunk1"})
	if err != nil {
		t.Fatalf("parseQuestions: %v", err)
	}
	if len(questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(questions))
	}
	if questions[0].Text != "What is Go?" {
		t.Errorf("unexpected text: %s", questions[0].Text)
	}
	if questions[0].CorrectIndex != 0 {
		t.Errorf("expected CorrectIndex 0, got %d", questions[0].CorrectIndex)
	}
	if questions[0].Options[0] != "A language" {
		t.Errorf("unexpected option: %s", questions[0].Options[0])
	}
	if len(questions[0].SourceChunks) != 1 || questions[0].SourceChunks[0] != "chunk1" {
		t.Errorf("unexpected source chunks: %v", questions[0].SourceChunks)
	}
}

func TestParseQuestionsWithMarkdownFence(t *testing.T) {
	raw := "```json\n[{\"text\":\"Q?\",\"options\":[\"a\",\"b\",\"c\",\"d\"],\"correct_index\":1,\"explanation\":\"because\"}]\n```"
	questions, err := parseQuestions(raw, nil)
	if err != nil {
		t.Fatalf("parseQuestions with fence: %v", err)
	}
	if len(questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(questions))
	}
}

func TestParseQuestionsInvalidJSON(t *testing.T) {
	_, err := parseQuestions("not json", nil)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseQuestionsWrongOptionCount(t *testing.T) {
	raw := `[{"text":"Q?","options":["a","b","c"],"correct_index":0,"explanation":""}]`
	_, err := parseQuestions(raw, nil)
	if err == nil {
		t.Error("expected error for 3 options instead of 4")
	}
}

func TestParseQuestionsIndexOutOfRange(t *testing.T) {
	raw := `[{"text":"Q?","options":["a","b","c","d"],"correct_index":5,"explanation":""}]`
	_, err := parseQuestions(raw, nil)
	if err == nil {
		t.Error("expected error for correct_index 5 out of range")
	}
}

func TestParseQuestionsMultiple(t *testing.T) {
	raw := `[
		{"text":"Q1?","options":["a","b","c","d"],"correct_index":0,"explanation":"e1"},
		{"text":"Q2?","options":["w","x","y","z"],"correct_index":3,"explanation":"e2"}
	]`
	questions, err := parseQuestions(raw, nil)
	if err != nil {
		t.Fatalf("parseQuestions: %v", err)
	}
	if len(questions) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(questions))
	}
	if questions[0].ID != "q1" || questions[1].ID != "q2" {
		t.Errorf("unexpected IDs: %s, %s", questions[0].ID, questions[1].ID)
	}
}

// ---------------------------------------------------------------------------
// Generator.Generate (integration with fakes)
// ---------------------------------------------------------------------------

func TestGeneratorGenerate(t *testing.T) {
	validJSON := `[{"text":"What is Go?","options":["A language","A game","A city","A food"],"correct_index":0,"explanation":"Go is a language"}]`

	g := NewGenerator(
		&embfake.Embedder{Dim: 4},
		memory.New(),
		&llmfake.LLM{Response: validJSON},
		GeneratorConfig{TopK: 3},
	)

	questions, err := g.Generate(context.Background(), "go", 1)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(questions))
	}
}

func TestGeneratorGenerateDefaultTopK(t *testing.T) {
	validJSON := `[{"text":"Q?","options":["a","b","c","d"],"correct_index":0,"explanation":"e"}]`
	// TopK=0 should default to 5
	g := NewGenerator(
		&embfake.Embedder{Dim: 4},
		memory.New(),
		&llmfake.LLM{Response: validJSON},
		GeneratorConfig{},
	)
	if g.topK != 5 {
		t.Errorf("expected topK=5, got %d", g.topK)
	}
}

func TestGeneratorEmbedError(t *testing.T) {
	g := NewGenerator(
		&embfake.Embedder{Err: errors.New("embed failed")},
		memory.New(),
		&llmfake.LLM{},
		GeneratorConfig{},
	)
	_, err := g.Generate(context.Background(), "go", 1)
	if err == nil {
		t.Error("expected error from embedder")
	}
}

func TestGeneratorLLMError(t *testing.T) {
	g := NewGenerator(
		&embfake.Embedder{Dim: 4},
		memory.New(),
		&llmfake.LLM{Response: ""},
		GeneratorConfig{},
	)
	// Default "fake response" from LLM is not valid JSON — expect parse error
	_, err := g.Generate(context.Background(), "go", 1)
	if err == nil {
		t.Error("expected error when LLM returns non-JSON")
	}
}
