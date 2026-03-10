<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-10
status:       active
tags:         [adr]
description:  "ADR-003: Expose /ping and /health via Go net/http"
-->

# ADR-003: Expose `/ping` and `/health` via Go `net/http`

- Date: 2026-03-10
- Status: Accepted

## Context

Hatch includes an HTTP server component (API + web assets). Container runtimes (Docker, Kubernetes,
Render) require **health signal endpoints** to determine whether a container is alive and ready to
serve traffic.

Two signals are needed:

- **Liveness** — "Is the process running?"
- **Readiness** — "Is the system healthy and ready?"

These serve different audiences and must not be conflated.

---

## Decision

Expose two endpoints via Go's standard `net/http`:

| Endpoint      | Signal    | Dependencies       |
| ------------- | --------- | ------------------ |
| `GET /ping`   | Liveness  | None               |
| `GET /health` | Readiness | SQLite, disk space |

### `/ping`

- Returns `200 OK` with `{"status":"ok"}` immediately
- Makes **no external calls**
- Used by load balancers and Kubernetes liveness probes

### `/health`

- Aggregates infrastructure checks
- Returns `200 OK` if all checks pass
- Returns non-200 if any critical check fails
- Used by monitoring systems and Kubernetes readiness probes

### Implementation

Both handlers are implemented as plain Go `http.HandlerFunc` — no framework required.

```go
func pingHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

The `Dockerfile` `HEALTHCHECK` calls `/health`.

---

## Consequences

### Positive

- No framework dependency (standard library only)
- Explicit, testable behavior
- Clear separation between liveness and readiness
- Compatible with Docker, Kubernetes, and Render health check configuration

### Trade-offs

- Health checks must be kept up to date as new infrastructure is added
- `/health` can be slow or flaky by design — must not be used as a liveness probe

## Related ADRs

- ADR-009: Deployment strategy
- ADR-005: Security model (public vs protected endpoints)
