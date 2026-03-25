# Sidebar

## ADDED Requirements

### Requirement: Sidebar layout position

The sidebar SHALL render to the right of the main content area, separated by a gap.

#### Scenario: Side-by-side rendering

- **WHEN** the session list view is rendered
- **THEN** the session list section box appears on the left
- **AND** the ACTIONS and STATUS section boxes appear on the right
- **AND** there is a gap between the left and right sections

### Requirement: ACTIONS section with button grid

The ACTIONS section box SHALL contain a 2-column grid of action buttons.

#### Scenario: Button grid layout

- **WHEN** the ACTIONS section is rendered
- **THEN** 7 action buttons are displayed in a 2-column grid (4 rows, last row has 1 button)
- **AND** each button has 2 lines: a 3-char ALL CAPS label and a keybind hint below

### Requirement: Button rendering

Each action button SHALL be rendered with a color-filled background containing the label and keybind.

#### Scenario: Standard button

- **WHEN** a button for "New" with keybind "C-n" is rendered
- **THEN** line 1 shows "NEW" in accent background color
- **AND** line 2 shows "C-n" in a dimmer color with the same background

#### Scenario: Kill button warning style

- **WHEN** the Kill button is rendered
- **THEN** it uses a red/warning background color instead of the accent color
- **AND** the label "KIL" and keybind "C-x" follow the same 2-line format

### Requirement: Button labels

The 7 action buttons SHALL use fixed 3-character ALL CAPS labels: NEW (C-n), PRJ (C-p), BKM (C-b), KIL (C-x), RMT (C-r), DWL (C-d), GIT (C-g).

#### Scenario: All buttons present

- **WHEN** the sidebar is rendered in session list mode
- **THEN** all 7 buttons are visible with their respective labels and keybinds

### Requirement: STATUS section

The STATUS section box SHALL display session count, current session name, and the last activity time of the current (self) session.

#### Scenario: Status display

- **WHEN** there are 13 sessions and the current session is "helm" with self-session last activity 1 minute ago
- **THEN** the STATUS section shows "13 sessions", "Current: helm", and "Active: 1m ago"
- **AND** "Active" refers to the self-session's last activity time

### Requirement: Sidebar visible in all modes

The sidebar SHALL be rendered in all TUI modes, not just the session list.

#### Scenario: Bookmarks mode

- **WHEN** the user switches to bookmarks mode
- **THEN** the sidebar is still visible with appropriate actions for that mode

#### Scenario: Projects mode

- **WHEN** the user enters the directory picker
- **THEN** the sidebar is still visible with appropriate actions for that mode

### Requirement: Mode-specific actions

The action buttons displayed in the sidebar SHALL change based on the current mode. Each Action is defined by a label (3-char ALL CAPS string), a keybind (string like "C-n"), and a style variant (normal or warning).

#### Scenario: Session mode actions

- **WHEN** the mode is ModeNormal (session list)
- **THEN** the ACTIONS section shows: NEW (C-n), PRJ (C-p), BKM (C-b), KIL (C-x, warning), RMT (C-r), DWL (C-d), GIT (C-g)

#### Scenario: Bookmarks mode actions

- **WHEN** the mode is ModeBookmarks
- **THEN** the ACTIONS section shows: OPN (Enter), ADD (C-a), MV↑ (C-p), MV↓ (C-n), RMV (C-x, warning), BCK (Esc)

#### Scenario: Projects mode actions

- **WHEN** the mode is ModePickDirectory
- **THEN** the ACTIONS section shows: SEL (Enter), BKM (C-a), RMV (C-x, warning), BCK (Esc)

#### Scenario: Clone mode actions

- **WHEN** the mode is ModeCloneRepo
- **THEN** the ACTIONS section shows: CLN (Enter), BCK (Esc)
