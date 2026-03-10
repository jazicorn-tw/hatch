<!--
created_by:   jazicorn-tw
created_date: 2026-03-05
updated_by:   jazicorn-tw
updated_date: 2026-03-09
status:       active
tags:         [devops, ci]
description:  "Publish Docker Image (Release Tags Only)"
-->
# 🐳 Publish Docker Image (Release Tags Only)

This repository publishes a Docker image to **GitHub Container Registry (GHCR)** only when
a **semantic-release tag** is pushed (example: `v1.2.3`).

Workflow file: `.github/workflows/publish.yml` (job: `docker`)

---

## ✅ When this job runs

The **docker** job triggers when a Git tag matching `v*.*.*` is pushed:

```yaml
on:
  push:
    tags: [ 'v*.*.*' ]
```

```yaml
if: >-
  github.repository == vars.CANONICAL_REPOSITORY &&
  vars.PUBLISH_DOCKER_IMAGE == 'true'
```

That means:

- A normal push to `main` **will not** publish an image unless semantic-release cuts a tag.
- A release tag created by **semantic-release** (or manually) **will** attempt to publish.

---

## 🔐 Permissions (minimal)

This job uses minimal GitHub Actions permissions:

- `contents: read` — to clone the repo
- `packages: write` — to push the image to GHCR

```yaml
permissions:
  contents: read
  packages: write
```

Authentication uses the built-in `GITHUB_TOKEN` (no PAT required).

---

## 🧠 Publish toggle: fork-safe + default-false

Before building/pushing, the job evaluates **gate conditions** that decide whether
publishing is allowed.

### Required repository variables

This job relies on GitHub **Repository variables**
(Settings → Secrets and variables → Actions → Variables):

| Variable               | Example value     | Purpose                                                               |
| ---------------------- | ----------------- | --------------------------------------------------------------------- |
| `PUBLISH_DOCKER_IMAGE` | `true` or `false` | Global "on/off" switch (defaults to `false` when missing).            |
| `PUBLISH_HELM_CHART`   | `true` or `false` | Controls Helm OCI push independently.                                 |
| `CANONICAL_REPOSITORY` | `owner/repo`      | Prevents publishing from non-canonical repos (hardens against forks). |

### Gate rules

Publishing is allowed **only if**:

1. `CANONICAL_REPOSITORY` is set **and** matches the current `${{ github.repository }}`, **and**
2. `PUBLISH_DOCKER_IMAGE == "true"` (for Docker) or `PUBLISH_HELM_CHART == "true"` (for Helm)

---

## 🏗️ What gets published

### Docker image

The image is published to GHCR under:

- `ghcr.io/<owner>/<repo>`

### Tags

Tags are derived from the Git tag (SemVer):

**Stable release** (`v1.2.3`):

- `1.2.3`, `1.2`, `1`, `latest`

**Canary release** (`v1.2.3-canary.1`):

- `1.2.3-canary.1`, `canary`

Configured via `docker/metadata-action`:

```yaml
tags:         []
  type=semver,pattern={{version}}
  type=semver,pattern={{major}}.{{minor}}
  type=semver,pattern={{major}}
  type=raw,value=latest,enable=${{ !contains(github.ref, 'canary') }}
  type=raw,value=canary,enable=${{ contains(github.ref, 'canary') }}
```

### Helm chart

When `PUBLISH_HELM_CHART == "true"`, the Helm chart is pushed as an OCI artifact to GHCR.

---

## 🧰 Build details

The publish job:

1. Checks out the repository with full history + tags (`fetch-depth: 0`)
2. Sets up Docker Buildx
3. Logs into GHCR using `GITHUB_TOKEN`
4. Extracts tags/labels from the Git tag via `docker/metadata-action`
5. Builds and pushes `linux/amd64` and `linux/arm64` (Apple Silicon native)

Build-push configuration:

```yaml
with:
  context: .
  file: ./Dockerfile
  push: true
  platforms: linux/amd64,linux/arm64
  tags: ${{ steps.meta.outputs.tags }}
  labels: ${{ steps.meta.outputs.labels }}
  cache-from: type=gha
  cache-to: type=gha,mode=max
```

---

## 🧪 Quick verification (after a release)

After semantic-release publishes a tag (e.g., `v1.2.3`), confirm:

1. GitHub Actions shows a successful run for **Publish** (the `docker` job)
2. GHCR has the new package + tags

You can also pull locally:

```bash
docker pull ghcr.io/<owner>/<repo>:1.2.3
docker pull ghcr.io/<owner>/<repo>:latest
```

---

## 🔧 Troubleshooting

### Publish job didn't run

Most common causes:

- `PUBLISH_DOCKER_IMAGE` is missing or not `"true"`
- `CANONICAL_REPOSITORY` is missing or doesn't match `${{ github.repository }}`
- The tag didn't match `v*.*.*`

See [`docs/faq/ci/WHY_NO_RELEASE.md`](../../faq/ci/WHY_NO_RELEASE.md) for the full
release diagnosis guide.

### Login to GHCR fails

Confirm:

- Workflow permissions include `packages: write`
- Package visibility/permissions in GHCR are correct for your org/user

---

## 🔒 Notes on safety

The **canonical repo check** is deliberate hardening. It prevents accidental or
malicious publishing from a fork where variables might differ.

---

## 🔁 Relationship to other workflows

| Workflow / Job                     | Responsibility                          |
| ---------------------------------- | --------------------------------------- |
| `release.yml` / **`docker-build`** | CI validation (Docker build), no push   |
| `release.yml` / **`helm-lint`**    | CI validation (Helm chart), no deploy   |
| `release.yml` / **`release`**      | semantic-release — version bump and tag |
| `publish.yml` / **`docker`**       | Build + push Docker image to GHCR       |
| `publish.yml` / **`helm`**         | Package + push Helm chart to GHCR       |
