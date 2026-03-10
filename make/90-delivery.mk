# -----------------------------------------------------------------------------
# 90-delivery.mk (90s — Delivery)
#
# Responsibility: Release tooling.
#
# Rule: High consequence. Require explicit intent and strong guards.
# -----------------------------------------------------------------------------

# -------------------------------------------------------------------
# Release
# -------------------------------------------------------------------

.PHONY: release-dry-run

release-dry-run: ## 🔍 Preview next semantic-release version (dry-run, no publish)
	$(call step,🔍 semantic-release dry-run)
	@yarn dlx semantic-release --dry-run
