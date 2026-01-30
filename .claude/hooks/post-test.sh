#!/bin/bash
# PostToolUse hook: Runs go test + go vet after editing Go files

input=$(cat)
CHANGED_FILE=$(echo "$input" | grep -o '"file_path":"[^"]*"' | head -1 | sed 's/"file_path":"//;s/"//')

[ -z "$CHANGED_FILE" ] && exit 0

# Only for Go files
echo "$CHANGED_FILE" | grep -qE '\.go$' || exit 0

MODULE_ROOT=$(cd "$(dirname "$0")/../.." && pwd)

# Only run for files in this project
echo "$CHANGED_FILE" | grep -q "$MODULE_ROOT" || exit 0

cd "$MODULE_ROOT" || exit 0

PKG_DIR=$(echo "$CHANGED_FILE" | sed "s|^${MODULE_ROOT}/||" | xargs dirname)

if [ -d "$PKG_DIR" ]; then
    echo "-- go test ./${PKG_DIR} --"
    go test "./${PKG_DIR}" -count=1 2>&1 | tail -20
    go vet "./${PKG_DIR}" 2>&1
fi

exit 0
