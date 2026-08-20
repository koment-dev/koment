#!/usr/bin/env bash
set -euo pipefail

test_directory=$(mktemp -d)
trap 'rm -rf "$test_directory"' EXIT HUP INT TERM
fake_bin="$test_directory/bin"
mkdir "$fake_bin"

cat >"$fake_bin/gh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

case "${1:-} ${2:-}" in
  "release view")
    count=0
    if [ -f "$GH_TEST_VIEW_COUNT" ]; then
      read -r count <"$GH_TEST_VIEW_COUNT"
    fi
    count=$((count + 1))
    printf '%s\n' "$count" >"$GH_TEST_VIEW_COUNT"
    if [ "$GH_TEST_MODE" = eventual ] && [ "$count" -ge 3 ]; then
      exit 0
    fi
    exit 1
    ;;
  "release upload")
    printf '%s\n' "$*" >"$GH_TEST_UPLOAD_LOG"
    ;;
  *)
    exit 2
    ;;
esac
EOF

cat >"$fake_bin/sleep" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$1" >>"$GH_TEST_SLEEP_LOG"
EOF

chmod +x "$fake_bin/gh" "$fake_bin/sleep"

eventual_directory="$test_directory/eventual"
mkdir "$eventual_directory"
PATH="$fake_bin:$PATH" \
GH_TOKEN=test-token \
GITHUB_REPOSITORY=koment-dev/koment \
GH_TEST_MODE=eventual \
GH_TEST_VIEW_COUNT="$eventual_directory/view-count" \
GH_TEST_UPLOAD_LOG="$eventual_directory/upload-log" \
GH_TEST_SLEEP_LOG="$eventual_directory/sleep-log" \
  ./scripts/upload-release-assets.sh v3.1.1 first.tar.gz second.json

test "$(cat "$eventual_directory/view-count")" = 3
test "$(cat "$eventual_directory/sleep-log")" = $'10\n20'
test "$(cat "$eventual_directory/upload-log")" = \
  "release upload v3.1.1 --repo koment-dev/koment first.tar.gz second.json --clobber"

bounded_directory="$test_directory/bounded"
mkdir "$bounded_directory"
set +e
PATH="$fake_bin:$PATH" \
GH_TOKEN=test-token \
GITHUB_REPOSITORY=koment-dev/koment \
GH_TEST_MODE=never \
GH_TEST_VIEW_COUNT="$bounded_directory/view-count" \
GH_TEST_UPLOAD_LOG="$bounded_directory/upload-log" \
GH_TEST_SLEEP_LOG="$bounded_directory/sleep-log" \
  ./scripts/upload-release-assets.sh v3.1.1 first.tar.gz >/dev/null 2>&1
status=$?
set -e

test "$status" -eq 1
test "$(cat "$bounded_directory/view-count")" = 6
test "$(cat "$bounded_directory/sleep-log")" = $'10\n20\n30\n40\n50'
test ! -e "$bounded_directory/upload-log"

echo "release asset upload helper: ok"
