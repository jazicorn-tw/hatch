# .dev/release.sh — Release task: changelog
# Sourced by ./dev — not executed directly.

run_changelog() {
  header "changelog"
  if [[ ! -f node_modules/.bin/semantic-release ]]; then
    log_error "semantic-release not installed — run: yarn install"
    exit 1
  fi
  $GUM log --level info "dry-run — no release will be created"
  npx semantic-release --dry-run --no-ci
  log_done "changelog"
  return 0
}
