# .dev/helpers.sh — GUM setup, logging, and shared helpers
# Sourced by ./dev — not executed directly.

# ── Gum availability ──────────────────────────────────────────────────────────

GUM="gum"
if ! command -v gum >/dev/null 2>&1; then
  _gopath_gum="$(go env GOPATH 2>/dev/null)/bin/gum"
  if [[ -x "$_gopath_gum" ]]; then
    GUM="$_gopath_gum"
  else
    echo "❌ gum is required but not installed."
    echo "   go install github.com/charmbracelet/gum@latest"
    echo "   Then add \$(go env GOPATH)/bin to your PATH:"
    echo "   export PATH=\"\$PATH:\$(go env GOPATH)/bin\""
    exit 1
  fi
fi

# ── Logging ───────────────────────────────────────────────────────────────────

log_info()  { $GUM log --level info  "$*"; return 0; }
log_warn()  { $GUM log --level warn  "$*"; return 0; }
log_error() { $GUM log --level error "$*"; return 0; }
log_done()  { $GUM log --level info  "$* ✓"; return 0; }

# ── Spinner ───────────────────────────────────────────────────────────────────
# TTY: run command in background, spin on its PID, print output only on failure.
# Non-TTY (CI / git hooks): run directly so output flows through.

spin() {
  local title="$1"; shift
  if [[ -t 1 ]]; then
    local _out _rc=0
    _out=$(mktemp)
    "$@" >"$_out" 2>&1 &
    local _pid=$!
    $GUM spin --spinner dot --title "$title" -- \
      bash -c "while kill -0 ${_pid} 2>/dev/null; do sleep 0.05; done" || true
    wait "$_pid" || _rc=$?
    if [[ $_rc -ne 0 ]]; then
      cat "$_out" >&2
    fi
    rm -f "$_out"
    return $_rc
  else
    "$@"
  fi
}

# ── Utilities ─────────────────────────────────────────────────────────────────

has_go_files() {
  find . -name '*.go' -not -path './vendor/*' 2>/dev/null | head -1 | grep -q .
  return $?
}

header() {
  $GUM style \
    --foreground 99 \
    --bold \
    "▶ $*"
  return 0
}

env_check() {
  if [[ ! -f .env ]]; then
    log_error ".env not found — run: ./dev env init"
    exit 1
  fi
  return 0
}
