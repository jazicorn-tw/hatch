package quiz

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/jazicorn/hatch/internal/embedder"
	"github.com/jazicorn/hatch/internal/genutil"
	"github.com/jazicorn/hatch/internal/llm"
	"github.com/jazicorn/hatch/internal/store"
)

//go:embed prompt/*.tmpl
var promptFS embed.FS

// Generator creates MCQ questions from a topic by:
//  1. Embedding the topic to a search vector
//  2. Retrieving the top-k relevant chunks from the store
//  3. Rendering a prompt template with those chunks
//  4. Calling the LLM and parsing the JSON response
type Generator struct {
	embedder embedder.Embedder
	store    store.Store
	llm      llm.Completer
	topK     int
}

// GeneratorConfig configures a Generator.
type GeneratorConfig struct {
	// TopK is the number of store chunks to retrieve for context. Defaults to 5.
	TopK int
}

// NewGenerator returns a Generator wired to the given dependencies.
func NewGenerator(emb embedder.Embedder, st store.Store, completer llm.Completer, cfg GeneratorConfig) *Generator {
	if cfg.TopK <= 0 {
		cfg.TopK = 5
	}
	return &Generator{
		embedder: emb,
		store:    st,
		llm:      completer,
		topK:     cfg.TopK,
	}
}

// promptData is the full template context passed to mcq.tmpl.
type promptData struct {
	Topic  string
	Count  int
	Chunks []genutil.ChunkData
}

// rawQuestion mirrors the JSON schema the LLM is asked to produce.
type rawQuestion struct {
	Text         string   `json:"text"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correct_index"`
	Explanation  string   `json:"explanation"`
}

// Generate returns count MCQ Questions for the given topic.
// It retrieves relevant chunks from the store using vector search, then asks
// the LLM to produce questions grounded in those chunks.
func (g *Generator) Generate(ctx context.Context, topic string, count int) ([]Question, error) {
	// 1. Embed the topic query.
	vecs, err := g.embedder.Embed(ctx, []string{topic})
	if err != nil {
		return nil, fmt.Errorf("quiz generator: embed topic: %w", err)
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("quiz generator: embedder returned no vectors")
	}

	// 2. Search the vector store. Fetch 4× more candidates than needed so the
	// diversity filter has enough material to pick one chunk per source file.
	candidates, err := g.store.Search(ctx, vecs[0], g.topK*4)
	if err != nil {
		return nil, fmt.Errorf("quiz generator: search: %w", err)
	}
	records := genutil.DiversifyBySource(candidates, g.topK)

	// 3. Build prompt.
	prompt, chunkIDs, err := g.buildPrompt(topic, count, records)
	if err != nil {
		return nil, fmt.Errorf("quiz generator: build prompt: %w", err)
	}

	// 4. Call LLM.
	raw, err := g.llm.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("quiz generator: llm: %w", err)
	}

	// 5. Parse JSON response.
	questions, err := parseQuestions(raw, chunkIDs)
	if err != nil {
		return nil, fmt.Errorf("quiz generator: parse: %w", err)
	}
	return questions, nil
}

// buildPrompt renders mcq.tmpl with the topic, count, and retrieved chunks.
// It also returns the chunk IDs so they can be attached to each Question.
func (g *Generator) buildPrompt(topic string, count int, records []store.Record) (string, []string, error) {
	tmplBytes, err := promptFS.ReadFile("prompt/mcq.tmpl")
	if err != nil {
		return "", nil, fmt.Errorf("read template: %w", err)
	}
	tmpl, err := template.New("mcq").Parse(string(tmplBytes))
	if err != nil {
		return "", nil, fmt.Errorf("parse template: %w", err)
	}

	chunks := make([]genutil.ChunkData, 0, len(records))
	ids := make([]string, 0, len(records))
	for _, r := range records {
		chunks = append(chunks, genutil.ChunkData{Text: r.Chunk.Text})
		ids = append(ids, r.Chunk.ID)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, promptData{Topic: topic, Count: count, Chunks: chunks}); err != nil {
		return "", nil, fmt.Errorf("render template: %w", err)
	}
	return buf.String(), ids, nil
}

// parseQuestions extracts Question values from the LLM's JSON response.
// The LLM may wrap the JSON in a markdown code fence — this is stripped.
func parseQuestions(raw string, chunkIDs []string) ([]Question, error) {
	cleaned := genutil.StripMarkdownFence(raw)

	var raws []rawQuestion
	if err := json.Unmarshal([]byte(cleaned), &raws); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w (raw: %.200s)", err, cleaned)
	}

	questions := make([]Question, 0, len(raws))
	for i, r := range raws {
		if len(r.Options) != 4 {
			return nil, fmt.Errorf("question %d: expected 4 options, got %d", i, len(r.Options))
		}
		if r.CorrectIndex < 0 || r.CorrectIndex > 3 {
			return nil, fmt.Errorf("question %d: correct_index %d out of range [0,3]", i, r.CorrectIndex)
		}
		var opts [4]string
		copy(opts[:], r.Options)
		questions = append(questions, Question{
			ID:           fmt.Sprintf("q%d", i+1),
			Text:         r.Text,
			Options:      opts,
			CorrectIndex: r.CorrectIndex,
			Explanation:  r.Explanation,
			SourceChunks: chunkIDs,
		})
	}
	return questions, nil
}
