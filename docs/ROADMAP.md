# Hatch — Roadmap

---

## v1 — Core Quiz Engine

Single binary, local ingestion, full TUI quiz loop

### Milestone 1 — Foundation

- [ ] Init repo: `go mod init`, `Makefile`, `.gitignore`
- [ ] Scaffold all `cmd/hatch/` and `internal/` packages
- [ ] Config layer: Viper + `~/.hatch/config.yaml`, env var overrides, `hatch config init`
- [ ] Core interfaces: `Source`, `Chunker`, `Embedder`, `LLM`, `Store`, `Agent`
- [ ] SQLite schema + migration runner; WAL mode on open
- [ ] In-memory store (`internal/store/memory/`) for tests
- [ ] Fake embedder + fake LLM for tests

### Milestone 2 — Ingestion Pipeline

- [ ] Filesystem source: directory walker with gitignore support
- [ ] Markdown chunker: heading-based recursive split
- [ ] Code chunker: fixed-size with configurable overlap (`.ts`, `.tsx`, `.go`, `.scss`)
- [ ] Ingestion pipeline: `Run(ctx, source, chunker, embedder, store, progressCh)`
- [ ] OpenAI embedder: batched API calls, `text-embedding-3-small` default
- [ ] sqlite-vec store: `Upsert`, cosine KNN `Search`, `DeleteBySource`
- [ ] CLI: `hatch ingest --source=<name>`, `hatch sources list/remove`

### Milestone 3 — Quiz Engine

- [ ] Anthropic LLM (`claude-sonnet-4-6` default)
- [ ] Question types: `Question{Text, Options[4], CorrectIndex, Explanation, SourceChunks}`
- [ ] Quiz generator: topic probe → `Store.Search(TopK=5)` → LLM MCQ prompt
- [ ] Prompt templates via `//go:embed` (`question_mcq.tmpl`, `question_explain.tmpl`)
- [ ] Answer evaluator: deterministic index comparison for MCQ
- [ ] Session model + SQLite persistence
- [ ] CLI: `hatch quiz --source=<name> --count=10`

### Milestone 4 — TUI

- [ ] Lip Gloss styles: dark/light/dracula themes
- [ ] Custom `tea.Msg` types in `internal/tui/msgs/`
- [ ] Root app model: state machine delegating to sub-models
- [ ] Welcome screen: ASCII wordmark + `huh.NewSelect`
- [ ] Source config screen: multi-group `huh.Form` with per-group validation
- [ ] Ingestion screen: `bubbles/spinner` → `bubbles/progress` bar + Glamour log panel
- [ ] Quiz screen: `Loading → Question → Feedback` inner state machine
- [ ] Results screen: `bubbles/viewport` with score breakdown
- [ ] Error screen: Glamour error box with Retry / Back to Home

### Milestone 5 — H-MAS Scaffold

- [ ] Wrap pipeline + generator behind `Agent` interface
- [ ] Leaf agents: `IngestAgent`, `ChunkAgent`, `EmbedAgent`, `RetrievalAgent`, `GeneratorAgent`, `EvaluatorAgent`
- [ ] `SimpleOrchestrator`: routes `Task` to registered agent by `Capability`
- [ ] Agent registry

---

## v2 — Multi-User SSH + Score Tracking

Juniors SSH in from anywhere; seniors look up scores via CLI

### Milestone 6 — Multi-User SSH Server

- [ ] `hatch serve`: start Wish SSH server on port 2222
- [ ] Auto-generate Ed25519 host key on first run
- [ ] Identity middleware: SHA256 SSH public key fingerprint → `users` table lookup
- [ ] First-login prompt: `huh.NewInput` for display name → insert `junior` user row
- [ ] Per-connection Bubble Tea session via `server/handler.go`
- [ ] All quiz sessions written with `user_id` from SSH connection context
- [ ] `hatch users list`: show all users (id, name, role, last seen)
- [ ] `hatch users role <id> senior`: promote a user

### Milestone 7 — Score Tracking CLI

- [ ] `hatch scores`: leaderboard — all juniors ranked by avg score %
- [ ] `hatch scores --user=<name|id>`: per-junior session history (date, source, score, duration)
- [ ] `hatch scores --source=<name>`: per-source avg score + per-question accuracy breakdown
- [ ] `hatch scores session <id>`: full Q&A drill-down (question, chosen answer, correct answer, explanation)
- [ ] Lip Gloss `lipgloss.NewTable()` for aligned terminal output
- [ ] `--json` flag on all `hatch scores` subcommands for scripting

### Milestone 7b — Knowledge Base

- [ ] TUI: "Search" screen — `huh.NewInput` query → `Store.Search(TopK=10)` → Glamour-rendered results with source citations
- [ ] TUI: "Browse" screen — list indexed sources; select source to page through chunk summaries via `bubbles/viewport`
- [ ] TUI: wire Knowledge Base entry point into Welcome screen `huh.NewSelect` (alongside Start Quiz / Configure Sources / Progress / Quit)
- [ ] CLI: `hatch search <query>` — headless semantic search; prints ranked chunks with similarity scores
- [ ] CLI: `hatch search --source=<name> <query>` — scoped to a single source
- [ ] FTS5 keyword fallback: full-text search when no embedding provider is configured
- [ ] `--json` flag on `hatch search` for scripting

---

## v3 — Web Dashboard

Seniors access leaderboard, progress, and session replays in a browser

### Milestone 8 — Web Dashboard

- [ ] `hatch serve` extended: Wish SSH + HTTP web server as goroutines under one command
- [ ] Basic Auth middleware: `HATCH_WEB_PASSWORD` env var; 401 on bad credentials
- [ ] `GET /api/leaderboard` → ranked list with avg score, best score, last active
- [ ] `GET /api/users` + `GET /api/users/:id` → aggregate stats + session history
- [ ] `GET /api/sources` + `GET /api/sources/:id/scores` → per-question accuracy heatmap
- [ ] `GET /api/sessions/:id` → full Q&A drill-down payload
- [ ] React + Vite frontend scaffolded in `web/`
- [ ] Auth wrapper: password prompt → stored in `sessionStorage` → injected as `Authorization` header
- [ ] Leaderboard page: sortable table (rank, name, sessions, avg %, best score, last active)
- [ ] User Progress page: session table + line chart of score over time (Recharts)
- [ ] Source Breakdown page: per-source accuracy; highlight sources with avg < 60%
- [ ] Session Drill-down page: full Q&A replay with user's choice highlighted
- [ ] Vite outputs to `internal/api/static/dist/`; embedded in Go binary via `//go:embed`
- [ ] `make web` → `make build` dependency chain in Makefile

---

## v4 — Expanded Sources (v1.5 features)

More ingestion targets and smarter search

- [ ] Web URL source: HTTP fetch, HTML→text, paragraph chunking
- [ ] Ollama provider: local LLM + embedder, no API cost
- [ ] LLM streaming: token-by-token feedback in quiz screen
- [ ] Hybrid search: FTS5 keyword + vector results merged via Reciprocal Rank Fusion
- [ ] Difficulty adaptation: calibrate question complexity from session history
- [ ] Confluence source: REST API connector
- [ ] Notion source: Notion API connector
- [ ] Free-text quiz questions: LLM semantic answer evaluation
- [ ] Export results: copy session summary to clipboard or write to `.md`

---

## v5 — Hierarchical Multi-Agent System

Three-tier orchestration, adaptive quizzing, curriculum mode

- [ ] Three-tier orchestrator: `StrategicOrchestrator → {IngestionOrchestrator, QuizOrchestrator} → leaf agents`
- [ ] Adaptive difficulty: per-user knowledge model drives question selection
- [ ] Multi-source questions: cross-reference content across sources in one question
- [ ] Knowledge graph: entity graph over indexed codebase for richer retrieval
- [ ] Curriculum mode: structured onboarding track vs. random drill
- [ ] LLM-driven `Dispatch`: task DAG built dynamically per-request (no interface changes required)
