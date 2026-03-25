## Context

Helm currently renders as a single vertical column: title bar, prompt, session list, and a 3-line footer (border + notification/state + 2-line hints). All view modes follow this pattern. The footer is overcrowded — action keybinds, navigation hints, and status info compete for space in ~2 lines.

The redesign introduces a section-box layout system and a right sidebar, splitting the UI into 5 distinct bordered sections. This is a rendering-only change — no behavioral changes to navigation, filtering, session management, or keybindings.

### Current layout structure

```
┌─ AppStyle border ──────────────────┐
│ TitleBar                           │
│ Prompt                             │
│ ─── border ───                     │
│ Table Header                       │
│ ··· dotted ···                     │
│ Session rows...                    │
│ ─── border ───                     │
│ Notification                       │
│ State line                         │
│ Hint line 1                        │
│ Hint line 2                        │
└────────────────────────────────────┘
```

### Target layout structure

```
┌────────────────────────────────────────────────────────────┐
│ BLACK ATOM HELM                                       SESS │
└────────────────────────────────────────────────────────────┘

 > _

 SESSIONS ────────────────────┐   ACTIONS ──────────┐
                              │                     │
  # CC ▸ SESS        ACT  GIT│    ┌─────┐ ┌─────┐  │
  * nikbrunner-dots   5s  1fi │    │ NEW │ │ PRJ │  │
 ┃ 0 ⠒ ▼ b-a-i-helm  1m  1f+2│    │ C-n │ │ C-p │  │
 │ 1   ▶ Penny        18m    │    └─────┘ └─────┘  │
 │ ...                        │    ...               │
 ─────────────────────────────┘   ──────────────────┘

                                  STATUS ───────────┐
                                   13 sessions      │
                                   Current: helm    │
                                  ──────────────────┘

 HINTS ──────────────────────────────────────────────┐
  C-j/k Nav · C-h/l Expand · Enter Switch            │
 ────────────────────────────────────────────────────┘
```

## Goals / Non-Goals

**Goals:**

- Introduce a reusable `SectionBox` rendering primitive in `internal/ui/`
- Render 5 section boxes: Header, Sessions, Actions, Status, Hints
- Move action keybinds into color-filled 2-column button grid in sidebar
- Show sidebar in all view modes (sessions, bookmarks, projects, clone, etc.)
- Keep all existing keybindings and navigation behavior unchanged

**Non-Goals:**

- No new keybindings or actions
- No changes to session data model, tmux integration, or git/claude status
- No interactive sidebar (buttons are display-only, not clickable — TUI uses keyboard)
- No responsive sidebar (hidden on narrow terminals) — fixed layout

## Decisions

### D1: SectionBox as a rendering function, not a Bubbletea component

The section box is a pure rendering function: `RenderSectionBox(label string, content string, width int) string`. It takes pre-rendered content and wraps it in borders with a label-on-top-line header.

**Rationale:** Section boxes have no state or interactivity — they're purely visual containers. A Bubbletea Model would add unnecessary complexity. Lipgloss string composition is sufficient.

**Alternative considered:** Bubbletea sub-model with its own Update/View — rejected because sections don't handle input.

### D2: Sidebar width is fixed at render time, not configurable

The sidebar width is calculated from the widest button label + padding + borders. With 3-char labels in a 2-column grid, this is ~15-17 chars. The session list gets `totalWidth - sidebarWidth - gap`.

**Rationale:** Dynamic sidebar sizing adds complexity without benefit — the content is static (7 fixed buttons + 3 status lines). A fixed calculation avoids layout negotiation.

### D3: Compose sidebar in the view function, not as a separate model

Each view function (`viewSessionList`, `viewBookmarks`, etc.) calls a shared `RenderSidebar(actions []Action, status StatusInfo, height int) string` function and uses `lipgloss.JoinHorizontal` to place it next to the content.

**Rationale:** The sidebar content varies by mode (different actions in bookmarks vs sessions), so it must be composed per-view. A shared render function avoids duplication.

### D4: Button rendering uses inverted/background style, not box-drawing

Buttons use Lipgloss background color fill (inverted text) rather than box-drawing characters. The 2-line button is:

- Line 1: 3-char label in ALL CAPS with accent background
- Line 2: Keybind hint in dimmer color with same background

This matches the Stitch reference design where buttons are solid colored blocks.

**Rationale:** Background-fill buttons are visually distinct and match the mockup aesthetic. Box-drawing buttons (┌─┐│└─┘) would consume more characters and look busier.

### D5: Section box borders use solid lines only

All section borders use `─` and `│` (no dotted `···` separators between sections). Dotted separators are only used within a section (e.g., between pinned self-session and regular sessions, between column headers and data rows).

**Rationale:** Solid borders create clear visual hierarchy between sections. Dotted lines remain for in-section sub-separators.

### D6: Kill button uses distinct warning style

The Kill action button uses a red/warning background instead of the standard accent color. This provides a visual safety cue.

### D7: Selected row uses accent background with dark foreground

The selected/highlighted row changes from a barely-visible black background (`tc.Black`) with yellow text to a **muted orange/accent background** with **dark foreground text**. This matches the Stitch reference design and makes the selection immediately visible. All per-column styles that apply `Colors.Bg.Selected` (git status, claude icon, spacers, etc.) automatically pick up the new background. Foreground colors for selected items (`Fg.Selected`, `Fg.SessionNameSelected`) change to dark/black for contrast on the accent background.

## Risks / Trade-offs

- **Width pressure on session names:** The sidebar consumes ~17 chars. At `-w60%` on a 160-col terminal, the session list gets ~79 chars — sufficient. At 80-col terminals, session names may need truncation. → Mitigation: Truncate long names with `..` mid-cut, as current code already handles.

- **Vertical space for sidebar:** At `-h35%` on a 50-row terminal, the visible area is ~17 rows. The sidebar needs ~15 rows (7 two-line buttons + separator + 3 status lines). This barely fits. → Mitigation: The sidebar is not scrollable; if the terminal is too short, the status section gets clipped (acceptable since it's informational).

- **View function complexity:** Every view function must now compose a sidebar. → Mitigation: The `RenderSidebar` function encapsulates all sidebar logic. Each view just calls it with mode-specific actions.

- **Testing visual output:** Section box rendering is hard to unit test pixel-perfectly. → Mitigation: Test the SectionBox function in isolation with known inputs; visual verification via tmux popup.
