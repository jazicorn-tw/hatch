<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [onboarding, go]
description:  "What a Go binary is and why hatch ships as one file"
-->
# Go Binary — What It Is

When you run `go build`, Go compiles your entire program into **one self-contained
executable file**. No interpreter, no runtime to install, no package folder — just `hatch`.

```bash
go build -o hatch ./cmd/hatch
# → produces one file: ./hatch
```

That single file contains:

- **Your application code** — every package under `cmd/` and `internal/`
- **All dependencies** — the OpenAI SDK, Cobra, Viper, Bubble Tea, sqlite-vec bindings, etc.
- **The Go runtime** — garbage collector, goroutine scheduler, panic handler

Copy that file to another machine running the same OS and CPU architecture, and it runs.
No `go install`, no `npm install`, nothing else required.

---

## How that compares to Python

In Python, `import requests` means the `requests` package must be installed separately
on every machine that runs the script:

```bash
# Python: install dependencies on every machine
pip install -r requirements.txt
python app.py
```

In Go, all imports are compiled in at build time:

```bash
# Go: copy one file, done
scp hatch server:/usr/local/bin/hatch
```

---

## What's inside the hatch binary

| What                   | How it got there                                      |
| ---------------------- | ----------------------------------------------------- |
| CLI commands           | `cmd/hatch/*.go` compiled in                          |
| Config + Viper         | `internal/config/` compiled in                        |
| Ingestion pipeline     | `internal/pipeline/` compiled in                      |
| OpenAI + Gemini SDKs   | imported packages compiled in                         |
| SQLite + sqlite-vec    | `mattn/go-sqlite3` via CGO compiled in                |
| SQL migration files    | `//go:embed migrations/*.sql` baked in                |
| Chunkers, embedders    | `internal/chunker/`, `internal/embedder/` compiled in |

What's **not** inside the binary:

- `~/.hatch/config.yaml` — user-specific settings, lives on disk
- `~/.hatch/hatch.db` — the SQLite database, created on first run
- API keys — secrets are never compiled in

---

## Building vs running

| Command                          | What it does                                            |
| -------------------------------- | ------------------------------------------------------- |
| `go build ./...`                 | Compile all packages; check for errors. No output file. |
| `go build -o hatch ./cmd/hatch`  | Produce the `hatch` executable                          |
| `go run ./cmd/hatch`             | Compile + run without writing a file to disk            |
| `./dev run`                      | Same as `go run ./...` but loads env / 1Password first  |
| `go test ./...`                  | Compile + run all tests                                 |

`go run` is convenient during development. `go build` is what you use when distributing.

---

## Related

- [`GO_EMBED.md`](GO_EMBED.md) — how hatch uses `//go:embed` to bake files into the binary
- [`CGO.md`](CGO.md) — why sqlite-vec requires a C compiler and what that means

## Resources

- [Go: How to Write Go Code](https://go.dev/doc/code) — official intro to packages, modules, and `go build`
- [Go: Executable size and what's inside](https://go.dev/blog/executable-size) — what's actually inside a Go binary
