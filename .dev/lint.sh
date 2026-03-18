# .dev/lint.sh — Lint tasks: lint, lint:docs
# Sourced by ./dev — not executed directly.

run_lint_docs() {
  spin "Running markdownlint..." \
    ./node_modules/.bin/markdownlint-cli2 '**/*.md' '#node_modules'
  log_done "lint:docs"
  return 0
}

run_lint() {
  header "lint"
  run_lint_docs
  if has_go_files; then
    spin "Running go vet..." go vet ./...
    log_done "lint"
  else
    log_warn "no Go files found, skipping go vet"
  fi
  return 0
}
