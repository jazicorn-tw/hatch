<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [adr, build, tooling, runtime]
description:  "ADR-013: Use Go as the implementation language"
-->
# ADR-013: Use Go as the Implementation Language

- **Status:** Accepted
- **Date:** 2026-03-14
- **Deciders:** Project maintainers
- **Scope:** Entire hatch codebase — CLI, TUI, ingestion pipeline, SSH server, and web API

---

## Context

Hatch is a developer onboarding tool distributed as a single binary. It needs to:

- Run as a CLI and SSH TUI on developer machines with no runtime installation
- Embed a database, migrations, and static assets in the binary
- Serve an SSH server and optionally an HTTP API from the same process
- Execute on macOS, Linux, and potentially Windows
- Be maintainable by a small team without a large dependency ecosystem

The language choice constrains nearly every other architectural decision — runtime
dependencies, distribution model, library ecosystem, and concurrency model.

---

## Decision

Implement hatch entirely in **Go**.

---

## Alternatives Considered

### Python

- Natural fit for the LLM/embedding ecosystem (LangChain, LlamaIndex, ChromaDB, OpenAI SDK)
- Requires a Python runtime on every developer's machine — breaks the single-binary goal
- Packaging a self-contained Python binary (PyInstaller, Nuitka) is fragile and produces
  large, slow-starting executables
- No first-class TUI library comparable to Bubble Tea / Lip Gloss
- GIL limits true parallelism for concurrent SSH sessions

### Node.js / TypeScript

- Strong ecosystem for HTTP APIs and tooling scripts
- Requires Node runtime — same distribution problem as Python
- No idiomatic SSH server library; no embedded SQLite with vector extension support
- Less suited for long-running server processes and systems-level work

### Rust

- Comparable distribution model (static binary, no runtime)
- Excellent performance and memory safety guarantees
- Steeper learning curve; smaller pool of contributors familiar with the language
- TUI and SSH server ecosystems less mature than Go's at project start
- Compile times longer, which increases iteration speed cost

---

## Consequences

### Positive

- **Single binary distribution** — `go build` produces one executable with no runtime
  dependency; users install with `go install` or download a release artifact
- **`//go:embed`** — SQL migrations, prompt templates, and web assets are bundled at
  compile time, keeping deployment trivial
- **Bubble Tea / Lip Gloss / Wish** — mature, idiomatic Go libraries for TUI and SSH server
  that align directly with hatch's feature set
- **Goroutines** — lightweight concurrency handles multiple simultaneous SSH sessions
  (Milestone 6) and parallel pipeline stages without callback complexity
- **`database/sql` + sqlite-vec via CGO** — first-class embedded database with vector
  search in the same binary
- **Fast compile times** — tight edit-compile-test loop without a separate build step
- **Cross-compilation** — `GOOS=linux GOARCH=amd64 go build` produces Linux binaries from
  macOS with no toolchain changes (CGO cross-compilation requires additional setup)
- **Strong standard library** — HTTP server, file I/O, crypto, and testing are all stdlib;
  fewer third-party dependencies to audit

### Trade-offs

- LLM/embedding ecosystem is Python-first; Go SDKs (OpenAI, Anthropic) are community-
  maintained and lag behind the official Python clients in features
- CGO (required for sqlite-vec) complicates cross-compilation and disables pure-Go builds
- Generics support is relatively recent (Go 1.18); some patterns common in other languages
  require more boilerplate in Go

---

## Related ADRs

- ADR-001: Use SQLite + sqlite-vec for embedded persistence
- ADR-006: Local developer experience (no external services required)
- ADR-012: Use sqlite-vec for vector search instead of a dedicated vector database
