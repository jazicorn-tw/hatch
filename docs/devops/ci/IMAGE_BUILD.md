<!--
created_by:   jazicorn-tw
created_date: 2026-03-05
updated_by:   jazicorn-tw
updated_date: 2026-03-09
status:       active
tags:         [devops, ci]
description:  "️ Docker Build Check (CI – No Push)"
-->
# 🏗️ Docker Build Check (CI – No Push)

The `docker-build` job in the `release.yml` workflow builds the Docker image
**for validation only**. It is designed to catch Dockerfile, dependency, and
Helm issues early — **without publishing anything**.

Workflow file: `.github/workflows/release.yml` (job: `docker-build`)

---

## ✅ When this job runs

The **docker-build** job runs on:

- Pull requests targeting `main`, `staging`, `dev`, or `canary`
- Direct pushes to `main`, `staging`, `dev`, or `canary`
- Manual trigger via **workflow_dispatch**

```yaml
if: ${{ !startsWith(github.ref, 'refs/tags/') }}
```

The job is skipped on tag pushes — those are handled by the `publish` job
once semantic-release creates a release.

This ensures Docker and Helm issues are detected **before** release tags
are created.

---

## 🔐 Permissions

This job is read-only:

```yaml
permissions:
  contents: read
```

It does **not** authenticate to any container registry and **never pushes images**.

---

## 🐳 Docker build (CI only)

### Purpose

The Docker build step validates that:

- `Dockerfile` is valid
- Application dependencies resolve correctly
- The image can be built on Linux (`amd64`)
- Layer caching behaves as expected

### Configuration

```yaml
- name: Build image (CI only, no push)
  uses: docker/build-push-action@v6
  with:
    context: .
    file: ./Dockerfile
    push: false
    platforms: linux/amd64
    tags: {{project-name}}:ci
    labels: |
      org.opencontainers.image.source=${{ github.repository }}
      org.opencontainers.image.revision=${{ github.sha }}
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

Key points:

- `push: false` guarantees **no registry interaction**
- A local-only tag (`{{project-name}}:ci`) is used for validation
- OCI labels link the image back to the repo and commit SHA
- GitHub Actions cache (`type=gha`) speeds up repeated builds

---

## ⎈ Helm lint (CI only)

### Helm lint Purpose

The `helm-lint` job (parallel to `docker-build`) validates Kubernetes manifests
**without deploying anything**.

It checks:

- Chart structure and metadata
- Template rendering
- Common Helm anti-patterns and errors

### Helm lint Configuration

```yaml
- name: Helm lint
  run: helm lint helm/app
```

Helm version is pinned for reproducibility:

```yaml
- uses: azure/setup-helm@v4
  with:
    version: v3.14.4
```

---

## 🧪 What this job does *not* do

This job intentionally does **not**:

- Push Docker images
- Require registry credentials
- Publish Helm charts
- Deploy to Kubernetes

Publishing is handled by the **`publish` job** in the same `release.yml` workflow,
which runs **only on release tags or after semantic-release produces a version**.

---

## 🔁 Relationship to other jobs in `release.yml`

| Job                | Responsibility                                   |
| ------------------ | ------------------------------------------------ |
| **`docker-build`** | CI validation (Docker build), no push            |
| **`helm-lint`**    | CI validation (Helm chart), no deploy            |
| **`release`**      | semantic-release — version bump and tag          |
| **`publish`**      | Build + push Docker image and Helm chart to GHCR |

This separation keeps CI **fast, safe, and predictable**.
