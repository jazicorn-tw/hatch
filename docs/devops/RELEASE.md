<!--
created_by:   jazicorn-tw
created_date: 2026-03-05
updated_by:   jazicorn-tw
updated_date: 2026-03-09
status:       active
tags:         [devops]
description:  "Releases (semantic-release)"
-->
<!-- markdownlint-disable-file MD036 -->
<!-- markdownlint-disable-file MD060 -->

# 🚀 Releases (semantic-release)

> **TL;DR**
>
> - Releases are created **only from `main`**
> - Releases are **gated** (off by default)
> - Commit **type** controls versioning (`feat` / `fix` / `perf`)
> - `CHANGELOG.md` and GitHub Releases are generated automatically
> - Artifact publishing (Docker / Helm) is optional and separately gated
> - The `release` commit scope is **reserved for automation**

---

## ⚡ TL;DR — How to get a release

1. Merge into `main` with a **releasable commit type**

   ```text
   feat: add resource inventory endpoint
   ```

2. Ensure one of the release gates is enabled:

   ```text
   ENABLE_SEMANTIC_RELEASE=true
   ```

   or run the workflow manually with:

   ```text
   enable_release=true
   ```

3. semantic-release:
   - calculates the next version
   - creates `vX.Y.Z`
   - updates `CHANGELOG.md`
   - publishes a GitHub Release

No releasable commits → **no release** (this is expected).

---

## 🏁 First-time setup (new repository)

Before semantic-release can version correctly, a `v0.0.0` baseline tag must
exist at the initial commit. Without it, the first release defaults to `v1.0.0`
regardless of commit type.

**Do this once, before your first push to `main`:**

```bash
git tag v0.0.0 $(git rev-list --max-parents=0 HEAD)
git push origin v0.0.0
```

Then push your branch normally. The `pre-push` hook will warn you if you forget.

After the baseline is in place, versioning works as expected:

| Commit type        | Next version |
| ------------------ | ------------ |
| `fix:`             | `v0.0.1`     |
| `feat:`            | `v0.1.0`     |
| `BREAKING CHANGE:` | `v1.0.0`     |

---

## 🧠 Deep dive — Release system design

### What happens on release

When a change lands on **`main`** *and releases are enabled*, the workflow performs:

1. Evaluate **release gates**
2. Analyze Conventional Commits
3. Preview the next version (dry run)
4. Create a Git tag (`vX.Y.Z`)
5. Generate GitHub Release notes
6. Update `CHANGELOG.md`
7. Commit the changelog back to `main` as:

   ```text
   chore(release): X.Y.Z [skip ci]
   ```

> ⚠️ The `release` scope is **reserved for automation only**.  
> Humans must never author `*(release): ...` commits.

---

### 🌿 Branch flow

- Work happens on `dev`
- Integrate via PR into `staging`
- Promote via PR from `staging` → `main`
- **Releases are cut only from `main`**

---

## ✍️ Commit messages (this drives versioning)

semantic-release reacts only to **commit types**, not scopes.

### Releasable types

| Type              | Result        |
| ----------------- | ------------- |
| `feat`            | Minor release |
| `fix`             | Patch release |
| `perf`            | Patch release |
| `BREAKING CHANGE` | Major release |

### Non-releasing types

| Type       | Result     |
| ---------- | ---------- |
| `docs`     | No release |
| `chore`    | No release |
| `test`     | No release |
| `ci`       | No release |
| `refactor` | No release |
| `style`    | No release |
| `build`    | No release |

Unknown or missing types are grouped under **🧩 Other** in release notes.

---

### Intentional override for refactors

Refactors **do not release by default**.

To intentionally cut a version (e.g. for rollback safety):

```text
fix: internal refactor + stability
```

or

```text
perf: refactor for performance
```

---

## 📝 Release notes strategy

This repository intentionally produces **two views** of release notes:

### CHANGELOG.md

- Grouped by category
- Includes short commit hashes
- Optimized for maintainers

### GitHub Releases

- Grouped by category
- No commit hashes
- Optimized for consumers

Both are generated from the same commits using different writer options.

---

## 🚦 Release gating

Releases are **disabled by default**.

The release job runs only when **one** is true:

```text
ENABLE_SEMANTIC_RELEASE=true
```

or manual workflow input:

```text
enable_release=true
```

This prevents accidental releases from routine merges.

---

## 📦 Artifact publishing gates

Docker and Helm publishing is handled by **`publish.yml`**, which triggers
independently on tag push (`v*.*.*`). It requires **all** of:

1. A tag matching `v*.*.*` pushed to the canonical repository
2. The corresponding feature flag enabled

```text
PUBLISH_DOCKER_IMAGE=true   # Docker image → ghcr.io/<owner>/<repo>
PUBLISH_HELM_CHART=true     # Helm chart  → ghcr.io/<owner>/charts
```

Forks can run CI safely but **cannot publish artifacts**.

---

## 🧪 Dry runs

Dry runs calculate the next version **without side effects**.

### Local

```bash
make release-dry-run
```

Or directly:

```bash
npm ci
npx semantic-release --dry-run
```

### CI

Every release workflow includes a dry-run preview step for visibility.

---

## 🔐 Required GitHub configuration

### GitHub App

Releases authenticate using a **GitHub App**, not `GITHUB_TOKEN`.

#### Step 1 — Create the App and add secrets

Required secrets:

- `GH_APP_ID`
- `GH_APP_PRIVATE_KEY`

#### Step 2 — Install the App on this repository

> ⚠️ Adding the secrets is **not enough**. The App must also be **installed** on
> this specific repository. Skipping this step produces a 404 error:
> `Could not retrieve installation`.

Go to your GitHub App → **Install App** → select this repository.
If already installed at the org level, confirm this repo is included
(Apps can be scoped to selected repos).

#### Step 3 — Allow the App to bypass branch protection

Because the changelog commit is pushed to `main`, the GitHub App must be allowed
to **bypass the `main` ruleset**:

1. **Settings → Rules → Rulesets → `main`**
2. **Bypass list → Add bypass**
3. Search for your App by name, select it, set role to **Always bypass**
4. Save

> For full setup instructions (creating the App, generating keys, installing on the repo),
> see [`GITHUB_APP_SETUP.md`](./GITHUB_APP_SETUP.md).

---

## 🐤 Canary releases

The `canary` branch is a pre-release staging channel for **features and breaking
changes** before they are promoted to stable. Patches (`fix:`) go directly to `main`
by convention.

### Version format

| Branch   | Commit             | Version              |
| -------- | ------------------ | -------------------- |
| `canary` | `feat:`            | `v0.1.0-canary.1`    |
| `canary` | `BREAKING CHANGE:` | `v1.0.0-canary.1`    |
| `main`   | merge from canary  | `v0.1.0` or `v1.0.0` |

### Docker image

Canary releases are tagged `:canary` (not `:latest`):

```bash
docker pull ghcr.io/<owner>/<repo>:canary
```

Stable releases continue to use `:latest`.

### GitHub Release

Canary releases appear as **pre-releases** in GitHub Releases (marked automatically
by semantic-release).

### Promotion to stable

```text
feature branch → canary (pre-release) → main (stable)
                  v0.1.0-canary.1        v0.1.0
```

Merge `canary` into `main` and push — semantic-release promotes the canary version
to the full stable release.

### Release gate

The same `ENABLE_SEMANTIC_RELEASE` gate applies to both `main` and `canary`.

---

## 🆘 Troubleshooting

### First release creates v1.0.0 instead of v0.0.1 or v0.1.0

**Cause:** semantic-release defaults to `1.0.0` when no git tags exist —
regardless of whether the triggering commit is `fix:` (patch) or `feat:` (minor).
This is hardcoded in semantic-release v22 and cannot be changed via configuration.

**Fix:** A `v0.0.0` tag must exist at the initial commit before the first release runs.
This gives semantic-release a baseline to increment from correctly.

Create it manually (one-time setup per repository):

```bash
git tag v0.0.0 $(git rev-list --max-parents=0 HEAD)
git push origin v0.0.0
```

After that, versioning works as expected:

| Commit type        | Next version |
| ------------------ | ------------ |
| `fix:`             | `v0.0.1`     |
| `feat:`            | `v0.1.0`     |
| `BREAKING CHANGE:` | `v1.0.0`     |

The `pre-push` hook warns you before pushing to `main` or `canary` if no
baseline tag exists, with the exact commands to run.

**If v1.0.0 was already created accidentally**, see
[`UNDO_ACCIDENTAL_RELEASE.md`](./UNDO_ACCIDENTAL_RELEASE.md) for the recovery steps,
which also covers this exact scenario.

---

### Nothing was released

- No releasable commits
- Expected behavior — check the Release Summary

### Artifacts not published

- Non-canonical repo
- Feature flag disabled
- No release version

### Release workflow fails at "Generate GitHub App token"

Symptom: `Could not create installation access token` / `Could not retrieve installation` (404)

Causes (check in order):

1. `GH_APP_ID` or `GH_APP_PRIVATE_KEY` secret not set → add them in repo Settings → Secrets
2. **App not installed on this repository** → go to GitHub App → Install App → select this repo
3. App installed org-wide but this repo excluded → update the App's repo access

### Changelog commit failed

- GitHub App not allowed to bypass branch protection

---

## 🎯 Design principles

- Explicit intent over automation magic
- Versioning decoupled from delivery
- Fork-safe by default
- CI explains *why* something happened (or didn’t)

---

## 🧭 Why didn’t a release happen? (decision tree)

Follow this top-to-bottom — you’ll always land on the answer.

```text
Did the workflow run?
 ├─ No → Releases are gated (ENABLE_SEMANTIC_RELEASE / manual run)
 └─ Yes
     └─ Did semantic-release find releasable commits?
         ├─ No
         │   ├─ Only docs/chore/refactor/ci/test commits → Expected (no release)
         │   └─ Squash commit message not Conventional → Fix commit message
         └─ Yes
             └─ Was this the canonical repository?
                 ├─ No → Release may run, artifacts will not publish
                 └─ Yes
                     └─ Did artifact publishing run?
                         ├─ No → Feature flag disabled (PUBLISH_*)
                         └─ Yes → ✅ Everything worked
```

### Common quick fixes

- **No releasable commits**
  - Use `feat:` / `fix:` / `perf:` in the *squash merge* commit
- **Workflow didn’t run**
  - Set `ENABLE_SEMANTIC_RELEASE=true` or use manual input
- **Artifacts skipped**
  - Confirm canonical repo + feature flag enabled

---

## 🔗 Related release documentation

- **Release contract (policy + rationale)**  
  See `.release.contract.json` — documents *why* the release system behaves the way it does.

- **semantic-release configuration**  
  See `.releaserc.cjs` — executable source of truth.

- **CHANGELOG.md**  
  Maintainer-facing, hash-inclusive release history.

These three files are intentionally kept separate:

- config (`.releaserc.cjs`)
- policy (`.release.contract.json`)
- behavior & usage (`docs/devops/RELEASE.md`)
