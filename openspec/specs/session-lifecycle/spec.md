# Session Lifecycle

## Purpose

Creating, switching to, and killing tmux sessions from the TUI.

## Requirements

### REQ-1: Create Session by Name

Pressing Ctrl+n SHALL enter create mode where the user types a session name.

#### Scenario: Create and switch

- **WHEN** the user presses Ctrl+n, types "myproject", and presses Enter
- **THEN** a new tmux session named "myproject" is created in the default session directory
- **AND** the configured layout script is applied (if any)
- **AND** helm switches to the new session and exits

#### Scenario: Empty name rejected

- **WHEN** the user submits an empty session name
- **THEN** an error message "Session name cannot be empty" is shown

#### Scenario: Name sanitization

- **WHEN** the session name contains dots, colons, slashes, or spaces
- **THEN** those characters SHALL be replaced with hyphens in the tmux session name

### REQ-2: Create Session at Path

When filtering produces no results, pressing Enter SHALL transition to a path input mode.

#### Scenario: Path input pre-filled

- **WHEN** the filter text "newproject" matches no sessions and the user presses Enter
- **THEN** the mode changes to ModeCreatePath
- **AND** the path input is pre-filled with the first project directory + the sanitized filter text

#### Scenario: Tab completion

- **WHEN** the user presses Tab in path input mode
- **THEN** the first matching directory completion is applied

#### Scenario: Folder creation

- **WHEN** the user enters a path that does not exist and presses Enter
- **THEN** the directory is created recursively
- **AND** a tmux session is created at that path with layout applied

### REQ-3: Create Session from Directory

Selecting a directory in the project picker SHALL create a new session there.

#### Scenario: New session from directory

- **WHEN** the user selects a directory in the project picker
- **THEN** a tmux session is created with a name derived from the path (using project_depth)
- **AND** the configured layout is applied and helm switches to the session

#### Scenario: Existing session switch

- **WHEN** the selected directory already has an active tmux session
- **THEN** helm switches to the existing session instead of creating a new one

### REQ-4: Kill Session

Pressing Ctrl+x SHALL enter kill confirmation mode. A second Ctrl+x confirms the kill.

#### Scenario: Kill confirmation

- **WHEN** the user presses Ctrl+x on a session
- **THEN** a confirmation message "Kill <name>?" is shown
- **AND** the mode changes to ModeConfirmKill

#### Scenario: Confirm kill

- **WHEN** in kill confirmation mode and the user presses Ctrl+x again
- **THEN** the session is killed via tmux
- **AND** the session list reloads

#### Scenario: Cancel kill

- **WHEN** in kill confirmation mode and the user presses Escape
- **THEN** the kill is cancelled and the mode returns to normal

#### Scenario: Kill window or pane

- **WHEN** the user confirms a kill on a window or pane (within an expanded session)
- **THEN** only that specific window or pane is killed, not the entire session

#### Scenario: Kill self session

- **WHEN** the user kills the self (current) session
- **THEN** helm switches to the most recent other session first
- **AND** then kills the self session and exits
