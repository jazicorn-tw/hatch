<!--
created_by:   jazicorn-tw
created_date: 2026-03-05
updated_by:   jazicorn-tw
updated_date: 2026-03-09
status:       active
tags:         [devops]
description:  "Release Gating & Local Simulation"
-->
<!-- markdownlint-disable-file MD036 -->
<!-- markdownlint-disable-file MD060 -->

# 🚦 Release Gating & Local Simulation

This document explains **how release and publish workflows are intentionally gated**
and how to **safely simulate them locally** using `act`.

This repository treats **releases as privileged operations**.
Nothing is released by accident.

---

## 🧠 Design principles

- Releases must be **explicitly enabled**
- Local behavior must **match GitHub Actions**
- Guardrails should fail **closed by default**
- Local overrides must be **obvious and reversible**

---

## 🔐 Release gates (high level)

The release workflow is protected by **two layers**:

1. **Release execution gate**
2. **Artifact publishing gate**

Both must pass for anything to be published.

---

## 1️⃣ Release execution gate

The `release` job only runs when **one** of the following is true:

```yaml
inputs.enable_release == 'true'
OR
vars.ENABLE_SEMANTIC_RELEASE == 'true'
```

### In GitHub Actions

- `ENABLE_SEMANTIC_RELEASE` is defined as a **repository variable**
- Default: `false` (release disabled)

### Locally with `act`

`act` does **not** know about GitHub repo variables.

To simulate them locally, you must provide a `.vars` file.

---

## 2️⃣ Artifact publishing gate

Publishing (Docker image, Helm chart) is handled by **`publish.yml`**, a
separate workflow triggered directly by tag push (`v*.*.*`). It is gated by:

```yaml
github.repository == vars.CANONICAL_REPOSITORY
AND
vars.PUBLISH_DOCKER_IMAGE == 'true'   # (Docker job)
vars.PUBLISH_HELM_CHART  == 'true'   # (Helm job)
```

This ensures:

- Forks can never publish artifacts
- Publishing only happens on a real `vX.Y.Z` tag push

> `publish.yml` is kept separate from `release.yml` to avoid a GitHub Actions
> limitation where a job with `needs: [release]` is skipped when `release` is
> skipped (which it is on tag pushes), even when `always()` is used.

---

## 🧪 Local setup with `act`

### Step 1: Create `.vars` (local only)

```bash
cp .vars.example .vars
```

⚠️ `.vars` **must be gitignored**.

```gitignore
.vars
```

### Step 2: Run the release workflow locally

```bash
act push -W .github/workflows/release.yml
```

If gates pass, the `release` job will execute exactly as in CI.

---

## 🧯 Safety notes

- `.vars` contains **non-secret values only**
- Secrets still come from `.secrets` or the environment
- CI remains the **authoritative execution environment**
- Local release simulation should be used for:
  - validating workflow logic
  - previewing semantic-release behavior
  - debugging gating conditions

---

## 3️⃣ CI job/step feature flags

Individual CI jobs and steps can be disabled via repository variables without
touching workflow files. All flags are **enabled by default** (unset = runs).
Set to `'false'` to skip.

| Variable                 | Skips                                      |
| ------------------------ | ------------------------------------------ |
| `ENABLE_STATIC_ANALYSIS` | Checkstyle/PMD/SpotBugs step in CI Quality |
| `ENABLE_SONAR`           | Sonar cache + analysis steps in CI Quality |
| `ENABLE_MD_LINT`         | Entire markdown-lint job in CI Quality     |
| `ENABLE_DOCTOR_SNAPSHOT` | Entire doctor job in Doctor Snapshot       |

See `docs/environment/ci/CI_FEATURE_FLAGS.md` for full details.

---

## ✅ Expected behavior matrix

| Scenario                             | Release | Publish |
| ------------------------------------ | ------- | ------- |
| No `.vars`, local act                | ❌      | ❌      |
| `.vars` with ENABLE_SEMANTIC_RELEASE | ✅      | ❌      |
| Canonical repo + published version   | ✅      | ✅      |
| Forked repo                          | ✅      | ❌      |
| GitHub repo vars disabled            | ❌      | ❌      |

---

## 🧭 Related docs

- `docs/environment/ci/CI_FEATURE_FLAGS.md`
- `docs/adr/ADR-008-ci-managed-releases.md`
- `.github/workflows/release.yml`
- `.github/workflows/publish.yml`

---

> If a release happens, it was intentional.
