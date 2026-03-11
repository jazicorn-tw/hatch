<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [adr, test, db]
description:  "ADR-002: Use in-memory SQLite for Go integration tests"
-->

# ADR-002: Use In-Memory SQLite for Go Integration Tests

- **Status:** Accepted
- **Date:** 2026-03-10
- **Deciders:** Project maintainers
- **Scope:** Integration testing strategy for all database-backed packages

---

## Context

The system stores state in SQLite (ADR-001). Integration tests must exercise real database behavior —
schema creation, query correctness, migration application — without:

- Requiring a running database server
- Coupling tests to Docker availability
- Introducing per-test cleanup complexity

Since the project uses SQLite (not PostgreSQL), there is no need for Testcontainers or any
container-based test database.

---

## Decision

Use **in-memory SQLite** for all integration tests via the `:memory:` DSN:

```go
db, err := sql.Open("sqlite", ":memory:")
```

Each test that requires a database gets its own isolated in-memory instance.
Migrations are applied programmatically at test startup.

Use **testify** (`github.com/stretchr/testify`) for assertions.

---

## Consequences

### Positive

- Tests are fully isolated — no shared state between runs
- No Docker required for testing
- Test startup is near-instant (no container pull or port binding)
- Identical behavior locally and in CI
- Zero manual database setup for contributors

### Trade-offs

- In-memory SQLite loses data on connection close — intentional for isolation
- Tests must apply migrations themselves (no shared persistent state)
- sqlite-vec behavior must be explicitly loaded for tests that require vector search

## Explicit Non-Goals

- Supporting Docker-based test databases
- Optimizing test speed at the cost of correctness
- Sharing database state between test cases

## Related ADRs

- ADR-001: SQLite + sqlite-vec for persistence
- ADR-006: Local developer experience
