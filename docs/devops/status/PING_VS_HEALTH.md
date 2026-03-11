<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [devops]
description:  "`/ping` vs `/health` — What's the Difference?"
-->
<!-- markdownlint-disable-file MD024 -->
# `/ping` vs `/health` — What's the Difference?

This document explains **why `/ping` and `/health` both exist**, what problem each one solves,
and how they are wired in this Go application.

---

## High-level overview

| Endpoint  | Purpose                 | Dependencies              | Intended audience              |
| --------- | ----------------------- | ------------------------- | ------------------------------ |
| `/ping`   | "Is the process alive?" | **None**                  | Load balancers, CI smoke tests |
| `/health` | "Is the app healthy?"   | **Many** (DB, disk, etc.) | Ops, monitoring, SRE           |

Think of it as:

> **`/ping` = liveness**
> **`/health` = readiness + health**

They solve **different problems** and are **not interchangeable**.

---

## `/ping` — Application liveness

### What `/ping` does

- Confirms the **application process is running**
- Confirms **HTTP routing works**
- Makes **no external calls**
- Never touches DB, cache, or APIs

### Example

```http
GET /ping
```

```json
{
  "status": "ok",
  "service": "hatch-api"
}
```

### When `/ping` should return `200`

- The process is running
- The HTTP server started successfully

### When `/ping` should fail

- The app crashed
- The app failed to start

### Design rules

- Must be **fast**
- Must be **reliable**
- Must **never depend on infrastructure**
- Must **never flake**

### Typical usage

- Load balancer health checks
- CI smoke tests
- Kubernetes **liveness probes**

### Go implementation

```go
func pingHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "ok",
        "service": "hatch-api",
    })
}
```

**Rules enforced here:**

- No DB calls
- No injected dependencies
- No external calls

---

## `/health` — Readiness & health

### What `/health` does

- Aggregates checks for:
  - Database connectivity (SQLite)
  - Disk space
  - Custom readiness conditions

### Example

```http
GET /health
```

```json
{
  "status": "ok",
  "checks": {
    "db": "ok",
    "disk": "ok"
  }
}
```

### When `/health` should return non-200

- Database is unreachable
- Disk is full
- A critical dependency is unavailable

### Design rules

- Infrastructure-aware
- Can be slow
- Can be flaky (by design)
- Intended for monitoring systems

### Typical usage

- Monitoring & alerting
- Readiness checks
- Ops dashboards
- Kubernetes **readiness probes**

---

## Key architectural difference

### `/ping`

- Implemented by **your code**
- Simple handler
- Zero dependencies
- Always safe to call

### `/health`

- Implemented by **your code**
- Checks real infrastructure
- Infrastructure-dependent
- Reflects system readiness

---

## Why you should use BOTH

| Scenario                      | Endpoint  |
| ----------------------------- | --------- |
| App is running?               | `/ping`   |
| App ready to receive traffic? | `/health` |
| Load balancer check           | `/ping`   |
| CI smoke test                 | `/ping`   |
| Monitoring & alerting         | `/health` |
| Kubernetes liveness probe     | `/ping`   |
| Kubernetes readiness probe    | `/health` |

Using only one leads to **false positives** or **false negatives**.

---

## What NOT to do

❌ Don't use `/health` as a liveness probe
❌ Don't make `/ping` check the database
❌ Don't expose `/health` publicly without auth in production
❌ Don't return `500` from `/ping` because DB is down

---

## TL;DR

| Question                                | Answer    |
| --------------------------------------- | --------- |
| Do I need both?                         | **Yes**   |
| Is `/ping` redundant?                   | **No**    |
| Should `/ping` touch DB?                | **Never** |
| Should `/health` touch DB?              | **Yes**   |
| Which endpoint can fail during outages? | `/health` |

---

## Final rule of thumb

> If it's about **"is the process alive?" → `/ping`**
> If it's about **"can the system safely operate?" → `/health`**
