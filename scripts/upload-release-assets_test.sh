#!/usr/bin/env bash
set -euo pipefail

test_directory=$(mktemp -d)
trap 'rm -rf "$test_directory"' EXIT HUP INT TERM
fake_bin="$test_directory/bin"
mkdir "$fake_bin"

cat >"$fake_bin/curl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

printf '%s\n' "$*" >>"$GH_TEST_UPLOAD_LOG"

output_file=
while [ "$#" -gt 0 ]; do
  if [ "$1" = --output ]; then
    output_file=$2
    shift 2
    continue
  fi
  shift
done

if [ "${GH_TEST_CURL_MODE:-success}" = failure ]; then
  printf '%s\n' '{"message":"upload denied"}' >"$output_file"
  exit 22
fi
EOF

chmod +x "$fake_bin/curl"

upload_directory="$test_directory/upload"
mkdir "$upload_directory"
touch "$upload_directory/first.tar.gz" "$upload_directory/second.json"
PATH="$fake_bin:$PATH" \
GH_TOKEN=test-token \
GITHUB_REPOSITORY=koment-dev/koment \
GH_TEST_UPLOAD_LOG="$upload_directory/upload-log" \
  ./scripts/upload-release-assets.sh \
    'https://uploads.github.com/repos/koment-dev/koment/releases/123/assets{?name,label}' \
    "$upload_directory/first.tar.gz" "$upload_directory/second.json"

test "$(wc -l <"$upload_directory/upload-log" | tr -d ' ')" = 2
grep -Fq -- '--data-binary @'"$upload_directory"'/first.tar.gz' "$upload_directory/upload-log"
grep -Fq -- 'https://uploads.github.com/repos/koment-dev/koment/releases/123/assets?name=first.tar.gz' \
  "$upload_directory/upload-log"
grep -Fq -- '--data-binary @'"$upload_directory"'/second.json' "$upload_directory/upload-log"
grep -Fq -- 'https://uploads.github.com/repos/koment-dev/koment/releases/123/assets?name=second.json' \
  "$upload_directory/upload-log"

invalid_directory="$test_directory/invalid"
mkdir "$invalid_directory"
touch "$invalid_directory/first.tar.gz"
set +e
PATH="$fake_bin:$PATH" \
GH_TOKEN=test-token \
GITHUB_REPOSITORY=koment-dev/koment \
GH_TEST_UPLOAD_LOG="$invalid_directory/upload-log" \
  ./scripts/upload-release-assets.sh \
    'https://uploads.github.com/repos/another/project/releases/123/assets{?name,label}' \
    "$invalid_directory/first.tar.gz" >/dev/null 2>&1
status=$?
set -e

test "$status" -eq 2
test ! -e "$invalid_directory/upload-log"

failure_directory="$test_directory/failure"
mkdir "$failure_directory"
touch "$failure_directory/first.tar.gz"
set +e
failure_output=$(PATH="$fake_bin:$PATH" \
GH_TOKEN=test-token \
GITHUB_REPOSITORY=koment-dev/koment \
GH_TEST_CURL_MODE=failure \
GH_TEST_UPLOAD_LOG="$failure_directory/upload-log" \
  ./scripts/upload-release-assets.sh \
    'https://uploads.github.com/repos/koment-dev/koment/releases/123/assets{?name,label}' \
    "$failure_directory/first.tar.gz" 2>&1)
status=$?
set -e

test "$status" -eq 1
test "$(wc -l <"$failure_directory/upload-log" | tr -d ' ')" = 1
printf '%s\n' "$failure_output" | grep -Fq 'upload denied'

echo "release asset upload helper: ok"
