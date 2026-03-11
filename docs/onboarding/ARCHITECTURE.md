<!--
created_by:   jazicorn-tw
created_date: 2026-03-11
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [onboarding, dev, planning]
description:  "High-level architecture overview for new contributors"
-->
# Hatch — Architecture Overview

A high-level map of the codebase for new contributors.

---

## What Hatch is

Hatch is a **developer onboarding tool** delivered as a single Go binary.

- **Juniors** SSH into a server and take quizzes through a terminal UI
- **Seniors** query scores and progress from the CLI or a web dashboard
- The binary embeds everything — no separate frontend server or database server required

---

## Tech stack

| Concern       | Choice                                                               |
| ------------- | -------------------------------------------------------------------- |
| Language      | Go 1.26+                                                             |
| TUI           | Charmbracelet: Bubble Tea, Huh, Glamour, Bubbles, Lip Gloss          |
| Database      | SQLite + sqlite-vec (WAL mode for concurrent SSH connections)        |
| LLM           | Provider-agnostic interface → Anthropic, OpenAI, Ollama              |
| Embeddings    | Provider-agnostic interface → same providers                         |
| Config        | Viper (YAML) + Cobra (CLI)                                           |
| SSH server    | Charmbracelet Wish — per-connection Bubble Tea session               |
| Web dashboard | React + Vite frontend embedded in Go binary via `//go:embed`         |

---

## Repository layout

```text
hatch/
├── cmd/
│   └── hatch/          # CLI entry point (main.go + Cobra commands)
├── internal/
│   ├── agent/          # Agent interfaces and orchestrator
│   ├── api/            # HTTP REST API handlers + embedded web assets
│   ├── chunker/        # Document chunking strategies
│   ├── config/         # Viper config struct, Load(), Validate()
│   ├── embed/          # Embedding providers (OpenAI, Ollama, fake)
│   ├── ingest/         # Ingestion pipeline and sources (filesystem, web)
│   ├── llm/            # LLM providers (Anthropic, OpenAI, Ollama, fake)
│   ├── quiz/           # Quiz engine, generator, evaluator, session model
│   ├── server/         # Wish SSH server and per-connection handler
│   ├── store/          # Storage interface + SQLite and in-memory impls
│   ├── tui/            # Bubble Tea models, screens, styles, messages
│   └── users/          # User identity, roles, SSH key fingerprinting
├── scripts/            # Local dev scripts (doctor, bootstrap, hooks, etc.)
├── web/                # React + Vite frontend (builds into internal/api/static/dist/)
├── docs/               # Project documentation
├── dev                 # gum-powered task runner
└── go.mod
```

---

## Key design decisions

- **SQLite only** — no external database server required for dev or tests ([ADR-001](../adr/ADR-001-database-postgresql.md))
- **In-memory SQLite for tests** — fast, isolated, no Docker needed ([ADR-002](../adr/ADR-002-testcontainers.md))
- **Single binary** — web assets embedded via `//go:embed`; one file ships everything
- **Provider-agnostic LLM/embeddings** — swap providers via config, no code changes
- **CI is authoritative** — `./dev` is a local convenience, never a CI replacement ([ADR-000](../adr/ADR-000-linting.md))
- **Phased security** — auth scaffolded early, enforcement deferred to Phase 7 ([ADR-005](../adr/ADR-005-security-phased.md))

---

## Data flow

```text
Source (filesystem / URL)
  └──► Chunker
         └──► Embedder  ──► sqlite-vec store
                                └──► Search (TopK)
                                       └──► LLM (question generation)
                                              └──► Quiz session
                                                     └──► SQLite (results)
```

---

## Related

- [`docs/adr/`](../adr/) — architecture decision records
- [`docs/onboarding/PROJECT_SETUP.md`](PROJECT_SETUP.md) — local setup
- [`docs/ROADMAP.md`](../ROADMAP.md) — what's built and what's next
