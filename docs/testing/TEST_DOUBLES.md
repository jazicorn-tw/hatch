<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [test, fakes, go]
description:  "Fake implementations used in tests to avoid live provider calls."
-->
# Test Doubles

Real LLM, embedder, source, and store providers are not called in tests. Each has a fake
implementation in its `fake/` subdirectory:

| Package                  | Fake                | Purpose                                                               |
| ------------------------ | ------------------- | --------------------------------------------------------------------- |
| `internal/llm/fake`      | `fake.LLM`          | Returns a configurable string; no network call                        |
| `internal/embedder/fake` | `fake.Embedder`     | Returns zero vectors at a configurable dimension                      |
| `internal/source/fake`   | `fake.Fetcher`      | Returns configurable docs; supports error injection                   |
| `internal/store/fake`    | `fake.VecStore`     | In-memory VecStore that records `Upsert`/`DeleteBySource` call counts |

Use these in any test that needs to exercise code depending on these interfaces without a
live provider:

```go
fake.LLM{Response: "..."}
fake.Embedder{Dim: N}
fake.Fetcher{Docs: []source.Document{...}}
fakestore.New()
```

---

## Related

- [`docs/TESTING.md`](../TESTING.md) — running tests and guide index
- [`docs/testing/COVERAGE.md`](COVERAGE.md) — per-package test inventory
