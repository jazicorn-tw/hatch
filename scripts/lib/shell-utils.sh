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
  return 0
}

warn() {
  if [[ -n "$_GUM" ]]; then
    $_GUM log --level warn "$*"
  else
    printf 'WARN %s\n' "$*" >&2
  fi
  return 0
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

have() {
  local cmd="$1"
  command -v "$cmd" >/dev/null 2>&1
  return $?
}

# setup_colors — sets ANSI color variables in the caller's scope.
# No-ops cleanly when stdout is not a terminal.
# Variables set: RESET BOLD DIM YELLOW CYAN GREEN RED
setup_colors() {
  if [[ -t 1 ]]; then
    RESET=$'\033[0m'
    BOLD=$'\033[1m'
    DIM=$'\033[2m'
    YELLOW=$'\033[1;33m'
    CYAN=$'\033[1;36m'
    GREEN=$'\033[1;32m'
    RED=$'\033[1;31m'
  else
    RESET="" BOLD="" DIM="" YELLOW="" CYAN="" GREEN="" RED=""
  fi
  return 0
}

# has_interactive_tty — returns 0 if /dev/tty can actually be opened, 1 otherwise.
# Performs a functional open (not just a permission check) so that a "Device not
# configured" /dev/tty is correctly treated as unavailable.
has_interactive_tty() { { true </dev/tty; } 2>/dev/null; return $?; }

# is_env_true — returns 0 if the argument equals "1", 1 otherwise.
# Usage: is_env_true "${MY_VAR:-0}"
is_env_true() { [[ "${1:-0}" == "1" ]]; return $?; }

# ── Styled output helpers (for scripts that don't use gum) ────────────────────
# Call init_output_helpers after sourcing to set _rule/_step/_pass/_fail/_skip/_indent.
init_output_helpers() {
  if [[ -n "$_GUM" ]]; then
    _rule()   { :; return 0; }
    _step()   { $_GUM style --foreground 240 "$*"; return 0; }
    _pass()   { $_GUM log --level info  "$*"; return 0; }
    _fail()   { $_GUM log --level error "$*"; failed=1; fail_count=$(( fail_count + 1 )); return 0; }
    _skip()   { $_GUM log --level warn  "$*"; return 0; }
    _indent() { cat; return 0; }
  else
    _RULE='  ────────────────────────────────────────────────────'
    _rule()   { printf '%s\n' "$_RULE"; return 0; }
    _step()   { printf '\n  %s\n' "$*"; return 0; }
    _pass()   { printf '  ✅  %s\n' "$*"; return 0; }
    _fail()   { printf '\n  ❌  %s\n' "$*"; failed=1; fail_count=$(( fail_count + 1 )); return 0; }
    _skip()   { printf '  ⏭   %s\n' "$*"; return 0; }
    _indent() { sed 's/^/    /'; return 0; }
  fi
}

# find_compose_file — prints path to first compose file found, returns 1 if none.
find_compose_file() {
  local f
  for f in docker-compose.yml docker-compose.yaml compose.yml compose.yaml; do
    [[ -f "$f" ]] && { printf '%s' "$f"; return 0; }
  done
  return 1
}
