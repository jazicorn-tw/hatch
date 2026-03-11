<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [tooling]
description:  "dev — gum-powered dev task runner"
-->
# dev — gum-powered dev task runner

`./dev` is the single entry point for all local developer tasks in this project.

It is a **gum-powered bash script** that wraps Go tools and shell scripts with
spinners, styled output, and confirmations — consistent with the Charmbracelet
ecosystem the project is built on.

> **Important**
> `./dev` is a local convenience tool. CI does **not** invoke it.
> CI remains the only authoritative quality gate.

---

## Usage

```bash
./dev                    # interactive menu (requires TTY)
./dev <task>             # run a task directly
./dev <task> <subtask>   # run a namespaced task (e.g. ./dev docker up)
```

---

## Tasks

### Quality

| Command            | What it does                                      |
| ------------------ | ------------------------------------------------- |
| `./dev format`     | Auto-format Go source with `gofmt`                |
| `./dev lint`       | Static analysis: `go vet` + `markdownlint-cli2`   |
| `./dev lint:docs`  | Markdown lint only                                |
| `./dev test`       | Run `go test ./...`                               |
| `./dev verify`     | `doctor` + `lint` + `test` (am I ready to push?)  |
| `./dev quality`    | `doctor` + `format` + `lint` + `test` (CI parity) |
| `./dev pre-commit` | Pre-commit gate: `format` + `lint` + `test`       |

### Setup

| Command           | What it does                                        |
| ----------------- | --------------------------------------------------- |
| `./dev bootstrap` | First-time setup: `hooks` + `doctor` + `quality`    |
| `./dev doctor`    | Validate local dev environment                      |
| `./dev hooks`     | Install repo-managed git hooks                      |

### Application

| Command     | What it does                                           |
| ----------- | ------------------------------------------------------ |
| `./dev run` | Load `.env` and start the application (`go run ./...`) |

### Environment

| Command            | What it does                                    |
| ------------------ | ----------------------------------------------- |
| `./dev env up`     | Start local dev environment (Colima + Docker)   |
| `./dev env down`   | Stop local dev environment                      |
| `./dev env status` | Show Docker context, Colima, running containers |
| `./dev env init`   | Initialise `.env` from template                 |

### Docker

| Command              | What it does                                           |
| -------------------- | ------------------------------------------------------ |
| `./dev docker up`    | Start Docker Compose services (`docker compose up -d`) |
| `./dev docker down`  | Stop Docker Compose services                           |
| `./dev docker reset` | down -v + up -d (destroys volumes, confirms first)     |

### Clean

| Command              | What it does                                          |
| -------------------- | ----------------------------------------------------- |
| `./dev clean local`  | Remove local build artifacts (confirms first)         |
| `./dev clean docker` | Prune Docker build cache (confirms first)             |
| `./dev clean colima` | Delete Colima containers and volumes (confirms first) |

---

## Interactive menu

Running `./dev` with no arguments opens an interactive picker powered by `gum choose`:

```bash
./dev
```

Select a task with arrow keys and press Enter. Requires a TTY — pipe-safe
(non-TTY invocations without a task argument will error).

---

## Overrides

One-off environment variable overrides for `./dev pre-commit`:

| Variable         | Default | Effect                              |
| ---------------- | ------- | ----------------------------------- |
| `AUTO_FORMAT=0`  | `1`     | Skip the auto-format step           |
| `SKIP_TESTS=1`   | `0`     | Skip the test step                  |

```bash
AUTO_FORMAT=0 ./dev pre-commit   # lint + test only, no format
SKIP_TESTS=1  ./dev pre-commit   # format + lint only, no tests
```

---

## How it works

`./dev` is a plain bash script (`set -euo pipefail`) that:

1. **Resolves gum** — checks `$PATH` first, then falls back to `$(go env GOPATH)/bin/gum`
2. **Parses the task** — first arg is the task, optional second arg is the subtask (e.g. `docker up` → `docker:up` internally)
3. **Falls back to interactive** — no args + TTY opens `gum choose` menu
4. **Dispatches via `case`** — each task calls the underlying Go tool or script, wrapped in a `gum spin` spinner

Go-specific tasks (`format`, `lint`, `test`) guard against projects with no `.go` files yet and skip gracefully with a warning instead of erroring.

---

## Dependency: gum

`gum` is required. Install it via Go modules (recommended — consistent with how this project manages tooling):

```bash
go install github.com/charmbracelet/gum@latest
```

Then ensure Go's bin directory is on your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add that line to your `~/.zshrc` or `~/.bashrc` to make it permanent.

---

## Related

- `./dev doctor` → [`docs/tooling/DOCTOR.md`](DOCTOR.md)
- `./dev bootstrap` / `./dev hooks` → [`docs/tooling/BOOTSTRAP.md`](BOOTSTRAP.md)
- `./dev pre-commit` → [`docs/commit/PRECOMMIT.md`](../commit/PRECOMMIT.md)
- [`docs/adr/ADR-006-local-dev-experience.md`](../adr/ADR-006-local-dev-experience.md) — decision record for this tooling approach
