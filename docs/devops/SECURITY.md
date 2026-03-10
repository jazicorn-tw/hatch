<!--
created_by:   jazicorn-tw
created_date: 2026-03-05
updated_by:   jazicorn-tw
updated_date: 2026-03-10
status:       active
tags:         [devops]
description:  "Security Model"
-->
# Security Model

This project uses **explicit Go HTTP middleware** for authentication and authorization.

---

## Public endpoints

- `GET /ping`
- `GET /health`

---

## Protected endpoints

- Everything else
- Returns **401 Unauthorized**

---

## Implementation

Security is enforced via HTTP middleware applied to the router.

No implicit or framework-inferred security rules are used — all rules are explicit and testable.

---

## Future work

- JWT authentication
- Role-based authorization
- Restricted health endpoint details in production
