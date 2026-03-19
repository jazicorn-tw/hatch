package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jazicorn/hatch/internal/chunker"
	"github.com/jazicorn/hatch/internal/config"
	"github.com/jazicorn/hatch/internal/embedder"
	"github.com/jazicorn/hatch/internal/pipeline"
	"github.com/jazicorn/hatch/internal/source"
	"github.com/jazicorn/hatch/internal/store"
)

const (
	wantNonNilEmbedder = "expected non-nil embedder"
	hatchConfigFile    = "config.yaml"
	hatchDirName       = ".hatch"
	envHatchDBPath     = "HATCH_DB_PATH"
	ollamaIngestYAML   = "llm_provider: anthropic\nembed_provider: ollama\ndb_path: \"\"\nsources:\n  - name: docs\n    path: %s\n    type: filesystem\n"
)

// ---------------------------------------------------------------------------
// findSource
// ---------------------------------------------------------------------------

func TestFindSourceFound(t *testing.T) {
	cfg := &config.Config{
		Sources: []config.SourceConfig{
			{Name: "docs", Path: "./docs", Type: "filesystem"},
			{Name: "src", Path: "./src", Type: "filesystem"},
		},
	}
	s, err := findSource(cfg, "docs")
	if err != nil {
		t.Fatalf("findSource: %v", err)
	}
	if s.Name != "docs" {
		t.Errorf("expected name docs, got %s", s.Name)
	}
}

func TestFindSourceNotFound(t *testing.T) {
	cfg := &config.Config{
		Sources: []config.SourceConfig{
			{Name: "docs", Path: "./docs", Type: "filesystem"},
		},
	}
	_, err := findSource(cfg, "missing")
	if err == nil {
		t.Error("expected error for missing source")
	}
}

func TestFindSourceEmpty(t *testing.T) {
	cfg := &config.Config{}
	_, err := findSource(cfg, "anything")
	if err == nil {
		t.Error("expected error when no sources configured")
	}
}

// ---------------------------------------------------------------------------
// resolveDBPath
// ---------------------------------------------------------------------------

func TestResolveDBPathAbsolute(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "test.db")
	got, err := resolveDBPath(path)
	if err != nil {
		t.Fatalf("resolveDBPath: %v", err)
	}
	if got != path {
		t.Errorf("want %s, got %s", path, got)
	}
}

func TestResolveDBPathTilde(t *testing.T) {
	got, err := resolveDBPath("~/.hatch/test.db")
	if err != nil {
		t.Fatalf("resolveDBPath with tilde: %v", err)
	}
	if got == "" || got == "~/.hatch/test.db" {
		t.Errorf("expected expanded path, got %s", got)
	}
}

func TestResolveDBPathEmpty(t *testing.T) {
	got, err := resolveDBPath("")
	if err != nil {
		t.Fatalf("resolveDBPath empty: %v", err)
	}
	if got == "" {
		t.Error("expected default path, got empty string")
	}
}

// ---------------------------------------------------------------------------
// resolvePath
// ---------------------------------------------------------------------------

func TestResolvePathAbsolute(t *testing.T) {
	abs := "/tmp/mydir"
	got, err := resolvePath(abs)
	if err != nil {
		t.Fatalf("resolvePath: %v", err)
	}
	if got != abs {
		t.Errorf("want %s, got %s", abs, got)
	}
}

func TestResolvePathRelative(t *testing.T) {
	got, err := resolvePath("docs")
	if err != nil {
		t.Fatalf("resolvePath relative: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("expected absolute path, got %s", got)
	}
}

// ---------------------------------------------------------------------------
// dispatchChunker.Chunk
// ---------------------------------------------------------------------------

func TestDispatchChunkerMarkdown(t *testing.T) {
	d := newDispatchChunker()
	doc := source.Document{ID: "README.md", Source: "src", Content: "# Hello\nworld"}
	chunks, err := d.Chunk(doc)
	if err != nil {
		t.Fatalf("Chunk .md: %v", err)
	}
	if len(chunks) == 0 {
		t.Error("expected at least one chunk for markdown doc")
	}
}

func TestDispatchChunkerCode(t *testing.T) {
	d := newDispatchChunker()
	doc := source.Document{ID: "main.go", Source: "src", Content: "package main\n\nfunc main() {}\n"}
	chunks, err := d.Chunk(doc)
	if err != nil {
		t.Fatalf("Chunk .go: %v", err)
	}
	if len(chunks) == 0 {
		t.Error("expected at least one chunk for Go file")
	}
}

func TestDispatchChunkerMDX(t *testing.T) {
	d := newDispatchChunker()
	doc := source.Document{ID: "page.mdx", Source: "src", Content: "# Page\nContent here."}
	chunks, err := d.Chunk(doc)
	if err != nil {
		t.Fatalf("Chunk .mdx: %v", err)
	}
	if len(chunks) == 0 {
		t.Error("expected at least one chunk for MDX file")
	}
}

// ---------------------------------------------------------------------------
// newEmbedder
// ---------------------------------------------------------------------------

func TestNewEmbedderOllama(t *testing.T) {
	cfg := &config.Config{EmbedProvider: "ollama"}
	emb, err := newEmbedder(cfg)
	if err != nil {
		t.Fatalf("newEmbedder ollama: %v", err)
	}
	if emb == nil {
		t.Fatal(wantNonNilEmbedder)
	}
}

func TestNewEmbedderGeminiWithKey(t *testing.T) {
	cfg := &config.Config{EmbedProvider: "gemini", GeminiAPIKey: "test-key"}
	emb, err := newEmbedder(cfg)
	if err != nil {
		t.Fatalf("newEmbedder gemini: %v", err)
	}
	if emb == nil {
		t.Fatal(wantNonNilEmbedder)
	}
}

func TestNewEmbedderGeminiNoKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	cfg := &config.Config{EmbedProvider: "gemini", GeminiAPIKey: ""}
	_, err := newEmbedder(cfg)
	if err == nil {
		t.Error("expected error when no gemini key provided")
	}
}

func TestNewEmbedderOpenAIWithKey(t *testing.T) {
	cfg := &config.Config{EmbedProvider: "openai", OpenAIAPIKey: "test-key"}
	emb, err := newEmbedder(cfg)
	if err != nil {
		t.Fatalf("newEmbedder openai: %v", err)
	}
	if emb == nil {
		t.Fatal(wantNonNilEmbedder)
	}
}

func TestNewEmbedderDefaultIsOpenAI(t *testing.T) {
	// Unknown provider falls to default (openai).
	t.Setenv("OPENAI_API_KEY", "")
	cfg := &config.Config{EmbedProvider: "unknown", OpenAIAPIKey: ""}
	_, err := newEmbedder(cfg)
	if err == nil {
		t.Error("expected error when no openai key provided for default")
	}
}

// ---------------------------------------------------------------------------
// drainProgressBar
// ---------------------------------------------------------------------------

func TestDrainProgressBar(t *testing.T) {
	progressCh := make(chan pipeline.Progress, 4)
	done := drainProgressBar("test-source", progressCh)
	progressCh <- pipeline.Progress{Done: 1, Total: 5}
	progressCh <- pipeline.Progress{Done: 5, Total: 5}
	close(progressCh)
	<-done // wait for goroutine to exit
}

// ---------------------------------------------------------------------------
// resolveDBPath — MkdirAll error
// ---------------------------------------------------------------------------

func TestResolveDBPathMkdirError(t *testing.T) {
	// /dev/null is a character device, so MkdirAll("/dev/null/sub") fails.
	_, err := resolveDBPath("/dev/null/sub/hatch.db")
	if err == nil {
		t.Error("expected error when parent directory cannot be created")
	}
}

// ---------------------------------------------------------------------------
// newIngestCmd — RunE closure
// ---------------------------------------------------------------------------

func TestNewIngestCmdRunE(t *testing.T) {
	// Trigger the RunE closure body with a malformed config so config.Load()
	// returns an error immediately.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, hatchDirName)
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hatchDir, hatchConfigFile), []byte("key: [unclosed"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := newIngestCmd()
	_ = cmd.Flags().Set("source", "docs")
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Error("expected error from RunE when config is malformed")
	}
}

// ---------------------------------------------------------------------------
// resolveDBPath — osUserHomeDir error
// ---------------------------------------------------------------------------

func TestResolveDBPathHomeDirError(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", fmt.Errorf("forced home dir error") }
	defer func() { osUserHomeDir = orig }()
	_, err := resolveDBPath("")
	if err == nil {
		t.Error("expected error when UserHomeDir fails")
	}
}

// ---------------------------------------------------------------------------
// resolvePath — osGetwd error
// ---------------------------------------------------------------------------

func TestResolvePathGetWdError(t *testing.T) {
	orig := osGetwd
	osGetwd = func() (string, error) { return "", fmt.Errorf("forced getwd error") }
	defer func() { osGetwd = orig }()
	_, err := resolvePath("relative/path")
	if err == nil {
		t.Error("expected error when getwd fails")
	}
}

// ---------------------------------------------------------------------------
// runIngest — various error and success paths
// ---------------------------------------------------------------------------

func TestRunIngestConfigLoadError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, hatchDirName)
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hatchDir, hatchConfigFile), []byte("key: [unclosed"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := runIngest(context.Background(), "test-source")
	if err == nil {
		t.Error("expected error for malformed config")
	}
}

func TestRunIngestSourceNotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, hatchDirName)
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := "llm_provider: anthropic\nembed_provider: ollama\nsources: []\n"
	if err := os.WriteFile(filepath.Join(hatchDir, hatchConfigFile), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	err := runIngest(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error when source not found")
	}
}

func TestRunIngestResolvePathError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, hatchDirName)
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// Use a relative source path so resolvePath calls osGetwd.
	yaml := "llm_provider: anthropic\nembed_provider: ollama\nsources:\n  - name: docs\n    path: relative/path\n    type: filesystem\n"
	if err := os.WriteFile(filepath.Join(hatchDir, hatchConfigFile), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	orig := osGetwd
	osGetwd = func() (string, error) { return "", fmt.Errorf("forced getwd error") }
	defer func() { osGetwd = orig }()
	err := runIngest(context.Background(), "docs")
	if err == nil {
		t.Error("expected error when resolvePath fails inside runIngest")
	}
}

func TestRunIngestFssourceNewError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	hatchDir := filepath.Join(tmp, hatchDirName)
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// Source path is absolute and nonexistent — fssource.New should fail.
	yaml := "llm_provider: anthropic\nembed_provider: ollama\nsources:\n  - name: docs\n    path: /nonexistent/path/xyz\n    type: filesystem\n"
	if err := os.WriteFile(filepath.Join(hatchDir, hatchConfigFile), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	err := runIngest(context.Background(), "docs")
	if err == nil {
		t.Error("expected error when source root does not exist")
	}
}

func TestRunIngestEmbedderError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("GEMINI_API_KEY", "")
	srcDir := filepath.Join(tmp, "src")
	if err := os.MkdirAll(srcDir, 0o700); err != nil {
		t.Fatal(err)
	}
	hatchDir := filepath.Join(tmp, hatchDirName)
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := fmt.Sprintf("llm_provider: anthropic\nembed_provider: gemini\ngemini_api_key: \"\"\nsources:\n  - name: docs\n    path: %s\n    type: filesystem\n", srcDir)
	if err := os.WriteFile(filepath.Join(hatchDir, hatchConfigFile), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	err := runIngest(context.Background(), "docs")
	if err == nil {
		t.Error("expected error when gemini embed key missing")
	}
}

func TestRunIngestResolveDBPathError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	srcDir := filepath.Join(tmp, "src")
	if err := os.MkdirAll(srcDir, 0o700); err != nil {
		t.Fatal(err)
	}
	hatchDir := filepath.Join(tmp, hatchDirName)
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := fmt.Sprintf("llm_provider: anthropic\nembed_provider: ollama\ndb_path: /dev/null/sub/hatch.db\nsources:\n  - name: docs\n    path: %s\n    type: filesystem\n", srcDir)
	if err := os.WriteFile(filepath.Join(hatchDir, hatchConfigFile), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	err := runIngest(context.Background(), "docs")
	if err == nil {
		t.Error("expected error for bad db path")
	}
}

func TestRunIngestSQLiteOpenError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	srcDir := filepath.Join(tmp, "src")
	if err := os.MkdirAll(srcDir, 0o700); err != nil {
		t.Fatal(err)
	}
	hatchDir := filepath.Join(tmp, hatchDirName)
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// Create a directory where the db file should be so sqlite.Open fails.
	dbPath := filepath.Join(hatchDir, "hatch.db")
	if err := os.MkdirAll(dbPath, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := fmt.Sprintf(ollamaIngestYAML, srcDir)
	if err := os.WriteFile(filepath.Join(hatchDir, hatchConfigFile), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	orig, set := os.LookupEnv(envHatchDBPath)
	os.Unsetenv(envHatchDBPath)
	if set {
		defer os.Setenv(envHatchDBPath, orig)
	} else {
		defer os.Unsetenv(envHatchDBPath)
	}
	err := runIngest(context.Background(), "docs")
	if err == nil {
		t.Error("expected error when db path is a directory")
	}
}

func TestRunIngestPipelineError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	srcDir := filepath.Join(tmp, "src")
	if err := os.MkdirAll(srcDir, 0o700); err != nil {
		t.Fatal(err)
	}
	hatchDir := filepath.Join(tmp, hatchDirName)
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := fmt.Sprintf(ollamaIngestYAML, srcDir)
	if err := os.WriteFile(filepath.Join(hatchDir, hatchConfigFile), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	orig, set := os.LookupEnv(envHatchDBPath)
	os.Unsetenv(envHatchDBPath)
	if set {
		defer os.Setenv(envHatchDBPath, orig)
	} else {
		defer os.Unsetenv(envHatchDBPath)
	}
	origRun := pipelineRun
	pipelineRun = func(ctx context.Context, src source.Fetcher, chk chunker.Chunker, emb embedder.Embedder, st store.VecStore, ch chan<- pipeline.Progress) error {
		return fmt.Errorf("forced pipeline error")
	}
	defer func() { pipelineRun = origRun }()
	err := runIngest(context.Background(), "docs")
	if err == nil {
		t.Error("expected error from pipeline")
	}
}

func TestRunIngestSuccess(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	srcDir := filepath.Join(tmp, "src")
	if err := os.MkdirAll(srcDir, 0o700); err != nil {
		t.Fatal(err)
	}
	hatchDir := filepath.Join(tmp, hatchDirName)
	if err := os.MkdirAll(hatchDir, 0o700); err != nil {
		t.Fatal(err)
	}
	yaml := fmt.Sprintf(ollamaIngestYAML, srcDir)
	if err := os.WriteFile(filepath.Join(hatchDir, hatchConfigFile), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	orig, set := os.LookupEnv(envHatchDBPath)
	os.Unsetenv(envHatchDBPath)
	if set {
		defer os.Setenv(envHatchDBPath, orig)
	} else {
		defer os.Unsetenv(envHatchDBPath)
	}
	origRun := pipelineRun
	pipelineRun = func(ctx context.Context, src source.Fetcher, chk chunker.Chunker, emb embedder.Embedder, st store.VecStore, ch chan<- pipeline.Progress) error {
		return nil
	}
	defer func() { pipelineRun = origRun }()
	err := runIngest(context.Background(), "docs")
	if err != nil {
		t.Errorf("runIngest: %v", err)
	}
}
