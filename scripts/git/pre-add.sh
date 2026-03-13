#!/usr/bin/env bash
set -euo pipefail

# pre-add.sh — lint, format, and exec-bit check before files are staged
#
# Called by the shell-level git() wrapper (not a native Git hook).
# .md files              → markdownlint-cli2 on staged files only (aborts on error)
# scripts/ / .githooks/  → exec-bits check  (chmod +x + stage, auto-fixes in place)
#
# Configuration hierarchy (highest → lowest priority):
#   1. SKIP_PRE_ADD_LINT=1 git add …        env var — skip this invocation
#   2. git config --local hooks.pre-add-lint developer override (not committed)
#      make pre-add-lint-on / make pre-add-lint-off
#   3. .config/local-settings.json           repo-committed defaults
#      .git.preAddLint.enabled  (bool, default true)
#   4. Hard defaults: enabled=true
#
# One-off env overrides:
#   SKIP_PRE_ADD_LINT=1    git add …  # skip everything this invocation
#
# See docs/tooling/PRE_ADD_LINT.md for shell function setup.

REPO_ROOT="$(git rev-parse --show-toplevel)"
# shellcheck source=scripts/lib/shell-utils.sh
source "$REPO_ROOT/scripts/lib/shell-utils.sh"

# ── 1. Env var: skip everything ───────────────────────────────────────────────
if [[ "${SKIP_PRE_ADD_LINT:-0}" == "1" ]]; then
  exit 0
fi

# ── 2. Read local-settings.json defaults ──────────────────────────────────────
ls_enabled="true"
ls_file=""
if [[ -f "$REPO_ROOT/.config/local-settings.json" ]]; then
  ls_file="$REPO_ROOT/.config/local-settings.json"
elif [[ -f "$REPO_ROOT/local-settings.json" ]]; then
  ls_file="$REPO_ROOT/local-settings.json"
fi
if [[ -n "$ls_file" ]] && command -v jq >/dev/null 2>&1; then
  val="$(jq -r '.git.preAddLint.enabled // empty' "$ls_file" 2>/dev/null || true)"
  [[ "$val" == "false" ]] && ls_enabled="false"
fi

# ── 3. git config --local: developer override (beats local-settings.json) ─────
gc_val="$(git config --bool --local hooks.pre-add-lint 2>/dev/null || echo "")"
if [[ "$gc_val" == "false" ]]; then
  exit 0
elif [[ "$gc_val" == "true" ]]; then
  ls_enabled="true"
fi

# ── 4. Apply enabled default ───────────────────────────────────────────────────
if [[ "$ls_enabled" == "false" ]]; then
  exit 0
fi

# ── Collect files from git add arguments ──────────────────────────────────────
args=("$@")
all_mode=false

for arg in "${args[@]:-}"; do
  case "${arg:-}" in
    -A|--all|.) all_mode=true ;;
    -u|--update) all_mode=true ;;
    *) ;;
  esac
done

md_files=()
script_files=()

collect_file() {
  local f="$1"
  [[ -f "$REPO_ROOT/$f" || -f "$f" ]] || return 0
  case "$f" in
    *.md)                          md_files+=("$f") ;;
    scripts/*|.githooks/*|*.sh|*.bash|*.mjs|dev) script_files+=("$f") ;;
    *) ;;
  esac
  return 0
}

if [[ "$all_mode" == "true" ]]; then
  while IFS= read -r f; do
    collect_file "$f"
  done < <(git -C "$REPO_ROOT" ls-files --modified --others --exclude-standard)
else
  # git ls-files (used in all_mode) returns repo-root-relative paths.
  # Single-file args are relative to CWD, which may be a subdirectory.
  # Prepend the git prefix (CWD relative to repo root) so paths stay
  # consistent after the later `cd "$REPO_ROOT"`.
  GIT_PREFIX="$(git rev-parse --show-prefix 2>/dev/null || true)"
  for arg in "${args[@]:-}"; do
    [[ "${arg:-}" == -* ]] && continue
    p="${GIT_PREFIX}${arg}"
    if [[ -d "$REPO_ROOT/$p" || -d "$p" ]]; then
      # Directory argument: enumerate changed files within it (same as all_mode)
      while IFS= read -r f; do
        collect_file "$f"
      done < <(git -C "$REPO_ROOT" ls-files --modified --others --exclude-standard -- "$p" || true)
    else
      collect_file "$p"
    fi
  done
fi

if [[ ${#md_files[@]} -eq 0 && ${#script_files[@]} -eq 0 ]]; then
  exit 0
fi

failed=0
fail_count=0
cd "$REPO_ROOT"

# ── Gum availability (for output only — optional, falls back to printf) ───────
_GUM=""
if command -v gum >/dev/null 2>&1; then
  _GUM="gum"
else
  _gopath_gum="$(go env GOPATH 2>/dev/null)/bin/gum"
  [[ -x "$_gopath_gum" ]] && _GUM="$_gopath_gum"
fi

# ── Output helpers ────────────────────────────────────────────────────────────
if [[ -n "$_GUM" ]]; then
  _rule()   { :; }
  _step()   { $_GUM style --foreground 240 "$*"; }
  _pass()   { $_GUM log --level info  "$*"; }
  _fail()   { $_GUM log --level error "$*"; failed=1; fail_count=$(( fail_count + 1 )); }
  _skip()   { $_GUM log --level warn  "$*"; }
  _indent() { cat; }
else
  _RULE='  ────────────────────────────────────────────────────'
  _rule()   { printf '%s\n' "$_RULE"; }
  _step()   { printf '\n  %s\n' "$*"; }
  _pass()   { printf '  ✅  %s\n' "$*"; }
  _fail()   { printf '\n  ❌  %s\n' "$*"; failed=1; fail_count=$(( fail_count + 1 )); }
  _skip()   { printf '  ⏭   %s\n' "$*"; }
  _indent() { sed 's/^/    /'; }
fi

# ── Header ────────────────────────────────────────────────────────────────────
_label="pre-add"
[[ ${#md_files[@]} -gt 0 ]]     && _label+="  ·  ${#md_files[@]} md"
[[ ${#script_files[@]} -gt 0 ]] && _label+="  ·  ${#script_files[@]} script"

if [[ -n "$_GUM" ]]; then
  $_GUM style --foreground 99 --bold "▶ $_label"
else
  printf '\n'
  _rule
  printf '  %s\n' "$_label"
  _rule
fi

# ── Frontmatter auto-patch ────────────────────────────────────────────────────
# Injects or refreshes HTML-comment frontmatter in staged .md files.
# created_by / created_date are written once (never overwritten).
# updated_by / updated_date are refreshed on every git add.
# status / tags / description are left untouched once written.
#
# Format: HTML comment block so metadata is hidden on GitHub.
#   <!--
#   created_by:   ...
#   ...
#   -->
if [[ ${#md_files[@]} -gt 0 ]]; then
  _fm_author=$(git config user.name 2>/dev/null \
    || git config user.email 2>/dev/null \
    || echo "unknown")
  _fm_date=$(date +%Y-%m-%d)
  _fm_patched=0

  for _f in "${md_files[@]}"; do
    _fp="$REPO_ROOT/$_f"
    [[ -f "$_fp" ]] || _fp="$_f"

    if head -1 "$_fp" | grep -q '^<!--$'; then
      # Has frontmatter — refresh updated_by and updated_date only
      sed -i '' \
        -e "s/^updated_by:.*$/updated_by:   $_fm_author/" \
        -e "s/^updated_date:.*$/updated_date: $_fm_date/" \
        "$_fp"
    else
      # No frontmatter — prepend full block (new files start as draft)
      _tmp=$(mktemp)
      {
        echo '<!--'
        printf 'created_by:   %s\n' "$_fm_author"
        printf 'created_date: %s\n' "$_fm_date"
        printf 'updated_by:   %s\n' "$_fm_author"
        printf 'updated_date: %s\n' "$_fm_date"
        echo 'status:       draft'
        echo 'tags:         []'
        echo 'description:  ""'
        echo '-->'
        cat "$_fp"
      } > "$_tmp" && mv "$_tmp" "$_fp"
    fi

    git add "$_f"
    _fm_patched=$(( _fm_patched + 1 ))
  done

  _pass "frontmatter (${_fm_patched} file(s) patched)"
fi

# ── Tag validation ────────────────────────────────────────────────────────────
# Enforces the canonical tag vocabulary defined in .github/tags.yml.
# Tags from both `both:` and `docs:` sections are valid for docs frontmatter.
# Empty tags: [] is always valid. Hard-fails on unknown tags.
if [[ ${#md_files[@]} -gt 0 ]]; then
  _TAGS_FILE="$REPO_ROOT/.github/tags.yml"
  _ALLOWED_TAGS=()
  if [[ -f "$_TAGS_FILE" ]]; then
    # Parse `both:` and `docs:` sections — `commits:` is scope-only, not valid for frontmatter
    while IFS= read -r _line; do
      _ALLOWED_TAGS+=("$_line")
    done < <(awk '/^  (both|docs):/{p=1;next} p && /^  [[:alpha:]]/{p=0} p && /^    - /{sub(/^    - /,""); print}' "$_TAGS_FILE")
  else
    printf '  ⚠️   validate-tags — .github/tags.yml not found, skipping\n'
  fi
  _tag_errors=()

  for _f in "${md_files[@]}"; do
    [[ ${#_ALLOWED_TAGS[@]} -eq 0 ]] && break   # no tags file — skip validation
    # Skip template files — they contain placeholder tags by design
    [[ "$_f" == *_TEMPLATE.md ]] && continue
    _fp="$REPO_ROOT/$_f"
    [[ -f "$_fp" ]] || _fp="$_f"

    _tags_line=$(grep '^tags:' "$_fp" 2>/dev/null | head -1)
    # Strip "tags:", brackets, split on commas
    _tags_raw=$(printf '%s' "$_tags_line" \
      | sed 's/^tags:[[:space:]]*//' \
      | tr -d '[]' \
      | tr ',' '\n')

    while IFS= read -r _tag; do
      _tag=$(printf '%s' "$_tag" | xargs)   # trim whitespace
      [[ -z "$_tag" ]] && continue
      _found=0
      for _allowed in "${_ALLOWED_TAGS[@]}"; do
        [[ "$_tag" == "$_allowed" ]] && { _found=1; break; }
      done
      [[ $_found -eq 0 ]] && _tag_errors+=("\"$_tag\" in $_f")
    done <<< "$_tags_raw"
  done

  if [[ ${#_ALLOWED_TAGS[@]} -eq 0 ]]; then
    : # skipped — no tags file
  elif [[ ${#_tag_errors[@]} -eq 0 ]]; then
    _pass "validate-tags (${#md_files[@]} file(s))"
  else
    _fail "validate-tags — unknown tag(s):"
    for _e in "${_tag_errors[@]}"; do
      printf '    %s\n' "$_e"
    done
    printf '    Allowed tags:\n'
    printf '      %-12s  %-12s  %-12s  %-12s  %-12s\n' "${_ALLOWED_TAGS[@]}"
  fi
fi

# ── Markdown lint (markdownlint-cli2) ────────────────────────────────────────
# Lints the full glob coverage defined by .markdownlint-cli2.jsonc (matching
# CI). The config file is required — abort if missing.
if [[ ${#md_files[@]} -gt 0 ]]; then
  _step "lint-docs"
  _cli2_cfg="$REPO_ROOT/.markdownlint-cli2.jsonc"
  if [[ ! -f "$_cli2_cfg" ]]; then
    _fail "lint-docs aborted — .markdownlint-cli2.jsonc not found in repo root"
    exit 1
  fi
  if (cd "$REPO_ROOT" && "$REPO_ROOT/node_modules/.bin/markdownlint-cli2") 2>&1 \
      | sed '/^Finding:\|^Linting:/d' | _indent; then
    _pass "lint-docs"
  else
    _fail "lint-docs failed — see errors above, then re-run git add"
  fi
fi

# ── Executable bits ───────────────────────────────────────────────────────────
# Auto-fixes chmod +x and stages the mode change (strict: 2, autoStage: true).
if [[ ${#script_files[@]} -gt 0 ]]; then
  _exec_out=$(./scripts/check/check-executable-bits.sh 2>&1); _exec_rc=$?
  if [[ $_exec_rc -eq 0 ]]; then
    _pass "exec-bits (${#script_files[@]} script file(s))"
  else
    _step "exec-bits (${#script_files[@]} script file(s))"
  fi
  printf '%s\n' "$_exec_out" | _indent
  if [[ $_exec_rc -ne 0 ]]; then
    _fail "exec-bits failed — see errors above, then re-run git add"
  fi
fi

# ── Footer ────────────────────────────────────────────────────────────────────
if [[ -n "$_GUM" ]]; then
  if [[ $failed -eq 0 ]]; then
    $_GUM log --level info "all checks passed ✓"
  elif [[ $fail_count -eq 1 ]]; then
    $_GUM log --level error "1 check failed — see errors above, then re-run git add"
  else
    $_GUM log --level error "${fail_count} checks failed — see errors above, then re-run git add"
  fi
else
  printf '\n'
  _rule
  if [[ $failed -eq 0 ]]; then
    printf '  ✅  all checks passed\n'
  elif [[ $fail_count -eq 1 ]]; then
    printf '  ❌  1 check failed — see errors above, then re-run git add\n'
  else
    printf '  ❌  %d checks failed — see errors above, then re-run git add\n' "$fail_count"
  fi
  _rule
  printf '\n'
fi

exit $failed
