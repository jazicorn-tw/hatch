# .dev/env.sh — Environment tasks: env:up, env:down, env:status, env:init
# Sourced by ./dev — not executed directly.

run_env_up() {
  header "env up"
  # Detect a stale Colima state that would cause a disk-lock failure on start.
  # This happens when a previous VM crashed without a clean shutdown:
  #   1. Orphaned colima/limactl processes still holding the disk
  #   2. A stale Lima in_use_by symlink that was never cleaned up
  local _stale_pids _disk_lock _stale
  _stale_pids=$(pgrep -d ' ' -x 'colima' 2>/dev/null || true)
  _stale_pids+=" $(pgrep -d ' ' -x 'limactl' 2>/dev/null || true)"
  _stale_pids="${_stale_pids// /}"
  _disk_lock="${HOME}/.colima/_lima/_disks/colima/in_use_by"
  _stale=false
  [[ -n "$_stale_pids" ]] && _stale=true
  [[ -L "$_disk_lock" ]]  && _stale=true
  if [[ "$_stale" == true ]] && ! colima status >/dev/null 2>&1; then
    [[ -n "$_stale_pids" ]] && \
      $GUM log --level warn "stale processes: ${_stale_pids// /,}"
    [[ -L "$_disk_lock" ]] && \
      $GUM log --level warn "stale disk lock: ${_disk_lock}"
    if $GUM confirm "Clean up stale Colima state and continue?"; then
      [[ -n "$_stale_pids" ]] && { kill ${_stale_pids} 2>/dev/null || true; sleep 1; }
      [[ -L "$_disk_lock" ]] && rm -f "$_disk_lock"
      $GUM log --level info "stale state cleared"
    else
      $GUM log --level error "aborting — stale state left in place"
      exit 1
    fi
  fi
  ./scripts/dev/start-dev.sh
}

run_env_down() {
  header "env down"
  ./scripts/dev/stop-dev.sh
}

run_env_status() {
  header "env status"
  echo "docker context: $(docker context show 2>/dev/null || echo 'n/a')"
  colima status 2>/dev/null || true
  docker ps 2>/dev/null | head -n 15 || true
}

run_env_init() {
  header "env init"
  ./scripts/bootstrap/init-env.sh
}
