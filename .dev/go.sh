# .dev/go.sh — Go tasks: format, test
# Sourced by ./dev — not executed directly.

run_format() {
  header "format"
  if has_go_files; then
    spin "Running gofmt..." go fmt ./...
    log_done "format"
  else
    log_warn "no Go files found, skipping gofmt"
  fi
}

run_test() {
  header "test"
  if has_go_files; then
    spin "Running go test..." go test ./...
    log_done "test"
  else
    log_warn "no Go files found, skipping go test"
  fi
}
