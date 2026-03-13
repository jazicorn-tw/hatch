<!--
created_by:   jazicorn-tw
created_date: 2026-03-11
updated_by:   jazicorn-tw
updated_date: 2026-03-13
status:       active
tags:         [onboarding]
description:  "How to build and run Hatch after cloning"
-->
# Hatch — Project Setup

How to get Hatch building and running locally after cloning the repository.

---

## Prerequisites

| Tool      | Version  | Install                                           |
| --------- | -------- | ------------------------------------------------- |
| Go        | 1.26+    | [go.dev/dl](https://go.dev/dl)                    |
| Node.js   | 20+      | `brew install node` or [nvm](https://nvm.sh)      |
| gum       | latest   | `go install github.com/charmbracelet/gum@latest`  |

> **gum** is required by the `./dev` task runner. After installing via `go install`,
> make sure Go's bin directory is on your PATH:
>
> ```bash
> export PATH="$PATH:$(go env GOPATH)/bin"
> ```
>
> Add that line to your `~/.zshrc` or `~/.bashrc` to make it permanent.

---

## First-time setup

```bash
git clone https://github.com/jazicorn/hatch.git
cd hatch
./dev bootstrap
```

`bootstrap` runs three steps in order:

1. **`hooks`** — installs repo-managed git hooks into `.git/hooks/`
2. **`doctor`** — validates your local environment (Go, Node, gum, Docker)
3. **`quality`** — runs the full quality gate (`format` + `lint` + `test`)

---

## Build

```bash
go build ./...
```

---

## Test

```bash
./dev test
# or directly:
go test ./...
```

Tests use in-memory SQLite — no database server or Docker required.

---

## Common dev tasks

| Command              | What it does                                           |
| -------------------- | ------------------------------------------------------ |
| `./dev format`       | Auto-format Go source with `gofmt`                     |
| `./dev lint`         | Static analysis: `go vet` + `markdownlint`             |
| `./dev lint:docs`    | Lint markdown files only                               |
| `./dev test`         | Run `go test ./...`                                    |
| `./dev pre-commit`   | Pre-commit gate: `format` + `lint` + `test`            |
| `./dev verify`       | `doctor` + `lint` + `test` (am I ready to push?)       |
| `./dev quality`      | `doctor` + `format` + `lint` + `test` (CI parity)      |
| `./dev doctor`       | Validate local dev environment                         |
| `./dev bootstrap`    | First-time setup: `hooks` + `doctor` + `quality`       |
| `./dev hooks`        | Install git hooks into `.git/hooks/`                   |
| `./dev exec-bits`    | Check and fix executable bits on scripts and hooks     |
| `./dev run`          | Load `.env` and start the application (`go run ./...`) |
| `./dev test-ci`      | Run local-safe CI workflows via `act`                  |
| `./dev changelog`    | Preview semantic-release changelog (dry-run)           |

Run `./dev` with no arguments to open an interactive task picker.

---

## Environment

Copy the env template before running the application:

```bash
./dev env init
```

This creates a `.env` file from the template. Edit it to add your API keys
and configuration before running `./dev run`.

---

## Verify everything works

```bash
./dev verify
```

This runs `doctor` + `lint` + `test` and exits 0 if everything is healthy.

---

## Related docs

- [`docs/tooling/DEV.md`](../tooling/DEV.md) — full `./dev` task reference
- [`docs/tooling/DOCTOR.md`](../tooling/DOCTOR.md) — what `doctor` checks
- [`docs/tooling/BOOTSTRAP.md`](../tooling/BOOTSTRAP.md) — bootstrap details
- [`docs/commit/PRECOMMIT.md`](../commit/PRECOMMIT.md) — pre-commit hook
- [`docs/providers/CONFIGURATION.md`](../providers/CONFIGURATION.md) — config file and env vars
