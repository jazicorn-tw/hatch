<!--
created_by:   jazicorn-tw
created_date: 2026-03-12
updated_by:   jazicorn-tw
updated_date: 2026-03-13
status:       active
tags:         [milestone, foundation, go, architecture]
description:  "Walkthrough of Milestone 1 — scaffolding the Go package structure, core interfaces, config layer, SQLite store, and test infrastructure."
-->
# Milestone 1 — Foundation

Walkthrough of the first Go milestone: establishing the package layout, defining core
abstractions, wiring configuration, and building the test infrastructure that all future
milestones depend on.

---

## Overview

Milestone 1 creates the skeleton of the hatch binary. No user-visible features are shipped —
the goal is a clean, compilable codebase where every future piece of work has a well-defined
home and a stable interface to program against.

By the end of this milestone:

- `hatch` builds as a single binary from `cmd/hatch/`
- All core domain interfaces are defined in `internal/`
- Configuration loads from `~/.hatch/config.yaml` with environment variable overrides
- SQLite opens with WAL mode and runs schema migrations automatically
- Tests can run without real LLM or embedding providers

---

## Checklist

- [x] Scaffold all `cmd/hatch/` and `internal/` packages
- [x] Config layer: Viper + `~/.hatch/config.yaml`, env var overrides, `hatch config init`
- [x] Core interfaces: `Source`, `Chunker`, `Embedder`, `LLM`, `Store`, `Agent`
- [x] SQLite schema + migration runner; WAL mode on open
- [x] In-memory store (`internal/store/memory/`) for tests
- [x] Fake embedder + fake LLM for tests

---

## Package Layout

```text
cmd/hatch/            Entry point. Wires cobra commands and calls Execute().
internal/agent/       Agent interface — orchestrates quiz/kata session lifecycle.
internal/chunker/     Chunker interface — splits Documents into indexable Chunks.
internal/config/      Viper-based config loader; hatch config init subcommand.
internal/embedder/    Embedder interface + fake/ test double.
internal/llm/         LLM interface + fake/ test double.
internal/source/      Source interface — fetches Documents from an origin.
internal/store/       Store interface + memory/ and sqlite/ implementations.
```

The `internal/` boundary means none of these packages are importable by external code —
everything is intentionally encapsulated until a public API is needed.

---

## 1. Entry Point (`cmd/hatch/`)

`main.go` is the only file in `cmd/hatch/`. It constructs a cobra root command, registers
subcommands, and calls `Execute()`. Keeping `main.go` thin means the binary's behaviour
is fully testable through the subcommand layer without invoking the binary directly.

The only subcommand added in this milestone is `hatch config init`.

---

## 2. Configuration (`internal/config/`)

Configuration is handled by [Viper](https://github.com/spf13/viper), which provides a
priority-ordered lookup across multiple sources:

1. Environment variables (highest priority)
2. `~/.hatch/config.yaml`
3. Compiled-in defaults (lowest priority)

The environment variable prefix is `HATCH_`. A key like `llm_provider` in the config file
maps to `HATCH_LLM_PROVIDER` as an env var override. This follows the twelve-factor app
convention and makes the binary easy to configure in CI or container environments without
touching the config file.

`hatch config init` writes a default `~/.hatch/config.yaml` if one does not already exist.
Running it twice is safe — it exits early if the file is present.

**Config keys** (from `.env.example`):

| Key              | Env var                | Default             |
| ---------------- | ---------------------- | ------------------- |
| `llm_provider`   | `HATCH_LLM_PROVIDER`   | `anthropic`         |
| `embed_provider` | `HATCH_EMBED_PROVIDER` | `ollama`            |
| `ssh_port`       | `HATCH_SSH_PORT`       | `2222`              |
| `http_port`      | `HATCH_HTTP_PORT`      | `8080`              |
| `web_password`   | `HATCH_WEB_PASSWORD`   | `changeme`          |
| `jwt_secret`     | `HATCH_JWT_SECRET`     | _(empty)_           |
| `db_path`        | `HATCH_DB_PATH`        | `~/.hatch/hatch.db` |

---

## 3. Core Interfaces (`internal/`)

All domain behaviour is expressed as Go interfaces rather than concrete types. This is the
central design decision of Milestone 1 — it means:

- Implementations can be swapped without changing call sites (e.g. swap OpenAI for Ollama)
- Tests can run with lightweight fakes instead of real network calls
- Future milestones add implementations without touching existing ones

### `Source`

Fetches raw documents from an origin — a local directory, a URL, a database, etc.
Returns a slice of `Document` values.

### `Chunker`

Splits a `Document` into smaller `Chunk` values that are suitable for embedding. Different
chunking strategies (heading-based, fixed-size with overlap) implement the same interface.

### `Embedder`

Converts a batch of text strings into dense float32 vectors. The caller does not need to
know whether the vectors come from OpenAI, Ollama, or a fake.

### `LLM`

Takes a prompt string and returns a completion string. Provider-agnostic by design.

### `Store`

Indexes `Record` values (a `Chunk` paired with its embedding vector) and supports nearest-
neighbour search. Two implementations ship in this milestone: `memory` and `sqlite`.

### `Agent`

Top-level orchestrator. Receives a context and runs a session to completion. Implemented
in later milestones.

---

## 4. SQLite Store (`internal/store/sqlite/`)

The SQLite store uses [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — a pure
Go port that requires no CGO and no system SQLite installation. This keeps the binary
self-contained and CI straightforward.

**WAL mode** is enabled on every connection via a DSN pragma (`_journal_mode=WAL`). WAL
allows concurrent readers alongside a single writer, which becomes important in Milestone 6
when multiple SSH sessions access the database simultaneously.

**Migration runner** — on open, the store reads all `*.sql` files embedded in
`internal/store/sqlite/migrations/` and applies any that have not been recorded in the
`schema_migrations` table. Migrations are applied in filename order, so they must be named
with a numeric prefix (e.g. `001_init.sql`, `002_add_index.sql`).

**Initial schema** (`001_init.sql`) creates:

- `schema_migrations` — tracks applied migration filenames and timestamps
- `chunks` — stores chunk text, source, and embedding blob alongside a timestamp

Vector search in the SQLite store is currently brute-force cosine similarity over all rows.
This will be replaced with [sqlite-vec](https://github.com/asg017/sqlite-vec) KNN search in
Milestone 2.

---

## 5. In-Memory Store (`internal/store/memory/`)

The memory store holds records in a plain slice and performs brute-force cosine similarity
search on every `Search` call. It has no persistence and resets when the process exits.

Its purpose is unit tests. Any test that needs a `store.Store` can use `memory.New()`
without touching the filesystem or running migrations.

---

## 6. Test Fakes

### `internal/embedder/fake/`

`FakeEmbedder` satisfies the `Embedder` interface by returning zero vectors of a
configurable dimension. It never makes a network call. Tests that need realistic vectors
can set specific values after calling `Embed`.

### `internal/llm/fake/`

`FakeLLM` satisfies the `LLM` interface by returning a configurable string. The default
response is `"fake response"`. Tests that check prompt-response logic can set `Response`
to whatever the scenario requires.

---

## Verification

```bash
go build ./...   # zero errors — full package graph compiles
go test ./...    # all tests pass
```

After confirming tests pass, remove `ENABLE_GO_ANALYSIS=FALSE` from GitHub repo variables
(or set it to `TRUE`) so the `quality` and `test` CI jobs run on future pushes.

---

## Technologies

| Technology             | Role in M1                                           |
| ---------------------- | ---------------------------------------------------- |
| Go 1.26                | Implementation language                              |
| Cobra                  | CLI framework — root command and subcommands         |
| Viper                  | Config loading, env var binding, YAML parsing        |
| modernc.org/sqlite     | Pure-Go SQLite driver, no CGO required               |
| SQLite WAL mode        | Embedded database with concurrent-read support       |
| Go `embed` package     | Bundles SQL migration files into the binary          |
| Go `database/sql`      | Standard database interface used by the SQLite store |
| Go `sync` package      | `RWMutex` for thread-safe in-memory store            |
| Go `testing` package   | Standard library unit test framework                 |

---

## Related

- [`docs/ROADMAP.md`](../ROADMAP.md) — full milestone plan
- [`docs/adr/ADR-010-local-ci-simulation-with-act.md`](../adr/ADR-010-local-ci-simulation-with-act.md) — running CI locally
- [`docs/devops/CI_VARIABLES.md`](../devops/CI_VARIABLES.md) — `ENABLE_GO_ANALYSIS` and other CI gates

---

## Resources

### Go Fundamentals

- [Effective Go](https://go.dev/doc/effective_go) — idiomatic Go patterns, including interfaces
- [Go by Example](https://gobyexample.com) — concise annotated examples for every core feature
- [The Go Programming Language (Donovan & Kernighan)](https://www.gopl.io) — definitive book

### Interfaces and Design

- [Go interfaces explained](https://jordanorelli.com/post/32665860244/how-to-use-interfaces-in-go) — Jordan Orelli's classic walkthrough
- [Interface satisfaction and the nil interface](https://go.dev/doc/faq#nil_error) — Go FAQ entry on a common gotcha
- [Accept interfaces, return structs](https://bryanftan.medium.com/accept-interfaces-return-structs-in-go-d4cab29a301b) — widely-cited design guideline

### Configuration

- [Viper documentation](https://github.com/spf13/viper#readme) — full feature reference
- [Cobra documentation](https://github.com/spf13/cobra#readme) — CLI framework used for `cmd/hatch/`
- [The Twelve-Factor App — Config](https://12factor.net/config) — rationale for env-var-based config

### SQLite and WAL

- [SQLite WAL mode](https://www.sqlite.org/wal.html) — official documentation on write-ahead logging
- [SQLite pragma reference](https://www.sqlite.org/pragma.html) — full list of connection pragmas including `journal_mode`
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — pure-Go SQLite driver used in this project
- [Database migrations in Go](https://pressly.github.io/goose/) — goose (not used here, but a common alternative pattern)

### Testing

- [Go testing package](https://pkg.go.dev/testing) — standard library reference
- [Test doubles in Go](https://go.dev/blog/testable-examples) — Go blog post on writing testable code
- [Table-driven tests](https://dave.cheney.net/2013/06/09/writing-table-driven-tests-in-go) — Dave Cheney on the idiomatic Go test pattern
