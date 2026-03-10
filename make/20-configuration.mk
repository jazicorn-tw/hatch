# -----------------------------------------------------------------------------
# 20-config.mk (20s — Configuration)
#
# Responsibility: Decide what should happen (no side effects).
# - feature flags, derived variables, CI/local toggles
#
# Rule: Safe to evaluate (make -pn) without mutating state.
# -----------------------------------------------------------------------------

LOCAL_SETTINGS ?= .config/local-settings.json
