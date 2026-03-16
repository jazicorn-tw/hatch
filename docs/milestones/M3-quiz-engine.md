<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-16
status:       active
tags:         [milestone, llm, pipeline, go, architecture, tui]
description:  "Walkthrough of Milestone 3 — the quiz engine: LLM providers, question generation, answer evaluation, sessions, and the kata engine with in-TUI code editing."
-->
# Milestone 3 — Quiz Engine

Walkthrough of the third Go milestone: wiring the LLM layer, building the quiz
generation and evaluation pipeline, and adding the kata engine with in-TUI code editing.

M3 ships in two parts:

- **M3** — Quiz engine: question generation, MCQ answer evaluation, session persistence, CLI
- **M3b** — Kata engine: code kata model, in-TUI editor, sandboxed evaluation

---

## Overview

M2 built the ingestion pipeline — documents are now chunked, embedded, and stored in
SQLite. M3 uses that vector store as the knowledge base for a quiz engine that generates
and evaluates multiple-choice questions grounded in the actual codebase documentation.

By the end of M3:

- `hatch quiz --topic=<name> --count=10` runs an interactive quiz session in the terminal
- Questions are either senior-provided (imported from a file) or AI-generated from the
  vector store
- Answers are evaluated deterministically (MCQ index comparison)
- Sessions are persisted to SQLite and tagged by topic
- `hatch kata --topic=<name>` runs a code kata with an in-TUI editor and sandboxed test runner

---

## Checklist

### M3 — Quiz Engine

- [ ] Anthropic LLM: `claude-sonnet-4-6` default; `ANTHROPIC_API_KEY` env var
- [ ] Google Gemini LLM provider: `gemini-2.0-flash` default; `GEMINI_API_KEY` env var
- [ ] Question model: `Question{ID, Text, Options[4], CorrectIndex, Explanation, SourceChunks}`
- [ ] Quiz generator: topic probe → `Store.Search(TopK=5)` → LLM MCQ prompt
- [ ] Prompt templates via `//go:embed` (`question_mcq.tmpl`, `question_explain.tmpl`)
- [ ] Answer evaluator: deterministic index comparison for MCQ
- [ ] Session model + SQLite migration; sessions tagged with topic
- [ ] Sr-provided quiz material: `hatch quiz create --topic=<name>` (import from file)
- [ ] AI-generated quiz: LLM generates questions from topic + source chunks
- [ ] CLI: `hatch quiz --topic=<name> --count=10`

### Milestone 3b — Kata Engine ✅

- [x] Kata model: `Kata{ID, Title, Description, StarterCode, Tests, Topic, Source}`
- [ ] Sr-provided katas: `hatch kata create --topic=<name>` (import from file)
- [x] AI-generated katas: LLM generates kata prompt + test cases from topic
- [x] Kata prompt template via `//go:embed` (`kata_generate.tmpl`)
- [x] In-TUI code editor: `bubbles/textarea` with syntax hint
- [x] Kata evaluator: run user solution against test cases; pass/fail per test
- [x] Sandbox execution: subprocess with timeout + resource limits; no network access
- [x] Kata session model + SQLite persistence; sessions tagged with topic
- [x] CLI: `hatch kata --topic=<name>`

---

## Package Layout

```text
internal/llm/                   Completer interface (defined in M1).
internal/llm/anthropic/         Anthropic API client — claude-sonnet-4-6 default.
internal/llm/gemini/            Google Gemini API client — gemini-2.0-flash default.
internal/llm/fake/              FakeLLM test double (defined in M1).
internal/quiz/                  Question model, quiz generator, answer evaluator.
internal/quiz/prompt/           Embedded prompt templates (question_mcq.tmpl etc.).
internal/kata/                  Kata model, kata generator, kata evaluator.
internal/kata/prompt/           Embedded prompt template (kata_generate.tmpl).
internal/kata/sandbox/          Subprocess executor with timeout + resource limits.
internal/session/               Session model + SQLite persistence for quiz and kata.
cmd/hatch/quiz.go               hatch quiz --topic=<name> --count=<n> CLI command.
cmd/hatch/kata.go               hatch kata --topic=<name> CLI command.
```

---

## 1. LLM Providers (`internal/llm/`)

The `llm.Completer` interface from M1 takes a prompt string and returns a completion:

```go
type Completer interface {
    Complete(ctx context.Context, prompt string) (string, error)
}
```

M3 adds two concrete implementations.

### Anthropic (`internal/llm/anthropic/`)

Calls the Anthropic Messages API. Default model: `claude-sonnet-4-6`. Requires
`ANTHROPIC_API_KEY`.

```yaml
llm_provider: anthropic
anthropic_api_key: sk-ant-...
```

### Google Gemini (`internal/llm/gemini/`)

Calls the Gemini GenerateContent API. Default model: `gemini-2.0-flash`. Reuses the
same `GEMINI_API_KEY` already set for the Gemini embedder.

```yaml
llm_provider: gemini
gemini_api_key: AIza...
```

Both providers are wired into the CLI via config — swap `llm_provider` to change models
without touching any code.

---

## 2. Question Model (`internal/quiz/`)

```go
type Question struct {
    ID           string
    Text         string
    Options      [4]string
    CorrectIndex int       // 0–3
    Explanation  string
    SourceChunks []string  // chunk IDs used to generate this question
}
```

`CorrectIndex` is an integer index into `Options`, making answer evaluation a simple
comparison — no LLM call needed at evaluation time.

`SourceChunks` links each question back to the chunks that grounded it, enabling
session drill-down (added in M7).

---

## 3. Quiz Generator (`internal/quiz/`)

```text
1. Embed the topic name → query vector
2. Store.Search(queryVec, TopK=5) → 5 nearest chunks
3. Render question_mcq.tmpl with chunks as context → prompt string
4. Completer.Complete(ctx, prompt) → raw LLM response
5. Parse response into Question struct
6. Repeat until count reached
```

Prompt templates live in `internal/quiz/prompt/` and are baked into the binary via
`//go:embed`. This keeps prompt iteration decoupled from Go code changes — editing a
template doesn't require recompiling, only re-running `go build`.

`question_mcq.tmpl` instructs the LLM to output a question in a structured JSON format
so parsing is deterministic. `question_explain.tmpl` generates a one-paragraph
explanation of the correct answer, shown after the user answers.

---

## 4. Answer Evaluator (`internal/quiz/`)

MCQ evaluation is deterministic — no LLM involved:

```go
func Evaluate(q Question, answerIndex int) bool {
    return answerIndex == q.CorrectIndex
}
```

After evaluation, the explanation from `q.Explanation` is displayed regardless of
whether the answer was correct.

---

## 5. Session Model (`internal/session/`)

A session groups a sequence of questions (or kata attempts) under a topic and user:

```go
type Session struct {
    ID        string
    UserID    string    // empty until M6 SSH auth
    Topic     string
    Type      string    // "quiz" | "kata"
    StartedAt time.Time
    EndedAt   time.Time
    Score     int       // correct answers
    Total     int       // questions attempted
}
```

Sessions and their individual answers are persisted to SQLite via migration `003_sessions.sql`.
The session ID is a UUID generated at start time.

---

## 6. Kata Engine (`internal/kata/`)

### Kata model

```go
type Kata struct {
    ID          string
    Title       string
    Description string
    StarterCode string
    Tests       string   // test file contents run against the user's solution
    Topic       string
    Source      string   // which ingested source this kata was derived from
}
```

### In-TUI code editor

The kata screen uses `bubbles/textarea` as a multi-line code input. The starter code is
pre-filled; the junior edits it and submits with a configurable key binding. The textarea
accepts both typed input and paste.

### Sandbox execution (`internal/kata/sandbox/`)

After submission, the user's solution is written to a temp file and executed as a
subprocess:

```text
1. Write user solution to os.TempDir()/kata-<id>/solution.go
2. Write test file to the same directory
3. exec.CommandContext(ctx, "go", "test", "./...") with timeout
4. Capture stdout/stderr — parse pass/fail per test case
5. Clean up temp directory
```

Resource constraints:

- **Timeout**: configurable, default 10 seconds
- **No network**: `GOPROXY=off GONOSUMCHECK=*` env vars prevent `go test` from fetching modules
- **No file system writes outside temp dir**: kata code runs in an isolated temp directory

---

## 7. CLI Commands

### `hatch quiz --topic=<name> --count=<n>`

Looks up the topic in config, runs the quiz generator for `n` questions, presents each
question in the terminal (plain text for M3 — no Bubble Tea yet), records the session.

### `hatch quiz create --topic=<name>`

Imports a YAML file of pre-written questions into the `questions` SQLite table, tagged
with the topic.

### `hatch kata --topic=<name>`

Fetches or generates a kata for the topic, opens the in-TUI editor, runs the sandbox
evaluator on submission, and records the kata session.

### `hatch kata create --topic=<name>`

Imports a YAML file of pre-written katas into the `katas` SQLite table.

---

## Verification

```bash
# Unit tests — fakes cover all LLM and store calls
go test ./internal/llm/...
go test ./internal/quiz/...
go test ./internal/kata/...
go test ./internal/session/...

# Full suite
go test ./...

# Integration (requires API key)
export ANTHROPIC_API_KEY=sk-ant-...
hatch quiz --topic=go-interfaces --count=5

# Kata (requires Go toolchain in PATH — used by sandbox)
hatch kata --topic=go-interfaces
```

---

## Technologies

| Technology               | Role in M3                                                       |
| ------------------------ | ---------------------------------------------------------------- |
| Anthropic SDK            | `claude-sonnet-4-6` completions for question and kata generation |
| Google Gemini SDK        | `gemini-2.0-flash` completions (reuses M2 dependency)            |
| `//go:embed`             | Bundles prompt templates into binary (reuses M1 pattern)         |
| `bubbles/textarea`       | In-TUI multi-line code editor for kata submissions               |
| `os/exec`                | Subprocess sandbox for running kata test cases                   |
| `text/template`          | Renders prompt templates with chunk context                      |
| `encoding/json`          | Parses structured LLM responses into `Question` structs          |
| `github.com/google/uuid` | Generates session IDs                                            |

---

## Related

- [`docs/ROADMAP.md`](../ROADMAP.md) — full milestone plan
- [`docs/milestones/M2-ingestion.md`](M2-ingestion.md) — vector store built in M2
- [`docs/onboarding/deeper/tui/BUBBLE_TEA.md`](../onboarding/deeper/tui/BUBBLE_TEA.md) — TUI framework (M4+)
- [`docs/onboarding/deeper/go/INTERFACES.md`](../onboarding/deeper/go/INTERFACES.md) — `Completer` and `Runner` interfaces
- [`docs/adr/ADR-014-per-agent-model-routing.md`](../adr/ADR-014-per-agent-model-routing.md) — per-agent model routing (M5)

---

## Resources

### LLM APIs

- [Anthropic API reference](https://docs.anthropic.com/en/api/getting-started) — Messages API used by the Anthropic completer
- [Gemini API reference](https://ai.google.dev/gemini-api/docs) — GenerateContent used by the Gemini completer
- [Prompt engineering guide](https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/overview) — structuring prompts for reliable JSON output

### Go patterns

- [text/template](https://pkg.go.dev/text/template) — standard library template package
- [os/exec](https://pkg.go.dev/os/exec) — subprocess execution used by the kata sandbox
- [Go embed](https://pkg.go.dev/embed) — baking static files into the binary

### Charmbracelet

- [bubbles/textarea](https://github.com/charmbracelet/bubbles/tree/master/textarea) — multi-line text input component
