<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-10
status:       active
tags:         [adr]
description:  "Architecture Decision Records (ADR)"
-->
# Architecture Decision Records (ADR)

This folder contains **Architecture Decision Records** for **Hatch**.

ADRs capture *why* we made a decision, not just *what* we built.

---

## ADR Index

> Keep this list in numeric order. Link each ADR file.

- **ADR-000** — [Linting & static analysis](./ADR-000-linting.md) — gofmt, go vet, staticcheck, markdownlint-cli2
- **ADR-001** — [SQLite + sqlite-vec for embedded persistence](./ADR-001-database-postgresql.md) — no database server required
- **ADR-002** — [In-memory SQLite for Go integration tests](./ADR-002-testcontainers.md) — replaces Testcontainers
- **ADR-003** — [Health endpoints via Go net/http](./ADR-003-actuator-health.md) — `/ping` liveness, `/health` readiness
- **ADR-004** — [Environment config with godotenv](./ADR-004-env-and-config.md) — `.env` + OS vars
- **ADR-005** — [Phased security implementation](./ADR-005-security-phased.md) — Go middleware + SSH key auth
- **ADR-006** — [Local developer experience](./ADR-006-local-dev-experience.md) — doctor, verify, quality, bootstrap
- **ADR-007** — [Commit message enforcement](./ADR-007-commit-msg.md) — Conventional Commits, cz optional
- **ADR-008** — [CI-managed releases with semantic-release](./ADR-008-semantic-release.md) — Go ldflags versioning
- **ADR-009** — [Deployment strategy](./ADR-009-deployment-strategy.md) — binary distribution + Docker SSH TUI
- **ADR-010** — [Local CI simulation with `act`](./ADR-010-local-ci-simulation-with-act.md)
- **ADR-012** — [Branching strategy](./ADR-012-branching-strategy.md) — four-tier: feature → staging → canary → main

---

## When to write an ADR

Write (or update) an ADR when a change affects any of the following:

### Architecture & boundaries

- Introducing a new module, layer, or major package boundary
- Changing service boundaries or responsibility splits
- Adding a new public API style (REST changes, versioning strategy, pagination rules)

### Data & persistence

- Changing database technology, schema ownership, or migration strategy
- Introducing new persistence patterns (CQRS, outbox, event sourcing)
- Decisions that affect transactionality, consistency, or performance

### Security & compliance

- Introducing authentication or authorization (JWT, sessions, OAuth)
- New security posture (public vs protected endpoints, CORS, rate limiting)
- Secrets handling, encryption, PII handling, audit requirements

### Infrastructure & operability

- Changing deployment topology (Docker / Compose / Kubernetes)
- Runtime, ports, or healthcheck strategy changes
- Observability decisions (logging format, metrics, tracing, alerting)
- CI/CD policy changes (branch protection, merge rules, release automation)

### Testing strategy

- Changing integration test strategy (in-memory SQLite isolation, test lifecycle)
- Adding or removing test categories or quality gates (coverage thresholds, smoke tests)

---

## Lightweight ADR rule

If you’re unsure, default to writing a *small* ADR:

- One page maximum
- Clear decision, context, and consequences
- Alternatives can be brief (2–3 bullets)

---

## ADR review checklist

Before marking an ADR as **Accepted**, confirm:

- The decision is clearly stated
- The context is specific to this repository (not generic advice)
- Alternatives were considered
- Consequences and tradeoffs are explicit
- The decision is reflected in documentation and code

---

## Naming & status conventions

Recommended format:

- **Filename**: `ADR-00X-short-title.md`
- **Title**: `ADR-00X: Short Title`
- **Status**: `Proposed` → `Accepted` → `Superseded` (with a link)

---

## Cross-links

- **PHASES**: phase-gated ADRs are referenced per phase
- **Pull requests**: PR templates include an ADR checklist to keep decisions explicit
