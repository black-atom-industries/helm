# Session Navigation

## MODIFIED Requirements

### Requirement: Session List Display

The TUI SHALL display all tmux sessions inside a SESSIONS section box with a label-on-border header, positioned to the left of the sidebar. Sessions are listed with their name, last activity time, and optional status indicators.

#### Scenario: Normal session list

- **WHEN** the user opens helm inside tmux
- **THEN** all sessions except the current one are listed inside a bordered SESSIONS section box
- **AND** the current session is pinned at the top as a separate "self" row
- **AND** the ACTIONS and STATUS section boxes render to the right

#### Scenario: Empty state

- **WHEN** there are no other tmux sessions
- **THEN** the message "No sessions. Press C-n to create one." SHALL be displayed inside the SESSIONS section box

## ADDED Requirements

### Requirement: Hints section box

Navigation hints SHALL render in their own HINTS section box at the bottom, spanning the full width.

#### Scenario: Hints rendering

- **WHEN** the session list view is displayed
- **THEN** a HINTS section box appears below the session list and sidebar
- **AND** it contains a single line of navigation keybind hints
- **AND** it spans the full available width

### Requirement: Header section box

The title bar SHALL render as its own section box at the top.

#### Scenario: Header rendering

- **WHEN** any view is displayed
- **THEN** the title bar appears in a bordered section box at the top
- **AND** it contains the app name on the left and mode name on the right

### Requirement: Width calculation with sidebar

The session list row width SHALL be calculated as: `totalWidth - AppBorderOverheadX - sidebarWidth - gapWidth`, where `gapWidth` is 1 character (the space between session list and sidebar section boxes) and `sidebarWidth` is derived from the sidebar section box width (label + button grid + borders + padding).

#### Scenario: Row width with sidebar

- **WHEN** the terminal is 80 columns wide, the app border overhead is 4, the sidebar is 20 columns, and the gap is 1
- **THEN** the session list section box content width is 55 columns (80 - 4 - 20 - 1)
