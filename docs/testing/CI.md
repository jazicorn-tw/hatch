<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [test, ci, actions]
description:  "GitHub Actions test job configuration and gating variables."
-->
# CI

The `test` job in `.github/workflows/ci.yml` runs:

```text
go build ./...
go test ./...
```

It is gated by `ENABLE_GO_ANALYSIS`. Set the variable to `FALSE` in GitHub
repo settings to skip Go analysis before source is scaffolded. Remove it
once `internal/` packages are in place.

See [`docs/devops/CI_VARIABLES.md`](../devops/CI_VARIABLES.md) for details.

---

## Related

- [`docs/TESTING.md`](../TESTING.md) — running tests and guide index
- [`docs/devops/CI_VARIABLES.md`](../devops/CI_VARIABLES.md) — CI gate variables
