<!--
created_by:   jazicorn-tw
created_date: 2026-03-11
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [onboarding, faq]
description:  "Frequently asked questions for new contributors"
-->
# FAQ

Common questions from new contributors.

---

## Setup

**`gum` is not found after `go install`**

Go installs binaries to `$GOPATH/bin`. Add it to your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add that line to your `~/.zshrc` or `~/.bashrc` to make it permanent.

---

**`./dev bootstrap` failed on the `doctor` step**

Run `./dev doctor` on its own to see exactly which check failed:

```bash
./dev doctor
```

Each failure includes a remediation message. Fix the flagged tool, then re-run `./dev bootstrap`.

---

**Do I need Docker to run tests?**

No. Tests use in-memory SQLite and have no external dependencies. `go test ./...` works with no services running.

---

**Do I need Docker at all for local development?**

Only if you want to test container builds or run the full local environment (`./dev env up`). For Go development — editing, testing, linting — Docker is not required.

---

## Commits

### My commit was rejected by the hook

The `commit-msg` hook enforces [Conventional Commits](https://www.conventionalcommits.org/). Use:

```bash
cz commit
```

Or fix your message to match the format:

```text
<type>(<optional scope>): <description>
```

See [`docs/commit/COMMIT_CHEAT_SHEET.md`](../commit/COMMIT_CHEAT_SHEET.md).

---

### My scope was rejected as invalid

Scopes must be in the list defined in [`.github/tags.yml`](../../.github/tags.yml). Either use a valid scope or omit the scope entirely:

```text
feat: description without scope
```

---

**`gofmt` changed my files after I committed**

The pre-commit hook runs `./dev format` automatically. If formatting changed files, re-stage and commit again:

```bash
git add -A
git commit -m "your message"
```

---

### I need to skip a hook in an emergency

```bash
SKIP_COMMIT_MSG_CHECK=1 git commit -m "..."   # skip commit-msg validation
SKIP_QUALITY=1 git commit -m "..."            # skip pre-commit gate entirely
git commit --no-verify                        # bypass all hooks
```

Use sparingly. CI will still catch any issues.

---

## Workflow

**Which branch do I create my work from?**

Always branch from `staging`:

```bash
git checkout staging && git pull
git checkout -b feature/<name>
```

See [`docs/onboarding/CONTRIBUTING.md`](CONTRIBUTING.md) for the full branching guide.

---

**Can I push directly to `main` or `staging`?**

No. All changes go through PRs. `main` is only updated via CI-promoted PRs from `canary`.

---

**What does `./dev verify` check?**

It runs `doctor` (environment checks) + `lint` (go vet + markdownlint) + `test` (go test ./...). Run it before opening a PR to catch issues locally before CI does.

---

## Related

- [`docs/onboarding/PROJECT_SETUP.md`](PROJECT_SETUP.md) — initial setup
- [`docs/onboarding/CONTRIBUTING.md`](CONTRIBUTING.md) — branching and PRs
- [`docs/tooling/DEV.md`](../tooling/DEV.md) — `./dev` task reference
- [`docs/commit/COMMIT_CHEAT_SHEET.md`](../commit/COMMIT_CHEAT_SHEET.md) — commit format
