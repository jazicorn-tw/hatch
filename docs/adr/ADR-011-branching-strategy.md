<!--
created_by:   jazicorn-tw
created_date: 2026-03-10
updated_by:   jazicorn-tw
updated_date: 2026-03-11
status:       active
tags:         [adr, ci, devops]
description:  "ADR-012: Branching Strategy"
-->
# ADR-012: Branching Strategy

- **Status:** Accepted
- **Date:** 2026-03-10
- **Deciders:** Project maintainers
- **Scope:** Branch naming, PR targets, release promotion flow

---

## Context

As the project moves beyond a single-developer bootstrap phase, a clear branching strategy is needed to:

- separate in-progress work from stable code
- provide an integration layer and a pre-release preview layer before production
- keep `main` clean and release-ready at all times
- support parallel feature and fix development

---

## Decision

We adopt a **four-tier branching model**:

```text
feature/<name>  ──┐
                  ├──► staging ──► canary ──► main
fix/<name>      ──┘
```

### Branch roles

| Branch           | Role                                                                | Releases          | Who pushes                    |
| ---------------- | ------------------------------------------------------------------- | ----------------- | ----------------------------- |
| `main`           | Production-ready code. Source of stable releases.                   | `v1.2.3`          | CI only (via canary PR)       |
| `canary`         | Pre-release preview. Canary artifacts published and validated here. | `v1.2.3-canary.1` | PRs from staging              |
| `staging`        | Integration layer. Fast CI feedback before promoting to canary.     | None              | PRs from feature/fix branches |
| `feature/<name>` | New functionality. One concern per branch.                          | None              | Developer                     |
| `fix/<name>`     | Bug fixes and corrections. One concern per branch.                  | None              | Developer                     |

### Branch naming

```bash
git checkout -b feature/<short-description>
git checkout -b fix/<short-description>
```

Rules:

- Branch **from `staging`**, not `main` or `canary`
- Use lowercase kebab-case descriptions
- One concern per branch — no mixing features and fixes

### PR targets

- `feature/*` and `fix/*` → PR targets **`staging`**
- `staging` → PR targets **`canary`** (after integration CI passes)
- `canary` → PR targets **`main`** (after canary artifacts are validated)

### Releases

| Branch   | Release type                 | Example tag       | Docker tag                        |
| -------- | ---------------------------- | ----------------- | --------------------------------- |
| `canary` | Pre-release (canary channel) | `v1.2.3-canary.1` | `:canary`                         |
| `main`   | Stable                       | `v1.2.3`          | `:latest`, `:1.2.3`, `:1.2`, `:1` |

- semantic-release runs on both `canary` and `main` (when `ENABLE_SEMANTIC_RELEASE=true`)
- Merging to `staging` does **not** trigger a release
- See [ADR-008](./ADR-008-semantic-release.md) for release authority

### Canary validation

Before opening a `canary → main` PR:

1. Confirm the `v*-canary.*` tag was created by CI

2. Pull and smoke-test the canary Docker image:

   ```bash
   docker pull ghcr.io/<org>/<repo>:canary
   ```

3. Verify the Helm chart published to `oci://ghcr.io/<org>/charts`

4. Only promote to `main` once the canary image is confirmed working

---

## CI behavior by branch

| Workflow         | `feature/*` / `fix/*` PRs | `staging` push | `canary` push        | `main` push          |
| ---------------- | ------------------------- | -------------- | -------------------- | -------------------- |
| CI               | ✅                        | ✅             | ✅                   | ✅                   |
| Doctor           | ✅                        | ✅             | ✅                   | ✅                   |
| Security         | ✗                         | ✅             | ✅                   | ✅                   |
| Release (build)  | ✅ (PR)                   | ✅             | ✅                   | ✅                   |
| semantic-release | ✗                         | ✗              | ✅ (canary tag)      | ✅ (stable tag)      |
| Publish          | ✗                         | ✗              | ✅ (`:canary` image) | ✅ (`:latest` image) |

---

## Consequences

### Positive

- `main` is always stable and release-ready
- `canary` acts as a pre-production release gate with real artifacts to validate
- `staging` provides fast integration feedback without the overhead of publishing artifacts
- Feature/fix isolation prevents interference between parallel workstreams
- Clear PR targets reduce mistakes

### Trade-offs

- Two PRs required to promote work from staging to production (`staging → canary`, `canary → main`)
- Developers must remember to branch from `staging`, not `canary` or `main`
- Canary validation is a manual step — no automated promotion to `main`

---

## Related Decisions

- **ADR-008** — semantic-release as sole release authority
- **ADR-009** — Deployment strategy
