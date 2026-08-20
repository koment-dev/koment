#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 2 ]; then
  echo "usage: upload-release-assets.sh <upload-url> <asset> [<asset> ...]" >&2
  exit 2
fi

: "${GH_TOKEN:?GH_TOKEN must be set}"
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY must be set}"

release_upload_url=${1%\{?name,label\}}
shift

expected_prefix="https://uploads.github.com/repos/${GITHUB_REPOSITORY}/releases/"
case "$release_upload_url" in
  "$expected_prefix"*/assets) ;;
  *)
    echo "release upload URL does not target $GITHUB_REPOSITORY" >&2
    exit 2
    ;;
esac

release_identifier=${release_upload_url#"$expected_prefix"}
release_identifier=${release_identifier%/assets}
case "$release_identifier" in
  ''|*[!0-9]*)
    echo "release upload URL does not contain a numeric release identifier" >&2
    exit 2
    ;;
esac

response_file=$(mktemp)
trap 'rm -f "$response_file"' EXIT HUP INT TERM

for asset in "$@"; do
  test -f "$asset"
  asset_name=${asset##*/}
  case "$asset_name" in
    *[!A-Za-z0-9._-]*)
      echo "release asset name contains unsupported characters: $asset_name" >&2
      exit 2
      ;;
  esac

  if ! curl --fail-with-body --location --silent --show-error \
      --request POST \
      --header 'Accept: application/vnd.github+json' \
      --header "Authorization: Bearer $GH_TOKEN" \
      --header 'X-GitHub-Api-Version: 2026-03-10' \
      --header 'Content-Type: application/octet-stream' \
      --data-binary "@$asset" \
      --output "$response_file" \
      "${release_upload_url}?name=${asset_name}"; then
    cat "$response_file" >&2
    exit 1
  fi
done
