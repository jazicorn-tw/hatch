<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-15
status:       active
tags:         [adr, llm, agent, providers]
description:  "ADR-014: Per-agent model routing — how to assign different LLM providers and models to different agent tasks"
-->
# ADR-014: Per-agent model routing

- **Status:** Accepted
- **Date:** 2026-03-14
- **Deciders:** Project maintainers
- **Scope:** `internal/llm/`, `internal/embedder/`, `internal/agent/`, M5 H-MAS scaffold

---

## Context

Hatch currently configures a single LLM provider and a single embedder provider globally
via `~/.hatch/config.yaml`. Every task — question generation, answer evaluation,
embedding — uses the same model.

The M5 H-MAS milestone introduces a hierarchy of specialised leaf agents
(`GeneratorAgent`, `EvaluatorAgent`, `EmbedAgent`, `RetrievalAgent`, etc.). These agents
have different cost, latency, and quality requirements:

| Agent             | Requirement                                 | Ideal model                    |
| ----------------- | ------------------------------------------- | ------------------------------ |
| `GeneratorAgent`  | High quality — generates quiz questions     | `claude-sonnet-4-6`            |
| `EvaluatorAgent`  | Fast + cheap — deterministic MCQ comparison | `gemini-2.0-flash`             |
| `EmbedAgent`      | Batch throughput                            | `text-embedding-3-small`       |
| `RetrievalAgent`  | No API cost preferred                       | Ollama (local)                 |
| `KataAgent`       | Code generation quality                     | `claude-sonnet-4-6` or similar |

A single global `llm.Completer` cannot express this routing. Every agent would use the
same model regardless of whether that model is the right fit for the task.

Charmbracelet's [Fantasy](https://github.com/charmbracelet/fantasy) library
(`charm.land/fantasy`) provides a multi-provider, multi-model Go API designed for
exactly this pattern. It is currently in **preview** (API may change).

---

## Decision

**Defer adoption of Fantasy until M5. Design the M3–M4 agent interfaces to be
routing-ready without committing to Fantasy today.**

Concretely:

1. **M3–M4:** Each agent struct holds its own `llm.Completer` instance (injected at
   construction time). The orchestrator wires each agent to its configured completer.
   Config grows to support per-agent provider overrides:

   ```yaml
   agents:
     generator:
       provider: anthropic
       model: claude-sonnet-4-6
     evaluator:
       provider: gemini
       model: gemini-2.0-flash
   ```

2. **M5:** Evaluate replacing the hand-wired per-agent injection with Fantasy's unified
   multi-provider API, if Fantasy has reached a stable release by then. If Fantasy is
   still in preview, continue with the per-agent injection approach.

---

## Alternatives Considered

### 1. Adopt Fantasy now (M3)

Gives per-agent routing immediately and offloads provider abstraction to a maintained
library.

**Rejected** — Fantasy is in preview; pulling an alpha dependency into the core LLM
path introduces risk of breaking API changes during active development milestones.

### 2. Keep single global provider forever

Simple. No routing complexity.

**Rejected** — Using `claude-sonnet-4-6` for every answer evaluation call is
unnecessarily expensive. Cost optimisation matters as junior headcount grows.

### 3. Add a model selector to `llm.Completer`

```go
type Completer interface {
    Complete(ctx context.Context, model string, prompt string) (string, error)
}
```

Lets callers specify the model per call without changing the injection pattern.

**Deferred** — Adds model name as a stringly-typed parameter; callers need to know
valid model strings per provider. Per-agent injection is cleaner and keeps model
selection at the config/wiring layer, not scattered through business logic.

---

## Consequences

### Positive

- M3–M4 agents gain routing-ready structure with minimal added complexity
- No alpha dependency risk during active feature development
- Fantasy adoption at M5 becomes a drop-in replacement if it stabilises, not a
  large-scale refactor
- Config-driven routing means operators can tune cost vs. quality per agent without
  code changes

### Negative

- Per-agent config adds surface area to `~/.hatch/config.yaml` — needs validation and
  sensible defaults
- If Fantasy stabilises before M5, there is some duplicated effort in the hand-wired
  injection layer

---

## Review trigger

Revisit this decision at the start of M5 planning. Check whether
`charm.land/fantasy` has cut a stable (non-preview) release. If yes, evaluate
replacing the per-agent injection wiring with Fantasy's unified API.
