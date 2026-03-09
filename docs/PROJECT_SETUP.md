# Hatch — Project Plan

## Overview

**Name:** Hatch
**Type:** New standalone Go repository
**Deployment:** Server-hosted. Juniors SSH in via Charmbracelet Wish (no local install). Seniors access scores via CLI or a browser-based web dashboard. Single binary ships everything.

---

## Tech Stack

| Concern       | Choice                                                                  |
| ------------- | ----------------------------------------------------------------------- |
| Language      | Go                                                                      |
| TUI           | Charmbracelet: Bubble Tea v2, Huh v2, Glamour v2, Bubbles, Lip Gloss    |
| Vector store  | SQLite + sqlite-vec (WAL mode for concurrent SSH connections)           |
| LLM           | Provider-agnostic interface → Anthropic, OpenAI, Ollama                 |
| Embeddings    | Provider-agnostic interface → same providers                            |
| Config        | Viper (YAML) + Cobra (CLI)                                              |
| SSH server    | Charmbracelet Wish — per-connection Bubble Tea session                  |
| User identity | SSH public key fingerprint + first-login display name prompt            |
| Web dashboard | React + Vite frontend; Go `net/http` REST API; build embedded in binary |
| Web auth      | Single shared password via `HATCH_WEB_PASSWORD` env var                 |

---

## Milestones

---

### Milestone 1 — Foundation

Repo scaffold, config, interfaces, and data layer

| Task             | Description                                                                           |
| ---------------- | ------------------------------------------------------------------------------------- |
| Init repo        | `go mod init github.com/<you>/hatch`, `Makefile`, `.gitignore`, `README.md`           |
| Directory layout | Scaffold all `cmd/hatch/` and `internal/` packages                                    |
| Config layer     | Viper `Config` struct, `Load()`, `Validate()`, env var overrides, `hatch config init` |
| Core interfaces  | Define `Source`, `Chunker`, `Embedder`, `LLM`, `Store`, `Agent` in their packages     |
| SQLite schema    | All tables + migrations runner; WAL mode on open                                      |
| In-memory store  | `internal/store/memory/` — `Store` impl for tests                                     |
| Fake providers   | `internal/embed/fake/` and `internal/llm/fake/` — deterministic responses             |

**Done when:** `go build ./...` and `go test ./...` pass with stubs.

---

### Milestone 2 — Ingestion Pipeline

Local filesystem source, chunking, embedding, storage

| Task                 | Description                                                                    |
| -------------------- | ------------------------------------------------------------------------------ |
| Filesystem source    | Walk directory tree with gitignore support; emit `Document` per file           |
| Markdown chunker     | Heading-based recursive split                                                  |
| Code chunker         | Fixed-size with configurable overlap (`.ts`, `.tsx`, `.go`, `.scss`)           |
| Ingestion pipeline   | `pipeline.Run(ctx, source, chunker, embedder, store, progressCh)`              |
| OpenAI embedder      | `internal/embed/openai/` — batched API calls, `text-embedding-3-small` default |
| sqlite-vec store     | `internal/store/sqlite/` — `Upsert`, cosine KNN `Search`, `DeleteBySource`     |
| CLI: `hatch ingest`  | Headless ingestion, prints chunk count + elapsed time                          |
| CLI: `hatch sources` | `list` and `remove <name>`                                                     |

**Done when:** `hatch ingest --source=my-project` indexes a directory and reports chunk count.

---

### Milestone 3 — Quiz Engine

RAG retrieval, question generation, answer evaluation, session tracking

| Task              | Description                                                                             |
| ----------------- | --------------------------------------------------------------------------------------- |
| Anthropic LLM     | `internal/llm/anthropic/` — `claude-sonnet-4-6` default                                 |
| Question types    | `Question{Text, Options[4], CorrectIndex, Explanation, SourceChunks}`                   |
| Quiz generator    | Embed topic probe → `Store.Search(TopK=5)` → LLM MCQ prompt → `Question`                |
| Prompt templates  | `internal/quiz/prompts/` via `//go:embed`; `question_mcq.tmpl`, `question_explain.tmpl` |
| Answer evaluator  | Deterministic index comparison for MCQ                                                  |
| Session model     | `Session{ID, UserID, SourceIDs, Questions, Score}` + SQLite persistence                 |
| CLI: `hatch quiz` | Headless quiz run — no TUI, for local admin use                                         |

**Done when:** `hatch quiz --source=my-project --count=5` produces 5 scored MCQ questions.

---

### Milestone 4 — TUI

Full Charmbracelet TUI wired to the quiz engine

| Task                 | Description                                                                             |
| -------------------- | --------------------------------------------------------------------------------------- |
| Styles               | `internal/tui/styles/` — Lip Gloss palettes, dark/light/dracula themes                  |
| Message types        | `internal/tui/msgs/msgs.go` — all `tea.Msg` types                                       |
| Root app model       | `app.go` — state machine, delegates to active sub-model                                 |
| Welcome screen       | ASCII wordmark + `huh.NewSelect` (Start Quiz / Configure Sources / Progress / Quit)     |
| Source config screen | Multi-group `huh.Form` — type → params → chunking config; per-group validation          |
| Ingestion screen     | `bubbles/spinner` (fetch) → `bubbles/progress` bar (embed/store) + Glamour log panel    |
| Quiz screen          | `Loading` → `Question` (`huh.NewSelect`) → `Feedback` (Glamour explanation + citations) |
| Results screen       | `bubbles/viewport` — score, per-question breakdown, replay/home prompts                 |
| Error screen         | Glamour error box with `[Retry]` / `[Back to Home]`                                     |
| Wire DI in `main.go` | Construct all deps, inject into `tui.NewApp(...)`, run `tea.NewProgram(model)`          |

**Done when:** Full TUI flow Welcome → Source Config → Ingest → Quiz → Results works end-to-end.

---

### Milestone 5 — H-MAS Scaffold

Agent interfaces wired up; SimpleOrchestrator replacing direct calls

| Task                 | Description                                                                                                             |
| -------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| Leaf agents          | `IngestAgent`, `ChunkAgent`, `EmbedAgent`, `RetrievalAgent`, `GeneratorAgent`, `EvaluatorAgent` wrapping existing logic |
| `SimpleOrchestrator` | Routes `Task` to registered agent by `Capability`                                                                       |
| Agent registry       | `registry.go` — register/lookup agents                                                                                  |

**Done when:** Quiz and ingestion run through `Orchestrator.Dispatch` with no behavior change.

---

### Milestone 6 — Multi-User SSH Server

Juniors SSH in; identity via key fingerprint; no local install needed

**New packages:** `internal/server/`, `internal/users/`

| Task                           | Description                                                                         |
| ------------------------------ | ----------------------------------------------------------------------------------- |
| Wish SSH server                | `hatch serve` starts Wish on `config.server.ssh.port` (default 2222)                |
| Host key                       | Auto-generate Ed25519 host key on first run; persist to `host_key_path`             |
| Identity middleware            | On connect: SHA256 SSH public key fingerprint → look up `users` table               |
| First-login prompt             | If user not found: `huh.NewInput` for display name → insert `users` row as `junior` |
| Per-connection TUI             | `server/handler.go` creates a fresh `tui.NewApp(...)` per SSH session               |
| SQLite WAL mode                | `PRAGMA journal_mode=WAL` on open; safe for concurrent connections                  |
| Sessions linked to user        | All quiz sessions write `user_id` from SSH connection context                       |
| `hatch users list`             | CLI: all users (id, name, role, last seen)                                          |
| `hatch users role <id> senior` | CLI: promote a user to senior                                                       |

**Done when:** Two simultaneous `ssh -p 2222 localhost` sessions run independent quizzes; both recorded in DB with distinct `user_id` values.

---

### Milestone 7 — Score Tracking CLI

Seniors query scores from the terminal

| Task                             | Description                                                                 |
| -------------------------------- | --------------------------------------------------------------------------- |
| `hatch scores`                   | Leaderboard: all juniors ranked by avg score %                              |
| `hatch scores --user=<name\|id>` | Per-junior: session list with date, source, score, duration                 |
| `hatch scores --source=<name>`   | Per-source: avg score and per-question accuracy breakdown                   |
| `hatch scores session <id>`      | Full drill-down: every question, chosen answer, correct answer, explanation |
| Tabular output                   | Lip Gloss `lipgloss.NewTable()` for aligned terminal rendering              |
| `--json` flag                    | All subcommands support `--json` for scripting                              |

**Done when:** All four `hatch scores` commands return accurate data from SQLite.

---

### Milestone 7b — Knowledge Base

Juniors search and browse indexed content directly from the TUI or CLI

| Task                                | Description                                                                                     |
| ----------------------------------- | ----------------------------------------------------------------------------------------------- |
| TUI: Search screen                  | `huh.NewInput` query → `Store.Search(TopK=10)` → Glamour-rendered results with source citations |
| TUI: Browse screen                  | List all indexed sources; select one to page through chunk summaries in a `bubbles/viewport`    |
| TUI: Welcome screen update          | Add "Knowledge Base" entry to the `huh.NewSelect` on the Welcome screen                         |
| CLI: `hatch search <query>`         | Headless semantic search; prints ranked chunks with similarity scores                           |
| CLI: `hatch search --source=<name>` | Scope search to a single named source                                                           |
| FTS5 keyword fallback               | Full-text search via `chunks_fts` when no embedding provider is configured                      |
| `--json` flag                       | All `hatch search` subcommands support `--json` for scripting                                   |

**Done when:** A junior SSHing in can select "Knowledge Base" from the welcome screen, search a query, and read relevant chunks with citations — and `hatch search <query>` returns ranked results headlessly.

---

### Milestone 8 — Web Dashboard

Seniors view leaderboard, progress, and session replays in a browser

#### Backend (Go — `internal/api/`)

| Task                          | Description                                                                                   |
| ----------------------------- | --------------------------------------------------------------------------------------------- |
| `hatch serve` extended        | Starts Wish SSH (2222) and HTTP web server (8080) as goroutines under one command             |
| Basic Auth middleware         | `HATCH_WEB_PASSWORD` env var; 401 on bad credentials                                          |
| `GET /api/leaderboard`        | `[{user_id, display_name, sessions, avg_score, best_score, last_active}]` sorted by avg score |
| `GET /api/users`              | All users with aggregate stats                                                                |
| `GET /api/users/:id`          | User detail + full session history                                                            |
| `GET /api/sources`            | All sources with per-source avg score                                                         |
| `GET /api/sources/:id/scores` | Per-question accuracy — which questions trip people up most                                   |
| `GET /api/sessions/:id`       | Full session: all questions, answers, correctness, explanations                               |
| Static file server            | `//go:embed dist/*` in `internal/api/static/`; serves React build at `/`                      |

#### Frontend (React + Vite + TypeScript — `web/`)

| Task                    | Description                                                                             |
| ----------------------- | --------------------------------------------------------------------------------------- |
| Project scaffold        | `web/`; Vite outputs to `internal/api/static/dist/`                                     |
| Auth wrapper            | Password prompt on load; stored in `sessionStorage`; injected as `Authorization` header |
| Leaderboard page        | Sortable table: rank, name, sessions, avg %, best score, last active                    |
| User Progress page      | Session history table + line chart of score over time (Recharts)                        |
| Source Breakdown page   | Per-source accuracy table; highlight sources with avg score < 60%                       |
| Session Drill-down page | Full Q&A replay: question, user's choice, correct answer, explanation                   |
| Makefile                | `make web` runs `yarn build` in `web/`; `make build` depends on `make web`              |

**Done when:** `hatch serve` starts both servers; `http://localhost:8080` shows leaderboard with real data; session drill-down loads full Q&A; React build is embedded in the binary (single file deploy).

---

### Milestone 9 — Expanded Sources (v1.5)

| Task                  | Description                                                        |
| --------------------- | ------------------------------------------------------------------ |
| Web URL source        | `internal/ingest/web/` — HTTP fetch, HTML→text, paragraph chunking |
| Ollama provider       | Local LLM + embedder support                                       |
| LLM streaming         | Token-by-token feedback in quiz screen                             |
| Hybrid search         | FTS5 keyword + vector results merged via Reciprocal Rank Fusion    |
| Difficulty adaptation | Calibrate question complexity from session history                 |
| Confluence source     | REST API connector                                                 |
| Notion source         | Notion API connector                                               |

---

### Milestone 10 — Hierarchical Multi-Agent System (v3)

| Task                     | Description                                                                       |
| ------------------------ | --------------------------------------------------------------------------------- |
| Three-tier orchestrator  | `StrategicOrchestrator → {IngestionOrchestrator, QuizOrchestrator} → leaf agents` |
| Adaptive difficulty      | Per-user knowledge model drives question selection                                |
| Multi-source questions   | Cross-reference content across sources in a single question                       |
| Knowledge graph          | Entity graph over indexed codebase for richer retrieval                           |
| Curriculum mode          | Structured onboarding track vs. random drill                                      |
| LLM-driven dispatch (v4) | `LLMOrchestrator` builds a task DAG per-request dynamically                       |

---

## SQLite Schema (multi-user)

```sql
PRAGMA journal_mode=WAL;

CREATE TABLE users (
    id            TEXT PRIMARY KEY,               -- SHA256 SSH key fingerprint
    display_name  TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'junior',  -- 'junior' | 'senior'
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at  DATETIME
);
CREATE TABLE sources (id, name, type, config_json, indexed_at, chunk_count);
CREATE TABLE chunks  (id, source_id, content, metadata JSON, position, created_at);
CREATE VIRTUAL TABLE chunk_vectors USING vec0(chunk_id TEXT, embedding FLOAT[{DIMS}]);
CREATE TABLE sessions (
    id, user_id REFERENCES users(id), source_ids JSON,
    started_at, completed_at, total_questions, correct_answers,
    score_pct REAL GENERATED ALWAYS AS (
        CASE WHEN total_questions > 0
             THEN ROUND(correct_answers * 100.0 / total_questions, 1)
             ELSE 0 END) VIRTUAL
);
CREATE TABLE session_questions (
    id, session_id, question_text, options JSON, correct_index,
    user_answer, is_correct, explanation, chunk_ids JSON, asked_at
);
CREATE VIRTUAL TABLE chunks_fts USING fts5(content, content=chunks);
CREATE INDEX idx_sessions_user    ON sessions(user_id);
CREATE INDEX idx_sessions_started ON sessions(started_at DESC);
```

---

## Verification Checklist

- [ ] `go build ./...` — compiles cleanly
- [ ] `go test ./...` — all tests pass with fakes
- [ ] `hatch config init` — scaffolds `~/.hatch/config.yaml`
- [ ] `hatch ingest --source=my-project` — indexes source, prints chunk count
- [ ] `hatch quiz --count=5` — headless quiz run, scores answers
- [ ] TUI end-to-end: Welcome → Source Config → Ingestion → Quiz → Results
- [ ] SSH flow: `ssh -p 2222 localhost` → name prompt → TUI → session recorded with `user_id`
- [ ] Concurrent SSH: two simultaneous connections run without DB errors
- [ ] `hatch scores` — leaderboard renders correctly
- [ ] `hatch scores session <id>` — full Q&A drill-down in terminal
- [ ] Web dashboard: leaderboard, user progress, source breakdown, session drill-down all load
- [ ] Single binary: no external static files required at runtime
