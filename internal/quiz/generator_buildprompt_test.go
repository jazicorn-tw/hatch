package quiz

// Internal tests (package quiz) covering buildPrompt error paths and the
// chunk-iteration loop body in buildPrompt.

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jazicorn/hatch/internal/chunker"
	embfake "github.com/jazicorn/hatch/internal/embedder/fake"
	llmfake "github.com/jazicorn/hatch/internal/llm/fake"
	"github.com/jazicorn/hatch/internal/store"
	"github.com/jazicorn/hatch/internal/store/memory"
)

func TestBuildPromptReadError(t *testing.T) {
	orig := readTemplateFile
	readTemplateFile = func(_ string) ([]byte, error) { return nil, errors.New("read error") }
	defer func() { readTemplateFile = orig }()

	g := NewGenerator(&embfake.Embedder{Dim: 4}, memory.New(), &llmfake.LLM{}, GeneratorConfig{})
	_, _, err := g.buildPrompt("go", 1, nil)
	if err == nil {
		t.Error("expected error when readTemplateFile fails")
	}
}

func TestBuildPromptParseError(t *testing.T) {
	orig := readTemplateFile
	readTemplateFile = func(_ string) ([]byte, error) {
		return []byte("{{ .Unclosed"), nil // invalid template syntax
	}
	defer func() { readTemplateFile = orig }()

	g := NewGenerator(&embfake.Embedder{Dim: 4}, memory.New(), &llmfake.LLM{}, GeneratorConfig{})
	_, _, err := g.buildPrompt("go", 1, nil)
	if err == nil {
		t.Error("expected error when template.Parse fails")
	}
}

func TestBuildPromptExecuteError(t *testing.T) {
	orig := readTemplateFile
	// Template that calls .Topic as a function — Execute will fail.
	readTemplateFile = func(_ string) ([]byte, error) {
		return []byte("{{ call .Topic }}"), nil
	}
	defer func() { readTemplateFile = orig }()

	g := NewGenerator(&embfake.Embedder{Dim: 4}, memory.New(), &llmfake.LLM{}, GeneratorConfig{})
	_, _, err := g.buildPrompt("go", 1, nil)
	if err == nil {
		t.Error("expected error when template.Execute fails")
	}
}

func TestBuildPromptWithRecords(t *testing.T) {
	// Covers the for-range loop body in buildPrompt (chunks and ids appended).
	validJSON := `[{"text":"Q?","options":["a","b","c","d"],"correct_index":0,"explanation":"e"}]`

	st := memory.New()
	_ = st.Add(context.Background(), []store.Record{
		{
			Chunk:     chunker.Chunk{ID: "c1", Source: "src", Text: "chunk content"},
			Embedding: []float32{0.1, 0.2, 0.3, 0.4},
		},
	})

	g := NewGenerator(&embfake.Embedder{Dim: 4}, st, &llmfake.LLM{Response: validJSON}, GeneratorConfig{TopK: 1})
	questions, err := g.Generate(context.Background(), "go", 1)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(questions) != 1 {
		t.Errorf("expected 1 question, got %d", len(questions))
	}
}

func TestGenerateBuildPromptError(t *testing.T) {
	orig := readTemplateFile
	readTemplateFile = func(_ string) ([]byte, error) { return nil, fmt.Errorf("template read failed") }
	defer func() { readTemplateFile = orig }()

	g := NewGenerator(&embfake.Embedder{Dim: 4}, memory.New(), &llmfake.LLM{}, GeneratorConfig{})
	_, err := g.Generate(context.Background(), "go", 1)
	if err == nil {
		t.Error("expected error when buildPrompt fails")
	}
}
