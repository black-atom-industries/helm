# Clone Repos

## Purpose

Clone repositories from GitHub into project directories via the TUI.

## Requirements

### REQ-1: Clone Choice Menu

Pressing Ctrl+d SHALL open a choice menu with two options: "Enter URL" and "My repos".

#### Scenario: Enter clone mode

- **WHEN** the user presses Ctrl+d
- **THEN** a two-option menu is displayed: "Enter URL" and "My repos"
- **AND** the user can navigate between them with Up/Down

#### Scenario: No project dirs configured

- **WHEN** no project_dirs are configured and the user presses Ctrl+d
- **THEN** an error message "No project_dirs configured" is shown

### REQ-2: Clone by URL

Selecting "Enter URL" SHALL present a text input for an arbitrary repository URL.

#### Scenario: Owner/repo shorthand

- **WHEN** the user enters "black-atom-industries/helm"
- **THEN** the repo is cloned to the first project directory under that owner/repo path

#### Scenario: Full URL

- **WHEN** the user enters a full git SSH or HTTPS URL
- **THEN** the owner/repo is extracted and the repo is cloned to the project directory

#### Scenario: Invalid URL

- **WHEN** the user enters an unresolvable URL
- **THEN** an error message is displayed and the user can try again

### REQ-3: Clone from My Repos

Selecting "My repos" SHALL fetch available repositories from GitHub and show them as a filterable list.

#### Scenario: Fetch available repos

- **WHEN** "My repos" is selected
- **THEN** repos are fetched via `gh` CLI
- **AND** already-cloned repos are filtered out
- **AND** the remaining repos are shown in a filterable list

#### Scenario: All repos cloned

- **WHEN** all fetched repos are already cloned
- **THEN** the message "All repositories are already cloned!" is shown

### REQ-4: Clone Execution

Selecting a repo (from either URL or list) SHALL clone it and create a tmux session.

#### Scenario: Successful clone

- **WHEN** a repo is selected for cloning
- **THEN** the repo is cloned to projectDir/owner/repo
- **AND** a tmux session is created at that path
- **AND** a success screen shows the repo name and session name

#### Scenario: Success confirmation

- **WHEN** the clone succeeds and the success screen is shown
- **THEN** pressing Enter applies the layout and switches to the new session
- **AND** pressing Escape returns to the session list without switching

#### Scenario: Clone failure

- **WHEN** cloning fails
- **THEN** an error message is shown
- **AND** the user can press Escape to go back

### REQ-5: Filter in Clone List

Typing characters in clone repo mode SHALL filter the repository list.

#### Scenario: Filter repos

- **WHEN** the user types characters in clone repo mode
- **THEN** the list is filtered by case-insensitive substring match on repo name
