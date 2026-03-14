<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [commit, devops]
description:  "How to rebase a branch onto main and resolve conflicts"
-->
# Rebase

How to rebase a feature branch onto `main` to incorporate upstream changes and maintain a clean linear history.

---

## Quick Reference

| Command                       | Purpose                                            |
|-------------------------------|----------------------------------------------------|
| `git fetch origin`            | Pull latest remote state without merging           |
| `git rebase origin/main`      | Rebase current branch onto latest main             |
| `git rebase -i HEAD~N`        | Interactive rebase — squash, reorder, drop commits |
| `git rebase --continue`       | Continue after resolving a conflict                |
| `git rebase --abort`          | Cancel and return to pre-rebase state              |
| `git push --force-with-lease` | Push rebased branch to remote                      |

---

## Standard Rebase onto Main

```bash
# 1. Fetch latest remote state
git fetch origin

# 2. Switch to your feature branch
git checkout your-branch

# 3. Rebase onto main
git rebase origin/main

# 4. Push the rebased branch (force required — history was rewritten)
git push origin your-branch --force-with-lease
```

---

## Pull with Rebase (instead of merge)

Keeps local branch history linear when pulling updates:

```bash
git pull --rebase origin main
```

To make this the default for all pulls:

```bash
git config --global pull.rebase true
```

---

## Interactive Rebase

Rewrite, squash, or drop commits before merging:

```bash
# Rebase the last N commits interactively
git rebase -i HEAD~3

# Or from a specific commit hash
git rebase -i <hash-before-target>
```

In the editor, change the action keyword for each commit:

| Keyword  | Shorthand | Action                               |
|----------|-----------|--------------------------------------|
| `pick`   | `p`       | Keep commit as-is                    |
| `reword` | `r`       | Keep commit, edit message            |
| `squash` | `s`       | Merge into previous commit           |
| `fixup`  | `f`       | Merge into previous, discard message |
| `drop`   | `d`       | Remove commit entirely               |

---

## Resolving Conflicts During Rebase

When a conflict occurs, git pauses and shows the conflicting files:

```bash
# 1. Open conflicting files and resolve manually
# Conflict markers look like:
# <<<<<<< HEAD
# your changes
# =======
# incoming changes
# >>>>>>> commit-hash

# 2. Stage resolved files
git add <resolved-file>

# 3. Continue the rebase
git rebase --continue

# If the conflict is too complex, abort and start over
git rebase --abort
```

---

## Notes

- Rebase rewrites commit hashes — always use `--force-with-lease` instead of `--force` when pushing.
- Never rebase `main` or any shared branch that others have pulled from.
- If a rebase produces an empty commit (all changes already exist upstream), skip it with `git rebase --skip`.

---

## Related

- [`docs/github/UNDO_COMMITS.md`](UNDO_COMMITS.md) — reset, revert, and drop commits
- [`docs/commit/COMMIT_CHEAT_SHEET.md`](../commit/COMMIT_CHEAT_SHEET.md) — conventional commit format reference
