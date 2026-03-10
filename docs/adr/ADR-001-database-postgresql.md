<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-10
status:       active
tags:         [adr]
description:  "ADR-001: Use SQLite + sqlite-vec for embedded persistence"
-->

# ADR-001: Use SQLite + sqlite-vec for Embedded Persistence

- Date: 2026-03-10
- Status: Accepted

## Context

Hatch is a **Go CLI/SSH TUI developer onboarding tool** — not a multi-user server application.
Its primary data needs are:

- Storing local developer environment state
- Caching onboarding progress and phase completion
- Semantic search over documentation and context (embeddings)

A traditional client-server database (PostgreSQL, MySQL) would:

- Require a running server process on every developer's machine
- Add Docker or Colima as hard dependencies for basic operation
- Introduce network configuration and connection management overhead
- Be disproportionate to the tool's actual data volume and concurrency needs

---

## Decision

Use **SQLite** as the embedded database engine for all persistence needs.

Use **sqlite-vec** as the SQLite extension for vector/embedding storage and similarity search.

### Why SQLite

- Zero-config embedded database — no server, no Docker, no network
- Single file on disk — easy to inspect, backup, and reset
- First-class Go support via `database/sql` + `modernc.org/sqlite` (pure Go, no CGo)
- More than sufficient for a local CLI tool's data volume

### Why sqlite-vec

- Native vector similarity search inside SQLite
- Enables semantic search over documentation, phases, and onboarding context
- No external vector database required (no Pinecone, Weaviate, or similar)
- Keeps the dependency footprint minimal

---

## Consequences

### Positive

- No database server required — `go run ./...` is the full dev loop
- Single binary can carry the DB schema and migrations
- Trivial test isolation (in-memory SQLite per test)
- Offline-first by design
- Drastically simpler onboarding for contributors

### Trade-offs

- Not suitable if Hatch ever needs multi-user concurrent writes at scale
- sqlite-vec extension must be available at runtime (bundled or installed)
- Migrations must be handled in-process (no Flyway; use embedded SQL or a Go migration library)

## Related ADRs

- ADR-002: Testing strategy (SQLite in-memory)
- ADR-006: Local developer experience (no database server required)
