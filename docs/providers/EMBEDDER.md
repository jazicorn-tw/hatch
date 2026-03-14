<!--
created_by:   jazicorn-tw
created_date: 2026-03-12
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [providers, embedder]
description:  "How to set up each embedding provider for use with hatch."
-->
# Embedder Providers

Step-by-step setup for each embedding provider supported by hatch.
Set `embed_provider` in `~/.hatch/config.yaml` or override with `HATCH_EMBED_PROVIDER`.
`~` refers to your home directory (e.g. `/Users/yourname`) — not the project
repo. See [`CONFIGURATION.md`](CONFIGURATION.md) for full details.

---

## What are embeddings?

Embeddings are numerical representations of text — each piece of content is converted
into a vector (a list of numbers) that captures its meaning. Similar content produces
similar vectors, which makes it possible to search by meaning rather than exact keywords.

In hatch, the embedder converts your ingested source material into vectors stored in
SQLite. When a junior searches the knowledge base or hatch retrieves context for a quiz
question, it compares vectors to find the most relevant chunks — this is called semantic
search.

hatch treats the embedder as a swappable dependency. You can run it locally via Ollama
or use a cloud API, without changing any application code.

---

## Ollama (default)

Runs entirely on-device. No API key or account required.

| Property     | Value                    |
| ------------ | ------------------------ |
| Provider key | `ollama`                 |
| Default      | yes                      |
| Required env | none                     |
| Default host | `http://localhost:11434` |
| Recommended  | `nomic-embed-text`       |
| Vector dim   | 768 (`nomic-embed-text`) |

### Ollama Setup

1. Install Ollama — download from [ollama.com](https://ollama.com) or via Homebrew:

   ```bash
   brew install ollama
   ```

2. Start the Ollama server (leave running in the background):

   ```bash
   ollama serve
   ```

   Ollama listens on `http://localhost:11434` by default.

3. Pull the embedding model:

   ```bash
   ollama pull nomic-embed-text
   ```

4. Set the provider in `~/.hatch/config.yaml`:

   ```yaml
   embed_provider: ollama
   ```

   Or via environment variable:

   ```bash
   export HATCH_EMBED_PROVIDER=ollama
   ```

5. Verify the model is available:

   ```bash
   ollama list
   ```

   `nomic-embed-text` should appear in the output. hatch connects to Ollama
   automatically when ingesting or searching.

---

## OpenAI

Uses the OpenAI Embeddings API. Requires an OpenAI account and API key.

| Property     | Value                           |
| ------------ | ------------------------------- |
| Provider key | `openai`                        |
| Default      | no                              |
| Required env | `OPENAI_API_KEY`                |
| Recommended  | `text-embedding-3-small`        |
| Vector dim   | 1536 (`text-embedding-3-small`) |

### OpenAI Setup

1. Create an API key at [platform.openai.com/api-keys](https://platform.openai.com/api-keys).

2. Add the key to your environment. In `.env` (created by `./dev env init`):

   ```bash
   OPENAI_API_KEY=sk-...
   ```

   Or export directly:

   ```bash
   export OPENAI_API_KEY=sk-...
   ```

3. Set the provider in `~/.hatch/config.yaml`:

   ```yaml
   embed_provider: openai
   ```

   Or via environment variable:

   ```bash
   export HATCH_EMBED_PROVIDER=openai
   ```

4. Verify by ingesting a small source. If the API key is missing or invalid,
   hatch exits with a clear error before any embeddings are generated.

---

## Google Gemini

Uses the Google Generative AI Embeddings API. Requires a Google Cloud account and API key.

| Property     | Value                      |
| ------------ | -------------------------- |
| Provider key | `gemini`                   |
| Default      | no                         |
| Required env | `GOOGLE_API_KEY`           |
| Recommended  | `text-embedding-004`       |
| Vector dim   | 768 (`text-embedding-004`) |

### Google Gemini Setup

1. Create an API key at [aistudio.google.com/app/apikey](https://aistudio.google.com/app/apikey).

2. Add the key to your environment. In `.env` (created by `./dev env init`):

   ```bash
   GOOGLE_API_KEY=AIza...
   ```

   Or export directly:

   ```bash
   export GOOGLE_API_KEY=AIza...
   ```

3. Set the provider in `~/.hatch/config.yaml`:

   ```yaml
   embed_provider: gemini
   ```

   Or via environment variable:

   ```bash
   export HATCH_EMBED_PROVIDER=gemini
   ```

4. Verify by ingesting a small source. If the API key is missing or invalid,
   hatch exits with a clear error before any embeddings are generated.

---

## Implementation Status

| Provider           | Status | Milestone |
| ------------------ | ------ | --------- |
| Fake (test double) | ✅     | M1        |
| OpenAI             | ✅     | M2        |
| Google Gemini      | 🔲     | M3        |
| Ollama             | 🔲     | v4        |

---

## Related

- [`docs/PROVIDERS.md`](../PROVIDERS.md) — provider overview
- [`docs/providers/LLM.md`](LLM.md) — LLM provider setup
- [`docs/providers/CONFIGURATION.md`](CONFIGURATION.md) — full config reference
- [`docs/milestones/M1-foundation.md`](../milestones/M1-foundation.md) — fake provider implementation
