#!/usr/bin/env bash
set -euo pipefail

# -----------------------------------------------------------------------------
# stop-dev.sh
#
# Idempotent local-dev teardown:
# - Optionally stops Docker Compose stack (if compose file exists)
# - Stops Colima (unless KEEP_COLIMA_RUNNING=1)
# -----------------------------------------------------------------------------

_LIB="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/../lib"
# shellcheck source=scripts/lib/shell-utils.sh
source "${_LIB}/shell-utils.sh"
# shellcheck source=scripts/lib/colima-utils.sh
source "${_LIB}/colima-utils.sh"

stop_compose_if_present() {
  local compose_file
  compose_file="$(find_compose_file || true)"

  if [[ -z "$compose_file" ]]; then
    log "no compose file found — skipping compose down"
    return 0
  fi

  if ! have docker; then
    warn "docker CLI not found; cannot run docker compose down"
    return 0
  fi

  log "docker compose down (${compose_file})…"
  docker compose -f "$compose_file" down
  log "docker compose stack is down"
}

stop_colima() {
  if [[ "${KEEP_COLIMA_RUNNING:-0}" == "1" ]]; then
    log "KEEP_COLIMA_RUNNING=1 — skipping colima stop"
    return 0
  fi

  if ! have colima; then
    warn "colima not found; nothing to stop"
    return 0
  fi

  if colima_running; then
    log "stopping colima…"
    colima stop
    log "colima stopped"
  else
    log "colima already stopped"
  fi
}

main() {
  if [[ -n "$_GUM" ]]; then
    $_GUM style --bold --border normal --padding "0 1" "🛑 Local dev stop"
  else
    printf '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n🛑 Local dev stop\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n'
  fi

  stop_compose_if_present
  stop_colima
}

main "$@"
