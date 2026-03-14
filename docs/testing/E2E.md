<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [test, e2e, ingest, cli]
description:  "Manual end-to-end testing of the hatch ingest CLI commands."
-->
# End-to-end CLI

Manual ingestion testing requires an OpenAI API key and a config file.

---

## Prerequisites

`~/.hatch/config.yaml`:

```yaml
openai_api_key: sk-...
sources:
  - name: myproject
    path: /path/to/repo
```

---

## Commands

```bash
# Ingest a named source
hatch ingest --source=myproject

# List configured sources
hatch sources list

# Remove a source and its indexed records
hatch sources remove --name=myproject
```

---

## Chunker dispatch

| File extensions               | Chunker    |
| ----------------------------- | ---------- |
| `.go`, `.ts`, `.tsx`, `.scss` | `code`     |
| All others                    | `markdown` |

---

## Related

- [`docs/TESTING.md`](../TESTING.md) — running tests and guide index
- [`docs/testing/CI.md`](CI.md) — GitHub Actions test job
