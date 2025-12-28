#!/usr/bin/env bash
# Install script for tmux-session-picker

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Create ~/.local/bin if it doesn't exist
mkdir -p "$HOME/.local/bin"

# Symlink the main script
ln -sf "$SCRIPT_DIR/tmux-session-picker" "$HOME/.local/bin/tmux-session-picker"
chmod +x "$SCRIPT_DIR/tmux-session-picker"

# Symlink the Claude status hook
ln -sf "$SCRIPT_DIR/hooks/claude-status-hook.sh" "$HOME/.local/bin/tmux-session-picker-hook"
chmod +x "$SCRIPT_DIR/hooks/claude-status-hook.sh"

echo "Installed:"
echo "  ~/.local/bin/tmux-session-picker"
echo "  ~/.local/bin/tmux-session-picker-hook"
echo ""
echo "Make sure ~/.local/bin is in your PATH."
echo ""
echo "To enable Claude Code status integration (optional):"
echo ""
echo "1. Add hooks to ~/.claude/settings.json:"
cat << 'EOF'

{
  "hooks": {
    "PreToolUse": [
      {"hooks": [{"type": "command", "command": "~/.local/bin/tmux-session-picker-hook PreToolUse"}]}
    ],
    "Stop": [
      {"hooks": [{"type": "command", "command": "~/.local/bin/tmux-session-picker-hook Stop"}]}
    ],
    "Notification": [
      {"hooks": [{"type": "command", "command": "~/.local/bin/tmux-session-picker-hook Notification"}]}
    ]
  }
}

EOF
echo "2. Enable in your ~/.tmux.conf:"
echo "   set-environment -g TMUX_SESSION_PICKER_CLAUDE_STATUS 1"
echo ""
echo "   Then reload: tmux source-file ~/.tmux.conf"
