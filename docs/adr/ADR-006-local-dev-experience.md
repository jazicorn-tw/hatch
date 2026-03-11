<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [adr, dx, local, tooling, onboarding]
description:  "ADR-006: Local Developer Experience – gum-powered dev script, Doctor Checks, and Verification"
-->
# ADR-006: Local Developer Experience – gum-powered dev script, Doctor Checks, and Verification

- **Status:** Accepted
- **Date:** 2026-03-10
- **Deciders:** Project maintainers
- **Scope:** Local developer experience, quality gates, onboarding

---

## Context

This project prioritizes **production parity**, **explicit quality gates**, and **fast, actionable feedback** for developers.

Key characteristics of the codebase:

- Go **1.22+**
- Built on the **Charmbracelet** stack (bubbletea, wish, lipgloss)
- SQLite (in-memory) for integration tests — no Docker required for testing
- Docker only needed for deployment image builds and local container testing
- CI as the **authoritative enforcer** of quality (ADR-000)

Historically, failures caused by missing or misconfigured local infrastructure
(Go version, Docker, Colima, missing tools) surfaced **late** and with poor error
messages.

We want local failures to be:

- Fast
- Explicit
- Actionable
- Clearly distinguished from CI enforcement

---

## Decision

### 1. Introduce a local environment sanity check (`doctor`)

A local-only script is added:

```bash
scripts/doctor.sh
```

Responsibilities:

- Validate **Go** availability
- Verify Docker CLI availability
- Verify Docker daemon reachability
- Validate Docker socket health
- Verify **gum** availability (required for `./dev`)
- On macOS:
  - Detect Colima
  - Warn if Colima is not running (configurable; see `DOCTOR_REQUIRE_COLIMA_RUNNING`)
- Perform best-effort Docker memory checks
- Verify Node.js availability (required for semantic-release and Commitizen)

Design constraints:

- **Fail fast**
- **Never auto-start services**
- **Print explicit remediation instructions**
- **Exit immediately when `--allow-ci` flag is passed** (CI mode)

This script is a **local convenience tool** and is not used directly by CI.

---

### 2. Expose the check via `./dev doctor`

A single `dev` script at the repo root serves as the entry point for all local
developer tasks, powered by **gum** (Charmbracelet's CLI tool for glamorous shell scripts).

| Command         | Purpose                                                    |
| --------------- | ---------------------------------------------------------- |
| `./dev doctor`  | Runs `scripts/doctor.sh` to validate local setup           |
| `./dev`         | Interactive gum menu — pick a task                         |

Rationale:

- `./dev doctor` is memorable and human-friendly (ideal for onboarding)
- Interactive menu (`./dev` with no args) reduces cognitive overhead for new developers
- Consistent with the Charmbracelet ecosystem the project is built on

---

### 3. Standardize local workflows using `./dev`

`./dev` is a **gum-powered bash script** that orchestrates Go tools and scripts.

Key tasks:

| Command            | Meaning                                                         |
| ------------------ | --------------------------------------------------------------- |
| `./dev format`     | Run `gofmt` and format Go source                                |
| `./dev lint`       | Static analysis only (`go vet` + `markdownlint-cli2`)           |
| `./dev test`       | Run `go test ./...` (unit + integration via in-memory SQLite)   |
| `./dev verify`     | "Am I good to push?" (`doctor` + `lint` + `test`)               |
| `./dev quality`    | Local CI approximation (`doctor` + `format` + `lint` + `test`)  |
| `./dev pre-commit` | Pre-commit gate (`format` + `lint` + `test`)                    |
| `./dev bootstrap`  | First-time setup (`hooks` + `doctor` + `quality`)               |

Design principles:

- Tasks are **memorable**
- Tasks **do not replace CI**
- Tasks encode **intent**, not implementation detail
- **gum** provides spinners, styled output, and confirmations for destructive operations

---

### 4. Define `verify` as a developer-experience umbrella

The `verify` task intentionally exists to answer a human question:

> "Is this good enough to push or open a PR?"

It runs:

1. Environment sanity (`doctor`)
2. Static analysis (`lint`)
3. Unit tests (`test`)

It does **not** run formatting or generate artifacts.

CI remains authoritative.

---

### 5. Keep CI authoritative and isolated

CI behavior remains unchanged:

- CI runs `go build`, `go test`, and `staticcheck` directly
- CI does **not** invoke `./dev`
- CI enforces the quality gate via workflow steps

Guards are added so that `doctor.sh` exits immediately via `--allow-ci` if invoked in CI context.

---

## Consequences

### Positive

- Faster, clearer local failures
- Smoother onboarding — `./dev` interactive menu for new developers
- Reduced "it works on my machine" ambiguity
- Strong alignment with the Charmbracelet ecosystem the project is built on
- No database server required for tests

### Trade-offs

- Requires `gum` to be installed locally (`brew install gum`)
- The `./dev` script must be kept in sync with CI workflow steps
- Best-effort checks (e.g. Docker memory) may vary by provider

---

## Non-goals

- Replacing CI with local tooling
- Auto-starting Docker or Colima
- Supporting unsupported Go versions
- Requiring Docker for Go integration tests

---

## Related ADRs

- ADR-000 — Quality gates & CI authority
- ADR-001 — SQLite for persistence (no database server required)
- ADR-002 — In-memory SQLite for integration tests
