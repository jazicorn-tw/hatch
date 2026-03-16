<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-16
status:       active
tags:         [test, ci, actions]
description:  "GitHub Actions CI job configuration and gating variables."
-->
# CI

`.github/workflows/ci.yml` runs the following jobs on every pull request and
push to `main`/`staging`:

| Job              | What it runs                                | Gate variable        |
| ---------------- | ------------------------------------------- | -------------------- |
| `quality`        | `go vet ./...` + `staticcheck`              | `ENABLE_GO_ANALYSIS` |
| `test`           | `go build ./...` + `go test ./...`          | `ENABLE_GO_ANALYSIS` |
| `sonar`          | SonarCloud quality gate (runs after `test`) | `ENABLE_SONAR`       |
| `markdown-lint`  | `markdownlint-cli2`                         | `ENABLE_MD_LINT`     |
| `docs-tags`      | frontmatter tag validation                  | `ENABLE_DOCS_TAGS`   |

**All jobs must pass for `release.yml` to trigger.** The `sonar` job is the
CI-side enforcement of the SonarCloud quality gate — if it fails, the release
bot will not run.

Set a gate variable to `FALSE` in GitHub Settings → Variables to skip that job.

See [`docs/devops/CI_VARIABLES.md`](../devops/CI_VARIABLES.md) for details on
each variable, including how to configure `SONAR_TOKEN`, `SONAR_ORGANIZATION`,
and `SONAR_PROJECT_KEY`.

---

## Related

- [`docs/TESTING.md`](../TESTING.md) — running tests and guide index
- [`docs/devops/CI_VARIABLES.md`](../devops/CI_VARIABLES.md) — CI gate variables
