# .dev/ci.sh — CI simulation task: test-ci
# Sourced by ./dev — not executed directly.
#
# Usage: run_test_ci <workflow>
#   workflow = "test-ci"  → run all local-safe workflows (ci, doctor, changelog-guard)
#   workflow = "ci"       → run that specific workflow

run_test_ci() {
  local _workflow="$1"
  header "test-ci"

  if ! command -v act >/dev/null 2>&1; then
    log_error "act is required: brew install act"
    exit 1
  fi

  local _tc_job="${ACT_JOB:-}"

  # ── Resolve Docker socket ──────────────────────────────────────────────────
  # act bind-mounts the socket into runner containers. Must be at a path the
  # host daemon can resolve — /var/run/docker.sock works; Colima's socket path
  # inside ~/.colima does not (macOS can't bind-mount VM-internal paths).
  local _std_sock="/var/run/docker.sock"
  local _colima_sock="${HOME}/.colima/default/docker.sock"
  local _tc_docker_sock=""

  if [[ -n "${DOCKER_HOST:-}" ]]; then
    _tc_docker_sock="$DOCKER_HOST"
  elif [[ -S "$_std_sock" ]]; then
    _tc_docker_sock="unix://${_std_sock}"
  elif [[ -S "$_colima_sock" ]]; then
    $GUM log --level warn "/var/run/docker.sock not found — act cannot mount the Colima socket into containers"
    $GUM log --level warn "Fix: sudo ln -sf ${_colima_sock} ${_std_sock}"
    if $GUM confirm "Create symlink now? (requires sudo)"; then
      sudo ln -sf "$_colima_sock" "$_std_sock"
      $GUM log --level info "symlink created: ${_std_sock} → ${_colima_sock}"
      _tc_docker_sock="unix://${_std_sock}"
    else
      $GUM log --level warn "continuing without socket mount — Docker-in-Docker steps will fail"
    fi
  fi
  [[ -n "$_tc_docker_sock" ]] && $GUM log --level info "docker socket: ${_tc_docker_sock}"

  # ── Build base act args ────────────────────────────────────────────────────
  local _tc_base_args=(
    --env ACT=true
    --container-architecture linux/amd64
    --container-options "--privileged"
  )
  [[ -n "$_tc_job" ]] && _tc_base_args+=(-j "$_tc_job")
  [[ -f .env ]]       && _tc_base_args+=(--secret-file .env)
  [[ -f .vars ]]      && _tc_base_args+=(--var-file .vars)

  # ── Run workflows ──────────────────────────────────────────────────────────
  if [[ "$_workflow" == "test-ci" ]]; then
    # No argument — run local-safe workflows only.
    # release.yml and publish.yml require Docker-in-Docker (DinD) which does
    # not work on macOS + Colima. Push a branch to test those on GitHub.
    local _tc_local=(ci doctor changelog-guard)
    $GUM log --level info "local workflows: ${_tc_local[*]}${_tc_job:+  job: ${_tc_job}}"
    for _wf in "${_tc_local[@]}"; do
      $GUM log --level info "── ${_wf}"
      DOCKER_HOST="${_tc_docker_sock:-}" act "${_tc_base_args[@]}" -W ".github/workflows/${_wf}.yml"
    done
  else
    # Specific workflow: ./dev test-ci release → release.yml
    $GUM log --level info "workflow: ${_workflow}${_tc_job:+  job: ${_tc_job}}"
    DOCKER_HOST="${_tc_docker_sock:-}" act "${_tc_base_args[@]}" -W ".github/workflows/${_workflow}.yml"
  fi

  log_done "test-ci"
  return 0
}
