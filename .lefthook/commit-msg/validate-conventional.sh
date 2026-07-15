#!/usr/bin/env bash
# Validate the commit message against Conventional Commits.
# releaser-pleaser relies on these types to compute releases, so keep them honest.
# No Node dependency: pure bash, matching the vault's lefthook note.
set -euo pipefail

msg="$(head -1 "$1")"
pattern='^(feat|fix|docs|style|refactor|test|chore|perf|build|ci|revert)(\(.+\))?!?: .+'

# Allow merge commits through untouched.
if [[ "$msg" =~ ^Merge ]]; then
  exit 0
fi

if ! [[ "$msg" =~ $pattern ]]; then
  echo "✗ Commit message does not follow Conventional Commits: '$msg'"
  echo "  Valid examples: 'feat: add ZFS access metric', 'fix(collector): omit metrics on API error'"
  exit 1
fi
