<!--
created_by:   jazicorn-tw
created_date: 2026-03-05
updated_by:   jazicorn-tw
updated_date: 2026-03-09
status:       active
tags:         [devops, deploy]
description:  "Helm (Preparation Only)"
-->
# ⛵ Helm (Preparation Only)

This repo currently deploys as a **single service** (modular monolith) while we prepare for a
future split into multiple services.

Helm is added now for **future Kubernetes readiness**, but Kubernetes is **not** required to develop or deploy today.

---

## ⚠️ Helm chart scope and intent

The Helm chart included in this repository exists for **future readiness only**.

References:

- Deployment strategy: **ADR-009**
- Release mechanics: **ADR-008**

Current state:

- Helm charts are linted in CI (`release.yml` / `helm-lint` job)
- Charts are published to GHCR as OCI artifacts on tag push (`publish.yml` / `helm` job), gated by `PUBLISH_HELM_CHART=true`
- Kubernetes is not required to run or deploy the application

All comments in `helm/` templates reference **ADR-009** when describing
deployment assumptions or future Kubernetes usage.

---

## What Helm is

Helm is a package manager for Kubernetes. A Helm "chart" is a versioned bundle of Kubernetes YAML templates + default configuration.

In this repo, the chart lives at:

```text
helm/app
```

## What we do today

- Keep a minimal chart in the repo
- Run `helm lint` in CI to ensure the chart stays valid
- Publish the chart as an OCI artifact to `ghcr.io/<owner>/charts` on each `vX.Y.Z` release (when `PUBLISH_HELM_CHART=true`)
- Default resource requests and limits are set in `helm/app/values.yaml`:

```yaml
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

Override per environment in your deployment values file.

No cluster is required.

## Local usage

Install Helm:

```bash
brew install helm
```

Lint:

```bash
helm lint helm/app
```

Environment variables are documented in `.env.example` and mirrored in
`helm/app/values.yaml`.

## CI

`release.yml` runs `helm lint helm/app` on every PR and branch push.

`publish.yml` packages and pushes the chart to GHCR on every `vX.Y.Z` tag push (gated by `PUBLISH_HELM_CHART=true`).

## Future plan

When we extract services (monorepo multi-service), we will:

- deploy per-service to Kubernetes using the published charts
