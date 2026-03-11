<!--
created_by:   jazicorn-tw
created_date: 2026-03-11
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [onboarding, commit, dev]
description:  "How to contribute to Hatch — branching, commits, and pull requests"
-->
# Contributing to Hatch

Everything you need to open a pull request.

---

## Branching

Always branch from `staging`:

```bash
git checkout staging
git pull
git checkout -b feature/<short-description>
# or
git checkout -b fix/<short-description>
```

Rules:

- Branch names use lowercase kebab-case
- One concern per branch — no mixing features and fixes
- **Never branch from `main` or `canary`**

### Branch flow

```text
feature/<name>  ──┐
                  ├──► staging ──► canary ──► main
fix/<name>      ──┘
```

| Branch      | PR targets  |
| ----------- | ----------- |
| `feature/*` | `staging`   |
| `fix/*`     | `staging`   |
| `staging`   | `canary`    |
| `canary`    | `main`      |

---

## Commit messages

This project enforces [Conventional Commits](https://www.conventionalcommits.org/).

```text
<type>(<optional scope>): <description>
```

Common types:

| Type       | When to use                  |
| ---------- | ---------------------------- |
| `feat`     | New user-facing capability   |
| `fix`      | Bug fix                      |
| `docs`     | Documentation only           |
| `test`     | Tests only                   |
| `refactor` | No behavior change           |
| `chore`    | Maintenance / housekeeping   |
| `ci`       | GitHub Actions changes       |
| `build`    | Build tooling / dependencies |

Use `cz commit` for an interactive prompt:

```bash
cz commit
```

Valid scopes are defined in [`.github/tags.yml`](../../.github/tags.yml).

See [`docs/commit/COMMIT_CHEAT_SHEET.md`](../commit/COMMIT_CHEAT_SHEET.md) for the full reference.

---

## Before pushing

Run the quality gate:

```bash
./dev verify
```

This checks your environment (`doctor`), lints, and runs tests. Fix any failures before opening a PR.

---

## Opening a pull request

1. Push your branch and open a PR targeting **`staging`**
2. CI runs automatically — all checks must pass
3. Request a review once CI is green
4. Squash-merge when approved

Do not open PRs directly to `canary` or `main`.

---

## Related

- [`docs/onboarding/PROJECT_SETUP.md`](PROJECT_SETUP.md) — local environment setup
- [`docs/commit/COMMIT_CHEAT_SHEET.md`](../commit/COMMIT_CHEAT_SHEET.md) — commit format reference
- [`docs/commit/PRECOMMIT.md`](../commit/PRECOMMIT.md) — pre-commit hook
- [`docs/adr/ADR-011-branching-strategy.md`](../adr/ADR-011-branching-strategy.md) — branching ADR
