<!--
created_by:   jazicorn-tw
created_date: 2026-03-05
updated_by:   jazicorn-tw
updated_date: 2026-03-09
status:       active
tags:         [devops]
description:  "Actuator & Healthchecks"
-->
# Actuator & Healthchecks

- Health endpoint: `GET /actuator/health`

The Dockerfile includes a HEALTHCHECK that calls `/actuator/health`.
`docker-compose.yml` also defines an app healthcheck.

## Health Endpoints

The `/actuator/health` endpoint is used for container healthchecks.

This replaces `/ping` for production readiness and liveness signaling.

See:

- [ADR-003](../adr/ADR-003-actuator-health.md): Expose health endpoints with Spring Boot Actuator
