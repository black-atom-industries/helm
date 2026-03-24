# Session Navigation

## Purpose

Browsing, fuzzy filtering, expand/collapse, jump-to-index, and pinned self-session in the TUI session list.

## Requirements

### REQ-1: Session List Display

The TUI SHALL display all tmux sessions (excluding the current session from the main list) with their name, last activity time, and optional status indicators.

#### Scenario: Normal session list

- **WHEN** the user opens helm inside tmux
- **THEN** all sessions except the current one are listed, sorted by last activity
- **AND** the current session is pinned at the top as a separate "self" row

#### Scenario: Empty state

- **WHEN** there are no other tmux sessions
- **THEN** the message "No sessions. Press C-n to create one." SHALL be displayed

### REQ-2: Pinned Self Session

The current (self) session SHALL always appear pinned at the top of the session list, visually separated by a dotted border.

#### Scenario: Self session visibility during filtering

- **WHEN** a filter is active
- **THEN** the self session row SHALL remain visible and unaffected by the filter
- **AND** only non-self sessions are filtered

### REQ-3: Fuzzy Filter

Typing characters SHALL progressively filter the session list using case-insensitive substring matching.

#### Scenario: Filter narrows results

- **WHEN** the user types "api"
- **THEN** only sessions whose name contains "api" (case-insensitive) are shown
- **AND** the state line shows "Showing N/M sessions"

#### Scenario: Clear filter with Escape

- **WHEN** a filter is active and the user presses Escape
- **THEN** the filter is cleared and all sessions are shown again

#### Scenario: Clear filter with Backspace

- **WHEN** the user presses Backspace
- **THEN** the last character of the filter is removed and the list is re-filtered

### REQ-4: Cursor Navigation

The user SHALL navigate the session list using Ctrl+j/Ctrl+k or arrow keys.

#### Scenario: Move cursor down

- **WHEN** the user presses Ctrl+j or Down arrow
- **THEN** the cursor moves to the next item in the list

#### Scenario: Move cursor up

- **WHEN** the user presses Ctrl+k or Up arrow
- **THEN** the cursor moves to the previous item in the list

#### Scenario: Scroll offset

- **WHEN** the cursor moves beyond the visible area
- **THEN** the list scrolls to keep the cursor visible

### REQ-5: Expand/Collapse Sessions

Pressing Ctrl+l (or Right arrow) SHALL expand the selected session to show its windows. Pressing Ctrl+h (or Left arrow) SHALL collapse it.

#### Scenario: Expand session

- **WHEN** the user presses Ctrl+l on a collapsed session
- **THEN** that session's windows are shown below it
- **AND** all other sessions are collapsed (accordion behavior)

#### Scenario: Expand window to show panes

- **WHEN** the user presses Ctrl+l on a window within an expanded session
- **THEN** that window's panes are shown below it
- **AND** other windows in the same session are collapsed

#### Scenario: Collapse session

- **WHEN** the user presses Ctrl+h on an expanded session
- **THEN** the session's windows are hidden

#### Scenario: Collapse from child

- **WHEN** the user presses Ctrl+h on a window
- **THEN** the parent session collapses and the cursor moves to the session row

### REQ-6: Number Jump

Pressing a digit key (0-9) without an active filter SHALL jump to and switch to the corresponding session.

#### Scenario: Jump to session

- **WHEN** no filter is active and the user presses "3"
- **THEN** helm switches to the 4th non-self session (0-indexed) and exits

#### Scenario: Jump to window within expanded session

- **WHEN** the cursor is on an expanded session and the user presses a digit
- **THEN** helm switches to the window with that index within the session

#### Scenario: Numbers disabled during filter

- **WHEN** a filter is active
- **THEN** digit keys are appended to the filter text instead of triggering jumps

### REQ-7: Select Item

Pressing Enter SHALL switch to the highlighted session, window, or pane.

#### Scenario: Select session

- **WHEN** the user presses Enter on a session
- **THEN** tmux switches to that session and helm exits

#### Scenario: Select window

- **WHEN** the user presses Enter on a window within an expanded session
- **THEN** tmux switches to that specific window and helm exits

#### Scenario: Filter with no results creates session

- **WHEN** a filter is active with no matching sessions and the user presses Enter
- **THEN** helm enters ModeCreatePath to create a new session with the filter text as the name
