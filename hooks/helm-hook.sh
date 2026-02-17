#!/usr/bin/env bash
# Claude Code hook - writes status to ~/.cache/helm/
# Used by helm to display Claude status per session

STATUS_DIR="$HOME/.cache/helm"
mkdir -p "$STATUS_DIR"

# Read JSON from stdin (required by Claude Code hooks)
INPUT=$(cat)

# Get tmux session name
TMUX_SESSION=$(tmux display-message -p '#{session_name}' 2>/dev/null)
[[ -z "$TMUX_SESSION" ]] && exit 0

HOOK_TYPE="$1"
STATUS_FILE="$STATUS_DIR/${TMUX_SESSION}.status"
TIMESTAMP=$(date +%s)

case "$HOOK_TYPE" in
    "SessionStart")
        echo "new:$TIMESTAMP" > "$STATUS_FILE"
        ;;
    "PreToolUse")
        # AskUserQuestion means Claude is asking the user something — treat as waiting
        TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null)
        if [[ "$TOOL_NAME" == "AskUserQuestion" ]]; then
            echo "waiting:$TIMESTAMP" > "$STATUS_FILE"
        else
            echo "working:$TIMESTAMP" > "$STATUS_FILE"
        fi
        ;;
    "Stop"|"SubagentStop"|"Notification")
        echo "waiting:$TIMESTAMP" > "$STATUS_FILE"
        ;;
    "SessionEnd")
        # Clean up status file when Claude session ends
        rm -f "$STATUS_FILE"
        ;;
esac

exit 0
