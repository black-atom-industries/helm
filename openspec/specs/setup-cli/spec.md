# Setup CLI

## Purpose

The `helm setup` subcommand for bulk-cloning repositories from the `ensure_cloned` config.

## Requirements

### REQ-1: Bulk Clone

`helm setup` SHALL clone all repositories listed in the `ensure_cloned` config section.

#### Scenario: Clone uncloned repos

- **WHEN** `helm setup` is run
- **THEN** already-cloned repos are skipped
- **AND** uncloned repos are cloned in parallel (max 4 concurrent)
- **AND** a summary shows cloned, skipped, and failed counts

#### Scenario: No ensure_cloned entries

- **WHEN** the config has no ensure_cloned entries
- **THEN** the message "No ensure_cloned entries in config." is shown with guidance to configure them

### REQ-2: Wildcard Expansion

Entries with `org/*` patterns SHALL expand to all repos in that organization.

#### Scenario: Expand SSH wildcard

- **WHEN** an ensure_cloned entry is `git@github.com:org/*`
- **THEN** all repos in the `org` organization are fetched via `gh repo list`
- **AND** each is cloned individually

#### Scenario: Expand HTTPS wildcard

- **WHEN** an ensure_cloned entry is `https://github.com/org/*`
- **THEN** the same expansion occurs using the HTTPS base URL

### REQ-3: Post-Clone Hooks

Entries with `post_clone` commands SHALL have them executed after successful cloning.

#### Scenario: Run post_clone

- **WHEN** a repo with `post_clone: "make install"` is successfully cloned
- **THEN** `make install` is run in the cloned repo's directory

#### Scenario: Post-clone failure

- **WHEN** a post_clone command fails
- **THEN** the failure is reported but does not prevent other repos from being processed

### REQ-4: Clone Target Selection

When multiple project_dirs are configured, the user SHALL be prompted to select one.

#### Scenario: Single project dir

- **WHEN** only one project_dirs entry exists
- **THEN** it is used as the clone target automatically

#### Scenario: Multiple project dirs

- **WHEN** multiple project_dirs entries exist
- **THEN** the user is prompted to select which directory to clone into

### REQ-5: Entry Formats

`ensure_cloned` entries SHALL support both string and object formats.

#### Scenario: String format

- **WHEN** an entry is a plain string like `git@github.com:user/repo.git`
- **THEN** it is treated as a URL with no post_clone hook

#### Scenario: Object format

- **WHEN** an entry is an object with `url` and `post_clone` fields
- **THEN** the URL is cloned and the post_clone command is executed after
