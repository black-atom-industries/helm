# Section Box

## ADDED Requirements

### Requirement: Label-on-border rendering

The section box SHALL render a bordered container where the label text sits on the top border line, with the horizontal rule continuing after the label to the right edge.

#### Scenario: Basic section box

- **WHEN** `RenderSectionBox("ACTIONS", content, 20)` is called
- **THEN** the output has the label "ACTIONS" followed by `─` characters to fill the width, ending with `┐`
- **AND** the content is indented inside with `│` on the right edge
- **AND** the bottom is closed with `─` characters ending with `┘`

### Requirement: Content padding

The section box SHALL include an empty line after the top border and before the bottom border to provide visual breathing room.

#### Scenario: Gap lines

- **WHEN** content "hello" is wrapped in a section box
- **THEN** there is an empty line between the top border and the content
- **AND** there is an empty line between the content and the bottom border

### Requirement: Width consistency

The section box SHALL render at a fixed width, padding or truncating content lines as needed.

#### Scenario: Content shorter than width

- **WHEN** a content line is shorter than the box width
- **THEN** the line is right-padded with spaces to maintain the right border alignment

### Requirement: No left border on session list section

The session list section box SHALL omit the left border to allow the scrollbar column to render in that position.

#### Scenario: Session list scrollbar

- **WHEN** the session list is rendered inside a section box
- **THEN** the left edge has no border character
- **AND** the scrollbar renders in the leftmost column position
