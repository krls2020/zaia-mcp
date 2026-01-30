#!/bin/bash
# PostToolUse hook: Runs go test + go vet after editing Go files
# Reads Claude Code hook protocol JSON from stdin

set -euo pipefail

# Read hook input
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_input',{}).get('file_path',''))" 2>/dev/null || echo "")

# Only run for Go files in this project
if [[ "$FILE_PATH" != *"/zaia-mcp/"* ]] || [[ "$FILE_PATH" != *.go ]]; then
    exit 0
fi

# Extract package directory
PKG_DIR=$(dirname "$FILE_PATH")
# Get path relative to module root
MODULE_ROOT=$(cd "$(dirname "$0")/../.." && pwd)

if [[ "$PKG_DIR" == "$MODULE_ROOT"* ]]; then
    REL_DIR="${PKG_DIR#$MODULE_ROOT/}"
    echo "🧪 Running tests for ./$REL_DIR"
    cd "$MODULE_ROOT"
    go test "./$REL_DIR" -count=1 2>&1 | tail -20
    go vet "./$REL_DIR" 2>&1
fi
