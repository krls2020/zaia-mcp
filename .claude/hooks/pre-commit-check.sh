#!/bin/bash
# PreToolUse hook: Before git commit, check CLAUDE.md is staged

input=$(cat)
COMMAND=$(echo "$input" | grep -o '"command":"[^"]*"' | head -1 | sed 's/"command":"//;s/"//')

echo "$COMMAND" | grep -qE 'git commit' || exit 0

MODULE_ROOT=$(cd "$(dirname "$0")/../.." && pwd)
cd "$MODULE_ROOT" || exit 0

STAGED=$(git diff --cached --name-only 2>/dev/null)
[ -z "$STAGED" ] && exit 0

KEY_PATTERNS="server/server\.go|executor/executor\.go|tools/convert\.go|tools/discover\.go|resources/knowledge\.go|go\.mod"

KEY_STAGED=$(echo "$STAGED" | grep -E "$KEY_PATTERNS" | head -1)

if [ -n "$KEY_STAGED" ]; then
    CLAUDE_STAGED=$(echo "$STAGED" | grep 'CLAUDE.md')
    if [ -z "$CLAUDE_STAGED" ]; then
        echo "CLAUDE.md CHECK: Key files staged but CLAUDE.md not included. Consider: git add CLAUDE.md"
    fi
fi

exit 0
