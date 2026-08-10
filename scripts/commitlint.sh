#!/usr/bin/env bash

set -euo pipefail

re='^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(!)?(\([a-z0-9 .\-]+\))?!?: .+'

failed=0
while IFS= read -r subject; do
  [ -z "$subject" ] && continue
  case "$subject" in
    'Merge '*) continue ;;
  esac
  if ! printf '%s\n' "$subject" | grep -qE "$re"; then
    echo "::error::non-conforming commit subject: $subject" >&2
    failed=1
  fi
done

if [ "$failed" -ne 0 ]; then
  cat >&2 <<'MSG'
Commit subjects MUST follow Conventional Commits 1.0.0:
  <type>(<scope>)?!: <description>
Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert
Use ! after the type (or after the scope) for breaking changes.
See https://www.conventionalcommits.org/ and docs/decisions/0128.
MSG
  exit 1
fi
