package kata

// Internal tests (package kata) covering buildPrompt error paths and the
// chunk-iteration loop body.

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
	_, err := g.buildPrompt("go", nil)
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
	_, err := g.buildPrompt("go", nil)
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
	_, err := g.buildPrompt("go", nil)
	if err == nil {
		t.Error("expected error when template.Execute fails")
	}
}

func TestBuildPromptWithRecords(t *testing.T) {
	// Covers the for-range loop body in buildPrompt (chunks appended).
	validJSON := `{"title":"T","description":"D","language":"go","starter_code":"x","tests":"t"}`

	st := memory.New()
	_ = st.Add(context.Background(), []store.Record{
		{
			Chunk:     chunker.Chunk{ID: "c1", Source: "src", Text: "chunk content"},
			Embedding: []float32{0.1, 0.2, 0.3, 0.4},
		},
	})

	g := NewGenerator(&embfake.Embedder{Dim: 4}, st, &llmfake.LLM{Response: validJSON}, GeneratorConfig{TopK: 1})
	k, err := g.Generate(context.Background(), "go")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if k.Title != "T" {
		t.Errorf("unexpected title: %s", k.Title)
	}
}

func TestGenerateBuildPromptError(t *testing.T) {
	orig := readTemplateFile
	readTemplateFile = func(_ string) ([]byte, error) { return nil, fmt.Errorf("template read failed") }
	defer func() { readTemplateFile = orig }()

	g := NewGenerator(&embfake.Embedder{Dim: 4}, memory.New(), &llmfake.LLM{}, GeneratorConfig{})
	_, err := g.Generate(context.Background(), "go")
	if err == nil {
		t.Error("expected error when buildPrompt fails")
	}
}
