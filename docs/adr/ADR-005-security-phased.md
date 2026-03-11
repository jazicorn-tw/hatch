<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [adr, api, phases]
description:  "ADR-005: Phase security implementation (scaffold first, enforcement later)"
-->
# ADR-005: Phase Security Implementation (Scaffold First, Enforcement Later)

- **Status:** Accepted
- **Date:** 2026-03-10
- **Deciders:** Project maintainers
- **Scope:** HTTP API and SSH TUI authentication scaffolding and enforcement timeline

## Context

Hatch has two security surfaces:

1. **HTTP API** — REST endpoints served by the Go HTTP server
2. **SSH TUI** — Interactive terminal sessions over SSH (Charmbracelet Wish)

Early development phases focus on domain modeling, onboarding flows, and correctness.
Full security enforcement from day one can:

- Slow iteration during early phases
- Obscure domain-level failures
- Increase cognitive load during TDD setup

Security remains a **non-negotiable production requirement**.
The project roadmap includes explicit security milestones.

---

## Decision

Introduce security infrastructure early as a **scaffold configuration**, without enforcing it.
Enable enforcement at a defined phase milestone.

### HTTP API security

- Wire Go HTTP middleware for authentication from the start (even if it allows all requests)
- All endpoints are public during early phases
- JWT-based authentication enforced at **Phase 7**

### SSH TUI security

- Charmbracelet Wish supports SSH key authentication natively
- SSH key validation scaffolded early, enforcement deferred
- Session isolation and audit logging added progressively

### Structure

```tree
internal/
  middleware/
    auth.go       # JWT validation middleware (wired, not enforced early)
  ssh/
    auth.go       # SSH key validation (scaffolded early)
```

Clearly document the current security posture in the README to avoid confusion during
open or development phases.

---

## Consequences

### Positive

- Faster early development and clearer domain validation
- Security infrastructure is present from the start (no retrofitting)
- Realistic production architecture without premature enforcement

### Trade-offs

- Requires discipline to enable enforcement at the planned phase
- Unsecured endpoints during early phases must be clearly documented
- Development instances must never be exposed publicly before enforcement

## Rejected Alternatives

### Full enforcement from day one

Rejected — unnecessary complexity during early phases. Slows feedback loops and obscures
core domain issues.

### No security until "later"

Rejected — retrofitting security after the fact is expensive and error-prone.
Scaffolding early prevents architectural rework.

## Related ADRs

- ADR-003: Health endpoints (public by default — `/ping`, `/health`)
- ADR-009: Deployment strategy (production requires enforcement before going live)
