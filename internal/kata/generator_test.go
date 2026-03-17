package kata

import (
	"context"
	"errors"
	"testing"

	embfake "github.com/jazicorn/hatch/internal/embedder/fake"
	llmfake "github.com/jazicorn/hatch/internal/llm/fake"
	"github.com/jazicorn/hatch/internal/store"
	"github.com/jazicorn/hatch/internal/store/memory"
)

// emptyVecEmbedder returns an empty (non-nil) slice with no error.
type emptyVecEmbedder struct{}

func (e *emptyVecEmbedder) Embed(_ context.Context, _ []string) ([][]float32, error) {
	return [][]float32{}, nil
}

// errSearchStore returns an error from Search.
type errSearchStore struct{ *memory.Store }

func (s *errSearchStore) Search(_ context.Context, _ []float32, _ int) ([]store.Record, error) {
	return nil, errors.New("search failed")
}

// errCompleter always returns an error from Complete.
type errCompleter struct{}

func (c *errCompleter) Complete(_ context.Context, _ string) (string, error) {
	return "", errors.New("llm failed")
}

// ---------------------------------------------------------------------------
// parseKata (internal)
// ---------------------------------------------------------------------------

func TestParseKataValid(t *testing.T) {
	raw := `{"title":"Hello World","description":"Write a hello world program","language":"go","starter_code":"package main\n","tests":"func TestHello(t *testing.T) {}"}`
	k, err := parseKata(raw, "go")
	if err != nil {
		t.Fatalf("parseKata: %v", err)
	}
	if k.Title != "Hello World" {
		t.Errorf("unexpected title: %s", k.Title)
	}
	if k.Language != Go {
		t.Errorf("expected language Go, got %s", k.Language)
	}
	if k.Topic != "go" {
		t.Errorf("expected topic go, got %s", k.Topic)
	}
	if k.StarterCode != "package main\n" {
		t.Errorf("unexpected starter code: %s", k.StarterCode)
	}
}

func TestParseKataWithMarkdownFence(t *testing.T) {
	raw := "```json\n{\"title\":\"T\",\"description\":\"D\",\"language\":\"python\",\"starter_code\":\"x\",\"tests\":\"t\"}\n```"
	k, err := parseKata(raw, "python")
	if err != nil {
		t.Fatalf("parseKata with fence: %v", err)
	}
	if k.Language != Python {
		t.Errorf("expected Python, got %s", k.Language)
	}
}

func TestParseKataInvalidJSON(t *testing.T) {
	_, err := parseKata("not json", "go")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseKataUnknownLanguageFallsBackToGo(t *testing.T) {
	raw := `{"title":"T","description":"D","language":"cobol","starter_code":"","tests":""}`
	k, err := parseKata(raw, "cobol")
	if err != nil {
		t.Fatalf("parseKata: %v", err)
	}
	if k.Language != Go {
		t.Errorf("expected Go fallback, got %s", k.Language)
	}
}

func TestParseKataKnownLanguages(t *testing.T) {
	cases := []struct {
		lang string
		want Language
	}{
		{"go", Go},
		{"python", Python},
		{"javascript", JavaScript},
		{"java", Java},
	}
	for _, tc := range cases {
		raw := `{"title":"T","description":"D","language":"` + tc.lang + `","starter_code":"","tests":""}`
		k, err := parseKata(raw, tc.lang)
		if err != nil {
			t.Fatalf("parseKata %s: %v", tc.lang, err)
		}
		if k.Language != tc.want {
			t.Errorf("lang %s: want %s, got %s", tc.lang, tc.want, k.Language)
		}
	}
}

func TestParseKataIDFromTitle(t *testing.T) {
	raw := `{"title":"Hello World","description":"","language":"go","starter_code":"","tests":""}`
	k, err := parseKata(raw, "go")
	if err != nil {
		t.Fatalf("parseKata: %v", err)
	}
	if k.ID != "kata-hello-world" {
		t.Errorf("expected id kata-hello-world, got %s", k.ID)
	}
}

// ---------------------------------------------------------------------------
// Generator.Generate (integration with fakes)
// ---------------------------------------------------------------------------

func TestGeneratorGenerate(t *testing.T) {
	validJSON := `{"title":"Hello World","description":"Write a hello world","language":"go","starter_code":"package main\n","tests":"func TestHello(t *testing.T) {}"}`

	g := NewGenerator(
		&embfake.Embedder{Dim: 4},
		memory.New(),
		&llmfake.LLM{Response: validJSON},
		GeneratorConfig{TopK: 3},
	)

	k, err := g.Generate(context.Background(), "go")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if k.Title != "Hello World" {
		t.Errorf("unexpected title: %s", k.Title)
	}
	if k.Topic != "go" {
		t.Errorf("expected topic go, got %s", k.Topic)
	}
}

func TestGeneratorGenerateDefaultTopK(t *testing.T) {
	// TopK=0 should default to 10
	g := NewGenerator(
		&embfake.Embedder{Dim: 4},
		memory.New(),
		&llmfake.LLM{},
		GeneratorConfig{},
	)
	if g.topK != 10 {
		t.Errorf("expected topK=10, got %d", g.topK)
	}
}

func TestGeneratorEmbedError(t *testing.T) {
	g := NewGenerator(
		&embfake.Embedder{Err: errors.New("embed failed")},
		memory.New(),
		&llmfake.LLM{},
		GeneratorConfig{},
	)
	_, err := g.Generate(context.Background(), "go")
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
	// Default "fake response" is not valid JSON — expect parse error
	_, err := g.Generate(context.Background(), "go")
	if err == nil {
		t.Error("expected error when LLM returns non-JSON")
	}
}

func TestGeneratorEmptyVectors(t *testing.T) {
	// Embedder returns empty slice (no error) → Generate should return error.
	g := NewGenerator(&emptyVecEmbedder{}, memory.New(), &llmfake.LLM{}, GeneratorConfig{})
	_, err := g.Generate(context.Background(), "go")
	if err == nil {
		t.Error("expected error when embedder returns no vectors")
	}
}

func TestGeneratorSearchError(t *testing.T) {
	g := NewGenerator(
		&embfake.Embedder{Dim: 4},
		&errSearchStore{Store: memory.New()},
		&llmfake.LLM{},
		GeneratorConfig{},
	)
	_, err := g.Generate(context.Background(), "go")
	if err == nil {
		t.Error("expected error when store search fails")
	}
}

func TestGeneratorLLMCompleteError(t *testing.T) {
	validJSON := `{"title":"T","description":"D","language":"go","starter_code":"","tests":""}`
	_ = validJSON
	g := NewGenerator(
		&embfake.Embedder{Dim: 4},
		memory.New(),
		&errCompleter{},
		GeneratorConfig{},
	)
	_, err := g.Generate(context.Background(), "go")
	if err == nil {
		t.Error("expected error when LLM complete fails")
	}
}
