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
  return 0
}

run_format() {
  header "format"
  if has_go_files; then
    spin "Running gofmt..." go fmt ./...
    log_done "format"
  else
    log_warn "no Go files found, skipping gofmt"
  fi
  return 0
}

run_test() {
  header "test"
  if has_go_files; then
    # Pre-compile all test binaries so per-package runs are fast and silent
    # compilation doesn't make the loop appear to hang.
    spin "Compiling test binaries..." go test -run '^$' ./...
    while IFS= read -r pkg; do
      short="${pkg##*/}"
      tmp=$(mktemp)
      if go test "$pkg" >"$tmp" 2>&1; then
        if grep -q '\[no test files\]' "$tmp"; then
          $GUM style --foreground 240 "  · ${short}"
        else
          $GUM style --foreground 2 "  ✓ ${short}"
        fi
      else
        cat "$tmp"
        $GUM style --foreground 1 "  ✗ ${short}"
        rm -f "$tmp"
        return 1
      fi
      rm -f "$tmp"
    done < <(go list ./...)
    log_done "test"
  else
    log_warn "no Go files found, skipping go test"
  fi
  return 0
}
