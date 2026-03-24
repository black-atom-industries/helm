# Directory Picker

## Purpose

Browse project directories to create sessions, with filtering and folder management.

## Requirements

### REQ-1: Open Directory Picker

Pressing Ctrl+p SHALL open the directory picker showing all subdirectories from configured project_dirs.

#### Scenario: Show project directories

- **WHEN** the user presses Ctrl+p
- **THEN** directories are scanned at the configured project_depth from all project_dirs
- **AND** results are shown as a filterable scrollable list
- **AND** any active filter from session mode is carried over

#### Scenario: No directories found

- **WHEN** no directories exist in the configured project_dirs
- **THEN** the message "No directories found" is shown

### REQ-2: Directory Scanning

Directories SHALL be scanned recursively to the configured `project_depth`.

#### Scenario: Depth 2 scanning

- **WHEN** project_depth is 2 and project_dirs contains ~/repos
- **THEN** directories at ~/repos/owner/repo level are listed

#### Scenario: Hidden directories excluded

- **WHEN** scanning encounters .git, .hg, .svn, .DS_Store, .Trash, .cache, .local, or .config directories
- **THEN** they are excluded from results

### REQ-3: Filter Directories

Typing characters SHALL filter the directory list by basename.

#### Scenario: Filter by name

- **WHEN** the user types "helm"
- **THEN** only directories whose basename contains "helm" (case-insensitive) are shown

### REQ-4: Select Directory

Pressing Enter SHALL create a session at the selected directory.

#### Scenario: Create from selected directory

- **WHEN** the user presses Enter on a directory
- **THEN** a tmux session is created with a name derived from the path
- **AND** layout is applied and helm switches to the session

#### Scenario: Directory with existing session

- **WHEN** the selected directory already has a matching tmux session
- **THEN** helm switches to the existing session

### REQ-5: Remove Folder

Pressing Ctrl+x SHALL initiate folder removal with confirmation.

#### Scenario: Confirm folder removal

- **WHEN** the user presses Ctrl+x on a directory
- **THEN** a confirmation message "Remove <name> from disk?" is shown

#### Scenario: Execute removal

- **WHEN** the user confirms with a second Ctrl+x
- **THEN** the folder is deleted from disk with os.RemoveAll
- **AND** any associated tmux session is killed
- **AND** the directory list rescans

#### Scenario: Cancel removal

- **WHEN** the user presses Escape during confirmation
- **THEN** the removal is cancelled
