# Repos CLI

## Purpose

The `helm repos` subcommand for managing repository sync state across all project directories.

## Requirements

### REQ-1: Status Command

`helm repos status` SHALL show the sync state of all repos in project_dirs.

#### Scenario: Human output

- **WHEN** `helm repos status` is run
- **THEN** each repo is listed with a state symbol (✓/~/↑/↓/↕/⊘), name, branch, and detail counts

#### Scenario: JSON output

- **WHEN** `helm repos status --json` is run
- **THEN** output is a JSON object with a `repos` array containing name, state, branch, ahead, behind, dirty fields

#### Scenario: No repos

- **WHEN** no repos exist in project_dirs
- **THEN** the message "No repos found in project_dirs." is shown (or empty JSON array with --json)

### REQ-2: Pull Command

`helm repos pull` SHALL fetch all repos and fast-forward-only pull those that are behind.

#### Scenario: Parallel fetch

- **WHEN** `helm repos pull` is run
- **THEN** all repos are fetched in parallel (max 4 concurrent network operations)
- **AND** only repos in "behind" state are pulled (fast-forward only)

#### Scenario: Skip non-pullable

- **WHEN** a repo is clean, ahead, dirty, or diverged
- **THEN** it is skipped during pull

#### Scenario: JSON output

- **WHEN** `helm repos pull --json` is run
- **THEN** output contains pulled, skipped, and failed arrays with a summary object

### REQ-3: Push Command

`helm repos push` SHALL push all repos that are ahead of their remote.

#### Scenario: Push ahead repos

- **WHEN** `helm repos push` is run
- **THEN** repos in "ahead" or "dirty+ahead" state are pushed in parallel (max 4 concurrent)

#### Scenario: No repos to push

- **WHEN** no repos are in "ahead" state
- **THEN** the message "No repos to push (none in 'ahead' state)." is shown

#### Scenario: JSON output

- **WHEN** `helm repos push --json` is run
- **THEN** output contains pushed and failed arrays with a summary object

### REQ-4: Dirty Command

`helm repos dirty` SHALL list repos with uncommitted changes.

#### Scenario: List dirty paths

- **WHEN** `helm repos dirty` is run without flags
- **THEN** the absolute path of each dirty repo is printed, one per line

#### Scenario: Walk mode

- **WHEN** `helm repos dirty --walk` is run
- **THEN** the configured `dirty_walkthrough_command` is executed for each dirty repo
- **AND** `{}` in the command is replaced with the repo path

#### Scenario: No walk command configured

- **WHEN** `--walk` is used but no `dirty_walkthrough_command` is configured
- **THEN** an error is returned referencing the config file

### REQ-5: Add Command

`helm repos add <repo>` SHALL clone a single repository.

#### Scenario: Clone by shorthand

- **WHEN** `helm repos add owner/repo` is run
- **THEN** the repo is cloned to projectDir/owner/repo

#### Scenario: Already cloned

- **WHEN** the target directory already contains a .git directory
- **THEN** the message "✓ <repo> already cloned at <path>" is shown

### REQ-6: Rebuild Command

`helm repos rebuild` SHALL re-run post_clone hooks from ensure_cloned config.

#### Scenario: Rebuild all

- **WHEN** `helm repos rebuild --all` is run
- **THEN** post_clone commands are executed for all ensure_cloned entries that have them

#### Scenario: Rebuild specific

- **WHEN** `helm repos rebuild --repos owner/repo1,owner/repo2` is run
- **THEN** post_clone commands are executed only for the specified repos

#### Scenario: No hook configured

- **WHEN** a targeted repo has no post_clone hook
- **THEN** the message "⊘ <repo>: no post_clone hook" is shown

#### Scenario: Repo not cloned

- **WHEN** a targeted repo is not cloned locally
- **THEN** the message "⊘ <repo>: not cloned" is shown
