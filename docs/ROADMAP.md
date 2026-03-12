<!--
created_by:   jazicorn-tw
created_date: 2026-03-11
updated_by:   jazicorn-tw
updated_date: 2026-03-12
status:       active
tags:         [planning]
description:  "Hatch development roadmap — versioned milestones"
-->
# Hatch — Roadmap

---

## v0 — Project Foundation ✅

Repo setup, CI/CD, developer tooling, and planning artifacts.

### Milestone 0 — Setup & Tooling

- [x] Init repo: `go mod init`, `./dev` task runner, `.gitignore`
- [x] CI: GitHub Actions workflows (ci, release, publish, doctor, changelog-guard)
- [x] Git hooks: pre-commit, commit-msg, pre-add
- [x] Doctor script: validate local dev environment
- [x] Bootstrap script: hooks + doctor + quality gate
- [x] Semantic release: automated versioning + changelog via CI
- [x] Branching strategy: `feature → staging → canary → main`
- [x] ADRs: architecture decisions documented (ADR-000 through ADR-011)
- [x] Docs: onboarding, contributing, architecture, roadmap, commit conventions
- [x] Docs: tooling reference (dev, doctor, bootstrap, pre-commit)
- [x] `.env.example`: env var template for local development
- [x] `docs/devops/CI_VARIABLES.md`: GitHub repo variables reference and quick-start checklist
- [x] `./dev test-ci`: local CI simulation via `act`
- [x] `./dev changelog`: semantic-release dry-run / changelog preview

---

## v1 — Core Quiz Engine

Single binary, local ingestion, full TUI quiz loop

### Milestone 1 — Foundation

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
- [ ] Session model + SQLite persistence; sessions tagged with topic
- [ ] Sr-provided quiz material: `hatch quiz create --topic=<name>` (import from file)
- [ ] AI-generated quiz: LLM generates questions from topic + source material
- [ ] CLI: `hatch quiz --topic=<name> --count=10`

### Milestone 3b — Kata Engine

- [ ] Kata model: `Kata{ID, Title, Description, StarterCode, Tests, Topic, Source}`
- [ ] Sr-provided katas: `hatch kata create --topic=<name>` (import from file)
- [ ] AI-generated katas: LLM generates kata prompt + test cases from topic
- [ ] Kata prompt template via `//go:embed` (`kata_generate.tmpl`)
- [ ] In-TUI code editor: `bubbles/textarea` with syntax hint
- [ ] Kata evaluator: run user solution against test cases; pass/fail per test
- [ ] Sandbox execution: subprocess with timeout + resource limits; no network access
- [ ] Kata session model + SQLite persistence; sessions tagged with topic
- [ ] CLI: `hatch kata --topic=<name>`

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

### Milestone 6b — Topics + Assignment

- [ ] Topic model: `Topic{ID, Name, Description}` + SQLite table
- [ ] CLI: `hatch topics create/list/remove`
- [ ] Sr assigns topics to juniors: `hatch assign topic <topic> --user=<name|id>`
- [ ] Jr sees assigned topics on TUI home screen
- [ ] Sr assigns quiz to junior by topic: `hatch assign quiz --topic=<name> --user=<name|id>` (Sr-provided or AI-generated)
- [ ] Sr assigns kata to junior by topic: `hatch assign kata --topic=<name> --user=<name|id>` (Sr-provided or AI-generated)
- [ ] Assignment model + SQLite persistence: `assignments{id, user_id, topic_id, type, source_id, assigned_by, assigned_at}`
- [ ] Jr TUI: "My Assignments" screen — list pending quizzes and katas by topic
- [ ] Assignment status: `pending → in_progress → complete`
- [ ] Role-based TUI routing: detect Sr vs Jr on SSH login; fork to Sr menu or Jr menu
- [ ] Sr TUI menu: manage topics, assign work, review sessions, view leaderboard
- [ ] Jr TUI menu: my assignments, take quiz, take kata, my scores

### Milestone 7 — Score Tracking CLI

- [ ] `hatch scores`: leaderboard — all juniors ranked by avg score %
- [ ] `hatch scores --user=<name|id>`: per-junior session history (date, topic, type, score, duration)
- [ ] `hatch scores --topic=<name>`: per-topic avg score across all juniors (quiz + kata)
- [ ] `hatch scores --user=<name|id> --topic=<name>`: one junior's score history for a topic
- [ ] `hatch scores session <id>`: full Q&A or kata drill-down (question/code, chosen answer, correct answer, explanation)
- [ ] Sr review: `hatch review --user=<name|id>` — browse any jr's completed quizzes and katas
- [ ] Sr feedback: `hatch review --user=<name|id> --session=<id> --comment="..."` — attach comment to session
- [ ] Jr sees Sr feedback: comments surfaced in "My Scores" TUI screen and `hatch scores session <id>`
- [ ] Lip Gloss `lipgloss.NewTable()` for aligned terminal output
- [ ] `--json` flag on all `hatch scores` subcommands for scripting
- [ ] CSV export: `hatch scores --user=<name|id> --export=csv` → writes `scores_<user>_<date>.csv`
- [ ] CSV export all: `hatch scores --export=csv` → all juniors, all topics, all sessions

### Milestone 7b — Knowledge Base

- [ ] TUI: "Search" screen — `huh.NewInput` query → `Store.Search(TopK=10)` → Glamour-rendered results with source citations
- [ ] TUI: "Browse" screen — list indexed sources; select source to page through chunk summaries via `bubbles/viewport`
- [ ] TUI: wire Knowledge Base entry point into Welcome screen `huh.NewSelect`
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
- [ ] `GET /api/topics` + `GET /api/topics/:id/scores` → per-topic avg score breakdown
- [ ] `GET /api/sources` + `GET /api/sources/:id/scores` → per-question accuracy heatmap
- [ ] `GET /api/sessions/:id` → full Q&A / kata drill-down payload
- [ ] `GET /api/scores/export?user=<id>&topic=<id>&format=csv` → CSV download
- [ ] React + Vite frontend scaffolded in `web/`
- [ ] Auth wrapper: password prompt → stored in `sessionStorage` → injected as `Authorization` header
- [ ] Leaderboard page: sortable table (rank, name, sessions, avg %, best score, last active)
- [ ] User Progress page: session table + line chart of score over time (Recharts)
- [ ] Source Breakdown page: per-source accuracy; highlight sources with avg < 60%
- [ ] Session Drill-down page: full Q&A / kata replay with user's choice highlighted
- [ ] Sr Review page: browse any jr's completed quizzes and katas by topic
- [ ] CSV export button on Leaderboard and User Progress pages
- [ ] Vite outputs to `internal/api/static/dist/`; embedded in Go binary via `//go:embed`
- [ ] Sr Review page: attach and view feedback comments on jr sessions

### Milestone 9 — Auth + Security Hardening (v1.0.0)

First intentional breaking change — enforces authentication across SSH and web.

- [ ] JWT issuance on SSH login: signed token stored in session context
- [ ] JWT middleware on all HTTP API routes: 401 on missing/expired token
- [ ] Role claims in JWT: `role: senior | junior` gates Sr-only endpoints
- [ ] Token refresh: short-lived access token + refresh flow
- [ ] `HATCH_JWT_SECRET` env var; rotate without restarting server
- [ ] Web dashboard: replace Basic Auth with JWT login page
- [ ] SSH: token-based re-auth after configurable idle timeout
- [ ] Audit log: `audit_log{id, user_id, action, resource, timestamp}` — all Sr actions recorded

---

## v4 — Expanded Sources

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
- [ ] Assignment notifications: Jr notified of new assignments on next TUI open (banner on welcome screen)
- [ ] Feedback notifications: Jr notified when Sr leaves a comment on their session

---

## v5 — Hierarchical Multi-Agent System

Three-tier orchestration, adaptive quizzing, curriculum mode

- [ ] Three-tier orchestrator: `StrategicOrchestrator → {IngestionOrchestrator, QuizOrchestrator} → leaf agents`
- [ ] Adaptive difficulty: per-user knowledge model drives question selection
- [ ] Multi-source questions: cross-reference content across sources in one question
- [ ] Knowledge graph: entity graph over indexed codebase for richer retrieval
- [ ] Curriculum mode: structured onboarding track vs. random drill
- [ ] LLM-driven `Dispatch`: task DAG built dynamically per-request
