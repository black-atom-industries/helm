## Why

The current helm footer is overcrowded — action keybinds, status info, and navigation hints are all crammed into 2-3 lines of text. This makes the UI feel dense and hard to scan. The Stitch mockups explored a landscape layout with a right sidebar, which separates concerns into distinct visual sections. This redesign introduces a section-box layout system to give each UI area its own bordered container with clear labels.

## What Changes

- Introduce a **section box** rendering primitive — a bordered container with a label on the top border line (fieldset-legend style), used to wrap all major UI areas
- Add a **right sidebar** containing two section boxes: ACTIONS (color-filled 2-column button grid with 3-char labels and keybind hints) and STATUS (session count, current session, last activity)
- Restructure the **main view layout** into 5 section boxes: Header, Sessions (left), Actions (right-top), Status (right-bottom), Hints (bottom full-width)
- Replace the current **2-line text footer hints** with a single-line hints section box at the bottom
- Move **action keybinds** from inline footer text to visual buttons in the sidebar (2-line per button: ALL CAPS label + dimmer keybind)
- **Sidebar is visible in all modes** (sessions, bookmarks, projects, clone) — not just the main session list
- Bump default **popup size** to `-w60%` to accommodate the sidebar

## Capabilities

### New Capabilities

- `section-box`: Reusable bordered container rendering with label-on-border headers, used to compose the layout
- `sidebar`: Right sidebar layout system with action buttons and status display, visible in all modes

### Modified Capabilities

- `session-navigation`: View layout changes from single-panel to split-panel with sidebar; session list renders inside a section box; footer restructured into separate hints section box

## Impact

- **UI package** (`internal/ui/`): New section box renderer, sidebar renderer, button grid renderer; modified styles for section borders, button fills, label-on-border headers
- **Model** (`internal/model/`): All view functions (`viewSessionList`, `viewBookmarks`, `viewPickDirectory`, `viewCloneRepo`, etc.) need layout restructuring to include the sidebar
- **Width calculations**: Row widths, content heights, and scroll calculations must account for the sidebar consuming ~20 chars of horizontal space
- **Config**: Default popup width recommendation changes from 50% to 60%
- **No breaking changes** to keybindings, session data, or external interfaces
