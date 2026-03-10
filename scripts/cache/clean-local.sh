#!/usr/bin/env bash
set -euo pipefail

# -----------------------------------------------------------------------------
# scripts/cache/clean-local.sh
#
# Responsibility: "one button" local hygiene runner.
#
# It coordinates:
# - docker cache hygiene (scripts/cache/cache-docker-.sh)
#
# NOTE:
# - Colima reset is intentionally NOT part of clean-local.
#   Use `make clean-colima` (or scripts/cache/clean-colima.sh reset) explicitly.
#
# Commands:
#   clean (default) - run docker prune (if enabled)
#   info            - print quick status summaries (docker + colima status)
# -----------------------------------------------------------------------------

cmd="${1:-clean}"

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/../.." && pwd)"

_LIB="${repo_root}/scripts/lib"
# shellcheck source=scripts/lib/shell-utils.sh
source "${_LIB}/shell-utils.sh"

docker_script="${repo_root}/scripts/cache/cache-docker.sh"
colima_script="${repo_root}/scripts/cache/clean-colima.sh"

require_file() {
  [[ -f "$1" ]] || die "Missing script: $1"
}

print_config() {
  echo "🧾 Configuration"
  echo "  docker:"
  echo "    CLEAN_DOCKER_MODE=${CLEAN_DOCKER_MODE:-false}"
  echo "    CLEAN_DOCKER_VOLUMES=${CLEAN_DOCKER_VOLUMES:-false}"
  echo "    CLEAN_DOCKER_VERBOSE=${CLEAN_DOCKER_VERBOSE:-false}"
  echo "    CLEAN_DOCKER_AUTO_MIN_FREE_GB=${CLEAN_DOCKER_AUTO_MIN_FREE_GB:-10}"
  echo "    CLEAN_DOCKER_AUTO_MIN_FREE_INODES=${CLEAN_DOCKER_AUTO_MIN_FREE_INODES:-5000}"
  echo "    CLEAN_DOCKER_COLIMA_PROFILE=${CLEAN_DOCKER_COLIMA_PROFILE:-default}"
  echo ""
}

info() {
  echo "🧼 clean-local (info)"
  echo ""

  print_config

  if [[ -f "${docker_script}" ]]; then
    "${docker_script}" info || true
  else
    echo "ℹ️ docker cache script not found (${docker_script})"
  fi

  echo ""
  # Info-only: show Colima status if script exists (not a pass-through reset).
  if [[ -f "${colima_script}" ]]; then
    "${colima_script}" info || true
  else
    echo "ℹ️ colima script not found (${colima_script})"
  fi
}

clean() {
  echo "🧼 clean-local"
  echo ""

  print_config

  require_file "${docker_script}"

  echo "▶ docker hygiene"
  "${docker_script}" prune
}

usage() {
  cat <<'EOF'
Usage:
  scripts/cache/clean-local.sh [command]

Commands:
  info    Print quick status summaries (docker + colima status)
  clean   Run docker hygiene (default)

Notes:
- Colima reset is intentionally excluded. Run it explicitly:
    make clean-colima CLEAN_COLIMA_RESET=true
EOF
}

main() {
  case "${cmd}" in
    info) info ;;
    clean) clean ;;
    -h|--help|help) usage ;;
    *) die "Unknown command: ${cmd} (use info|clean)" ;;
  esac
}

main "$@"
