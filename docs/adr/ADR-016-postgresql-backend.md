<!--
created_by:   jazicorn-tw
created_date: 2026-03-15
updated_by:   jazicorn-tw
updated_date: 2026-03-15
status:       active
tags:         [adr, db, deploy, architecture]
description:  "ADR-016: Optional PostgreSQL backend for enterprise deployments (M6c)"
-->
# ADR-016: Optional PostgreSQL Backend for Enterprise Deployments

- **Status:** Proposed
- **Date:** 2026-03-15
- **Deciders:** Project maintainers
- **Scope:** `internal/store/`, `internal/embedder/`, `~/.hatch/config.yaml`, Helm chart — M6c

---

## Context

Hatch's default persistence layer is SQLite + sqlite-vec (ADR-001, ADR-012). This is the
right default for the majority of deployments: zero external dependencies, single-file
database, works out of the box for local CLI and self-hosted Docker/Kubernetes.

However, some enterprise teams deploying hatch internally already operate PostgreSQL
infrastructure. For these teams:

- Spinning up a new SQLite-backed pod introduces an unfamiliar operational pattern
- Existing DBA tooling (backups, monitoring, replication) does not apply to SQLite
- PostgreSQL with `pgvector` extension handles concurrent writes better at scale — relevant
  when a team has 10+ juniors actively taking quizzes simultaneously
- Some enterprise security policies require all persistent state to live in a managed
  database with audit logging

A configurable backend — SQLite (default) or PostgreSQL (opt-in) — covers both
deployment profiles without forcing enterprise teams to adopt an embedded database.

---

## Decision

**Introduce an optional PostgreSQL backend as an enterprise deployment option at M6c.**
SQLite remains the default. PostgreSQL is opt-in via a config flag.

### Store interface

Define clean interfaces in `internal/store/` that both backends implement:

```go
// internal/store/interface.go
type ChunkStore interface { ... }
type SessionStore interface { ... }
type EmbedStore interface { ... }
```

### Package layout

```text
internal/store/
  interface.go          ← ChunkStore, SessionStore, EmbedStore interfaces
  sqlite/               ← current SQLite + sqlite-vec implementation
  postgres/             ← new PostgreSQL + pgvector implementation
```

### Configuration

```yaml
# ~/.hatch/config.yaml
database:
  driver: sqlite          # sqlite (default) | postgres
  dsn: ~/.hatch/hatch.db  # file path for SQLite; connection string for PostgreSQL
```

### PostgreSQL driver

Use `pgx/v5` — pure Go, no CGO. PostgreSQL support does not reintroduce CGO requirements.

### Vector search

Replace `sqlite-vec` KNN search with `pgvector` cosine similarity:

```sql
-- pgvector equivalent of sqlite-vec KNN
SELECT id, content FROM chunks
ORDER BY embedding <=> $1
LIMIT $2;
```

### Helm chart

Add an optional `postgresql` subchart dependency (Bitnami) that teams can enable:

```yaml
# values.yaml
postgresql:
  enabled: false   # set true for PostgreSQL backend
database:
  driver: sqlite
```

---

## Alternatives Considered

### 1. PostgreSQL only — drop SQLite

Simplifies the codebase to one backend, but breaks the zero-dependency local CLI and
self-hosted Docker use cases. Enterprise teams are a minority of the target audience.

**Rejected** — SQLite is a core deployment advantage for the majority of users.

### 2. Support both backends from M1

Clean architecture from day one, but premature for a project still building core features.
Requires upfront interface design before the store contract is stable.

**Rejected** — deferred to M6c when the store interface has been proven through M1–M6.

### 3. Use `database/sql` with a pluggable driver

`database/sql` works for standard SQL queries, but sqlite-vec uses a non-standard
extension API (`sqlite3_load_extension`) and pgvector uses non-standard operators (`<=>`).
A shared `database/sql` layer cannot abstract these differences cleanly.

**Rejected** — requires two separate query layers regardless; explicit backend packages
are clearer.

---

## Consequences

### Positive

- Enterprise teams with existing PostgreSQL can adopt hatch without new infrastructure
- pgvector handles concurrent vector search at scale — better for large junior cohorts
- `pgx/v5` is pure Go — no CGO complexity, Docker build stays simple for PG backend
- Helm chart `postgresql` subchart gives a complete K8s deployment with one values flag
- SQLite default is unchanged — zero regression for existing users

### Negative

- Two store implementations to maintain — schema changes must be applied to both backends
- pgvector requires the `vector` extension installed in PostgreSQL (not always available
  in managed PG offerings)
- Local development with PostgreSQL requires a running server (mitigated by the existing
  Testcontainers setup in ADR-002)

### Follow-up

- Document which managed PostgreSQL offerings include pgvector (AWS RDS, Supabase, Neon)
- Add a `hatch doctor` check that validates database connectivity and pgvector availability
  when `database.driver: postgres` is configured

---

## Related ADRs

- **ADR-001** — SQLite + sqlite-vec: the default backend this extends
- **ADR-002** — In-memory SQLite for tests: test strategy for SQLite; PostgreSQL tests use Testcontainers
- **ADR-012** — sqlite-vec: vector search implementation replaced by pgvector for PG backend
- **ADR-015** — CGO cross-compilation: pgx/v5 is pure Go, so PG backend does not reintroduce CGO
