<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [milestone, ingest, pipeline, go, architecture]
description:  "Walkthrough of Milestone 2 — the ingestion pipeline: filesystem source, chunkers, OpenAI embedder, sqlite-vec KNN store, and CLI commands."
-->
# Milestone 2 — Ingestion Pipeline

Walkthrough of the second Go milestone: building the full fetch → chunk → embed → upsert
pipeline that indexes local codebases and documentation into the vector store.

---

## Overview

Milestone 2 adds the first user-visible feature: `hatch ingest`. A developer points hatch
at a local directory, and the pipeline walks every file, splits it into chunks, generates
embeddings via OpenAI, and stores the vectors in SQLite for later retrieval.

By the end of this milestone:

- `hatch ingest --source=<name>` indexes a configured source end-to-end
- `hatch sources list` and `hatch sources remove` manage the source registry
- All pipeline stages are interface-driven and fully testable with fakes
- SQLite uses sqlite-vec KNN for fast approximate nearest-neighbour search

---

## Checklist

- [x] Filesystem source: directory walker with `.gitignore` support and binary-file skip
- [x] Markdown chunker: heading-based recursive split (H1 / H2 / H3 boundaries)
- [x] Code chunker: fixed-size sliding window with configurable overlap
- [x] OpenAI embedder: batched API calls, `text-embedding-3-small` default
- [x] Ingestion pipeline: `Run(ctx, source, chunker, embedder, store, progressCh)`
- [x] `VecStore` interface extending `Store` with `Upsert` and `DeleteBySource`
- [x] sqlite-vec migration (`002_vec.sql`): `vec0` virtual table for KNN search
- [x] Replace brute-force `TopK` with sqlite-vec KNN in `Store.Search`
- [x] Swap SQLite driver: `modernc.org/sqlite` → `mattn/go-sqlite3` (CGO) for vec extension
- [x] Config: extend with `sources []SourceConfig` and `openai_api_key`
- [x] CLI: `hatch ingest --source=<name>`, `hatch sources list`, `hatch sources remove --name=<name>`
- [x] Fakes: `source/fake`, `store/fake` for pipeline tests

---

## Package Layout

```text
internal/source/fs/         Filesystem Fetcher — walks a directory, respects .gitignore,
                            skips hidden dirs and binary files, returns relative-path IDs.
internal/chunker/markdown/  Heading-based chunker — splits on H1/H2/H3 boundaries;
                            chunk ID is "docID#heading-anchor".
internal/chunker/code/      Fixed-size sliding window chunker — configurable window size
                            and overlap; chunk metadata carries line ranges.
internal/embedder/openai/   OpenAI Embeddings API client — batches chunks, retries on
                            rate limit, defaults to text-embedding-3-small.
internal/pipeline/          Run() — orchestrates fetch → chunk → embed → upsert with an
                            optional progress channel for reporting chunk counts.
internal/store/vecstore.go  VecStore interface — extends Store with Upsert and DeleteBySource.
internal/store/fake/        In-memory VecStore test double that records call counts.
internal/source/fake/       Configurable Fetcher test double that supports error injection.
cmd/hatch/ingest.go         hatch ingest --source=<name> CLI command.
cmd/hatch/sources.go        hatch sources list / remove subcommands.
```

---

## 1. Filesystem Source (`internal/source/fs/`)

The `Fetcher` walks a directory recursively using `filepath.WalkDir`. It applies three
filters before returning a document:

1. **Hidden directories** — any directory prefixed with `.` is skipped entirely (e.g.
   `.git`, `.claude`).
2. **`.gitignore` patterns** — the fetcher loads the root `.gitignore` and skips any
   path that matches a pattern.
3. **Binary files** — the first 512 bytes of each file are sniffed with
   `http.DetectContentType`; files that are not `text/*` or `application/json` are skipped.

Document IDs are the file's path relative to the root directory, making them stable across
different machines.

---

## 2. Chunkers

### Markdown (`internal/chunker/markdown/`)

Splits on heading boundaries (lines beginning with `#`, `##`, or `###`). Each heading
starts a new chunk; the heading text becomes the chunk's section label and contributes to
the anchor portion of the chunk ID (`docID#section-anchor`). Plain text documents with no
headings are returned as a single chunk.

### Code (`internal/chunker/code/`)

Uses a fixed-size sliding window measured in bytes. A configurable overlap keeps context
across boundaries — consecutive chunks share `overlap` bytes so that a construct that spans
a boundary appears (partially) in both. An overlap equal to or greater than the window size
is rejected at construction time. Each chunk carries `lines_start` and `lines_end` metadata
for display purposes.

**Chunker dispatch** — the CLI selects the chunker by file extension:

| Extension                     | Chunker    |
| ----------------------------- | ---------- |
| `.go`, `.ts`, `.tsx`, `.scss` | `code`     |
| All others                    | `markdown` |

---

## 3. OpenAI Embedder (`internal/embedder/openai/`)

Calls the [OpenAI Embeddings API](https://platform.openai.com/docs/guides/embeddings) in
batches. The default model is `text-embedding-3-small` (1536 dimensions). Construction
requires a non-empty API key — no lazy validation.

Configuration via `~/.hatch/config.yaml`:

```yaml
openai_api_key: sk-...
embed_provider: openai
```

---

## 4. Ingestion Pipeline (`internal/pipeline/`)

`Run` is the single entry point:

```go
func Run(ctx context.Context, cfg Config) error
```

Stages execute sequentially per-document:

1. **Fetch** — `source.Fetch(ctx)` returns `[]Document`
2. **Chunk** — each document is split by the appropriate `Chunker`
3. **Embed** — chunks are batched and passed to `Embedder.Embed(ctx, texts)`
4. **Upsert** — embedded records are written to `VecStore.Upsert(ctx, records)`

An optional `ProgressCh chan<- int` receives chunk counts after each embed batch so callers
can display a progress bar. Passing `nil` is safe.

Context cancellation is respected at the fetch boundary — if the context is cancelled
before a document is fetched, the pipeline stops and surfaces the context error.

---

## 5. sqlite-vec Store (`internal/store/sqlite/`)

### Driver swap

`modernc.org/sqlite` (pure Go) was replaced by `mattn/go-sqlite3` (CGO) to gain access to
the [sqlite-vec](https://github.com/asg017/sqlite-vec) extension, which cannot be loaded
into the pure-Go driver.

### Migration `002_vec.sql`

Creates a `vec0` virtual table that indexes the float32 embedding vectors:

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS chunk_vecs USING vec0(
  chunk_id TEXT PRIMARY KEY,
  embedding FLOAT[1536]
);
```

### KNN search

`Store.Search` now issues a single KNN query against `chunk_vecs` rather than scanning
all rows:

```sql
SELECT chunk_id, distance
FROM chunk_vecs
WHERE embedding MATCH ?
ORDER BY distance
LIMIT ?
```

Results are joined back to the `chunks` table to retrieve text and metadata.

### `VecStore` interface

`internal/store/vecstore.go` defines:

```go
type VecStore interface {
  Store
  Upsert(ctx context.Context, records []Record) error
  DeleteBySource(ctx context.Context, source string) error
}
```

`Upsert` writes to both `chunks` and `chunk_vecs` in a single transaction.
`DeleteBySource` removes all records for a given source name.

---

## 6. CLI Commands

### `hatch ingest --source=<name>`

Looks up the named source in `~/.hatch/config.yaml`, opens the SQLite store, builds the
pipeline, and calls `Run`. Progress is reported to stdout.

### `hatch sources list`

Prints all configured sources with their `name` and `path`.

### `hatch sources remove --name=<name>`

Removes all indexed records for the named source (`DeleteBySource`) and removes the source
entry from the config file.

---

## Verification

```bash
# Unit tests — all pipeline stages covered with fakes
go test ./internal/source/fs/...
go test ./internal/chunker/...
go test ./internal/embedder/fake/...
go test ./internal/pipeline/...
go test ./internal/store/sqlite/...

# Full suite
go test ./...

# End-to-end (requires OpenAI API key)
hatch ingest --source=myproject
hatch sources list
```

---

## Technologies

| Technology          | Role in M2                                              |
| ------------------- | ------------------------------------------------------- |
| `mattn/go-sqlite3`  | CGO SQLite driver required for sqlite-vec extension     |
| sqlite-vec          | KNN vector search inside SQLite via `vec0` virtual table |
| OpenAI Embeddings   | `text-embedding-3-small` — 1536-dim dense vectors       |
| `go-gitignore`      | `.gitignore` pattern matching in filesystem source      |
| `net/http`          | `DetectContentType` for binary-file detection           |

---

## Related

- [`docs/ROADMAP.md`](../ROADMAP.md) — full milestone plan
- [`docs/milestones/M1-foundation.md`](M1-foundation.md) — interfaces established in M1
- [`docs/TESTING.md`](../TESTING.md) — how to run tests
- [`docs/testing/E2E.md`](../testing/E2E.md) — manual ingest CLI testing
- [`docs/testing/COVERAGE.md`](../testing/COVERAGE.md) — per-package test inventory
