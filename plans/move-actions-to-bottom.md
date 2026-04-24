# Plan: Move Action Buttons from Sidebar to Bottom Bar

## Context

The **ACTIONS sidebar** currently renders as a tall vertical column on the right side of every view. It consumes ~17px of width and visually dominates the screen. The user wants buttons moved to the **bottom** in **two left-aligned rows**, freeing the full width for the session/bookmark/project list.

**Design decision from visual companion:** Two rows, left-aligned, compact (no border box).

## Scope Assessment

| Aspect | Assessment |
|--------|------------|
| Scope | Medium — multi-file layout refactor, no data model changes |
| Risk | Low — purely presentational; all view functions use a single shared helper |
| Breaking changes | None for users; internal API changes are localized |

## Approach

Replace the right-side sidebar with a bottom action bar rendered by the existing shared helper `renderWithSidebar`. The helper is already called by every view (sessions, bookmarks, projects, clones, create, confirm-kill). Changing only the helper and the `*MaxVisibleItems()` calculations updates all views at once.

**Visual result:**
```
┌─ helm ───────────────────────── SESSIONS ──────────────┐
│ > filter text                                          │
│ ───────────────────────────────────────────────────────│
│ ▼ SESS   Git    Claude   Time                          │
│ ·······················································│
│ ● self-session                              2m         │
│ ·······················································│
│ ┃ 0  black-atom-industries/helm  +2 -1      2m         │
│ │ 1  dotfiles                    clean      1h         │
│ │ 2  website                     clean      3h         │
│ │ 3  api-server                  +12 -4     5h         │
│                                                        │
│ (more padding when list is short)                      │
│                                                        │
│ SWITCH   EXPAND   BOOKMARKS  PROJECTS  DOWNLOAD        │
│ Enter    C-h/l    C-b        C-p       C-d             │
│                                                        │
│ NEW      LAZYGIT  REMOTE     KILL                      │
│ C-n      C-g      C-r        C-x                       │
│ ───────────────────────────────────────────────────────│
│ 4 sessions                                             │
│ C-j/k ↕ Nav · Type filter · Esc Back                   │
└────────────────────────────────────────────────────────┘
```

## Files to Modify

| File | What changes |
|------|-------------|
| `internal/ui/layout.go` | Add `ActionBarHeight = 5` constant |
| `internal/ui/sidebar.go` | Add `RenderButtonLabel`, `RenderButtonKeybind`, `RenderButtonBar`, `renderButtonRow`; refactor `RenderButton` to reuse new helpers |
| `internal/model/model.go` | `sidebarWidth()` → return `0`; `sessionListWidth()` → return `contentWidth()`; `rowWidth()` → use full width; `renderWithSidebar()` → render bottom bar instead of right sidebar, adjust padding target |
| `internal/model/session.go` | `sessionMaxVisibleItems()` → add `ActionBarHeight` to overhead |
| `internal/model/directory.go` | `projectMaxVisibleItems()` → add `ActionBarHeight` to overhead |
| `internal/model/bookmarks.go` | Inline visible-items calculation → add `ActionBarHeight` to overhead |
| `internal/model/clone.go` | `cloneMaxVisibleItems()` → add `ActionBarHeight` to overhead |
| `internal/model/layout_test.go` | Update expected values for `sessionMaxVisibleItems` and `projectMaxVisibleItems` tests |
| `internal/ui/styles_test.go` (or new `sidebar_test.go`) | Add tests for `RenderButtonBar` |

## Reuse

- `RenderButton` logic is preserved but refactored to call new `RenderButtonLabel` / `RenderButtonKeybind` helpers
- `ButtonStyle`, `ButtonWarningStyle`, `ButtonKeybindStyle`, `ButtonWarnKbStyle` are reused as-is
- `centerText` is reused for button text centering
- `ButtonInnerWidth = 12` is preserved
- `RenderSimpleFooter` is reused as-is
- `AppStyle.Height()` continues to enforce exact terminal height

## Detailed Implementation Steps

### Step 1: Add layout constant

In `internal/ui/layout.go`, add:
```go
ActionBarHeight = 5 // 2 rows × 2 lines each + 1 gap line
```

### Step 2: Build horizontal button rendering primitives

In `internal/ui/sidebar.go`:

1. **Add `RenderButtonLabel`** — returns the styled label line of a single button (extracted from existing `RenderButton`).
2. **Add `RenderButtonKeybind`** — returns the styled keybind line of a single button.
3. **Refactor `RenderButton`** to call the two new helpers joined by `\n`.
4. **Add `renderButtonRow(actions, width)`** (unexported) — renders one row of buttons side-by-side with 1-space gaps, left-aligned, padded/truncated to exact `width`.
   - Builds label row by joining `RenderButtonLabel` of each action with `" "`
   - Builds keybind row by joining `RenderButtonKeybind` of each action with `" "`
   - Pads or truncates each row to `width` using `lipgloss.Width` and `lipgloss.NewStyle().MaxWidth()`
   - Returns `labelRow + "\n" + kbRow`
   - If `len(actions) == 0`, returns two space-filled lines
5. **Add `RenderButtonBar(actions, width)`** (exported) — renders the full action bar:
   - Splits actions: row1 = first `min(5, len(actions))`, row2 = remainder
   - Calls `renderButtonRow` for each row
   - Joins rows with `"\n\n"` (1 blank line gap)
   - Always returns exactly 5 lines (4 `\n` characters)

### Step 3: Eliminate right-sidebar width deductions

In `internal/model/model.go`:

1. **`sidebarWidth()`** → return `0`
2. **`sessionListWidth()`** → return `m.contentWidth()`
3. **`rowWidth()`** → return `m.contentWidth() - ui.ScrollbarColumnWidth`

These are the *only* width changes needed. All view functions already call `sessionListWidth()` / `rowWidth()`, so the list automatically expands to full width.

### Step 4: Rewrite `renderWithSidebar` as bottom-bar composer

In `internal/model/model.go`, replace the body of `renderWithSidebar`:

1. Write `header` to builder (unchanged)
2. Write `listContent` to builder (unchanged — now full width)
3. Count lines in builder via `strings.Count(content, "\n")`
4. **New padding target:** `contentHeight - ui.ActionBarHeight - 3` (was `contentHeight - 3`)
5. Pad with newlines to reach target
6. **New:** Write `ui.RenderButtonBar(actions, m.contentWidth())`
7. Write `ui.RenderSimpleFooter(notification, hints, isError, m.width)` (unchanged)
8. Return `ui.AppStyle.Height(m.contentHeight()).Render(b.String())`

Remove the entire line-by-line list+sidebar joining block, the `SidebarGap` usage, and the `ui.RenderSidebar` call.

### Step 5: Update visible-item calculations for all views

Each `*MaxVisibleItems()` function currently accounts for header + footer overhead. Add `ui.ActionBarHeight` to the overhead in all of them:

| Function | File | Current overhead | New overhead |
|----------|------|-----------------|--------------|
| `sessionMaxVisibleItems` | `session.go` | `6 + tableHeader + selfSession` | `+ ActionBarHeight` |
| `projectMaxVisibleItems` | `directory.go` | `6` | `+ ActionBarHeight` |
| `cloneMaxVisibleItems` | `clone.go` | `6` | `+ ActionBarHeight` |
| Bookmarks inline calc | `bookmarks.go` | `8` | `+ ActionBarHeight` |

### Step 6: Update layout tests

In `internal/model/layout_test.go`:

- `TestSessionMaxVisibleItems`: adjust `want` values to account for `ActionBarHeight = 5`
- `TestProjectMaxVisibleItems`: same
- Update comments to reflect new overhead math
- `TestViewLinesHaveConsistentWidth` and `TestViewLineCountNotShorterThanHeight` should pass as-is (still fill exact height/width)

### Step 7: Add unit tests for new UI helpers

Add tests for `RenderButtonBar` and `renderButtonRow`:

- **Empty actions** → returns 5 blank lines
- **Single action** → 1 button in row 1, row 2 empty, both padded to width
- **5 actions** → fills row 1, row 2 empty
- **6 actions** → 3+3 split, both rows padded to width
- **9 actions** → 5+4 split (SessionActions), verifies KILL lands in row 2
- **Width smaller than buttons** → verifies truncation via `MaxWidth`
- **Warning button styling** → verifies row 2 danger button uses `ButtonWarningStyle`

### Step 8: Build and manual verification

1. `make test` — all unit tests pass
2. `go build -o helm ./cmd/helm/` — compiles cleanly
3. Run `./helm` inside tmux and verify:
   - Session list uses full width (no sidebar)
   - Bottom bar shows 2 rows of buttons, left-aligned
   - Action bar appears in **all** modes: sessions, bookmarks, projects, downloads, create, confirm-kill
   - Footer sits directly below action bar
   - List scrolling still works correctly
   - No layout shift or width inconsistency

## Edge Cases & Decisions

| Case | Decision |
|------|----------|
| Very narrow terminal (< 40 cols) | Buttons may truncate; acceptable — old code hid sidebar entirely, new code always shows bar |
| Modes with < 5 actions (e.g. Clone: 1 action) | Row 1 shows the action, row 2 is empty (padded to width). Keeps layout stable across mode switches. |
| Modes with exactly 5 actions | Row 1 fills, row 2 empty. Clean. |
| Modes with > 10 actions | Not currently possible (max is 9). If added later, row 2 would show actions 6–10 and truncate. |
| Padding between list and action bar | Natural via `renderWithSidebar` padding logic. When list is short, empty lines push bar down. When list is long, bar sits immediately below last item. |
| Background gaps between buttons | 1-space gap has terminal default background. Consistent with current sidebar blank lines between buttons. |
| "ACTIONS" label | Omitted for compactness. The button visual style is self-explanatory. |

## Verification Checklist

- [ ] `make test` passes (including updated layout tests)
- [ ] Binary builds successfully
- [ ] Manual test: `./helm` in tmux shows bottom action bar in all modes
- [ ] Manual test: narrow terminal (60 cols) renders without panic
- [ ] Manual test: expand a session (windows visible) — action bar still shows correctly
- [ ] Manual test: filter mode — action bar still shows correctly
