<!--
created_by:   jazicorn-tw
created_date: 2026-03-05
updated_by:   jazicorn-tw
updated_date: 2026-03-09
status:       active
tags:         [devops, deploy]
description:  "Deploy to Render (Option A: Single Service)"
-->
# 🚀 Deploy to Render (Option A: Single Service)

This project deploys to Render as a **single Docker-based web service** while we build out the modular monolith.

## Recommended approach (start simple)

Use Render's **Dockerfile deploy**:

- Render builds the container directly from this repo
- No registry wiring required initially

You can later switch to pulling a pinned GHCR image tag (release-aligned) if desired.

## Render setup (high level)

1. Create a **Web Service** in Render
2. Connect your GitHub repo
3. Choose:
   - Runtime: Docker
   - Branch: `main`
4. Set environment variables (below)
5. Add a Health Check (below)
6. Deploy

## Environment variables

At minimum, set:

- `ENV=production`

Security:

- `JWT_SECRET` (or your app’s secret name)

## Health check

Use the `/health` endpoint:

- `/health` — readiness check (database, disk)
- `/ping` — liveness check (process alive, no dependencies)

Render recommends `/health` for the service health check URL.

## Notes

- Keep all runtime config in environment variables (12-factor)
- Prefer pinning images to a release tag for production stability (later enhancement)

## Future enhancement (release-aligned deploy)

Once you publish images to GHCR on semantic-release tags, you can configure Render to deploy from a specific image tag
(e.g., `ghcr.io/your-org/{{project-name}}:1.2.3`) rather than building from the repo.
