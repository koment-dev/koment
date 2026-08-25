#!/usr/bin/env bash

set -euo pipefail

repository_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
test_directory=$(mktemp -d)

trap 'rm -r "$test_directory"' EXIT

binary_directory=$test_directory/bin
inactive_repository=$test_directory/inactive
mkdir -p "$binary_directory" "$inactive_repository"
git -C "$inactive_repository" init -q

(
  cd "$repository_root"
  go build -o "$binary_directory/koment" ./cmd/koment
)

export PATH="$binary_directory:$PATH"

pre_tool_output=$(
  cd "$inactive_repository"
  printf '%s\n' 'not JSON' | koment agents hook pre-tool
)
test "$pre_tool_output" = '{}'

stop_output=$(
  cd "$inactive_repository"
  printf '%s\n' 'not JSON' | koment agents hook stop
)
test "$stop_output" = '{}'

for session_start in \
  "$repository_root/integrations/agent-plugins/claude/scripts/session-start.sh" \
  "$repository_root/integrations/agent-plugins/codex/plugins/koment/scripts/session-start.sh"
do
  session_output=$(cd "$inactive_repository" && "$session_start")
  test -z "$session_output"
done

(
  cd "$repository_root"
  node --input-type=module - "$inactive_repository" <<'EOF'
import plugin from "./integrations/agent-plugins/opencode/index.js";

const integration = await plugin({ directory: process.argv[2] });
await integration.dispose();
EOF
)

printf '%s\n' 'Claude, OpenCode and Codex stayed inactive without .koment/policy.yaml'
