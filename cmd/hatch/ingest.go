package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/jazicorn/hatch/internal/chunker"
	codechunker "github.com/jazicorn/hatch/internal/chunker/code"
	mdchunker "github.com/jazicorn/hatch/internal/chunker/markdown"
	"github.com/jazicorn/hatch/internal/config"
	"github.com/jazicorn/hatch/internal/embedder"
	gemiembed "github.com/jazicorn/hatch/internal/embedder/gemini"
	oaiembed "github.com/jazicorn/hatch/internal/embedder/openai"
	"github.com/jazicorn/hatch/internal/pipeline"
	"github.com/jazicorn/hatch/internal/source"
	fssource "github.com/jazicorn/hatch/internal/source/fs"
	"github.com/jazicorn/hatch/internal/store/sqlite"
)

func newIngestCmd() *cobra.Command {
	var sourceName string
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Ingest a configured source into the vector store",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIngest(cmd.Context(), sourceName)
		},
	}
	cmd.Flags().StringVar(&sourceName, "source", "", "Name of the source to ingest (required)")
	_ = cmd.MarkFlagRequired("source")
	return cmd
}

func runIngest(ctx context.Context, sourceName string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("ingest: load config: %w", err)
	}

	srcCfg, err := findSource(cfg, sourceName)
	if err != nil {
		return err
	}

	srcPath, err := resolvePath(srcCfg.Path)
	if err != nil {
		return fmt.Errorf("ingest: %w", err)
	}

	src, err := fssource.New(fssource.Config{Root: srcPath, SourceName: srcCfg.Name})
	if err != nil {
		return fmt.Errorf("ingest: create source: %w", err)
	}

	emb, err := newEmbedder(cfg)
	if err != nil {
		return fmt.Errorf("ingest: create embedder: %w", err)
	}

	dbPath := cfg.DBPath
	if dbPath == "" {
		home, _ := os.UserHomeDir()
		dbPath = filepath.Join(home, ".hatch", "hatch.db")
	}
	st, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("ingest: open store: %w", err)
	}
	defer st.Close()

	progressCh := make(chan pipeline.Progress, 16)
	barDone := drainProgressBar(sourceName, progressCh)

	runErr := pipeline.Run(ctx, src, newDispatchChunker(), emb, st, progressCh)
	close(progressCh)
	<-barDone

	if runErr != nil {
		return fmt.Errorf("ingest: %w", runErr)
	}
	fmt.Fprintf(os.Stderr, "\nIngestion complete for source %q\n", sourceName)
	return nil
}

// findSource returns the SourceConfig with the given name, or an error.
func findSource(cfg *config.Config, name string) (*config.SourceConfig, error) {
	for i := range cfg.Sources {
		if cfg.Sources[i].Name == name {
			return &cfg.Sources[i], nil
		}
	}
	return nil, fmt.Errorf("ingest: source %q not found in config (run: hatch sources list)", name)
}

// resolvePath returns an absolute path, resolving relative paths against cwd.
func resolvePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working dir: %w", err)
	}
	return filepath.Join(wd, path), nil
}

// newEmbedder constructs the appropriate embedder based on cfg.EmbedProvider.
func newEmbedder(cfg *config.Config) (embedder.Embedder, error) {
	switch cfg.EmbedProvider {
	case "gemini":
		apiKey := cfg.GoogleAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("GOOGLE_API_KEY")
		}
		return gemiembed.New(gemiembed.Config{APIKey: apiKey})
	default: // "openai" and unset
		apiKey := cfg.OpenAIAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		return oaiembed.New(oaiembed.Config{APIKey: apiKey})
	}
}

// drainProgressBar starts a goroutine that updates a progress bar from ch.
// It returns a channel closed when the goroutine exits.
func drainProgressBar(sourceName string, ch <-chan pipeline.Progress) <-chan struct{} {
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(fmt.Sprintf("ingesting %q", sourceName)),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(),
	)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer bar.Finish() //nolint:errcheck
		for p := range ch {
			bar.ChangeMax(p.Total)
			_ = bar.Set(p.Done)
		}
	}()
	return done
}

// dispatchChunker routes documents to the markdown or code chunker by extension.
type dispatchChunker struct {
	md   *mdchunker.Chunker
	code *codechunker.Chunker
}

func newDispatchChunker() *dispatchChunker {
	codeChk, _ := codechunker.New(codechunker.DefaultConfig())
	return &dispatchChunker{
		md:   mdchunker.New(),
		code: codeChk,
	}
}

// codeExtensions are file extensions routed to the code chunker.
var codeExtensions = map[string]bool{
	".go": true, ".ts": true, ".tsx": true, ".scss": true,
}

// Chunk implements chunker.Chunker.
func (d *dispatchChunker) Chunk(doc source.Document) ([]chunker.Chunk, error) {
	ext := strings.ToLower(filepath.Ext(doc.ID))
	if codeExtensions[ext] {
		return d.code.Chunk(doc)
	}
	return d.md.Chunk(doc)
}
