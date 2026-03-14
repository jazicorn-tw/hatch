<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [deploy, docker, release]
description:  "How to delete published packages and container image versions from GitHub Container Registry"
-->
# Delete Published Packages

How to delete container image versions and Helm chart packages from GitHub Container Registry (ghcr.io).

---

## GitHub UI

Package settings live on your **profile**, not the repository.

1. Go to `github.com/YOUR_USERNAME` → **Packages** tab
2. Click the package name (e.g. `hatch`)
3. Click a **specific version** in the version list
4. Open **Package settings** in the right sidebar
5. Scroll to the danger zone → **Delete this version**

> **Note:** The delete option only appears on individual version pages, not the package overview.

---

## Via curl (no CLI required)

You need a GitHub Personal Access Token (PAT) with `read:packages` and `delete:packages` scopes.
Create one at: `github.com/settings/tokens`

### Standard package (e.g. `hatch`)

```bash
# 1. List versions and their numeric IDs
curl -s \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/user/packages/container/hatch/versions" \
  | python3 -c "
import sys, json
for v in json.load(sys.stdin):
    print(v['id'], v['metadata']['container']['tags'])
"

# 2. Delete by numeric ID (not the version tag)
curl -X DELETE \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/user/packages/container/hatch/versions/VERSION_ID"
```

### Nested package (e.g. `charts/hatch`)

Packages with a `/` in the name must be URL-encoded as `%2F`.

```bash
# 1. List versions
curl -s \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/user/packages/container/charts%2Fhatch/versions" \
  | python3 -c "
import sys, json
for v in json.load(sys.stdin):
    print(v['id'], v['metadata']['container']['tags'])
"

# 2. Delete by numeric ID
curl -X DELETE \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/user/packages/container/charts%2Fhatch/versions/VERSION_ID"
```

`VERSION_ID` is the numeric `"id"` field from the list response (e.g. `287364521`), **not** the version tag string like `0.2.1`.

---

## Notes

- Deleting a package version from ghcr.io does **not** delete the associated GitHub Release.
- Deleting a GitHub Release does **not** remove the ghcr.io image version.
- Public package versions may be restricted from deletion if they have been downloaded recently — unpublish via API if the UI blocks deletion.

---

## Related

- [`docs/adr/ADR-015-cgo-cross-compilation.md`](../adr/ADR-015-cgo-cross-compilation.md) — Dockerfile multi-platform build strategy
- [`docs/github/UNDO_COMMITS.md`](UNDO_COMMITS.md) — how to undo commits and rewrite remote history
