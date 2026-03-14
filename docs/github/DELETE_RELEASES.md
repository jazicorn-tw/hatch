<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [release, deploy, devops]
description:  "How to delete GitHub Releases and associated tags via UI and curl"
-->
# Delete GitHub Releases

How to delete a GitHub Release and its associated git tag. Deleting a release does **not** automatically delete the tag, and vice versa — both steps are usually needed.

---

## GitHub UI

1. Go to `github.com/YOUR_USERNAME/YOUR_REPO/releases`
2. Find the release you want to delete
3. Click the **pencil icon** (edit) on that release
4. Scroll to the bottom → **Delete this release**

> This removes the release but leaves the git tag intact. Delete the tag separately (see below).

---

## Delete the Associated Tag

```bash
# Delete locally
git tag -d v1.2.3

# Delete on remote
git push origin --delete v1.2.3
```

Or via the GitHub UI:

1. Go to `github.com/YOUR_USERNAME/YOUR_REPO/tags`
2. Click the tag name
3. Click the **trash icon** to delete

---

## Via curl (no CLI required)

You need a GitHub Personal Access Token (PAT) with `repo` scope.
Create one at: `github.com/settings/tokens`

### List releases

```bash
curl -s \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/YOUR_USERNAME/YOUR_REPO/releases" \
  | python3 -c "
import sys, json
for r in json.load(sys.stdin):
    print(r['id'], r['tag_name'], r['name'])
"
```

### Delete a release by ID

```bash
curl -X DELETE \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/YOUR_USERNAME/YOUR_REPO/releases/RELEASE_ID"
```

`RELEASE_ID` is the numeric `"id"` field from the list response, **not** the tag string like `v0.2.1`.

### Delete the tag via API

```bash
curl -X DELETE \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/YOUR_USERNAME/YOUR_REPO/git/refs/tags/v1.2.3"
```

---

## Notes

- Deleting a release does **not** delete the ghcr.io container image for that version.
- Deleting a tag does **not** delete the release — GitHub will show it as a "tag release" without a title.
- If semantic-release created the release, re-running the release workflow may recreate it. Disable the workflow or skip the tag before re-running.

---

## Related

- [`docs/github/DELETE_PACKAGES.md`](DELETE_PACKAGES.md) — how to delete ghcr.io container image versions
- [`docs/github/UNDO_COMMITS.md`](UNDO_COMMITS.md) — how to undo commits and rewrite remote history
- [`docs/commit/CONFIRM_RELEASE_COMMIT.md`](../commit/CONFIRM_RELEASE_COMMIT.md) — confirming and triggering a release commit
