# Open Git Remote in Browser (DEV-264)

## Summary

Add `Ctrl+r` binding to open the GitHub remote URL for the selected session in the default browser. Reassign clone from `Ctrl+r` to `Ctrl+d`.

## Keybinding Changes

| Binding  | Before       | After        |
| -------- | ------------ | ------------ |
| `Ctrl+r` | Clone repo   | Remote (new) |
| `Ctrl+d` | (unused)     | Download     |

Help line: `C-n New · C-p Projects · C-b Bookmarks · C-a Bookmark · C-r Remote · C-d Download · C-g Lazygit`

## Behavior

1. Get session path via `git.GetSessionPath(sessionName)`
2. Run `git -C <path> remote get-url origin`
3. Normalize URL: convert SSH (`git@github.com:org/repo.git`) to HTTPS (`https://github.com/org/repo`)
4. Run `open <url>` in background
5. Show status message (`"Opened: org/repo"`) — helm stays open
6. Clear message after 5 seconds

## Error Handling

- Not a git repo: `"Not a git repository"`
- No remote configured: `"No git remote found"`
- Can't resolve session path: `"Could not get session path"`

## Files to Change

- `internal/git/status.go` — add `GetRemoteURL(dir string) (string, error)`
- `internal/ui/keys.go` — rename `CloneRepo` to `DownloadRepo`, add `OpenRemote`, update help text
- `internal/model/model.go` — add `openRemote()`, rewire `Ctrl+r`/`Ctrl+d`

## Out of Scope

- Multi-remote support (only `origin`)
- Cross-platform `open` command (macOS only)
- Config toggle
