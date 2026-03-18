# .dev/tags.sh — Tags task: list allowed tags from .github/tags.yml
# Sourced by ./dev — not executed directly.

run_tags() {
  header "tags"
  local _tags_file=".github/tags.yml"
  if [[ ! -f "$_tags_file" ]]; then
    log_error "tags file not found: $_tags_file"
    exit 1
  fi
  local _both _docs _scopes
  _both=$(awk '/^  both:/{p=1;next} p && /^  [[:alpha:]]/{p=0} p && /^    - /{sub(/^    - /,""); print}' "$_tags_file")
  _docs=$(awk '/^  docs:/{p=1;next}  p && /^  [[:alpha:]]/{p=0} p && /^    - /{sub(/^    - /,""); print}' "$_tags_file")
  _scopes=$(awk '/^  scopes:/{p=1;next} p && /^  [[:alpha:]]/{p=0} p && /^    - /{sub(/^    - /,""); print}' "$_tags_file")
  $GUM style --bold "both  (frontmatter + commit scope)"
  $GUM style --foreground 99 "$(printf '%s\n' $_both | paste - - - - - | column -t)"
  echo ""
  $GUM style --bold "docs  (frontmatter only)"
  $GUM style --foreground 99 "$(printf '%s\n' $_docs | paste - - - - - | column -t)"
  echo ""
  $GUM style --bold "scopes  (commit scope only)"
  $GUM style --foreground 99 "$(printf '%s\n' $_scopes)"
  return 0
}
