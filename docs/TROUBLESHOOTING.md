<!--
created_by:   jazicorn-tw
created_date: 2026-03-15
updated_by:   jazicorn-tw
updated_date: 2026-03-15
status:       active
tags:         [errors, ingest, quiz, sqlite, embedder, llm, ollama]
description:  "Known issues and fixes encountered running hatch locally."
-->
# Troubleshooting

Common errors and their fixes when running `hatch` locally.

---

## Ingest / Quiz

### `openai embedder: api key is required` when `HATCH_EMBED_PROVIDER=ollama`

```text
Error: quiz: create embedder: openai embedder: api key is required
```

The `newEmbedder` / `newQuizEmbedder` switch statement was missing a `case "ollama"` and fell
through to the OpenAI default.

The `"ollama"` case now exists in both switch statements and routes to
`internal/embedder/ollama`. No API key is required for the Ollama embedder.

---

### `sqlite: ping: unable to open database file: no such file or directory`

```text
Error: quiz: open store: sqlite: ping: unable to open database file: no such file or directory
```

`HATCH_DB_PATH=~/.hatch/hatch.db` — Go does not expand `~`, so the path was taken literally.
The `~/.hatch/` directory also may not exist until `hatch config init` is run.

`resolveDBPath()` in `cmd/hatch/ingest.go` now expands a leading `~/` to the real home
directory via `os.UserHomeDir()` and calls `os.MkdirAll` on the parent directory before
opening the store.

---

### `Dimension mismatch … Expected 1536 dimensions but received 768`

```text
Error: quiz: generate: quiz generator: search: sqlite: rows:
Dimension mismatch for query vector for the "embedding" column.
Expected 1536 dimensions but received 768.
```

or during ingest:

```text
Error: ingest: pipeline: upsert …: sqlite: insert vec …:
Dimension mismatch for inserted vector for the "embedding" column.
Expected 1536 dimensions but received 768.
```

The database was previously created with an OpenAI embedder (1536 dims). Switching to Ollama's
`nomic-embed-text` (768 dims) produces vectors that don't match the schema. The `vec_chunks`
virtual table schema is fixed at creation time. Additionally,
`internal/store/sqlite/migrations/002_vec.sql` had `float[1536]` hardcoded.

`002_vec.sql` has been updated to `float[768]` (matching `nomic-embed-text`). Delete the old
database and re-ingest so the schema is recreated:

```bash
rm ~/.hatch/hatch.db
./dev build run ingest --source <source-name>
```

> If you switch embed providers again, repeat this process — the vector dimensions must match
> across ingest and query.

---

### `ollama embed: api error: the input length exceeds the context length`

```text
Error: ingest: pipeline: embed <file>: ollama embed: api error:
the input length exceeds the context length
```

`nomic-embed-text` has an 8192-token context limit. Files that are not `.md`/`.mdx` were
routed to the markdown chunker, which returns the entire file as a single chunk when no
headings are found — e.g. `.json`, `.mk`, `.yaml`, `.xml`.

The dispatch logic in `cmd/hatch/ingest.go` has been inverted: only `.md` and `.mdx` use the
heading-based markdown chunker. All other file types use the sliding-window code chunker
(50-line windows, 10-line overlap), which always produces context-safe chunks.

---

## Quiz quality

### Quiz questions are all about the same file (e.g. only Makefiles)

The KNN search returns the top-k closest vectors regardless of which file they come from.
If several chunks from the same file score highest, all questions end up about that file.

The generator now fetches `TopK × 4` candidates and applies a diversity filter
(`diversifyBySource`) that keeps at most one chunk per file, ensuring the LLM has context
from up to 10 different files.

---

### Quiz topic doesn't match ingested content

Asking about `"Go interfaces"` on a Java project returns Makefile questions; asking about
topics unrelated to the codebase produces generic or hallucinated questions.

The vector search finds the closest chunks to the topic embedding. If the ingested source
doesn't contain files relevant to the topic, unrelated files (build files, config files)
score highest by default.

Use topics that match what was actually ingested. For a Java API project, topics like
`"REST API"`, `"Spring Boot"`, or `"Docker"` will return relevant chunks. The `--topic`
flag is optional — omitting it runs a general quiz across all ingested content:

```bash
./dev build run quiz --count 5                    # general quiz
./dev build run quiz --topic "REST API" --count 5 # topic-focused
```

---

## Sources

### `source "<name>" not found in config`

```text
Error: ingest: source "/path/to/project" not found in config
```

`--source` takes a **name** defined in `~/.hatch/config.yaml`, not a file path.

Add the source to `~/.hatch/config.yaml` first:

```yaml
sources:
  - name: my-project
    path: /absolute/path/to/project
    type: filesystem
```

Then ingest by name:

```bash
./dev build run ingest --source my-project
```

---

## Related

- [`docs/providers/EMBEDDER.md`](providers/EMBEDDER.md) — embedding provider configuration
- [`docs/providers/CONFIGURATION.md`](providers/CONFIGURATION.md) — full config reference
- [`docs/testing/E2E.md`](testing/E2E.md) — end-to-end CLI testing
