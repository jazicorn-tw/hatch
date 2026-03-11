<!--
created_by:   jazicorn-tw
created_date: 2026-03-11
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       draft
tags:         []
description:  ""
-->
# Hatch

A self-hosted developer onboarding tool. Junior engineers SSH in and get
interactive quizzes and coding exercises built from your actual codebase.
Seniors assign topics, review progress, and export reports — no local install
required on the junior's end.

---

## How it works

1. A senior creates topics and ingests sources (local directories, docs, URLs)
2. Hatch chunks and embeds the content into a local SQLite vector store
3. The senior assigns quizzes and katas to juniors by topic
4. Juniors SSH in — they get a role-aware TUI with:
   - **My Assignments** — pending quizzes and katas assigned by their senior
   - **Quiz** — multiple choice questions generated from source material
   - **Kata** — coding exercises with automated test evaluation
   - **Knowledge Base** — semantic search and browsable chunk viewer
   - **My Scores** — session history, scores by topic, and senior feedback
5. Seniors review progress, leave feedback, and export CSV reports via CLI or web dashboard

---

## Tech stack

| Concern       | Choice                                                                  |
| ------------- | ----------------------------------------------------------------------- |
| Language      | Go                                                                      |
| TUI           | Charmbracelet: Bubble Tea v2, Huh v2, Glamour v2, Bubbles, Lip Gloss    |
| Vector store  | SQLite + sqlite-vec (WAL mode for concurrent SSH connections)           |
| LLM           | Provider-agnostic — Anthropic, OpenAI, Ollama                           |
| Embeddings    | Provider-agnostic — same providers                                      |
| SSH server    | Charmbracelet Wish — per-connection Bubble Tea session                  |
| Web dashboard | React + Vite + TypeScript; Go `net/http` REST API; embedded in binary   |

---

## Usage

### Ingest a source

```bash
hatch ingest --source=my-project --path=./src
```

### Manage topics

```bash
hatch topics create "Go Concurrency"
hatch topics list
hatch assign topic "Go Concurrency" --user=alice
```

### Assign quizzes and katas

```bash
# Assign an AI-generated quiz
hatch assign quiz --topic="Go Concurrency" --user=alice

# Assign a kata (Sr-provided or AI-generated)
hatch assign kata --topic="Go Concurrency" --user=alice
```

### Start the server

```bash
hatch serve
# SSH server on :2222, web dashboard on :8080
```

### Juniors connect

```bash
ssh -p 2222 your-server
```

On first login they enter a display name and land in their assignment queue.

### Score tracking and review (seniors)

```bash
hatch scores                                      # leaderboard
hatch scores --user=alice                         # per-junior history
hatch scores --topic="Go Concurrency"             # per-topic breakdown
hatch scores --user=alice --topic="Go Concurrency"
hatch scores session <id>                         # full Q&A or kata drill-down

hatch review --user=alice                         # browse alice's sessions
hatch review --user=alice --session=<id> --comment="Good approach, but consider error wrapping"

hatch scores --user=alice --export=csv            # export alice's scores
hatch scores --export=csv                         # export all juniors
```

### Knowledge base search (headless)

```bash
hatch search "how does auth work"
hatch search --source=my-project "middleware"
```

---

## Configuration

```bash
hatch config init   # creates ~/.hatch/config.yaml
```

Key config options:

```yaml
llm:
  provider: anthropic          # anthropic | openai | ollama
  model: claude-sonnet-4-6

embeddings:
  provider: openai
  model: text-embedding-3-small

server:
  ssh:
    port: 2222
  http:
    port: 8080
```

Environment variables:

| Variable            | Purpose                                                |
| ------------------- | ------------------------------------------------------ |
| `HATCH_JWT_SECRET`  | Signs JWT tokens for SSH + web auth                    |
| `HATCH_WEB_PASSWORD`| Web dashboard password (pre-JWT, deprecated in v1.0.0) |

---

## Web dashboard

Browse to `http://your-server:8080` after running `hatch serve`.

Screens: leaderboard, per-user progress by topic, per-source accuracy breakdown,
full session replay, Sr review with feedback, CSV export.

---

## Building

```bash
./dev bootstrap   # first-time setup: hooks + doctor + quality gate
./dev build       # compile binary (includes embedded web UI)
./dev test        # run tests
go build ./...    # direct Go build
```

---

## Roadmap

See [`docs/ROADMAP.md`](docs/ROADMAP.md) for the full milestone plan.
