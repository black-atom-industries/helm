# Configuration

## Purpose

Config file loading, `helm init`, `helm tmux-bindings`, and environment variable overrides.

## Requirements

### REQ-1: Config File Loading

helm SHALL load configuration from `~/.config/helm/config.yml` with sensible defaults.

#### Scenario: Load with defaults

- **WHEN** no config file exists
- **THEN** default values are used: appearance=dark, project_depth=2, project_dirs=[~/repos], lazygit_popup=90%x90%

#### Scenario: Load from file

- **WHEN** a config file exists at ~/.config/helm/config.yml
- **THEN** values from the file override defaults

#### Scenario: Path expansion

- **WHEN** config values contain `~`
- **THEN** it is expanded to the user's home directory

### REQ-2: Environment Variable Overrides

Environment variables SHALL take priority over config file values.

#### Scenario: Layout override

- **WHEN** `TMUX_LAYOUT` is set
- **THEN** it overrides the `layout` config value

#### Scenario: Layout dir override

- **WHEN** `TMUX_LAYOUTS_DIR` is set
- **THEN** it overrides the `layout_dir` config value

#### Scenario: Claude status override

- **WHEN** `TMUX_SESSION_PICKER_CLAUDE_STATUS=1` is set
- **THEN** claude_status_enabled is forced to true

#### Scenario: Git status override

- **WHEN** `TMUX_SESSION_PICKER_GIT_STATUS=1` is set
- **THEN** git_status_enabled is forced to true

### REQ-3: Init Command

`helm init` SHALL create a config file with commented defaults.

#### Scenario: Create config

- **WHEN** `helm init` is run and no config exists
- **THEN** a config file is created at ~/.config/helm/config.yml with all options commented out
- **AND** the config directory is created if needed

#### Scenario: Config already exists

- **WHEN** `helm init` is run and a config already exists
- **THEN** an error "config file already exists at <path>" is returned

### REQ-4: Tmux Bindings Export

`helm tmux-bindings` SHALL output tmux bind commands for bookmark quickstart.

#### Scenario: Generate bindings

- **WHEN** `helm tmux-bindings` is run
- **THEN** 10 tmux bind commands are printed (Alt+Shift+0 through Alt+Shift+9)
- **AND** each binds to `helm bookmark <N>`

### REQ-5: Tmux Requirement

The TUI SHALL only run inside a tmux session.

#### Scenario: Outside tmux

- **WHEN** helm is run without `$TMUX` set
- **THEN** the error "helm must be run from within tmux" is printed and helm exits with code 1

#### Scenario: HOME required

- **WHEN** `$HOME` is not set
- **THEN** the error "HOME environment variable not set" is printed and helm exits with code 1

### REQ-6: Project Depth

The `project_depth` config SHALL control how directory paths are converted to session names.

#### Scenario: Depth 2

- **WHEN** project_depth is 2 and the path is ~/repos/owner/repo
- **THEN** the session name is derived from "owner/repo" (sanitized to "owner-repo")

#### Scenario: Minimum depth

- **WHEN** project_depth is set to 0 or negative
- **THEN** it defaults to 2
