<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-10
status:       active
tags:         [adr]
description:  "ADR-004: Support .env for local dev without overriding CI/prod"
-->

# ADR-004: Support `.env` for Local Dev Without Overriding CI/Prod

- Date: 2026-03-10
- Status: Accepted

## Context

- Developers need a simple and predictable way to configure local environments.
- Go does not load `.env` files automatically.
- CI and production environments rely on OS-level environment variables.
- Configuration precedence must be explicit to prevent accidental overrides.

---

## Decision

Support optional local `.env` loading via `joho/godotenv` (or equivalent):

```go
// Load .env only when present — never fail if absent
_ = godotenv.Load()
```

Treat OS-level environment variables as the **source of truth** in all non-local environments.
`.env` is strictly a local development convenience — it must not be committed to the repository.

### Precedence (highest to lowest)

1. OS / CI environment variables
2. `.env` file (local dev only)
3. Application defaults

This ensures that CI and production always use their own explicitly set values,
and that `.env` can never accidentally override them.

---

## Consequences

### Positive

- Simplifies local onboarding (no shell-level tooling required for most cases)
- Predictable and auditable configuration precedence
- `.env.example` in the repo documents required variables without exposing secrets

### Trade-offs

- Developers must understand that OS vars always win
- `.env` files must be excluded from version control (enforced via `.gitignore`)

## Rejected Alternatives

### `direnv` / Shell injection

Rejected — shell pollution and onboarding complexity. Makes configuration harder to reason
about across machines.

### Committed `.env` files

Rejected — risks leaking secrets and accidental CI/prod overrides.

### No `.env` support at all

Rejected — unnecessarily burdens local setup with manual `export` commands.

## Related ADRs

- ADR-001: SQLite (configuration for DB path)
- ADR-006: Local developer experience
