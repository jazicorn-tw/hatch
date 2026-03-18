# .dev/go.sh — Go tasks: format, test
# Sourced by ./dev — not executed directly.

run_build() {
  header "build"
  if has_go_files; then
    spin "Building hatch binary..." go build -o ./hatch ./cmd/hatch
    log_done "build — binary written to ./hatch"
  else
    log_warn "no Go files found, skipping build"
  fi
}

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
    while IFS= read -r pkg; do
      spin "  ${pkg##*/}" go test "$pkg" || return 1
    done < <(go list ./...)
    log_done "test"
  else
    log_warn "no Go files found, skipping go test"
  fi
}
