<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [onboarding, commit]
description:  "How conventional commits work in hatch — format, scopes, types, and what happens at release time"
-->
# Conventional Commits — How Hatch Commits Work

Every commit in hatch follows the **Conventional Commits** format. This isn't just style
preference — the CI release pipeline reads commit messages to decide whether to cut a
new version, what version number to use, and what goes in the changelog.

---

## The format

```text
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Examples:

```text
feat(ingest): add Google Gemini embedder provider
fix(store): handle empty chunk slice in Upsert
docs(onboarding): add deeper Go binary explainer
refactor(config): extract validateProvider helper
chore(deps): update go-sqlite3 to v1.14.25
```

---

## Types

| Type       | When to use                 | Triggers a release?  |
| ---------- | --------------------------- | -------------------- |
| `feat`     | New user-facing capability  | Yes — **minor** bump |
| `fix`      | Bug fix                     | Yes — **patch** bump |
| `docs`     | Documentation only          | No                   |
| `test`     | Tests only                  | No                   |
| `refactor` | No behaviour change         | No                   |
| `perf`     | Performance improvement     | Usually patch        |
| `build`    | Build tooling, dependencies | No                   |
| `ci`       | GitHub Actions changes      | No                   |
| `chore`    | Maintenance, housekeeping   | No                   |
| `revert`   | Revert a previous commit    | Depends              |

Breaking changes (`feat!` or a `BREAKING CHANGE:` footer) trigger a **major** bump.
Hatch is currently `0.x` — the first planned major release is Phase 7 (JWT auth).

---

## Scopes

A scope is the subsystem you're changing. It goes in parentheses after the type:

```text
feat(embedder): ...
fix(store): ...
docs(onboarding): ...
```

Scopes are optional but strongly recommended. They show up in the changelog grouped by
subsystem, so reviewers can scan what changed without reading every message.

### Where scopes come from

Valid scopes are defined in `.github/tags.yml` under the `both:` and `scopes:` keys.
The commit-msg hook rejects scopes not in that file.

**Selected valid scopes:**

```text
agent      api        build      chunker    ci         cli
config     db         deploy     embedder   env        hooks
ingest     llm        onboarding pipeline   providers  qa
release    security   source     store      test       tooling
deps       ...
```

Run `cat .github/tags.yml` to see the full list.

---

## How releases work

You **never** run `cz bump` or tag a release locally. The CI pipeline owns releases:

```text
you push a commit to main
  └──► GitHub Actions: ci.yml runs tests
         └──► semantic-release reads commit messages since last tag
                └──► if feat or fix commits exist → cut a new release
                       └──► bumps version, generates CHANGELOG.md, creates GitHub Release
```

Semantic-release uses `scripts/git/semantic-release-impact.mjs` to determine the impact
of each commit type. If you only push `docs` and `chore` commits, no release is cut.

---

## The pre-commit hooks

Hatch uses two commit-time hooks that run before your commit is recorded:

| Hook           | What it checks                                                                                           |
| -------------- | -------------------------------------------------------------------------------------------------------- |
| `commit-msg`   | Validates the commit message against the conventional commit schema and the allowed scopes in `tags.yml` |
| `pre-commit`   | Runs `gofmt`, imports formatting, and lint checks on staged files                                        |

If either hook fails, the commit is rejected. Fix the issue and try again.

The easiest way to write a valid commit message is with Commitizen:

```bash
cz commit
```

Commitizen walks you through type → scope → description interactively and constructs
the message for you.

---

## Imperative mood

Always write the description in imperative mood — as if completing the sentence
"This commit will...":

```text
✅  add Gemini embedder
✅  fix null pointer in Upsert
✅  update onboarding docs

❌  added Gemini embedder
❌  fixes null pointer
❌  updating docs
```

Keep the subject line under ~72 characters. Add a body if you need to explain _why_.

---

## Quick self-check before pushing

```bash
# Lint + format
go vet ./...
gofmt -l .

# Run tests
go test ./...
```

---

## Related

- [`docs/commit/COMMIT_CHEAT_SHEET.md`](../../../commit/COMMIT_CHEAT_SHEET.md) — one-page reference card
- [`docs/commit/PRECOMMIT.md`](../../../commit/PRECOMMIT.md) — what the pre-commit hook does
- [`docs/commit/COMMITIZEN.md`](../../../commit/COMMITIZEN.md) — how to use `cz commit`

## Resources

- [Conventional Commits spec](https://www.conventionalcommits.org/en/v1.0.0/) — the full specification
- [semantic-release](https://github.com/semantic-release/semantic-release) — the CI release tool that reads commit messages
- [Commitizen](https://commitizen-tools.github.io/commitizen/) — interactive commit message helper
