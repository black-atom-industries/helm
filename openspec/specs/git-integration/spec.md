# Git Integration

## Purpose

Per-session git status indicators, open remote in browser, and lazygit integration.

## Requirements

### REQ-1: Git Status Indicators

When `git_status_enabled` is true, each session row SHALL display git status information.

#### Scenario: Dirty repo indicator

- **WHEN** a session's working directory is a git repo with uncommitted changes
- **THEN** the session row shows the dirty file count and line additions/deletions

#### Scenario: Clean repo hidden

- **WHEN** a session's working directory is a clean git repo (or not a repo)
- **THEN** no git status indicator is shown for that session

#### Scenario: Async loading

- **WHEN** sessions are loaded
- **THEN** git statuses are fetched asynchronously in parallel (one command per session)
- **AND** the UI updates incrementally as each status arrives

#### Scenario: Loading indicator delay

- **WHEN** git statuses are still loading after 500ms
- **THEN** a loading indicator is shown for pending sessions

### REQ-2: Open Remote in Browser

Pressing Ctrl+r SHALL open the selected session's git remote URL in the browser.

#### Scenario: Open remote

- **WHEN** the user presses Ctrl+r on a session with a git remote
- **THEN** the remote URL is opened in the default browser (via `open` command)
- **AND** a message "Opened: owner/repo" is shown temporarily

#### Scenario: No remote

- **WHEN** the session has no git remote
- **THEN** an error message "No git remote found" is shown

#### Scenario: No path

- **WHEN** the session's working directory cannot be determined
- **THEN** an error message "Could not get session path" is shown

### REQ-3: Lazygit Integration

Pressing Ctrl+g SHALL open lazygit in the selected session's directory.

#### Scenario: Open lazygit

- **WHEN** the user presses Ctrl+g on a session
- **THEN** helm exits
- **AND** lazygit opens in a tmux popup at the session's working directory
- **AND** after lazygit closes, helm reopens with the same dimensions

#### Scenario: Popup dimensions

- **WHEN** lazygit is launched
- **THEN** the popup uses the dimensions from `lazygit_popup` config (default 90%x90%)
