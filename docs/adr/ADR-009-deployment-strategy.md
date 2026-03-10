<!--
created_by:   jazicorn-tw
created_date: 2026-03-05
updated_by:   jazicorn-tw
updated_date: 2026-03-10
status:       active
tags:         [adr]
description:  "ADR-009: Deployment Strategy"
-->
# ADR-009: Deployment Strategy

- **Status:** Accepted
- **Date:** 2026-03-10
- **Deciders:** Project maintainers
- **Scope:** Application deployment and runtime environment

---

## Context

Hatch is a **Go CLI/SSH TUI developer onboarding tool**. It is distributed and run in two modes:

- **Local CLI**: installed directly on a developer's machine, run as a binary
- **SSH TUI server**: run as a Docker container, accessible via SSH from any terminal

The project prioritizes:

- low operational overhead
- CI-first quality gates
- production parity without premature complexity
- offline-first local operation (SQLite, no external database)

---

## Decision

### Phase 1: Binary distribution + Docker SSH server

**Local CLI**:

- Built as a single Go binary via CI
- Distributed via GitHub Releases (attached to semantic-release tags)
- Version embedded at build time via Go `ldflags` (see ADR-008)

**SSH TUI server**:

- Packaged as a Docker image published to `ghcr.io`
- Run via Docker Compose or as a standalone container
- Runtime configuration supplied via environment variables (`.env` or container env)
- Health checks use `/ping` (liveness) and `/health` (readiness) — see ADR-003

This phase validates:

- binary correctness and portability
- container build and SSH entrypoint
- configuration discipline via environment variables
- release stability via semantic-release

### Phase 2: Kubernetes + Helm (future)

When operational scale requires it:

- Adopt Kubernetes as the runtime platform
- Use Helm charts for packaging and configuration
- Align Helm chart versions with semantic-release tags

Helm is introduced early (linting only) to keep the chart valid without requiring
Kubernetes today.

---

## Consequences

### Positive

- Zero external dependencies for local CLI users
- SSH TUI is accessible from any terminal with no local install
- Clear separation between release and deployment concerns
- No Kubernetes tax before it provides real value

### Trade-offs

- Binary distribution requires per-platform builds (darwin/amd64, linux/amd64, etc.)
- Kubernetes features (HPA, rolling strategies) are deferred
- SSH key management is a new operational concern compared to HTTP-only tools

---

## Notes

- Docker image publishing is toggled via `PUBLISH_DOCKER_IMAGE` repository variable (see ADR-008)
- Helm charts are linted in CI but not deployed in Phase 1
- Kubernetes adoption will be revisited when operational scale requires it

---

## Relationship to release artifacts

Release artifacts (Go binaries + Docker images) are produced independently via CI, as defined
in **ADR-008**.

- semantic-release remains the sole authority for version creation
- Docker image publishing may be enabled or disabled via repository variables
- Deployment and release lifecycles are intentionally decoupled

This separation ensures that:

- releases remain reproducible
- deployments remain reversible
- operational changes do not require source code changes

---

## Related ADRs

- **ADR-003** — Health endpoints (`/ping`, `/health`)
- **ADR-008** — CI-managed releases with semantic-release
- **ADR-012** — Branching strategy (canary → main promotion)
