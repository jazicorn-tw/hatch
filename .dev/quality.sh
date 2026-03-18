# .dev/quality.sh — Quality tasks: verify, quality, pre-commit, bootstrap
# Sourced by ./dev — not executed directly.

run_verify() {
  header "verify"
  run_doctor
  run_lint
  run_test
  log_done "verify"
  return 0
}

run_quality() {
  header "quality"
  run_doctor
  run_format
  run_lint
  run_test
  log_done "quality gate passed"
  return 0
}

run_pre_commit() {
  header "pre-commit"
  if [[ "${AUTO_FORMAT:-1}" != "0" ]]; then
    run_format
  else
    log_warn "AUTO_FORMAT=0 set, skipping format"
  fi
  run_lint
  if [[ "${SKIP_TESTS:-0}" != "1" ]]; then
    run_test
  else
    log_warn "SKIP_TESTS=1 set, skipping tests"
  fi
  log_done "pre-commit"
  return 0
}

run_bootstrap() {
  header "bootstrap"
  run_hooks
  run_doctor
  run_quality
  log_done "bootstrap"
  return 0
}
