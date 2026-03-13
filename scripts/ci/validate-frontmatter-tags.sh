#!/usr/bin/env bash
# scripts/ci/validate-frontmatter-tags.sh
#
# Validates that all docs/**/*.md frontmatter tags are in the canonical
# vocabulary defined by .github/tags.yml (both: + docs: sections).
#
# Usage:
#   bash scripts/ci/validate-frontmatter-tags.sh
#
# Exit codes:
#   0 — all tags valid
#   1 — one or more unknown tags found

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TAGS_FILE="$REPO_ROOT/.github/tags.yml"
DOCS_DIR="$REPO_ROOT/docs"

# ── Load allowed tags ──────────────────────────────────────────────────────────
if [[ ! -f "$TAGS_FILE" ]]; then
  printf 'ERROR: .github/tags.yml not found at %s\n' "$TAGS_FILE" >&2
  exit 1
fi

ALLOWED_TAGS=()
while IFS= read -r _line; do
  ALLOWED_TAGS+=("$_line")
done < <(awk '/^  (both|docs):/{p=1;next} p && /^  [[:alpha:]]/{p=0} p && /^    - /{sub(/^    - /,""); print}' "$TAGS_FILE")

if [[ ${#ALLOWED_TAGS[@]} -eq 0 ]]; then
  printf 'ERROR: no tags extracted from %s — check file format\n' "$TAGS_FILE" >&2
  exit 1
fi

# ── Find docs markdown files ───────────────────────────────────────────────────
MD_FILES=()
while IFS= read -r _f; do
  MD_FILES+=("$_f")
done < <(find "$DOCS_DIR" -name '*.md' | sort)

if [[ ${#MD_FILES[@]} -eq 0 ]]; then
  printf 'No markdown files found under %s\n' "$DOCS_DIR"
  exit 0
fi

# ── Validate tags in each file ─────────────────────────────────────────────────
_errors=()

for _f in "${MD_FILES[@]}"; do
  # Skip template files — they contain placeholder tags by design
  [[ "$_f" == *_TEMPLATE.md ]] && continue

  _tags_line=$(grep '^tags:' "$_f" 2>/dev/null | head -1)
  [[ -z "$_tags_line" ]] && continue   # no frontmatter tags line — skip

  _tags_raw=$(printf '%s' "$_tags_line" \
    | sed 's/^tags:[[:space:]]*//' \
    | tr -d '[]' \
    | tr ',' '\n')

  while IFS= read -r _tag; do
    _tag=$(printf '%s' "$_tag" | xargs)   # trim whitespace
    [[ -z "$_tag" ]] && continue

    _found=0
    for _allowed in "${ALLOWED_TAGS[@]}"; do
      [[ "$_tag" == "$_allowed" ]] && { _found=1; break; }
    done

    [[ $_found -eq 0 ]] && _errors+=("${_f##"$REPO_ROOT/"}: unknown tag '${_tag}'")
  done <<< "$_tags_raw"
done

# ── Report ─────────────────────────────────────────────────────────────────────
if [[ ${#_errors[@]} -eq 0 ]]; then
  printf '✓ frontmatter tags valid (%d file(s) checked)\n' "${#MD_FILES[@]}"
  exit 0
else
  printf 'frontmatter tag validation failed:\n\n' >&2
  for _e in "${_errors[@]}"; do
    printf '  %s\n' "$_e" >&2
  done
  printf '\nAllowed tags: %s\n' "${ALLOWED_TAGS[*]}" >&2
  exit 1
fi
