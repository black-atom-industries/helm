# Claude Status

## Purpose

Claude Code activity indicator per session, driven by hook-based status files.

## Requirements

### REQ-1: Status File Format

Status files SHALL be stored at `<cache_dir>/<session-name>.status` in the format `state:timestamp`.

#### Scenario: Parse status file

- **WHEN** a status file contains "working:1711234567"
- **THEN** the state is "working" and the timestamp is the corresponding Unix time

#### Scenario: Missing status file

- **WHEN** no status file exists for a session
- **THEN** no Claude status indicator is shown for that session

#### Scenario: Invalid format

- **WHEN** a status file has invalid content (wrong format, bad timestamp)
- **THEN** it is treated as no status

### REQ-2: Status States

The status indicator SHALL show different symbols based on the Claude Code state and elapsed time.

#### Scenario: Working state

- **WHEN** the state is "working" and not stale
- **THEN** an animated spinner (⠤⠆⠒⠰) is shown, cycling every 300ms

#### Scenario: Waiting state (0-5 minutes)

- **WHEN** the state is "waiting" and less than 5 minutes have elapsed
- **THEN** a "?" indicator is shown

#### Scenario: Waiting state (5-15 minutes)

- **WHEN** the state is "waiting" and 5-15 minutes have elapsed
- **THEN** a "!" indicator is shown

#### Scenario: Waiting state (>15 minutes)

- **WHEN** the state is "waiting" and more than 15 minutes have elapsed
- **THEN** a "Z" indicator is shown

### REQ-3: Staleness

Status files SHALL be considered stale after a threshold period.

#### Scenario: Working state staleness

- **WHEN** the "working" state hasn't been updated in over 2 minutes
- **THEN** the status is treated as empty (no indicator shown)

#### Scenario: Waiting state staleness

- **WHEN** the "waiting" state hasn't been updated in over 30 minutes
- **THEN** the status is treated as empty

### REQ-4: Configuration

Claude status integration SHALL be opt-in via configuration.

#### Scenario: Disabled by default

- **WHEN** `claude_status_enabled` is not set or false
- **THEN** no Claude status files are read and no indicators are shown

#### Scenario: Enable via config

- **WHEN** `claude_status_enabled: true` is set in config.yml
- **THEN** status files are read for all sessions on load

#### Scenario: Enable via environment

- **WHEN** the environment variable `TMUX_SESSION_PICKER_CLAUDE_STATUS=1` is set
- **THEN** Claude status is enabled regardless of config file setting

### REQ-5: Hook Integration

The status file SHALL be written by the Claude Code hook script (`hooks/helm-hook.sh`).

#### Scenario: Hook writes status

- **WHEN** Claude Code triggers the hook
- **THEN** the hook writes the current state and Unix timestamp to the session's status file
