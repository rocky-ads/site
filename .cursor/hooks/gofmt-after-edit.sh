#!/usr/bin/env bash
set -euo pipefail

input=$(cat)
file_path=$(
  printf '%s' "$input" |
    sed -n 's/.*"file_path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' |
    head -1
)

if [[ -n "$file_path" && "$file_path" == *.go && -f "$file_path" ]]; then
  gofmt -w "$file_path"
fi

exit 0
