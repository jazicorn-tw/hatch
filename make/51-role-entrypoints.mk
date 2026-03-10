# -----------------------------------------------------------------------------
# 51-role-entrypoint.mk (Roles entrypoints)
#
# Responsibility: Role-based orchestration targets (e.g., contributor/maintainer).
#
# Placement note:
# - If these targets are primarily "public entrypoints", treat as Interface.
# - If they’re implementation glue that calls other targets, treat as Library.
# -----------------------------------------------------------------------------

# -------------------------------------------------------------------
# WORKFLOWS / ROLE GATES
# -------------------------------------------------------------------
#
# Opinionated, executable entrypoints that run gates.
# These are NOT help commands.
#
# Examples:
#   make contributor
#   make reviewer
#   make maintainer
# -------------------------------------------------------------------

.PHONY: dev-up dev-down dev-status contributor reviewer maintainer

dev-up: ## 🔼 Start local dev prerequisites (env-up)
	@$(MAKE) --no-print-directory env-up

dev-down: ## 🔽 Stop local dev prerequisites (env-down)
	@$(MAKE) --no-print-directory env-down

dev-status: ## 📋 Show local dev env status (env-status)
	@$(MAKE) --no-print-directory env-status

contributor: ## 🧑‍💻 Run contributor gate (verify)
	@$(MAKE) --no-print-directory format verify

reviewer: ## 🧑‍🔍 Run reviewer gate (CI-parity)
	@$(MAKE) --no-print-directory quality

maintainer: ## 🧑‍🔧 Run maintainer gate (heaviest local confidence)
	@$(MAKE) --no-print-directory quality
