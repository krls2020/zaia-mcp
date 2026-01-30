#!/bin/bash
# PreToolUse hook: Checks CLAUDE.md is staged when key files are committed
# Reads Claude Code hook protocol JSON from stdin

set -euo pipefail

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_input',{}).get('command',''))" 2>/dev/null || echo "")

# Only check for git commit commands
if [[ "$COMMAND" != *"git commit"* ]]; then
    exit 0
fi

MODULE_ROOT=$(cd "$(dirname "$0")/../.." && pwd)
cd "$MODULE_ROOT"

# Check if key Go files are staged
KEY_STAGED=$(git diff --cached --name-only 2>/dev/null | grep -E '\.(go|mod)$' | head -1 || true)

if [[ -n "$KEY_STAGED" ]]; then
    # Check if CLAUDE.md is also staged
    CLAUDE_STAGED=$(git diff --cached --name-only 2>/dev/null | grep 'CLAUDE.md' || true)
    if [[ -z "$CLAUDE_STAGED" ]]; then
        echo "⚠️  Key files changed but CLAUDE.md not staged. Consider: git add CLAUDE.md"
    fi
fi
