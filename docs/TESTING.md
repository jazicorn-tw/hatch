<!--
created_by:   jazicorn-tw
created_date: 2026-03-12
updated_by:   jazicorn-tw
updated_date: 2026-03-13
status:       active
tags:         [test, qa, go]
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
```

### Via dev script

```bash
./dev test
```

---

## Test Coverage

### `internal/config`

| Test                               | Covers                                                  |
| ---------------------------------- | ------------------------------------------------------- |
| `TestLoadDefaults`                 | Default values applied when no file or env vars are set |
| `TestLoadEnvOverride`              | `HATCH_*` env vars override file and defaults           |
| `TestValidateDefaults`             | Default config passes `Validate()`                      |
| `TestValidateUnknownLLMProvider`   | Unknown `llm_provider` rejected                         |
| `TestValidateUnknownEmbedProvider` | Unknown `embed_provider` rejected                       |
| `TestValidatePortOutOfRange`       | Port 0 or > 65535 rejected                              |

### `internal/store`

| Test                                | Covers                                         |
| ----------------------------------- | ---------------------------------------------- |
| `TestCosineOrthogonal`              | Orthogonal vectors return similarity 0         |
| `TestCosineIdentical`               | Identical vectors return similarity 1.0        |
| `TestCosineZeroVector`              | Zero vector returns 0 without dividing by zero |
| `TestCosineDimensionMismatchPanics` | Mismatched dimensions panic immediately        |

### `internal/store/memory`

| Test               | Covers                                                 |
| ------------------ | ------------------------------------------------------ |
| `TestAddAndSearch` | Records added and returned ranked by cosine similarity |
| `TestSearchEmpty`  | Search on empty store returns no results without error |

### `internal/store/sqlite`

| Test                     | Covers                                                  |
| ------------------------ | ------------------------------------------------------- |
| `TestOpen`               | Database file created and migrations applied            |
| `TestOpenRunsMigrations` | Opening the same DB twice is idempotent                 |
| `TestAddAndSearch`       | Records persisted and ranked by cosine similarity       |
| `TestSearchEmpty`        | Search on empty DB returns no results without error     |
| `TestAddReplaces`        | Inserting a duplicate ID replaces the existing record   |
| `TestEncodeDecodeVec`    | `float32` slice survives a round-trip encode/decode     |
| `TestDecodeVecOddBytes`  | Malformed blob (non-multiple of 4 bytes) returns nil    |

### `internal/embedder/fake`

| Test                  | Covers                                                   |
| --------------------- | -------------------------------------------------------- |
| `TestEmbed`           | Returns correct number of vectors at specified dimension |
| `TestEmbedDefaultDim` | Defaults to dimension 4 when `Dim` is unset              |

### `internal/llm/fake`

| Test                  | Covers                               |
| --------------------- | ------------------------------------ |
| `TestCompleteDefault` | Returns `"fake response"` by default |
| `TestCompleteCustom`  | Returns custom `Response` when set   |

---

## Test Doubles

Real LLM and embedder providers are not called in tests. Each has a fake
implementation in its `fake/` subdirectory:

| Package                  | Fake            | Purpose                                          |
| ------------------------ | --------------- | ------------------------------------------------ |
| `internal/llm/fake`      | `fake.LLM`      | Returns a configurable string; no network call   |
| `internal/embedder/fake` | `fake.Embedder` | Returns zero vectors at a configurable dimension |

Use `fake.LLM{Response: "..."}` and `fake.Embedder{Dim: N}` in any test that
needs to exercise code depending on these interfaces without a live provider.

---

## CI

The `test` job in `.github/workflows/ci.yml` runs:

```text
go build ./...
go test ./...
```

It is gated by `ENABLE_GO_ANALYSIS`. Set the variable to `FALSE` in GitHub
repo settings to skip Go analysis before source is scaffolded. Remove it
once `internal/` packages are in place.

See [`docs/devops/CI_VARIABLES.md`](devops/CI_VARIABLES.md) for details.

---

## Related

- [`docs/devops/CI_VARIABLES.md`](devops/CI_VARIABLES.md) — CI gate variables
- [`docs/milestones/M1-foundation.md`](milestones/M1-foundation.md) — M1 scope and fake implementations
