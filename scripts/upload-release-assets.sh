#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 2 ]; then
  echo "usage: upload-release-assets.sh <tag> <asset> [<asset> ...]" >&2
  exit 2
fi

: "${GH_TOKEN:?GH_TOKEN must be set}"
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY must be set}"

release_tag=$1
shift

attempt=1
maximum_attempts=6
while ! gh release view "$release_tag" --repo "$GITHUB_REPOSITORY" >/dev/null 2>&1; do
  if [ "$attempt" -eq "$maximum_attempts" ]; then
    echo "release $release_tag was not visible in $GITHUB_REPOSITORY after $maximum_attempts attempts" >&2
    exit 1
  fi
  wait_seconds=$((attempt * 10))
  echo "release $release_tag is not visible yet; retrying in $wait_seconds seconds" >&2
  sleep "$wait_seconds"
  attempt=$((attempt + 1))
done

gh release upload "$release_tag" --repo "$GITHUB_REPOSITORY" "$@" --clobber
