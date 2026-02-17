# helm

> Take the helm of your workspaces.

A TUI for quickly switching between tmux sessions, creating new ones, and managing your workspace. Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) for a fast, responsive interface.

Part of the [Black Atom Industries](https://github.com/black-atom-industries) cockpit - pairs with [radar.nvim](https://github.com/black-atom-industries/radar.nvim) for file navigation.

## Features

- Fuzzy filtering (just start typing)
- Ctrl-based navigation (`Ctrl+j/k`) to preserve filter input
- Number shortcuts for instant session switching (`1`-`9`)
- Expandable sessions to view windows
- Quick kill with confirmation (`Ctrl+x`)
- Create new sessions inline (`Ctrl+n`)
- Project picker (`Ctrl+p`)
- Bookmarks (`Ctrl+b`)
- Claude Code status integration (animated spinner)
- Git status per session (dirty/ahead/behind)

## Installation

### Prerequisites

- Go 1.21+
- tmux

### Build and Install

```sh
git clone https://github.com/black-atom-industries/helm.git
cd helm
make install
```

This builds the `helm` binary and installs it to `~/.local/bin/`.

## Setup

Add a key binding to your `~/.tmux.conf`:

```tmux
bind -n M-w display-popup -w50% -h35% -B -E "helm"
```

Reload your tmux configuration: `tmux source-file ~/.tmux.conf`

## Keybindings

| Key | Action |
|-----|--------|
| Type letters | Fuzzy filter sessions |
| `Ctrl+j/k` or `↓`/`↑` | Navigate up/down |
| `Ctrl+h/l` or `←`/`→` | Collapse/Expand session windows |
| `1`-`9` | Jump to session (when no filter active) |
| `Enter` | Switch to selected session/window |
| `Ctrl+x` | Kill with confirmation |
| `Ctrl+n` | Create new session |
| `Ctrl+p` | Project picker |
| `Ctrl+b` | Bookmarks |
| `Ctrl+a` | Add/remove bookmark |
| `Ctrl+r` | Clone repo from GitHub |
| `Ctrl+g` | Open lazygit |
| `q`/`Esc` | Quit |

## Configuration

Initialize config file:

```sh
helm init
```

Config location: `~/.config/helm/config.yml`

## Repository Management

helm includes CLI subcommands for managing all repos under your configured `project_dirs`.

### Bulk Clone

```sh
helm setup
```

Clones all repositories listed in `ensure_cloned` config. Supports wildcards (`org/*`) via `gh` CLI and post-clone hooks.

### Sync Commands

```sh
helm repos status              # Show sync state of all repos
helm repos pull                # Fetch and pull (ff-only) clean repos
helm repos push                # Push all ahead repos (including dirty+ahead)
helm repos dirty               # Print paths of dirty repos
helm repos dirty --walk        # Run configured command on each dirty repo
helm repos rebuild             # Re-run post_clone hooks
```

All commands support `--json` for machine-readable output.

### Dirty Walkthrough

Configure a command to run on each dirty repo:

```yaml
# ~/.config/helm/config.yml
dirty_walkthrough_command: "lazygit -p {}"
```

Then `helm repos dirty --walk` steps through each dirty repo with lazygit. Use `{}` as the path placeholder — works with any command.

## Claude Code Status Integration

Display Claude Code status for each session with an animated indicator.

### Setup

1. Copy the hook script:

   ```sh
   cp hooks/helm-hook.sh ~/.local/bin/
   chmod +x ~/.local/bin/helm-hook.sh
   ```

2. Add hooks to your `~/.claude/settings.json`:

   ```json
   {
     "hooks": {
       "SessionStart": [{ "hooks": [{ "type": "command", "command": "~/.local/bin/helm-hook.sh SessionStart" }] }],
       "PreToolUse": [{ "hooks": [{ "type": "command", "command": "~/.local/bin/helm-hook.sh PreToolUse" }] }],
       "Stop": [{ "hooks": [{ "type": "command", "command": "~/.local/bin/helm-hook.sh Stop" }] }],
       "Notification": [{ "hooks": [{ "type": "command", "command": "~/.local/bin/helm-hook.sh Notification" }] }],
       "SessionEnd": [{ "hooks": [{ "type": "command", "command": "~/.local/bin/helm-hook.sh SessionEnd" }] }]
     }
   }
   ```

3. Enable in config (`~/.config/helm/config.yml`):

   ```yaml
   claude_status_enabled: true
   ```

### Display

Sessions show Claude status as a single animated character:
- `⠤⠆⠒⠰` (spinner) - Claude actively processing
- `?` - Claude waiting for input
- `!` - Claude waiting for input > 5 minutes (needs attention)

## Project Tracking

Issues and roadmap are tracked in [Linear](https://linear.app/black-atom-industries) under the Development team with the `helm` label.

## License

MIT
