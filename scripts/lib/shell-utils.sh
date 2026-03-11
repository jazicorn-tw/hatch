#!/usr/bin/env bash
# scripts/lib/shell-utils.sh — shared output helpers. Source; do not execute.

# ── Gum resolver ──────────────────────────────────────────────────────────────
# Resolve gum once at source time; scripts use $_GUM for pretty output.
# Falls back to plain printf helpers when gum is unavailable.
_GUM=""
if command -v gum >/dev/null 2>&1; then
  _GUM="gum"
else
  _gopath_gum="$(go env GOPATH 2>/dev/null)/bin/gum"
  [[ -x "$_gopath_gum" ]] && _GUM="$_gopath_gum"
fi

# ── Output helpers ────────────────────────────────────────────────────────────
log() {
  if [[ -n "$_GUM" ]]; then
    $_GUM log --level info "$*"
  else
    printf 'INFO %s\n' "$*"
  fi
}

warn() {
  if [[ -n "$_GUM" ]]; then
    $_GUM log --level warn "$*"
  else
    printf 'WARN %s\n' "$*" >&2
  fi
}

# die [exit_code] message  (exit_code defaults to 1)
die() {
  local _code=1
  if [[ "${1:-}" =~ ^[0-9]+$ ]]; then _code="$1"; shift; fi
  if [[ -n "$_GUM" ]]; then
    $_GUM log --level error "$*" >&2
  else
    printf 'ERROR %s\n' "$*" >&2
  fi
  exit "${_code}"
}

have() { command -v "$1" >/dev/null 2>&1; }

# find_compose_file — prints path to first compose file found, returns 1 if none.
find_compose_file() {
  local f
  for f in docker-compose.yml docker-compose.yaml compose.yml compose.yaml; do
    [[ -f "$f" ]] && { printf '%s' "$f"; return 0; }
  done
  return 1
}
