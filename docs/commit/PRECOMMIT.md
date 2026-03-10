<!--
created_by:   jazicorn-tw
created_date: 2026-03-09
updated_by:   jazicorn-tw
updated_date: 2026-03-10
status:       active
tags:         [commit]
description:  "Pre-commit hook"
-->
# Pre-commit hook

This repo uses a **`make pre-commit`** smart gate to enforce quality checks *before* code leaves your machine.

The goal is **fast, deterministic feedback** that prevents CI failures and catches common issues early.

---

## What runs on `git commit`

The gate applies a **branch-aware strategy**:

| Branch | What runs                      | Why                                    |
| ------ | ------------------------------ | -------------------------------------- |
| `main` | `make quality` (full gate)     | CI parity — doctor + go vet + go test  |
| Other  | `make format lint test` (fast) | Faster feedback on feature branches    |

---

## Make targets

| Target            | Description                                   |
| ----------------- | --------------------------------------------- |
| `make pre-commit` | Smart gate — strict on main, fast on branches |
| `make format`     | Auto-format with `gofmt`                      |
| `make lint`       | `go vet` + markdown lint                      |
| `make test`       | `go test ./...`                               |
| `make quality`    | doctor + `go vet` + `go test` (CI parity)     |

---

## Run the same checks without committing

```bash
make pre-commit
```

Or target a specific step:

```bash
make format     # gofmt only
make lint       # go vet + markdownlint
make test       # go test ./...
make quality    # doctor + go vet + go test (matches CI)
```

---

## Overrides (one-off per commit)

Skip the pre-commit gate once:

```bash
SKIP_QUALITY=1 git commit -m "..."
```

Hard bypass (skips *all* Git hooks):

```bash
git commit --no-verify
```

---

## Setup

The pre-commit hook requires Git hooks to be configured:

```bash
make hooks
```

or during first-time setup:

```bash
make bootstrap
```
