<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [adr, tooling, qa, build]
description:  "ADR-000: Establish Formatting, Linting, and Static Analysis First"
-->
<!-- markdownlint-disable-file MD036 -->

# ADR-000: Establish Formatting, Linting, and Static Analysis First

**Status:** Accepted
**Date:** 2026-03-10
**Decision Makers:** Project Maintainers
**Scope:** Entire codebase (current and future projects)

---

## Context

Early stages of this project focused on functionality, environment setup, and architecture.
Formatting, linting, and static analysis were introduced **after** initial development had already begun.

This resulted in:

- Inconsistent formatting and style decisions
- Delayed discovery of avoidable issues
- Extra cleanup work once tooling was enabled
- Unclear quality expectations for early contributions

In hindsight, formatting + linting + static analysis define the **baseline quality contract** of a project
and should be established **before or alongside the first production code**.

---

## Decision

We will **establish formatting, linting, and static analysis as ADR-000**, representing a foundational,
non-negotiable decision that precedes all other architectural choices.

For this project and all future projects:

- Formatting and linting **must be set up first**
- Rules define the minimum quality bar
- CI must enforce them as part of the default quality gate

---

## Chosen Tooling (Go)

This project uses:

- **`gofmt`** — deterministic, opinionated Go formatting (no configuration)
- **`go vet`** — reports suspicious constructs and common mistakes
- **`staticcheck`** — advanced static analysis (unused code, API misuse, performance)
- **`markdownlint-cli2`** — documentation quality gate for all Markdown files

### Scope

- ✅ All `.go` files (`./...`)
- ✅ All `.md` files in `docs/`

> Documentation is held to the same quality standard as production code.
> Markdown lint runs in CI alongside Go analysis.

### How it runs

```bash
# Auto-format (local dev)
make format         # runs gofmt -w ./...

# Static analysis
make lint           # runs go vet ./... + markdownlint

# Full quality gate (matches CI)
make quality        # doctor + go vet + go test
```

> CI uses `gofmt -l` (list-only, non-mutating) to verify formatting without modifying files.

---

## Rationale

### Why formatting + linting come first

Formatting and linting:

- Define *how* code should look and behave before anyone writes it
- Prevent subjective debates in PRs ("tabs vs spaces", "import order", etc.)
- Catch issues earlier than tests alone
- Encourage small, incremental changes

By establishing them first:

- All contributors share the same expectations
- Architectural discussions happen on top of a clean baseline
- Refactoring cost is reduced long-term

### Why lint documentation as well

Markdown documentation:

- Encodes design intent and operational procedures
- Lives alongside code and must remain readable
- Is frequently updated during refactoring and onboarding

Linting docs:

- Prevents broken links, malformed tables, and inconsistent heading levels
- Encourages clear, navigable documentation
- Surfaces formatting issues before they accumulate

### Why `gofmt` (no configuration)

`gofmt` provides:

- Deterministic formatting (same output everywhere)
- Zero configuration (no debates about style)
- Go-native, universally supported
- CI-friendly non-mutating check mode

---

## Alternatives Considered

### 1. Add tooling later (post-MVP)

**Rejected** — causes rework, creates inconsistent legacy code, weakens early quality culture.

### 2. Rely only on IDE inspections / auto-formatting

**Rejected** — not enforceable, inconsistent across contributors, not CI-verifiable.

### 3. Use a third-party linter aggregator only

**Rejected** — adds dependency complexity before establishing the baseline. `go vet` + `staticcheck`
cover the essential cases. Can be extended later once the baseline is stable.

### 4. Exclude documentation from linting

**Rejected** — documentation drift is a real cost. Treating docs as second-class encourages neglect.

---

## Consequences

### Positive

- Clear quality baseline from the start
- Faster PR reviews
- Reduced cognitive load when reading code
- CI enforces standards automatically
- Less "style churn" (formatting is deterministic)
- Documentation stays readable and consistent

### Trade-offs

- Occasional false positives (managed via narrow suppressions)
- Requires discipline to evolve rules intentionally
- Markdown lint may flag stylistic choices that need local overrides

---

## Follow-ups

- Periodically review staticcheck rules for relevance
- Prefer narrow suppressions over broad exclusions
- Extend to `golangci-lint` when the project grows
