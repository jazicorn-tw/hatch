<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-10
status:       active
tags:         [adr]
description:  "ADR-006: Local Developer Experience – Makefile Commands, Doctor Checks, and Verification"
-->
# ADR-006: Local Developer Experience – Makefile Commands, Doctor Checks, and Verification

- **Status:** Accepted
- **Date:** 2026-03-10
- **Deciders:** Project maintainers
- **Scope:** Local developer experience, quality gates, onboarding

---

## Context

This project prioritizes **production parity**, **explicit quality gates**, and **fast, actionable feedback** for developers.

Key characteristics of the codebase:

- Go **1.22+**
- `make`-driven builds and workflows
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

### 2. Expose the check via a single human-facing Make target: `doctor`

The following Make targets are defined for environment readiness:

| Target         | Purpose                                                    |
| -------------- | ---------------------------------------------------------- |
| `doctor`       | Runs `scripts/doctor.sh` to validate local setup           |
| `doctor-json`  | Outputs structured JSON report of environment state        |
| `check-env`    | Verifies required environment variables are present        |
| `help`         | Lists all Make targets by category                         |

Rationale:

- `doctor` is memorable and human-friendly (ideal for onboarding)
- A single command avoids redundancy and cognitive overhead
- JSON output enables scripted checks and CI doctor gates

---

### 3. Standardize local workflows using Make

Make is used as a **thin orchestration layer** over Go tools and scripts.

Key targets:

| Target       | Meaning                                                              |
| ------------ | -------------------------------------------------------------------- |
| `format`     | Run `gofmt` and format Go source                                     |
| `lint`       | Static analysis only (`go vet`, `staticcheck`)                       |
| `lint-docs`  | Lint Markdown docs (`markdownlint-cli2`)                             |
| `test`       | Run `go test ./...` (unit + integration via in-memory SQLite)        |
| `test-ci`    | Run tests in CI mode (strict output, no TTY)                         |
| `verify`     | "Am I good to push?" (`doctor` + `lint` + `test`)                    |
| `quality`    | Local CI approximation (`doctor` + `format` + `lint` + `test`)       |
| `pre-commit` | Branch-aware pre-commit gate (runs `quality` on main, else `verify`) |
| `bootstrap`  | First-time setup (`hooks` + `doctor` + `quality`)                    |

Design principles:

- Make targets are **memorable**
- Make targets **do not replace CI**
- Make targets encode **intent**, not implementation detail

---

### 4. Define `verify` as a developer-experience umbrella

The `verify` target intentionally exists to answer a human question:

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
- CI does **not** invoke Make
- CI enforces the quality gate via workflow steps

Guards are added so that even if Make targets are accidentally invoked in CI,
local-only helpers (`doctor.sh`) exit immediately via `--allow-ci`.

---

## Consequences

### Positive

- Faster, clearer local failures
- Smoother onboarding — `go run ./...` is the full dev loop
- Reduced "it works on my machine" ambiguity
- Strong alignment between docs, tooling, and CI
- No database server required for tests

### Trade-offs

- Additional scripts and documentation must be maintained
- Best-effort checks (e.g. Docker memory) may vary by provider
- Makefile introduces a small abstraction layer over Go tools

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
