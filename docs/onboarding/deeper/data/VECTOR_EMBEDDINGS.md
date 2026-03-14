<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [onboarding, embedder, store, sqlite]
description:  "What vector embeddings are, how KNN search works, and how hatch stores and queries them"
-->
# Vector Embeddings — What They Are and How Hatch Uses Them

A vector embedding is a list of numbers that represents the meaning of a piece of text.
Similar texts produce similar number lists. That similarity is what makes semantic search work.

---

## From text to numbers

A sentence like `"how do I set up SSH keys"` gets converted by an embedding model into a
list of floats — for example, 1536 of them for OpenAI `text-embedding-3-small`, or 768 for
Google Gemini `text-embedding-004`.

```text
"how do I set up SSH keys"
  → [0.021, -0.104, 0.083, 0.197, ..., -0.042]   (1536 numbers)
```

The model is trained so that texts with similar meanings land close together in this
high-dimensional space. `"configure SSH authentication"` and `"set up SSH keys"` would
produce vectors that are very close to each other. `"what is the weather today"` would be
far away.

---

## KNN search — finding the nearest neighbours

Given a query vector, **KNN (K-Nearest Neighbours) search** finds the k stored vectors
that are most similar to it. "Similar" is measured by cosine similarity — roughly, the
angle between two vectors. Cosine = 1 means identical direction (most similar),
cosine = 0 means completely unrelated.

```text
query: "how do I set up SSH keys"   → query vector q
store: 500 chunks already embedded

KNN(q, k=5) → top 5 chunks whose embeddings are closest to q
```

Those 5 chunks get passed to the LLM as context so it can generate a question or answer
that is grounded in the actual codebase docs.

---

## How hatch stores embeddings

### The `chunks` table

`001_init.sql` creates the main text table:

```sql
CREATE TABLE IF NOT EXISTS chunks (
    id          TEXT PRIMARY KEY,
    source      TEXT NOT NULL,
    text        TEXT NOT NULL,
    embedding   BLOB,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);
```

`embedding` is the raw float array stored as a binary blob. The `source` column is the
name of the ingested source (e.g. `"hatch-docs"`), used to scope deletes when you
re-ingest.

### The `vec_chunks` virtual table

`002_vec.sql` creates the vector search index powered by **sqlite-vec**:

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS vec_chunks USING vec0(
    chunk_id TEXT PRIMARY KEY,
    embedding float[1536]
);
```

`vec0` is the virtual table type provided by the sqlite-vec C extension. It maintains an
index optimised for KNN queries. `float[1536]` declares the dimension — it must match
the output of whichever embedding provider you use.

When you call `Search(ctx, queryVec, k)`, hatch queries `vec_chunks` for the top k
nearest `chunk_id` values, then fetches the full text from `chunks`.

---

## The ingestion flow

```text
Source (filesystem)
  └──► Chunker        splits docs into ~500-token chunks
         └──► Embedder   calls OpenAI / Gemini / Ollama API
                └──► Store.Upsert()   writes to chunks + vec_chunks
```

`pipeline.Run` in `internal/pipeline/` wires these stages together. The
`store.VecStore` interface exposes `Upsert` (insert-or-replace keyed on chunk ID)
and `DeleteBySource` (remove all chunks from a source before re-ingesting).

---

## The query flow

```text
User quiz question
  └──► Embed query text   (same model as ingestion)
         └──► VecStore.Search(vec, k=5)   → top 5 chunks
                └──► LLM (question generator)   with chunks as context
```

The embedding model used at query time **must match** the one used at ingestion — you
can't mix a Gemini-produced 768-dim query against OpenAI-produced 1536-dim stored vectors.

---

## Embedding dimensions by provider

| Provider | Model                    | Dimensions |
| -------- | ------------------------ | ---------- |
| OpenAI   | `text-embedding-3-small` | 1536       |
| Gemini   | `text-embedding-004`     | 768        |
| Ollama   | depends on model         | varies     |

If you switch providers, you need to drop and re-create `vec_chunks` with the new
dimension and re-ingest all sources.

---

## Related

- [`CGO.md`](../go/CGO.md) — why sqlite-vec requires a C compiler
- [`INTERFACES.md`](../go/INTERFACES.md) — the `Embedder`, `Store`, and `VecStore` interfaces
- [`SQLITE_WAL.md`](SQLITE_WAL.md) — how SQLite handles concurrent reads during search

## Resources

- [sqlite-vec](https://github.com/asg017/sqlite-vec) — the vector extension used by hatch
- [OpenAI embeddings guide](https://platform.openai.com/docs/guides/embeddings) — how text-embedding-3-small works
- [Gemini embeddings](https://ai.google.dev/gemini-api/docs/embeddings) — text-embedding-004 reference
- [Cosine similarity explained](https://en.wikipedia.org/wiki/Cosine_similarity) — the math behind KNN scoring
