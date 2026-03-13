#!/usr/bin/env bash
# scripts/lib/doctor-check-utils.sh — shared helpers for doctor check scripts.
# Source; do not execute.

# ANSI colors (safe even if Make disables color upstream)
ORANGE="\033[38;5;208m"
RED="\033[1;31m"
GREEN="\033[1;32m"
GRAY="\033[90m"
RESET="\033[0m"

status="pass"
errors=()
warnings=()

fail() {
  local message="$1"
  status="fail"
  errors+=("$message")
  return 0
}

warn() {
  local message="$1"
  warnings+=("$message")
  return 0
}

# emit_json <check-name>
# Reads $status, $errors[], $warnings[] from caller scope.
emit_json() {
  local _check="${1}"
  jq -n \
    --arg check  "${_check}" \
    --arg status "${status}" \
    --argjson errors   "$(printf '%s\n' "${errors[@]:-}"   | jq -R . | jq -s .)" \
    --argjson warnings "$(printf '%s\n' "${warnings[@]:-}" | jq -R . | jq -s .)" \
    '{ check: $check, status: $status, errors: $errors, warnings: $warnings }'
  return 0
}
