#!/usr/bin/env bash
# Claude Code hook - writes status to ~/.cache/helm/
# Used by helm to display Claude status per session
#
# Status file format is JSON: {"state":"working","ts":1730000000,"tool":"Bash",...}
# helm also still parses the legacy "state:timestamp" format.

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

# write_status <state> — persist state plus context from the hook payload
# (tool name, session id, transcript path). Falls back to a minimal JSON
# object if jq is unavailable or the payload is malformed.
write_status() {
    local state="$1"
    echo "$INPUT" | jq -c --arg state "$state" --argjson ts "$TIMESTAMP" \
        '{state: $state, ts: $ts, tool: (.tool_name // ""), session_id: (.session_id // ""), transcript: (.transcript_path // ""), cwd: (.cwd // "")}' \
        >"$STATUS_FILE" 2>/dev/null ||
        echo "{\"state\":\"$state\",\"ts\":$TIMESTAMP}" >"$STATUS_FILE"
}

case "$HOOK_TYPE" in
    "SessionStart")
        write_status "new"
        ;;
    "PreToolUse")
        # AskUserQuestion means Claude is asking the user something — treat as waiting
        TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null)
        if [[ "$TOOL_NAME" == "AskUserQuestion" ]]; then
            write_status "waiting"
        else
            write_status "working"
        fi
        ;;
    "Stop"|"SubagentStop"|"Notification")
        # Running background tasks keep the session logically working
        BG_COUNT=$(echo "$INPUT" | jq -r '(.background_tasks // []) | length' 2>/dev/null)
        if [[ "$BG_COUNT" =~ ^[0-9]+$ ]] && ((BG_COUNT > 0)); then
            write_status "working"
        else
            write_status "waiting"
        fi
        ;;
    "SessionEnd")
        # Clean up status file when Claude session ends
        rm -f "$STATUS_FILE"
        ;;
esac

exit 0
