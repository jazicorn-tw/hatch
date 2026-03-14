<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [onboarding, go, architecture]
description:  "What Go interfaces are and every interface hatch defines — where they live, what they do, and what implements them"
-->
# Interfaces — Go's Answer to Swappable Implementations

In Go, an interface is a set of method signatures. Any type that has those methods
automatically satisfies the interface — no declaration required. This is called
**implicit satisfaction**.

```go
// Define an interface
type Greeter interface {
    Greet(name string) string
}

// This type satisfies Greeter automatically — no "implements Greeter" keyword
type EnglishGreeter struct{}
func (g EnglishGreeter) Greet(name string) string { return "Hello, " + name }

// So does this one
type SpanishGreeter struct{}
func (g SpanishGreeter) Greet(name string) string { return "Hola, " + name }
```

Code that depends on `Greeter` works with either type. Swap one for the other in one
place (usually the constructor), and nothing else changes.

---

## Why hatch uses interfaces everywhere

Hatch needs to support multiple LLM providers, multiple embedding providers, and multiple
storage backends — all switchable via config. Interfaces make this work:

- The **pipeline** doesn't know whether it's talking to OpenAI or Gemini — it holds an
  `embedder.Embedder` and calls `Embed()`
- Tests use **fake implementations** (`embedder/fake`, `store/fake`) so they run without
  network calls or a real database
- New providers can be added without changing the pipeline code

---

## All interfaces in hatch

### `source.Fetcher`

**File:** `internal/source/source.go`

```go
type Fetcher interface {
    Fetch(ctx context.Context) ([]Document, error)
}
```

Reads raw documents from a source. The only current implementation is
`internal/source/fs` — it walks a directory and returns one `Document` per file.
Future: a web fetcher (URLs), a Git fetcher.

---

### `chunker.Chunker`

**File:** `internal/chunker/chunker.go`

```go
type Chunker interface {
    Chunk(doc source.Document) ([]Chunk, error)
}
```

Splits a `Document` into smaller `Chunk`s ready for embedding. Hatch has two
implementations and one dispatcher:

| Implementation    | Package                     | What it does                                  |
| ----------------- | --------------------------- | --------------------------------------------- |
| Markdown chunker  | `internal/chunker/markdown` | Splits on headings                            |
| Code chunker      | `internal/chunker/code`     | Splits on function/class boundaries           |
| `dispatchChunker` | `cmd/hatch/ingest.go`       | Routes by file extension to the right chunker |

---

### `embedder.Embedder`

**File:** `internal/embedder/embedder.go`

```go
type Embedder interface {
    Embed(ctx context.Context, texts []string) ([][]float32, error)
}
```

Converts a batch of text strings into a batch of float vectors. Input: `n` strings.
Output: `n` vectors (one per string).

| Implementation | Package                    | Model                                |
| -------------- | -------------------------- | ------------------------------------ |
| OpenAI         | `internal/embedder/openai` | `text-embedding-3-small` (1536 dims) |
| Gemini         | `internal/embedder/gemini` | `text-embedding-004` (768 dims)      |
| Fake           | `internal/embedder/fake`   | returns zero vectors — for tests     |

---

### `store.Store`

**File:** `internal/store/store.go`

```go
type Store interface {
    Add(ctx context.Context, records []Record) error
    Search(ctx context.Context, vec []float32, k int) ([]Record, error)
    Close() error
}
```

Persists embedded chunks and retrieves the k nearest by vector. `Record` wraps a `Chunk`
with its `[]float32` embedding.

| Implementation      | Package                 | When used                 |
| ------------------- | ----------------------- | ------------------------- |
| SQLite + sqlite-vec | `internal/store/sqlite` | Production                |
| In-memory           | `internal/store/memory` | Unit tests                |
| Fake                | `internal/store/fake`   | Controlled test scenarios |

---

### `store.VecStore`

**File:** `internal/store/vecstore.go`

```go
type VecStore interface {
    Store   // embeds all three Store methods
    Upsert(ctx context.Context, records []Record) error
    DeleteBySource(ctx context.Context, source string) error
}
```

Extends `Store` with two methods needed by the ingestion pipeline:

- `Upsert` — insert-or-replace keyed on `Chunk.ID` so re-ingesting a source is
  idempotent
- `DeleteBySource` — wipe all records from a source before re-ingesting it

Only the SQLite store satisfies `VecStore`. The in-memory store only satisfies the base
`Store` interface — that's enough for unit tests that don't exercise the full pipeline.

---

### `llm.Completer`

**File:** `internal/llm/llm.go`

```go
type Completer interface {
    Complete(ctx context.Context, prompt string) (string, error)
}
```

Sends a prompt to an LLM and returns the generated text. Used by the quiz engine to
generate questions and evaluate answers. Providers: Anthropic, OpenAI, Ollama (M3).
A `fake` implementation returns canned responses for tests.

---

### `agent.Runner`

**File:** `internal/agent/agent.go`

```go
type Runner interface {
    Run(ctx context.Context) error
}
```

Orchestrates the quiz or kata session lifecycle — kicks off the Bubble Tea TUI, manages
the session loop. The SSH server in `internal/server/` holds a `Runner` and calls `Run`
for each incoming connection.

---

## Interface embedding

`VecStore` shows a Go pattern called **interface embedding**: it includes `Store` by
name, which means any type that satisfies `VecStore` must also satisfy all of `Store`.

```go
type VecStore interface {
    Store                                         // all methods of Store
    Upsert(ctx context.Context, records []Record) error
    DeleteBySource(ctx context.Context, source string) error
}
```

This lets you use a `VecStore` anywhere a `Store` is expected — the full ingestion
pipeline uses `VecStore`, but simpler components that only need search can accept `Store`.

---

## Related

- [`VECTOR_EMBEDDINGS.md`](../data/VECTOR_EMBEDDINGS.md) — how embeddings and KNN search work
- [`CGO.md`](CGO.md) — why the SQLite store requires a C compiler

## Resources

- [Go Tour: Interfaces](https://go.dev/tour/methods/9) — interactive intro to Go interfaces
- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces) — idiomatic interface usage
- [Go blog: The Laws of Reflection](https://go.dev/blog/laws-of-reflection) — deep dive into how interfaces work internally
