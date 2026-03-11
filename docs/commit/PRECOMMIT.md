<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [commit]
description:  "Pre-commit hook"
-->
# Pre-commit hook

This repo uses a **`./dev pre-commit`** gate to enforce quality checks *before* code leaves your machine.

The goal is **fast, deterministic feedback** that prevents CI failures and catches common issues early.

---

## What runs on `git commit`

All branches run the same three steps:

| Step | Command        | What it does             |
| ---- | -------------- | ------------------------ |
| 1    | `./dev format` | Auto-format with `gofmt` |
| 2    | `./dev lint`   | `go vet` + markdown lint |
| 3    | `./dev test`   | `go test ./...`          |

---

## Dev tasks

| Command            | Description                               |
| ------------------ | ----------------------------------------- |
| `./dev pre-commit` | Run the full pre-commit gate              |
| `./dev format`     | Auto-format with `gofmt`                  |
| `./dev lint`       | `go vet` + markdown lint                  |
| `./dev test`       | `go test ./...`                           |
| `./dev quality`    | doctor + format + lint + test (CI parity) |
| `./dev exec-bits`  | Check + fix executable bits on scripts    |

---

## Run the same checks without committing

```bash
./dev pre-commit
```

Or target a specific step:

```bash
./dev format     # gofmt only
./dev lint       # go vet + markdownlint
./dev test       # go test ./...
./dev exec-bits  # check + fix executable bits on scripts
./dev quality    # doctor + format + lint + test (matches CI)
```

---

## Overrides (one-off per commit)

Skip the pre-commit gate once:

```bash
SKIP_QUALITY=1 git commit -m "..."
```

Skip only auto-format:

```bash
AUTO_FORMAT=0 git commit -m "..."
```

Skip only tests:

```bash
SKIP_TESTS=1 git commit -m "..."
```

Hard bypass (skips *all* Git hooks):

```bash
git commit --no-verify
```

---

## Setup

The pre-commit hook requires Git hooks to be configured:

```bash
./dev hooks
```

or during first-time setup:

```bash
./dev bootstrap
```
