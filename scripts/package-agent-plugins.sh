#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 2 ]]; then
  printf 'usage: %s VERSION OUTPUT_DIRECTORY\n' "$0" >&2
  exit 2
fi

version=$1
output_directory=$2
plugin_root=integrations/agent-plugins
staging_directory=$(mktemp -d)

trap 'rm -r "$staging_directory"' EXIT

mkdir -p "$output_directory"

for plugin in claude hermes opencode; do
  source_directory=$plugin_root/$plugin
  packaged_directory=$staging_directory/$plugin
  archive=$output_directory/koment-plugin-${plugin}_v${version}.tar.gz

  case $plugin in
    claude)
      source_entries=".claude-plugin .mcp.json README.md commands hooks scripts skills"
      required_files=".claude-plugin/plugin.json .mcp.json README.md hooks/hooks.json scripts/session-start.sh skills/koment/SKILL.md"
      ;;
    hermes)
      source_entries="README.md __init__.py plugin.yaml"
      required_files="README.md plugin.yaml __init__.py"
      ;;
    opencode)
      source_entries="README.md index.js package.json plugin.json"
      required_files="README.md plugin.json package.json index.js"
      ;;
  esac

  mkdir -p "$packaged_directory"
  for source_entry in $source_entries; do
    test -e "$source_directory/$source_entry" || {
      printf '%s is missing %s\n' "$source_directory" "$source_entry" >&2
      exit 1
    }
    cp -R "$source_directory/$source_entry" "$packaged_directory/"
  done
  cp LICENSE "$packaged_directory/LICENSE"

  if [[ $plugin == claude ]] && ! compgen -G "$source_directory/commands/*.md" >/dev/null; then
    printf '%s contains no slash commands\n' "$source_directory" >&2
    exit 1
  fi

  tar -czf "$archive" -C "$staging_directory" "$plugin"

  for required_file in $required_files LICENSE; do
    tar -tzf "$archive" "$plugin/$required_file" >/dev/null
  done

  printf '%s\n' "$archive"
done

npm_directory=$output_directory/opencode-npm
test ! -e "$npm_directory" || {
  printf '%s already exists\n' "$npm_directory" >&2
  exit 1
}
cp -R "$staging_directory/opencode" "$npm_directory"
printf '%s\n' "$npm_directory"
