<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [tooling]
description:  "Doctor (Local Environment Sanity)"
-->
# 🩺 Doctor (Local Environment Sanity)

Doctor is a **local-first environment sanity check** designed to catch setup issues
*before* you run `go build`, `go test`, or any other local dev commands.

> **Important**
> Doctor does **not** replace CI.
> CI remains the only authoritative quality gate.

---

## What this is

`./dev doctor` answers one question:

> *“Is my machine correctly set up to run this project?”*

If the answer is **no**, it exits early with **clear, actionable instructions**
(e.g. “install Go” or “run `colima start`”).

This avoids confusing failures later in:

- `go build` / `go test`
- Docker-dependent workflows
- Colima or container startup

---

## How it’s implemented

Doctor is implemented as a standalone script:

```bash
scripts/doctor.sh
```

The naming is intentional:

- **Script:** `scripts/doctor.sh` — technical, explicit, reusable
- **Command:** `./dev doctor` — human-friendly entry point

The script can also be run directly.

---

## What it checks

### Required (hard failures)

These must pass for the project to work locally.

- **Git**
  - `git` on `PATH`
- **Docker**
  - Docker CLI installed
  - Docker daemon reachable
  - Docker socket healthy

If any of these fail, Doctor **exits immediately**.

---

### macOS-specific (conditional)

- If **Colima** is installed but not running:
  - Warns by default (not required for Go dev)
  - Hard fails only if `doctor.requireColimaRunning: true` is set in `local-settings.json`
    or `DOCTOR_REQUIRE_COLIMA_RUNNING=1` env var is set

This supports both Docker Desktop and Colima workflows.

---

### Best-effort / advisory checks

These checks provide **guidance**, not hard failures
(unless strict mode is enabled):

- **Go** — required for `go build`, `go test`, `go vet`
- **Node.js 20+** — required by semantic-release scripts and markdown lint
- Docker provider detection (Desktop, Colima, Rancher, Podman)
- Docker CPU inspection
- Docker memory inspection
- Docker context mismatch detection (macOS)
- Colima resource inspection and **actionable suggestions**

When resources are low, Doctor prints **exact commands** to fix them:

```bash
colima stop
colima start --cpu 6 --memory 8
```

---

## How to run it

```bash
./dev doctor
```

Use Doctor:

- After cloning
- During onboarding
- When something “feels wrong”
- Before long-running checks (`./dev quality`)

---

## Optional configuration (advanced)

Doctor can be tuned **per invocation** via environment variables.

### `DOCTOR_STRICT`

```bash
DOCTOR_STRICT=1 ./dev doctor
```

Treats **warnings as failures**.
Useful when you want a fully clean environment.

---

### `DOCTOR_MIN_DOCKER_MEM_GB`

```bash
DOCTOR_MIN_DOCKER_MEM_GB=6 ./dev doctor
```

Sets the *recommended* Docker memory threshold (GiB).
Doctor will warn (or fail in strict mode) if below this value.

Defaults (without env var) are read from `.config/local-settings.json`:

```json
{ "doctor": { "minDockerMemGb": 4, "minDockerCpus": 2 } }
```

---

### `DOCTOR_MIN_DOCKER_CPUS`

```bash
DOCTOR_MIN_DOCKER_CPUS=4 ./dev doctor
```

Sets the *recommended* Docker CPU count.
Also falls back to `doctor.minDockerCpus` in `.config/local-settings.json`.

---

### `DOCTOR_REQUIRE_COLIMA_RUNNING` (macOS only)

```bash
DOCTOR_REQUIRE_COLIMA_RUNNING=1 ./dev doctor
```

Fails if Colima is installed but not running.
By default, a stopped Colima is a warning only (not required for Go dev).
Useful for teams standardizing on Colima for Docker.

Can also be set persistently in `.config/local-settings.json`:

```json
{ "doctor": { "requireColimaRunning": true } }
```

---

## CI behavior

When `CI=true` is set, Doctor exits immediately by default — it is a local diagnostic tool, not a CI gate.

The `doctor.yml` workflow runs Doctor with `--allow-ci` to produce a JSON snapshot for visibility:

```bash
./scripts/doctor.sh --json --allow-ci > build/doctor/doctor.json
```

This captures environment info as a CI artifact without blocking the build.

---

## Summary

If Doctor passes, your environment is sane.
If it fails, it tells you **exactly what to fix** — early, clearly, and locally.
