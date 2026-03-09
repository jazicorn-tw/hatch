# Hatch

A self-hosted developer onboarding tool. Junior engineers SSH in and get an
interactive quiz and knowledge base built from your actual codebase. No local
install required on their end.

---

## How it works

1. A senior ingests one or more sources (local directories, docs, URLs)
2. Hatch chunks and embeds the content into a local SQLite vector store
3. Juniors SSH in — they get a TUI with two modes:
   - **Quiz** — RAG-generated multiple choice questions with explanations
   - **Knowledge Base** — semantic search and browsable chunk viewer
4. Seniors track progress via the CLI or a web dashboard

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

### Start the server

```bash
hatch serve
# SSH server on :2222, web dashboard on :8080
```

### Juniors connect

```bash
ssh -p 2222 your-server
```

On first login they enter a display name. After that they land directly in the TUI.

### Score tracking (seniors)

```bash
hatch scores                        # leaderboard
hatch scores --user=alice           # per-junior history
hatch scores --source=my-project    # per-source breakdown
hatch scores session <id>           # full Q&A drill-down
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

Set `HATCH_WEB_PASSWORD` to protect the web dashboard.

---

## Web dashboard

Browse to `http://your-server:8080` after running `hatch serve`. Requires the
password set in `HATCH_WEB_PASSWORD`.

Screens: leaderboard, per-user progress, per-source accuracy breakdown, full
session replay.

---

## Building

```bash
make build        # compiles binary (includes embedded web UI)
make web          # builds React frontend only (yarn build in web/)
go test ./...     # run tests
```

---

## Roadmap

See [ROADMAP.md](ROADMAP.md) for the full milestone plan.
