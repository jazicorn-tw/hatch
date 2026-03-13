<!--
created_by:   jazicorn-tw
created_date: 2026-03-12
updated_by:   jazicorn-tw
updated_date: 2026-03-13
status:       active
tags:         [providers, llm, embeddings, configuration]
description:  "Supported LLM and embedding providers, how to configure them, and what each one requires."
-->
# Providers

Reference for LLM and embedding providers supported by hatch. Providers are selected via
config or environment variable and swapped without changing any application code.

---

## Overview

hatch separates two provider concerns:

| Concern    | Interface                    | Config key       | Env var                |
| ---------- | ---------------------------- | ---------------- | ---------------------- |
| LLM        | `internal/llm.LLM`           | `llm_provider`   | `HATCH_LLM_PROVIDER`   |
| Embeddings | `internal/embedder.Embedder` | `embed_provider` | `HATCH_EMBED_PROVIDER` |

Both are provider-agnostic by design — the interfaces accept any conforming implementation.
Set the provider name via environment variable, or run `hatch config init` to create
`~/.hatch/config.yaml` and edit it there. The `~/.hatch/` folder does not exist until
that command is run for the first time.

---

## LLM Providers

### Anthropic

Generates completions via the Anthropic Messages API.

| Property     | Value                       |
| ------------ | --------------------------- |
| Provider key | `anthropic`                 |
| Default      | yes                         |
| Required env | `ANTHROPIC_API_KEY`         |
| Recommended  | `claude-3-5-haiku-20241022` |

Set `HATCH_LLM_PROVIDER=anthropic` and provide `ANTHROPIC_API_KEY` in `.env` or the
shell environment. No additional dependencies required.

### OpenAI

Generates completions via the OpenAI Chat Completions API.

| Property     | Value              |
| ------------ | ------------------ |
| Provider key | `openai`           |
| Default      | no                 |
| Required env | `OPENAI_API_KEY`   |
| Recommended  | `gpt-4o-mini`      |

Set `HATCH_LLM_PROVIDER=openai` and provide `OPENAI_API_KEY`.

### Ollama

Generates completions via a locally running [Ollama](https://ollama.com) instance.
No API key required — Ollama runs entirely on-device.

| Property     | Value                    |
| ------------ | ------------------------ |
| Provider key | `ollama`                 |
| Default      | no                       |
| Required env | none                     |
| Default host | `http://localhost:11434` |
| Recommended  | `llama3.2`, `mistral`    |

Set `HATCH_LLM_PROVIDER=ollama`. Ollama must be running locally with the target model
pulled (`ollama pull llama3.2`).

---

## Embedding Providers

### OpenAI Embeddings

Generates embeddings via the OpenAI Embeddings API.

| Property     | Value                           |
| ------------ | ------------------------------- |
| Provider key | `openai`                        |
| Default      | no                              |
| Required env | `OPENAI_API_KEY`                |
| Recommended  | `text-embedding-3-small`        |
| Vector dim   | 1536 (`text-embedding-3-small`) |

Set `HATCH_EMBED_PROVIDER=openai` and provide `OPENAI_API_KEY`.

### Ollama Embeddings

Generates embeddings via a locally running Ollama instance.
No API key required — runs entirely on-device.

| Property     | Value                    |
| ------------ | ------------------------ |
| Provider key | `ollama`                 |
| Default      | yes                      |
| Required env | none                     |
| Default host | `http://localhost:11434` |
| Recommended  | `nomic-embed-text`       |
| Vector dim   | 768 (`nomic-embed-text`) |

Set `HATCH_EMBED_PROVIDER=ollama`. Pull the embedding model first:
`ollama pull nomic-embed-text`.

---

## Configuration Examples

### Local / offline (Ollama for both)

```yaml
llm_provider: ollama
embed_provider: ollama
```

### Cloud LLM + cloud embeddings (OpenAI)

```yaml
llm_provider: openai
embed_provider: openai
```

### Mixed (Anthropic LLM + Ollama embeddings)

```yaml
llm_provider: anthropic
embed_provider: ollama
```

---

## Implementation Status

| Provider           | LLM | Embeddings | Milestone |
| ------------------ | --- | ---------- | --------- |
| Fake (test double) | ✅  | ✅         | M1        |
| Anthropic          | 🔲  | —          | M2        |
| OpenAI             | 🔲  | 🔲         | M2        |
| Ollama             | 🔲  | 🔲         | M2        |

---

## Related

- [`docs/milestones/M1-foundation.md`](milestones/M1-foundation.md) — fake provider implementation
- [`docs/ROADMAP.md`](ROADMAP.md) — M2 ingestion pipeline adds real provider implementations
- [`internal/llm/`](../internal/llm/) — LLM interface definition
- [`internal/embedder/`](../internal/embedder/) — Embedder interface definition
