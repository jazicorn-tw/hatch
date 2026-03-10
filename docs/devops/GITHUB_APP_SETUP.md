<!--
created_by:   jazicorn-tw
created_date: 2026-03-07
updated_by:   jazicorn-tw
updated_date: 2026-03-09
status:       active
tags:         [devops]
description:  "GitHub App Setup (Branch Bypass for CI)"
-->
# GitHub App Setup (Branch Bypass for CI)

The release workflow pushes a changelog commit directly to `main`. Because `main` has branch
protection rules, a regular `GITHUB_TOKEN` cannot push to it. A **GitHub App** is used instead —
it authenticates as a trusted bot identity that can be added to the ruleset bypass list.

This is a **one-time setup per repository**.

---

## Overview

| Step | What you do                                       |
| ---- | ------------------------------------------------- |
| 1    | Create a GitHub App (or reuse an existing one)    |
| 2    | Generate and store a private key as a repo secret |
| 3    | Install the App on this repository                |
| 4    | Add the App to the `main` ruleset bypass list     |

---

## Step 1 — Create the GitHub App

1. Go to **GitHub → Settings → Developer settings → GitHub Apps → New GitHub App**
   (or your org: **Org Settings → Developer settings → GitHub Apps**)
2. Fill in:
   - **Name:** something like `your-project-release-bot`
   - **Homepage URL:** your repo URL (required but not used)
   - **Webhook:** uncheck "Active" — this App does not need webhooks
3. Under **Permissions → Repository permissions**, set:
   - **Contents:** Read & write (to push the changelog commit and tags)
   - **Workflows:** Read & write (required — tags that point to commits containing
     workflow files will be rejected without this permission)
   - **Pull requests:** Read-only (optional, for PR status)
4. Under **Where can this GitHub App be installed**, select:
   - **Only on this account** (unless you want org-wide reuse)
5. Click **Create GitHub App**
6. Note the **App ID** shown on the App's settings page — you will need it

---

## Step 2 — Generate a private key and add secrets

1. On the App's settings page, scroll to **Private keys → Generate a private key**
2. A `.pem` file downloads automatically — keep it safe
3. In your repository, go to **Settings → Secrets and variables → Actions**
4. Add two secrets:

| Secret name          | Value                                |
| -------------------- | ------------------------------------ |
| `GH_APP_ID`          | The App ID from Step 1               |
| `GH_APP_PRIVATE_KEY` | The full contents of the `.pem` file |

The workflow passes `GH_APP_PRIVATE_KEY` through `fromJSON(format(...))` to
handle the multiline PEM value — paste the raw `.pem` file contents as-is,
including the `-----BEGIN RSA PRIVATE KEY-----` header and footer.

### Local `act` use: generate `.tmp_key.b64`

`act` does not support multiline secret values in `.secrets`. Base64-encode the
key into a single line before adding it:

```bash
# Run once — save alongside your .pem file, outside the repo
base64 github-app.pem | tr -d '\n' > .tmp_key.b64

# Append the encoded key to your .secrets file
echo "GH_APP_PRIVATE_KEY=$(cat .tmp_key.b64)" >> .secrets
```

> ⚠️ `.tmp_key.b64` is still a secret — base64 is encoding, not encryption.
> Delete it after use and never commit it.

---

## Step 3 — Install the App on this repository

> Adding secrets alone is not enough. The App must be **installed** on the repository
> or CI will fail with: `Could not retrieve installation`.

1. Go to your GitHub App → **Install App**
2. Click **Install** next to your account or org
3. Choose **Only select repositories** and select this repo
4. Click **Install**

If the App is already installed org-wide, verify this repo is included:
GitHub App → **Install App** → configure → confirm the repo is in the selected list.

---

## Step 4 — Add the App to the `main` ruleset bypass list

This allows the App to push the changelog commit directly to `main` without triggering
branch protection rules.

1. In your repository, go to **Settings → Rules → Rulesets**
2. Click the **`main`** ruleset
3. Scroll to **Bypass list → Add bypass**
4. Search for your App by name (e.g. `your-project-release-bot`)
5. Select it and set the role to **Always bypass**
6. Save the ruleset

---

## Verifying the setup

After setup, trigger the release workflow manually (or merge a `feat:` commit to `main`
with `ENABLE_SEMANTIC_RELEASE=true`). The "Generate GitHub App token" step should
succeed and the changelog commit should land on `main` without errors.

**Common failure messages and causes:**

| Error                                        | Cause                                                           |
| -------------------------------------------- | --------------------------------------------------------------- |
| `Could not create installation access token` | `GH_APP_ID` or `GH_APP_PRIVATE_KEY` secret missing or malformed |
| `Could not retrieve installation` (404)      | App not installed on this repo (Step 3 skipped)                 |
| `Changelog commit failed`                    | App not in the ruleset bypass list (Step 4 skipped)             |

---

## Reusing the App across projects

If you have multiple projects using this template, you can reuse the same GitHub App —
just repeat Steps 3 and 4 for each repository. Each repo needs its own
`GH_APP_ID` and `GH_APP_PRIVATE_KEY` secrets, but they can point to the same App.
