# Bookmarks

## Purpose

Bookmark mode for quick-access to frequently used project directories, with TUI management and CLI quickstart.

## Requirements

### REQ-1: Enter Bookmarks Mode

Pressing Ctrl+b SHALL switch to bookmarks mode showing all configured bookmarks.

#### Scenario: Show bookmarks

- **WHEN** the user presses Ctrl+b
- **THEN** the view switches to bookmarks mode
- **AND** all bookmarks are listed with their derived session names
- **AND** any active filter from session mode is carried over

#### Scenario: No bookmarks

- **WHEN** bookmarks mode is entered with no bookmarks configured
- **THEN** the message "No bookmarks configured" and "Press C-a to add a bookmark" are shown

### REQ-2: Open Bookmark

Pressing Enter on a bookmark SHALL create a session (if needed) and switch to it.

#### Scenario: Bookmark with existing session

- **WHEN** the selected bookmark has an active tmux session
- **THEN** helm switches to that session and exits

#### Scenario: Bookmark without session

- **WHEN** the selected bookmark has no active session
- **THEN** a new session is created at the bookmark's path
- **AND** the configured layout is applied
- **AND** helm switches to the new session and exits

### REQ-3: Add Bookmark

Pressing Ctrl+a SHALL add a bookmark.

#### Scenario: Add from session list

- **WHEN** the user presses Ctrl+a in normal mode on a session
- **THEN** the session's working directory path is added to bookmarks
- **AND** the config is saved

#### Scenario: Add from bookmarks mode

- **WHEN** the user presses Ctrl+a in bookmarks mode
- **THEN** the project directory picker opens
- **AND** selecting a directory adds it as a bookmark and returns to bookmarks mode

#### Scenario: Duplicate prevention

- **WHEN** the user tries to bookmark an already-bookmarked path
- **THEN** an error message "Already bookmarked" or "Session already bookmarked" is shown

### REQ-4: Remove Bookmark

Pressing Ctrl+x in bookmarks mode SHALL remove the selected bookmark.

#### Scenario: Remove bookmark

- **WHEN** the user presses Ctrl+x on a bookmark
- **THEN** the bookmark is removed from the list
- **AND** the config is saved
- **AND** the message "Bookmark removed" is shown

### REQ-5: Reorder Bookmarks

Ctrl+p SHALL move the selected bookmark up; Ctrl+n SHALL move it down.

#### Scenario: Move bookmark up

- **WHEN** the user presses Ctrl+p on a bookmark
- **THEN** it swaps position with the bookmark above it
- **AND** the config is saved

#### Scenario: Move bookmark down

- **WHEN** the user presses Ctrl+n on a bookmark
- **THEN** it swaps position with the bookmark below it

#### Scenario: Boundary

- **WHEN** the bookmark is at the top and Ctrl+p is pressed (or bottom and Ctrl+n)
- **THEN** nothing happens

### REQ-6: CLI Quickstart

The `helm bookmark <N>` command SHALL open the bookmark at slot N (0-9).

#### Scenario: Quickstart existing session

- **WHEN** `helm bookmark 0` is run and slot 0's session exists
- **THEN** tmux switches to that session

#### Scenario: Quickstart new session

- **WHEN** `helm bookmark 0` is run and slot 0's session does not exist
- **THEN** a new session is created at the bookmark's path with layout applied
- **AND** tmux switches to the new session

#### Scenario: Invalid slot

- **WHEN** `helm bookmark 15` is run
- **THEN** an error "invalid slot: 15 (must be 0-9)" is returned

### REQ-7: Bookmark Persistence

Bookmarks SHALL be stored in `~/.config/helm/bookmarks.yml` separately from the main config.

#### Scenario: Save bookmarks

- **WHEN** bookmarks are modified (add, remove, reorder)
- **THEN** the changes are written to bookmarks.yml
- **AND** the main config.yml is not modified
