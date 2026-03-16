package kata

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/jazicorn/hatch/internal/embedder"
	"github.com/jazicorn/hatch/internal/genutil"
	"github.com/jazicorn/hatch/internal/llm"
	"github.com/jazicorn/hatch/internal/store"
)

//go:embed prompt/*.tmpl
var promptFS embed.FS

// Generator creates katas from a topic by:
//  1. Embedding the topic to a search vector
//  2. Retrieving top-k relevant chunks from the store
//  3. Rendering the kata_generate.tmpl prompt with those chunks
//  4. Calling the LLM and parsing the JSON response into a Kata
type Generator struct {
	embedder embedder.Embedder
	store    store.Store
	llm      llm.Completer
	topK     int
}

// GeneratorConfig configures a Generator.
type GeneratorConfig struct {
	// TopK is the number of chunks to retrieve for context. Defaults to 10.
	TopK int
}

// NewGenerator returns a Generator wired to the given dependencies.
func NewGenerator(emb embedder.Embedder, st store.Store, completer llm.Completer, cfg GeneratorConfig) *Generator {
	if cfg.TopK <= 0 {
		cfg.TopK = 10
	}
	return &Generator{
		embedder: emb,
		store:    st,
		llm:      completer,
		topK:     cfg.TopK,
	}
}

// chunkData is the template context for a single retrieved chunk.
type chunkData struct {
	Text string
}

// promptData is the full template context passed to kata_generate.tmpl.
type promptData struct {
	Topic  string
	Chunks []chunkData
}

// rawKata mirrors the JSON schema the LLM is asked to produce.
type rawKata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Language    string `json:"language"`
	StarterCode string `json:"starter_code"`
	Tests       string `json:"tests"`
}

// Generate returns a Kata for the given topic.
func (g *Generator) Generate(ctx context.Context, topic string) (*Kata, error) {
	// 1. Embed the topic query.
	vecs, err := g.embedder.Embed(ctx, []string{topic})
	if err != nil {
		return nil, fmt.Errorf("kata generator: embed topic: %w", err)
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("kata generator: embedder returned no vectors")
	}

	// 2. Search the vector store with diversity filter.
	candidates, err := g.store.Search(ctx, vecs[0], g.topK*4)
	if err != nil {
		return nil, fmt.Errorf("kata generator: search: %w", err)
	}
	records := genutil.DiversifyBySource(candidates, g.topK)

	// 3. Build prompt.
	prompt, err := g.buildPrompt(topic, records)
	if err != nil {
		return nil, fmt.Errorf("kata generator: build prompt: %w", err)
	}

	// 4. Call LLM.
	raw, err := g.llm.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("kata generator: llm: %w", err)
	}

	// 5. Parse JSON response.
	k, err := parseKata(raw, topic)
	if err != nil {
		return nil, fmt.Errorf("kata generator: parse: %w", err)
	}
	return k, nil
}

// buildPrompt renders kata_generate.tmpl with topic and chunks.
func (g *Generator) buildPrompt(topic string, records []store.Record) (string, error) {
	tmplBytes, err := promptFS.ReadFile("prompt/kata_generate.tmpl")
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}
	tmpl, err := template.New("kata").Parse(string(tmplBytes))
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	chunks := make([]chunkData, 0, len(records))
	for _, r := range records {
		chunks = append(chunks, chunkData{Text: r.Chunk.Text})
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, promptData{Topic: topic, Chunks: chunks}); err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}
	return buf.String(), nil
}

// parseKata extracts a Kata from the LLM's JSON response.
func parseKata(raw, topic string) (*Kata, error) {
	cleaned := genutil.StripMarkdownFence(raw)

	var rk rawKata
	if err := json.Unmarshal([]byte(cleaned), &rk); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w (raw: %.200s)", err, cleaned)
	}

	lang := Language(strings.ToLower(strings.TrimSpace(rk.Language)))
	switch lang {
	case Go, Python, JavaScript, Java:
	default:
		lang = Go // safe fallback
	}

	return &Kata{
		ID:          fmt.Sprintf("kata-%s", strings.ReplaceAll(strings.ToLower(rk.Title), " ", "-")),
		Title:       rk.Title,
		Description: rk.Description,
		Language:    lang,
		StarterCode: rk.StarterCode,
		Tests:       rk.Tests,
		Topic:       topic,
	}, nil
}
