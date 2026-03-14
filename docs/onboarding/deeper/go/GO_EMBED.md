<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [onboarding, go]
description:  "How //go:embed bakes files into the hatch binary at compile time"
-->
# `//go:embed` — Baking Files Into the Binary

Go lets you pack files directly into your binary at compile time using the `//go:embed`
directive. The files become part of the executable — no separate files needed at runtime.

---

## How it works

You add a special comment directly above a variable declaration. The Go compiler reads
that comment and packs the matching files into the binary when you run `go build`.

```go
import "embed"

//go:embed migrations/*.sql
var migrationsFS embed.FS
```

At build time: the compiler finds all `.sql` files under `migrations/` and stores their
contents inside the binary.

At runtime: `migrationsFS` behaves like a normal filesystem — you can open files, read
them, walk directories — but the data comes from memory, not disk.

---

## How hatch uses it today

In `internal/store/sqlite/sqlite.go`:

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS
```

The migration files live at:

```text
internal/store/sqlite/migrations/
  001_init.sql    ← creates chunks table, indexes
  002_vec.sql     ← creates chunk_vecs virtual table for KNN search
```

When `hatch` starts up and opens the database, it reads those SQL files from `migrationsFS`
and runs any migrations that haven't been applied yet. This works even on a machine that
has never seen the hatch source code — the SQL is baked into the binary.

---

## What hatch will embed in future milestones

| File(s)                              | Milestone | Purpose                            |
| ------------------------------------ | --------- | ---------------------------------- |
| `migrations/*.sql`                   | M1 / M2   | Database schema setup ✅ done      |
| `question_mcq.tmpl`, etc.            | M3        | LLM prompt templates               |
| `kata_generate.tmpl`                 | M3b       | Kata generation prompt             |
| `internal/api/static/dist/`          | M8        | React + Vite web dashboard         |

The same pattern applies to all of them — one `//go:embed` line, and the files ship with
the binary automatically.

---

## Why this matters

Without `//go:embed`, you'd need to distribute extra files alongside the binary:

```text
hatch            ← binary
migrations/      ← also needed at runtime
  001_init.sql
  002_vec.sql
```

With `//go:embed`, there's just one file:

```text
hatch            ← binary (migrations included inside)
```

No risk of the files getting out of sync with the binary version, no missing files on
the server, no extra deployment steps.

---

## Related

- [`GO_BINARY.md`](GO_BINARY.md) — what a Go binary is and what's inside it
- [`CGO.md`](CGO.md) — why sqlite-vec requires a C compiler

## Resources

- [embed package docs](https://pkg.go.dev/embed) — full reference for `//go:embed`
- [Go blog: Using go embed](https://go.dev/blog/embed) — practical examples and patterns
