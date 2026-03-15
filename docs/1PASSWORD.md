<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-15
status:       active
tags:         [security, config, tooling]
description:  "1Password CLI integration for managing API keys in hatch"
-->
# 1Password CLI Integration

This project uses the 1Password CLI (`op`) to manage API keys. This document
covers every approach, when to use each one, and how to troubleshoot common
errors.

## How It Works

API keys (Anthropic, OpenAI, Google) are stored as items in your 1Password
vault. The CLI fetches them at runtime so secrets never live in plain text on
disk or in your shell history.

Hatch reads `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, and `GEMINI_API_KEY`
directly from the environment. When a required key is missing or empty, `hatch`
exits with a configuration error.

## Approaches

### Option A — `.env.op` file (Recommended)

The project ships a `.env.op` file containing `op://` URI references:

```text
ANTHROPIC_API_KEY=op://Private/ANTHROPIC_API_KEY/credential
OPENAI_API_KEY=op://Private/OPENAI_API_KEY/credential
# GEMINI_API_KEY=op://Private/GEMINI_API_KEY/credential   ← uncomment if you have this
```

These are URIs, not actual keys — the file is safe to commit. Secrets are
resolved by `op run` at the moment you run the command, so 1Password only
needs to be unlocked then (not at terminal startup).

> **Important:** `op run` errors on any `op://` reference it cannot resolve.
> Only uncomment lines for keys that actually exist in your 1Password vault.

```bash
# Run hatch with resolved secrets
op run --env-file .env.op -- hatch ingest --source=<name>

# Run tests with live keys
op run --env-file .env.op -- go test ./...

# Run any hatch command with resolved secrets
op run --env-file .env.op -- hatch <command>
```

### Option B — `.zshrc` with `op read`

Keys are resolved once when you open a terminal via `$(op read ...)` in your
`~/.zshrc`:

```bash
export ANTHROPIC_API_KEY=$(op read "op://Private/ANTHROPIC_API_KEY/credential")
export OPENAI_API_KEY=$(op read "op://Private/OPENAI_API_KEY/credential")
export GEMINI_API_KEY=$(op read "op://Private/GEMINI_API_KEY/credential")
```

1Password must be unlocked at terminal startup. If it is not, `op read` fails
silently and the variable is exported as an empty string — which triggers the
error described in [Troubleshooting](#troubleshooting) below.

### Option C — `.zshrc` with `op://` URI + `op run`

Export the `op://` URI directly (no `$(op read ...)`) and always run via
`op run --`:

```bash
# ~/.zshrc
export ANTHROPIC_API_KEY="op://Private/ANTHROPIC_API_KEY/credential"
```

```bash
op run -- hatch <command>
```

`op run` sees the `op://` value in the environment and resolves it before
spawning the child process. 1Password does not need to be unlocked at terminal
startup.

## Comparison

|                                 | Option A `.env.op` | Option B `.zshrc op read` | Option C `.zshrc URI` |
| ------------------------------- | ------------------ | ------------------------- | --------------------- |
| 1Password locked at startup OK  | ✅                 | ❌                        | ✅                    |
| No `op run` wrapper needed      | ❌                 | ✅                        | ❌                    |
| File documents required secrets | ✅                 | ❌                        | ❌                    |
| Safe to commit                  | ✅                 | n/a                       | n/a                   |
| Empty-string risk               | ❌                 | ✅                        | ❌                    |

## One-Time Setup

```bash
# 1. Install 1Password CLI
brew install --cask 1password/tap/1password-cli

# 2. Sign in (links CLI to your 1Password desktop app)
eval $(op signin)

# 3. Verify
op whoami
```

After this, 1Password CLI uses biometric unlock via the desktop app — you
rarely need to run `op signin` again unless your session expires.

## Daily Usage

```bash
# Recommended: use .env.op (works regardless of terminal age)
op run --env-file .env.op -- hatch <command>

# Or if using .zshrc op read: just open a fresh terminal
# The key is loaded automatically on shell startup
```

## Troubleshooting

### `ANTHROPIC_API_KEY is set but empty`

```text
OSError: Environment variable ANTHROPIC_API_KEY is set but empty.
Your 1Password CLI integration may not be authenticated.

Option A — use the project .env.op file (avoids this issue):
  op run --env-file .env.op -- hatch <command>

Option B — sign in and reload your shell:
  eval $(op signin) && source ~/.zshrc
```

**Cause:** Your `.zshrc` runs `op read` at terminal startup but 1Password was
not unlocked at that moment. The variable was exported as an empty string.

**Fix (Option A — avoids this permanently):**

```bash
op run --env-file .env.op -- hatch <command>
```

**Fix (Option B — for the current terminal):**

```bash
eval $(op signin) && source ~/.zshrc
hatch <command>
```

### `ANTHROPIC_API_KEY contains a 1Password URI`

```text
OSError: Environment variable ANTHROPIC_API_KEY contains a 1Password URI,
not an actual key. Run your command via op run so it is resolved.
```

**Cause:** You exported `ANTHROPIC_API_KEY=op://...` directly but ran the
command without `op run`.

**Fix:**

```bash
op run --env-file .env.op -- hatch <command>
# or
op run -- hatch <command>
```

### `op: command not found`

1Password CLI is not installed or not in `PATH`:

```bash
# Install
brew install --cask 1password/tap/1password-cli

# Verify
which op
op --version
```

### `[AUTH] error: you are not currently signed in`

```bash
eval $(op signin)
```

If you have the 1Password desktop app open, it will prompt for biometric
confirmation and the CLI will be authorized automatically.

## Key not loading in fresh terminal

Confirm your `.zshrc` contains the export:

```bash
grep ANTHROPIC ~/.zshrc
# Should show: export ANTHROPIC_API_KEY=$(op read "op://...")
```

If it is missing, add it or switch to Option A (`.env.op`).
