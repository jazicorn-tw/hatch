<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [devops]
description:  "Healthchecks"
-->
# Healthchecks

- Liveness endpoint: `GET /ping`
- Health endpoint: `GET /health`

The Dockerfile includes a `HEALTHCHECK` that calls `/health`.

## Health Endpoints

| Endpoint  | Purpose          | Used by                        |
| --------- | ---------------- | ------------------------------ |
| `/ping`   | Liveness check   | Load balancers, K8s liveness   |
| `/health` | Readiness check  | Monitoring, K8s readiness      |

## `/ping`

Returns `200 OK` if the process is alive and the HTTP server is responding.
Makes no external calls — always fast, never flaky.

## `/health`

Aggregates infrastructure checks (database, disk).
Returns `200 OK` if all checks pass, non-200 if any fail.

See [PING_VS_HEALTH.md](./PING_VS_HEALTH.md) for full design rationale.
