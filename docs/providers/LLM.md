<!--
created_by:   jazicorn-tw
created_date: 2026-03-12
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [providers, llm]
description:  "How to set up each LLM provider for use with hatch."
-->
# LLM Providers

Step-by-step setup for each LLM provider supported by hatch.
Set `llm_provider` in `~/.hatch/config.yaml` or override with `HATCH_LLM_PROVIDER`.
`~` refers to your home directory (e.g. `/Users/yourname`) — not the project
repo. See [`CONFIGURATION.md`](CONFIGURATION.md) for full details.

---

## What is an LLM?

A Large Language Model (LLM) is an AI model trained on large amounts of text that can
generate human-like responses to natural language prompts. In hatch, the LLM is
responsible for generating quiz questions, evaluating kata answers, and producing
explanations — all from your ingested source material.

hatch treats the LLM as a swappable dependency. You point it at whichever provider
fits your setup (cloud API or local) and the rest of the application stays the same.

---

## Anthropic (default)

Uses the Anthropic Messages API. Requires an Anthropic account and API key.

| Property     | Value                       |
| ------------ | --------------------------- |
| Provider key | `anthropic`                 |
| Default      | yes                         |
| Required env | `ANTHROPIC_API_KEY`         |
| Recommended  | `claude-3-5-haiku-20241022` |

### Anthropic Setup

1. Create an API key at [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys).

2. Add the key to your environment. In `.env` (created by `./dev env init`):

   ```bash
   ANTHROPIC_API_KEY=sk-ant-...
   ```

   Or export directly:

   ```bash
   export ANTHROPIC_API_KEY=sk-ant-...
   ```

3. Set the provider in `~/.hatch/config.yaml`:

   ```yaml
   llm_provider: anthropic
   ```

   Or via environment variable:

   ```bash
   export HATCH_LLM_PROVIDER=anthropic
   ```

4. Verify by running hatch. If the API key is missing or invalid, hatch exits
   with a clear error before any completions are requested.

---

## OpenAI

Uses the OpenAI Chat Completions API. Requires an OpenAI account and API key.

| Property     | Value            |
| ------------ | ---------------- |
| Provider key | `openai`         |
| Default      | no               |
| Required env | `OPENAI_API_KEY` |
| Recommended  | `gpt-4o-mini`    |

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
   llm_provider: openai
   ```

   Or via environment variable:

   ```bash
   export HATCH_LLM_PROVIDER=openai
   ```

4. Verify by running hatch. If the API key is missing or invalid, hatch exits
   with a clear error before any completions are requested.

---

## Google Gemini

Uses the Google Generative AI API. Requires a Google Cloud account and API key.

| Property     | Value              |
| ------------ | ------------------ |
| Provider key | `gemini`           |
| Default      | no                 |
| Required env | `GOOGLE_API_KEY`   |
| Recommended  | `gemini-2.0-flash` |

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
   llm_provider: gemini
   ```

   Or via environment variable:

   ```bash
   export HATCH_LLM_PROVIDER=gemini
   ```

4. Verify by running hatch. If the API key is missing or invalid, hatch exits
   with a clear error before any completions are requested.

---

## Ollama

Runs entirely on-device. No API key or account required.

| Property     | Value                    |
| ------------ | ------------------------ |
| Provider key | `ollama`                 |
| Default      | no                       |
| Required env | none                     |
| Default host | `http://localhost:11434` |
| Recommended  | `llama3.2`, `mistral`    |

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

3. Pull a model:

   ```bash
   ollama pull llama3.2
   ```

4. Set the provider in `~/.hatch/config.yaml`:

   ```yaml
   llm_provider: ollama
   ```

   Or via environment variable:

   ```bash
   export HATCH_LLM_PROVIDER=ollama
   ```

5. Verify the model is available:

   ```bash
   ollama list
   ```

   The pulled model should appear in the output. hatch connects to Ollama
   automatically when generating completions.

---

## Implementation Status

| Provider           | Status | Milestone |
| ------------------ | ------ | --------- |
| Fake (test double) | ✅     | M1        |
| Anthropic          | 🔲     | M3        |
| Google Gemini      | 🔲     | M3        |
| OpenAI             | 🔲     | M3        |
| Ollama             | 🔲     | v4        |

---

## Related

- [`docs/PROVIDERS.md`](../PROVIDERS.md) — provider overview
- [`docs/providers/EMBEDDER.md`](EMBEDDER.md) — embedding provider setup
- [`docs/providers/CONFIGURATION.md`](CONFIGURATION.md) — full config reference
- [`docs/milestones/M1-foundation.md`](../milestones/M1-foundation.md) — fake provider implementation
