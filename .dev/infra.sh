# .dev/infra.sh — Infrastructure tasks: doctor, hooks, exec-bits, run
# Sourced by ./dev — not executed directly.

run_doctor() {
  header "doctor"
  spin "Checking environment..." ./scripts/doctor.sh
  log_done "doctor"
}

run_hooks() {
  header "hooks"
  ./scripts/bootstrap/install-hooks.sh
  log_done "hooks"
}

run_exec_bits() {
  header "exec-bits"
  spin "Checking executable bits..." ./scripts/check/check-executable-bits.sh
  log_done "exec-bits"
}

run_build_run() {
  header "build:run"
  env_check
  run_build
  local args=("${@}")
  if command -v op >/dev/null 2>&1 && [[ -f .env.op ]]; then
    log_info "resolving secrets via 1Password CLI"
    op run --env-file .env.op -- ./hatch "${args[@]}"
  else
    log_warn "op CLI or .env.op not found — falling back to .env"
    set -a; source .env; set +a
    ./hatch "${args[@]}"
  fi
}

run_run() {
  header "run"
  env_check
  if command -v op >/dev/null 2>&1 && [[ -f .env.op ]]; then
    log_info "resolving secrets via 1Password CLI"
    op run --env-file .env.op -- go run ./...
  else
    log_warn "op CLI or .env.op not found — falling back to .env"
    set -a; source .env; set +a
    go run ./...
  fi
}
