<!--
created_by:   jazicorn-tw
created_date: 2026-03-12
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [test, qa, go, ingest, pipeline, embedder]
description:  "Testing strategy, test coverage, and how to run tests locally and in CI."
-->
# Testing

---

## Running Tests

### All tests

```bash
go test ./...
```

### Verbose output

```bash
go test -v ./...
```

### Single package

```bash
go test ./internal/config/...
go test ./internal/store/...
go test ./internal/store/sqlite/...
go test ./internal/source/fs/...
go test ./internal/chunker/...
go test ./internal/pipeline/...
```

### Via dev script

```bash
./dev test
```

---

## Guides

- [Coverage](testing/COVERAGE.md) — per-package test inventory
- [Test Doubles](testing/TEST_DOUBLES.md) — fake implementations and usage
- [End-to-end CLI](testing/E2E.md) — manual ingest testing
- [CI](testing/CI.md) — GitHub Actions test job

---

## Related

- [`docs/devops/CI_VARIABLES.md`](devops/CI_VARIABLES.md) — CI gate variables
- [`docs/milestones/M1-foundation.md`](milestones/M1-foundation.md) — M1 scope and fake implementations
- [`docs/milestones/M2-ingestion.md`](milestones/M2-ingestion.md) — M2 ingestion pipeline scope
