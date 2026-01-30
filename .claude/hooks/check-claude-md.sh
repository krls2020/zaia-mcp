#!/bin/bash
# PostToolUse hook: Reminds to update CLAUDE.md when key files change
# Reads Claude Code hook protocol JSON from stdin

set -euo pipefail

INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_input',{}).get('file_path',''))" 2>/dev/null || echo "")

# Only check for key files in this project
if [[ "$FILE_PATH" != *"/zaia-mcp/"* ]]; then
    exit 0
fi

# Key files that should trigger CLAUDE.md check
KEY_FILES=(
    "server/server.go"
    "executor/executor.go"
    "tools/"
    "resources/"
    "go.mod"
)

for pattern in "${KEY_FILES[@]}"; do
    if [[ "$FILE_PATH" == *"$pattern"* ]]; then
        echo "📝 Reminder: Check if CLAUDE.md needs updating after changing $(basename "$FILE_PATH")"
        exit 0
    fi
done
