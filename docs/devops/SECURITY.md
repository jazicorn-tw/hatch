<!--
created_by:   jazicorn-tw
created_date: 2026-03-05
updated_by:   jazicorn-tw
updated_date: 2026-03-09
status:       active
tags:         [devops]
description:  "Security Model"
-->
# Security Model

This project uses **explicit, testable Spring Security configuration**.

---

## Public endpoints

- `GET /ping`
- `GET /actuator/health`
- `GET /actuator/info`

---

## Protected endpoints

- Everything else
- Returns **401 Unauthorized**

---

## Implementation

Security is enforced via `SecurityConfig.java`.

No dispatcher-type or implicit exceptions are used.

---

## Future work

- JWT authentication
- Role-based authorization
- Restricted actuator endpoints in production
