#!/usr/bin/env bash
set -euo pipefail

input=$(cat)
status=$(
  printf '%s' "$input" |
    sed -n 's/.*"status"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' |
    head -1
)

if [[ "$status" == "aborted" ]]; then
  exit 0
fi

if command -v go >/dev/null 2>&1; then
  go fmt ./...
fi

exit 0
