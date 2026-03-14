<!--
created_by:   jazicorn-tw
created_date: 2026-03-11
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [onboarding, llm, providers, configuration, go]
description:  "Overview, usage, configuration, and build instructions for hatch."
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

| Concern       | Choice                                                                |
| ------------- | --------------------------------------------------------------------- |
| Language      | Go                                                                    |
| TUI           | Charmbracelet: Bubble Tea v2, Huh v2, Glamour v2, Bubbles, Lip Gloss  |
| Vector store  | SQLite (WAL mode for concurrent SSH connections)                      |
| LLM           | Provider-agnostic — Anthropic (default), OpenAI, Ollama               |
| Embeddings    | Provider-agnostic — Ollama (default), OpenAI                          |
| SSH server    | Charmbracelet Wish — per-connection Bubble Tea session                |
| Web dashboard | React + Vite + TypeScript; Go `net/http` REST API; embedded in binary |

See [`docs/providers/PROVIDERS.md`](docs/providers/PROVIDERS.md) for provider configuration details.

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

Key config options (`~/.hatch/config.yaml`, where `~` is your home directory — e.g. `/Users/yourname`):

```yaml
llm_provider:   anthropic        # anthropic | openai | ollama
embed_provider: ollama           # ollama | openai
ssh_port:       2222
http_port:      8080
web_password:   changeme
jwt_secret:     ""
db_path:        ~/.hatch/hatch.db
```

Any key can be overridden with a `HATCH_<KEY>` environment variable:

| Variable               | Purpose                                        |
| ---------------------- | ---------------------------------------------- |
| `HATCH_LLM_PROVIDER`   | LLM provider (`anthropic`, `openai`, `ollama`) |
| `HATCH_EMBED_PROVIDER` | Embedding provider (`ollama`, `openai`)        |
| `HATCH_SSH_PORT`       | SSH server port (default `2222`)               |
| `HATCH_HTTP_PORT`      | Web dashboard port (default `8080`)            |
| `HATCH_JWT_SECRET`     | Signs JWT tokens for SSH + web auth            |
| `HATCH_WEB_PASSWORD`   | Web dashboard password                         |
| `HATCH_DB_PATH`        | Path to the SQLite database file               |

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

## Docs

| Document                                                               | Description                                         |
| ---------------------------------------------------------------------- | --------------------------------------------------- |
| [`docs/providers/PROVIDERS.md`](docs/providers/PROVIDERS.md)           | LLM and embedding provider overview                 |
| [`docs/providers/LLM.md`](docs/providers/LLM.md)                       | LLM providers and recommended models                |
| [`docs/providers/EMBEDDER.md`](docs/providers/EMBEDDER.md)             | Embedding providers and recommended models          |
| [`docs/providers/CONFIGURATION.md`](docs/providers/CONFIGURATION.md)   | Full config file and environment variable reference |
| [`docs/TESTING.md`](docs/TESTING.md)                                   | Test coverage, test doubles, and how to run tests   |
| [`docs/ROADMAP.md`](docs/ROADMAP.md)                                   | Milestone plan                                      |
| [`docs/devops/CI_VARIABLES.md`](docs/devops/CI_VARIABLES.md)           | CI gate variables                                   |
