#!/bin/bash
# PostToolUse hook: Remind to check CLAUDE.md when key files change

input=$(cat)
CHANGED_FILE=$(echo "$input" | grep -o '"file_path":"[^"]*"' | head -1 | sed 's/"file_path":"//;s/"//')

[ -z "$CHANGED_FILE" ] && exit 0

if echo "$CHANGED_FILE" | grep -qE "(server/server\.go|executor/executor\.go|tools/|resources/|go\.mod)"; then
    echo "CLAUDE.md CHECK: Key file changed: $(basename "$CHANGED_FILE"). Check if CLAUDE.md or README.md needs updating."
fi

exit 0
