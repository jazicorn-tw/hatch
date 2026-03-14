<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [onboarding, sqlite, db, concurrency]
description:  "Why hatch uses SQLite, what WAL mode is, and how it handles concurrent SSH connections"
-->
# SQLite and WAL Mode

Hatch uses a single SQLite file as its entire database — no separate database server
process, no Docker container, no connection pool config. One file: `~/.hatch/hatch.db`.

---

## Why SQLite

Most apps reach for PostgreSQL or MySQL. Hatch doesn't need to because:

| Concern            | Hatch's situation                                                |
| ------------------ | ---------------------------------------------------------------- |
| Concurrent writers | Low — ingestion is a CLI command, not continuous traffic         |
| Concurrent readers | Multiple SSH sessions read quiz content simultaneously           |
| Deployment         | Single binary on a VPS — no database server to install           |
| Dev setup          | `go build` is the only setup step                                |
| Ops                | Zero maintenance — no WAL archiving, no vacuum jobs, no replicas |

SQLite is embedded in the `hatch` binary via CGO (`mattn/go-sqlite3`). The file it
writes is just a file — back it up with `cp`, restore it with `cp`.

---

## The concurrency problem

By default, SQLite uses **rollback journal mode**. When a write happens:

1. SQLite locks the entire database file
2. All readers block until the write completes

For hatch, this means: if 10 juniors SSH in at the same time and load their quiz sessions
(reads), and someone triggers an ingestion (write), the readers would block and feel
slow or unresponsive.

---

## WAL mode fixes this

**WAL (Write-Ahead Log)** is an alternative journaling mode that separates reads from
writes:

- Writes go to a separate `hatch.db-wal` file
- Readers continue reading from the original `hatch.db` — **reads never block writes**
- Periodically, WAL content is checkpointed (merged) back into the main file

The practical result: multiple SSH sessions can read quiz content concurrently while an
ingestion write is in progress, with no blocking.

```text
hatch.db         ← main database — readers always access this
hatch.db-wal     ← pending writes — exists while WAL mode is active
hatch.db-shm     ← shared memory index — coordination between processes
```

All three files are managed automatically by SQLite. You don't need to touch them
directly. They disappear after a clean shutdown.

---

## How hatch enables WAL mode

In `internal/store/sqlite/sqlite.go`, WAL mode is set via the DSN (data source name)
passed to `sql.Open`:

```go
db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
```

The `?_journal_mode=WAL` query parameter is a `mattn/go-sqlite3` extension that runs
`PRAGMA journal_mode=WAL` immediately when the connection opens. No separate
configuration step required.

---

## Migrations

When `Open` is called, hatch automatically runs any pending SQL migrations before
returning. The migration files are baked into the binary (see [`GO_EMBED.md`](../go/GO_EMBED.md))
and tracked in a `schema_migrations` table:

```text
001_init.sql  → creates chunks table and schema_migrations
002_vec.sql   → creates vec_chunks virtual table for KNN search
```

If you delete `hatch.db` and start fresh, the migrations run again from scratch.

---

## SQLite limitations to be aware of

- **One writer at a time** — WAL allows concurrent reads, but writes still serialize.
  For hatch (occasional ingestion + SSH sessions), this is fine.
- **No network access** — SQLite is a local file. Remote machines can't connect to it
  directly. Hatch's SSH server and web dashboard run on the same machine as the binary.
- **File permissions** — the user running `hatch` must have write access to
  `~/.hatch/hatch.db`. CI tests use a temp directory (`t.TempDir()`).

---

## Related

- [`CGO.md`](../go/CGO.md) — why the sqlite-vec extension requires a C compiler
- [`VECTOR_EMBEDDINGS.md`](VECTOR_EMBEDDINGS.md) — the `vec_chunks` table and KNN search
- [`GO_EMBED.md`](../go/GO_EMBED.md) — how migration SQL files are baked into the binary

## Resources

- [SQLite WAL mode documentation](https://www.sqlite.org/wal.html) — full technical explanation
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) — the CGO driver used by hatch
- [SQLite: Appropriate uses](https://www.sqlite.org/whentouse.html) — when to use SQLite vs a server database
