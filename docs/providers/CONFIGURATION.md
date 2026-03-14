<!--
created_by:   jazicorn-tw
created_date: 2026-03-12
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [providers, configuration, config]
description:  "Full configuration reference for hatch providers, ports, and runtime settings."
-->
# Configuration

Full reference for `~/.hatch/config.yaml` and the `HATCH_*` environment variables
that override it.

> `~` is shorthand for your home directory — for example, `/Users/yourname`
> on macOS or `/home/yourname` on Linux. `~/.hatch/` is a folder hatch creates
> inside your home directory, not inside the project repo. It does not exist
> until you run `hatch config init` for the first time.

---

## Initialise

Run this once after installing hatch:

```bash
hatch config init
```

Creates the `~/.hatch/` directory and writes `config.yaml` with default values.
If the file already exists, the command does nothing.

> This step is optional. hatch works without a config file — all keys fall back
> to their built-in defaults. Only run `hatch config init` if you want a file
> to edit.

---

## Config File

All keys are optional — defaults are applied for any key that is absent.

```yaml
llm_provider:   anthropic        # anthropic | openai | ollama
embed_provider: ollama           # ollama | openai
ssh_port:       2222
http_port:      8080
web_password:   changeme
jwt_secret:     ""
db_path:        ~/.hatch/hatch.db
```

---

## Environment Variables

Any config key can be overridden with a `HATCH_<KEY>` environment variable.
Environment variables take precedence over the config file.

| Variable               | Default              | Purpose                                        |
| ---------------------- | -------------------- | ---------------------------------------------- |
| `HATCH_LLM_PROVIDER`   | `anthropic`          | LLM provider (`anthropic`, `openai`, `ollama`) |
| `HATCH_EMBED_PROVIDER` | `ollama`             | Embedding provider (`ollama`, `openai`)        |
| `HATCH_SSH_PORT`       | `2222`               | SSH server port                                |
| `HATCH_HTTP_PORT`      | `8080`               | Web dashboard port                             |
| `HATCH_WEB_PASSWORD`   | `changeme`           | Web dashboard password                         |
| `HATCH_JWT_SECRET`     | _(empty)_            | Signs JWT tokens for SSH + web auth            |
| `HATCH_DB_PATH`        | `~/.hatch/hatch.db`  | Path to the SQLite database file               |

---

## Validation

hatch validates config at startup and exits with a clear error for any invalid value:

- `llm_provider` must be one of `anthropic`, `openai`, `ollama`
- `embed_provider` must be one of `ollama`, `openai`
- `ssh_port` and `http_port` must be in the range `[1, 65535]`

---

## Examples

### Local / offline (Ollama for both)

```yaml
llm_provider:   ollama
embed_provider: ollama
```

No API keys required. Requires a running Ollama instance with the target models pulled.

### Cloud LLM + cloud embeddings (OpenAI)

```yaml
llm_provider:   openai
embed_provider: openai
```

Requires `OPENAI_API_KEY`.

### Mixed (Anthropic LLM + Ollama embeddings)

```yaml
llm_provider:   anthropic
embed_provider: ollama
```

Requires `ANTHROPIC_API_KEY`. Default configuration.

---

## Related

- [`docs/PROVIDERS.md`](../PROVIDERS.md) — provider overview
- [`docs/providers/LLM.md`](LLM.md) — LLM provider details
- [`docs/providers/EMBEDDER.md`](EMBEDDER.md) — embedding provider details
