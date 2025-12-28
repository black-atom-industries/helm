# tmux-session-picker

Pop up fzf in tmux to quickly switch between sessions, create new ones, or manage existing sessions.

## Installation

### Option 1: Clone and Install (recommended)

```sh
git clone https://github.com/nikbrunner/tmux-session-picker.git
cd tmux-session-picker
./install.sh
```

This creates symlinks in `~/.local/bin/`, so updates are as simple as `git pull`.

### Option 2: Manual Installation

1.  Download the script:
    ```sh
    curl -O https://raw.githubusercontent.com/nikbrunner/tmux-session-picker/main/tmux-session-picker
    ```
2.  Make it executable:
    ```sh
    chmod +x tmux-session-picker
    ```
3.  Move it to a directory in your `PATH`:
    ```sh
    mv tmux-session-picker ~/.local/bin/
    ```

## Dependencies

- [fzf](https://github.com/junegunn/fzf): For the fuzzy-finding interface.
- [gum](https://github.com/charmbracelet/gum) (optional): For enhanced confirmations.

## Setup

Add a key binding to your `~/.tmux.conf` to launch the session picker:

```tmux
# ~/.tmux.conf
bind -n M-w display-popup -w65% -h35% -B -E "tmux-session-picker"
```

Reload your tmux configuration: `tmux source-file ~/.tmux.conf`

## Usage

1. Press your configured key binding (e.g., `Alt-w`).
2. An fzf window will pop up listing all sessions (except the current one).
3. Sessions are sorted by recency (most recently used first) and show relative time (e.g., "5m ago", "2h ago") to help identify stale sessions.

### Keybindings inside the picker

- `Enter`: Switch to the selected session, or create a new session if you typed a name
- `Ctrl-O`: Open window picker for the selected session
- `Ctrl-X`: Kill the selected session (picker stays open for more actions)
- `Esc`: Cancel and close the picker

### Window and Pane Navigation

When you press `Ctrl-O` on a session, you get a window picker with:

- `Enter`: Switch to the selected window
- `Ctrl-O`: Open pane picker for the selected window
- `Ctrl-X`: Kill the selected window
- `Esc`: Go back to session picker

### Creating New Sessions

Simply type a name that doesn't match any existing session and press `Enter` to create and switch to a new session.

## Claude Code Status Integration

Optionally display Claude Code status for each session, showing whether Claude is working or idle and for how long.

### Setup

The hook is already installed if you used `./install.sh`. Just configure Claude Code:

1. Add hooks to your `~/.claude/settings.json`:

   ```json
   {
     "hooks": {
       "PreToolUse": [
         {
           "hooks": [
             {
               "type": "command",
               "command": "~/.local/bin/tmux-session-picker-hook PreToolUse"
             }
           ]
         }
       ],
       "Stop": [
         {
           "hooks": [
             {
               "type": "command",
               "command": "~/.local/bin/tmux-session-picker-hook Stop"
             }
           ]
         }
       ],
       "Notification": [
         {
           "hooks": [
             {
               "type": "command",
               "command": "~/.local/bin/tmux-session-picker-hook Notification"
             }
           ]
         }
       ]
     }
   }
   ```

2. Enable the feature in your `~/.tmux.conf`:

   ```tmux
   set-environment -g TMUX_SESSION_PICKER_CLAUDE_STATUS 1
   ```

   Then reload: `tmux source-file ~/.tmux.conf`

### Display

When enabled, sessions show Claude status:

```
my-project (2m ago) [CC: working 30s]   # Claude actively working
other-work (15m ago) [CC: done 5m]      # Claude idle/finished
new-session (1m ago) [CC: unknown]      # Claude present, status unknown
plain-session (1h ago)                   # No Claude in this session
```

- `[CC: working 30s]` - Claude actively processing (working in yellow)
- `[CC: done 5m]` - Claude finished, waiting for input (done in green)
- `[CC: unknown]` - Claude detected but hooks not yet fired (restart Claude to enable)

A notification sound plays when Claude finishes (macOS).

## Layout Support

The picker supports automatic layout application for new sessions via environment variables:

```bash
# Set default layout (default: ide)
export TMUX_LAYOUT="ide"

# Set layouts directory (default: ~/.config/tmux/layouts)
export TMUX_LAYOUTS_DIR="$HOME/.config/tmux/layouts"

# Disable layouts
export TMUX_LAYOUT=""
```

Layout scripts receive the session name and working directory as arguments.

## License

MIT
