<!--
created_by:   jazicorn-tw
created_date: 2026-03-11
updated_by:   jazicorn-tw
updated_date: 2026-03-16
status:       active
tags:         [devops, ci, config]
description:  "GitHub Actions repository variables — what each controls and when to set it"
-->
# CI Variables

GitHub Actions repository variables for this project. Set these in:
**GitHub → Settings → Secrets and variables → Actions → Variables**

---

## Go analysis

| Variable             | Type     | Default | Values          |
| -------------------- | -------- | ------- | --------------- |
| `ENABLE_GO_ANALYSIS` | Variable | enabled | `FALSE` / unset |

Gates both `go vet` and `staticcheck` in the `quality` job, and the `test` job
(`go build` + `go test`).

Set to `FALSE` **before Go source exists** to prevent CI from failing with
"matched no packages". Remove or set to `TRUE` once `internal/` is scaffolded.

---

## Markdown lint

| Variable          | Type     | Default | Values          |
| ----------------- | -------- | ------- | --------------- |
| `ENABLE_MD_LINT`  | Variable | enabled | `FALSE` / unset |

Gates the `markdown-lint` job. Disable temporarily if markdownlint is being
configured or if bulk doc changes are in progress.

---

## Frontmatter tag validation

| Variable            | Type     | Default | Values          |
| ------------------- | -------- | ------- | --------------- |
| `ENABLE_DOCS_TAGS`  | Variable | enabled | `FALSE` / unset |

Gates the `docs-tags` job in CI, which validates that all `docs/**/*.md`
frontmatter tags appear in the canonical vocabulary defined in
`.github/tags.yml`. Disable temporarily when adding new tags to `tags.yml`
before the updated list is merged.

---

## Doctor snapshot

| Variable                  | Type     | Default | Values          |
| ------------------------- | -------- | ------- | --------------- |
| `ENABLE_DOCTOR_SNAPSHOT`  | Variable | enabled | `FALSE` / unset |

Gates the `doctor` workflow entirely. The doctor snapshot runs on push to
`main`/`staging`/`canary` and manual dispatch — not on PRs. Disable if the
workflow is producing noise or consuming unnecessary CI minutes.

---

## SonarCloud

| Variable               | Type       | Default  | Values           |
| ---------------------- | ---------- | -------- | ---------------- |
| `ENABLE_SONAR`         | Variable   | enabled  | `FALSE` / unset  |
| `SONAR_ORGANIZATION`   | Variable   | —        | org slug         |
| `SONAR_PROJECT_KEY`    | Variable   | —        | project key      |
| `SONAR_TOKEN`          | **Secret** | —        | SonarCloud token |

The `sonar` job in `ci.yml` runs after `test` passes. It generates a Go
coverage report and runs the SonarCloud scanner. **When this job fails, CI
fails and semantic-release does not run.**

- `ENABLE_SONAR` — set to `FALSE` to skip the job entirely (e.g. before
  SonarCloud is configured).
- `SONAR_ORGANIZATION` — your SonarCloud organization slug (visible in the
  sonarcloud.io URL: `https://sonarcloud.io/organizations/<slug>`).
- `SONAR_PROJECT_KEY` — your SonarCloud project key (shown in project
  Administration → Project Key).
- `SONAR_TOKEN` — go to sonarcloud.io → My Account → Security → Generate
  Token, then add it as a **Secret** (not a variable) in GitHub Settings →
  Secrets and variables → Actions.

> **Important:** disable SonarCloud automatic analysis in your project's
> Administration → Analysis Method settings to avoid running the scanner
> twice per commit.

---

## Semantic release

| Variable                  | Type     | Default  | Values           |
| ------------------------- | -------- | -------- | ---------------- |
| `ENABLE_SEMANTIC_RELEASE` | Variable | disabled | `TRUE` / unset   |

Must be explicitly set to `TRUE` to allow `semantic-release` to cut a tag,
update `CHANGELOG.md`, and create a GitHub Release. Leave unset during active
development to prevent accidental releases.

---

## Publishing

| Variable               | Type     | Default  | Values     |
| ---------------------- | -------- | -------- | ---------- |
| `PUBLISH_DOCKER_IMAGE` | Variable | disabled | `TRUE`     |
| `PUBLISH_HELM_CHART`   | Variable | disabled | `TRUE`     |
| `CANONICAL_REPOSITORY` | Variable | —        | `org/repo` |

`CANONICAL_REPOSITORY` must match `github.repository` exactly for publish jobs
to run. Set to your GitHub org and repo name (e.g. `jazicorn/hatch`).

`PUBLISH_DOCKER_IMAGE` and `PUBLISH_HELM_CHART` each gate their respective
publish jobs independently.

---

## Changelog guard

| Variable                  | Type     | Default | Values                    |
| ------------------------- | -------- | ------- | ------------------------- |
| `GUARD_RELEASE_ARTIFACTS` | Variable | enabled | `FALSE` / unset           |
| `RELEASE_BOT_NAMES`       | Variable | —       | comma-separated bot names |

`GUARD_RELEASE_ARTIFACTS` blocks PRs that modify `CHANGELOG.md` unless the
author is in `RELEASE_BOT_NAMES`. Set `RELEASE_BOT_NAMES` to your semantic-
release bot's GitHub username (e.g. `github-actions[bot],release-bot`).

---

## Local CI override (`ACT`)

When running workflows locally via `act` (`./dev test-ci`), the runner exports
`ACT=true`. Workflows can branch on this to skip steps that require real GitHub
infrastructure (secrets, token scopes, publishing).

```yaml
- name: Skip publishing under act
  if: ${{ !env.ACT }}
  run: ...
```

---

## Quick-start variable checklist

For a fresh repository setup, set these variables before the first push:

```text
CANONICAL_REPOSITORY     = <org>/<repo>
ENABLE_GO_ANALYSIS       = FALSE        # until Go source is written
RELEASE_BOT_NAMES        = github-actions[bot]
```

Enable as the project progresses:

| Milestone      | Variable / secret to enable                                          |
| -------------- | -------------------------------------------------------------------- |
| M1 (Go src)    | Remove `ENABLE_GO_ANALYSIS=false`                                    |
| SonarCloud     | `SONAR_ORGANIZATION`, `SONAR_PROJECT_KEY`, `SONAR_TOKEN` (secret)    |
| First release  | `ENABLE_SEMANTIC_RELEASE=TRUE`, `PUBLISH_DOCKER_IMAGE=TRUE`          |
