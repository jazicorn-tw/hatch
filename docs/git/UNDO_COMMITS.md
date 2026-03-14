<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [commit, devops]
description:  "How to undo local and remote commits using revert, reset, and interactive rebase"
-->
# Undo Commits

How to undo commits locally and on remote branches. Choose the approach based on whether others have pulled your commits.

---

## Quick Reference

| Situation                               | Command                   | Rewrites history? |
|-----------------------------------------|---------------------------|-------------------|
| Undo last commit, keep changes staged   | `git reset --soft HEAD~1` | Yes (local only)  |
| Undo last commit, keep changes unstaged | `git reset HEAD~1`        | Yes (local only)  |
| Undo last commit, discard all changes   | `git reset --hard HEAD~1` | Yes (local only)  |
| Remove a specific middle commit         | `git rebase -i <hash>`    | Yes (local only)  |
| Safely reverse a commit (shared branch) | `git revert <hash>`       | No                |

---

## Undo the Most Recent Commit

```bash
# Keep changes staged (ready to re-commit)
git reset --soft HEAD~1

# Keep changes in working tree (unstaged)
git reset HEAD~1

# Discard changes entirely (irreversible)
git reset --hard HEAD~1
```

---

## Undo Multiple Recent Commits

```bash
# Find the commit you want to roll back to
git log --oneline

# Reset to that commit (replace HEAD~3 with the target hash)
git reset --soft HEAD~3
```

---

## Remove a Specific Commit in the Middle

```bash
# Start interactive rebase from just before the target commit
git rebase -i <hash-of-commit-before-target>

# In the editor, change 'pick' to 'drop' for commits to remove
# Save and close — git will replay the remaining commits
```

---

## Push Changes to Remote

After any `reset` or `rebase`, the local branch has diverged from remote. Force push to overwrite:

```bash
# Safer: fails if someone else pushed since your last fetch
git push origin YOUR_BRANCH --force-with-lease

# Less safe: always overwrites (avoid on shared branches)
git push origin YOUR_BRANCH --force
```

> ⚠️ Never force push to `main` or a shared branch without coordinating with the team.
> Anyone who has pulled the old commits will need to reset their local branch.

---

## Safely Reverse a Commit on a Shared Branch

If others have already pulled the commits, rewrite with a new reversal commit instead:

```bash
# Creates a new commit that undoes the target commit
git revert <commit-hash>

# Revert a range of commits
git revert HEAD~3..HEAD

# Push normally — no force push needed
git push origin YOUR_BRANCH
```

`revert` preserves history and is safe for `main` and shared branches.

---

## Delete a Remote Tag

```bash
# Delete local tag
git tag -d v1.2.3

# Delete remote tag
git push origin --delete v1.2.3
```

---

## Related

- [`docs/commit/COMMIT_CHEAT_SHEET.md`](../commit/COMMIT_CHEAT_SHEET.md) — conventional commit format reference
- [`docs/commit/PRECOMMIT.md`](../commit/PRECOMMIT.md) — pre-commit hook setup
- [`docs/github/DELETE_PACKAGES.md`](../github/DELETE_PACKAGES.md) — how to delete published ghcr.io image versions
