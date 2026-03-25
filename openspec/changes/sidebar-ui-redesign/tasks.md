## 1. Section Box Primitive

- [x] 1.1 Add color tokens to `Colors` struct in `internal/ui/colors.go`: `Fg.SectionLabel`, `Bg.ButtonAccent`, `Bg.ButtonWarning`, `Fg.ButtonLabel`, `Fg.ButtonKeybind` for both dark and light palettes. Change `Bg.Selected` from `tc.Black` to a muted accent/orange background, and update `Fg.Selected` and `Fg.SessionNameSelected` to a dark foreground (e.g. `tc.Black`) for contrast on the new accent background. Light mode needs the inverse treatment. Then add corresponding styles `SectionLabelStyle`, `SectionBorderStyle`, `ButtonStyle`, `ButtonKeybindStyle`, `ButtonWarningStyle` to `internal/ui/styles.go` referencing these tokens.
- [x] 1.2 Create `internal/ui/section.go` with `RenderSectionBox(label string, content string, width int, opts SectionBoxOpts) string` — renders label-on-border header (`LABEL ────┐`), right edge `│`, bottom border (`────┘`), with gap lines (empty line after top border, empty line before bottom border). `SectionBoxOpts` has `OmitLeftBorder bool` for the session list (where scrollbar occupies the left edge).
- [x] 1.3 Test section box rendering manually in tmux popup — verify label sits on top border line, right border aligns, gap lines appear

## 2. Action Button and Sidebar Renderer

- [x] 2.1 Define `Action` type in `internal/ui/sidebar.go`: `Label string` (3-char ALL CAPS), `Keybind string`, `Warning bool`
- [x] 2.2 Implement `RenderButton(action Action, width int) string` — 2-line output: line 1 is label with accent background fill (`ButtonStyle`), line 2 is keybind in dimmer color with same background (`ButtonKeybindStyle`). Warning variant uses `ButtonWarningStyle` background.
- [x] 2.3 Implement `RenderButtonGrid(actions []Action, colWidth int) string` — arranges buttons in 2-column grid using `lipgloss.JoinHorizontal` for pairs, `lipgloss.JoinVertical` for rows
- [x] 2.4 Define `StatusInfo` type: `SessionCount int`, `CurrentSession string`, `LastActivity time.Time`. Note: `LastActivity` is the self-session's `tmux.Session.LastActivity` — the time of the pinned current session, not the selected session.
- [x] 2.5 Implement `RenderSidebar(actions []Action, status StatusInfo, height int) string` — composes ACTIONS section box (button grid) and STATUS section box (3 info lines) vertically with gap between them
- [ ] 2.6 Test button and sidebar rendering in isolation — verify color fill, 2-column layout, warning style on Kill

## 3. Mode-Specific Action Sets

- [x] 3.1 Define action set vars in `internal/ui/sidebar.go`: `SessionActions`, `BookmarkActions`, `ProjectActions`, `CloneActions` with the exact labels and keybinds from the sidebar spec (session: NEW/PRJ/BKM/KIL/RMT/DWL/GIT; bookmarks: OPN/ADD/MV↑/MV↓/RMV/BCK; projects: SEL/BKM/RMV/BCK; clone: CLN/BCK)
- [x] 3.2 No `ActionsForMode` helper needed — each view function in `internal/model/` already knows its mode and passes the correct `ui.*Actions` slice directly to `RenderSidebar`. This avoids a circular import (`ui` cannot import `model` for the `Mode` type).

## 4. Layout Restructure — Session List View

- [ ] 4.1 Add layout constants in `internal/ui/layout.go`: `SidebarWidth` (calculated from button grid + section box borders), `SidebarGap = 1`. Update `FooterOverhead` from 5 to 3 (border + notification + single-line hints). Update `BaseOverhead` and `WithTableHeaderOverhead` accordingly.
- [ ] 4.2 Add `sidebarWidth()` and `sessionListWidth()` methods to Model that compute the split: `sessionListWidth = contentWidth - SidebarWidth - SidebarGap`
- [ ] 4.3 Refactor `viewSessionList()`: render session content into a string, wrap in SESSIONS section box, render sidebar via `ui.RenderSidebar(ui.SessionActions, statusInfo, height)`, join horizontally with `lipgloss.JoinHorizontal(lipgloss.Top, sessionsBox, " ", sidebarStr)`, render header as section box on top, render hints as section box at bottom
- [ ] 4.4 Update `rowWidth()` to subtract sidebar width and gap
- [ ] 4.5 Update `sessionMaxVisibleItems()` to account for new overhead: section box top border (1) + top gap (1) + bottom gap (1) + bottom border (1) = 4 lines of section box overhead replacing the old header/footer overhead model
- [ ] 4.6 Replace `RenderTitleBar` call with header section box
- [ ] 4.7 Replace `RenderFooter` call with notification line + hints section box
- [ ] 4.8 Visual test: `tmux display-popup -w60% -h35% -B -E "./helm"` — verify full layout with sidebar, section boxes, buttons, status

## 5. Layout Restructure — All Other Views

- [ ] 5.1 Refactor `viewBookmarks()` (bookmarks.go:248) to use section box layout with sidebar (`ui.BookmarkActions`)
- [ ] 5.2 Refactor `viewPickDirectory()` (directory.go:117) to use section box layout with sidebar (`ui.ProjectActions`)
- [ ] 5.3 Refactor `viewCloneRepo()` (clone.go:247) to use section box layout with sidebar (`ui.CloneActions`)
- [ ] 5.4 Refactor `viewCloneChoice()` (clone.go:183) to use section box layout with sidebar
- [ ] 5.5 Refactor `viewCloneURL()` (clone.go:309) to use section box layout with sidebar
- [ ] 5.6 Refactor `viewCreatePath()` (session.go:947) to use section box layout with sidebar
- [ ] 5.7 Visual test: verify each mode renders correctly with sidebar in tmux popup

## 6. Cleanup and Polish

- [ ] 6.1 Remove unused functions no longer called: old 2-line `HelpNormal`/`HelpFiltering` variants, `RenderFooter` if fully replaced, `RenderTitleBar` if fully replaced
- [ ] 6.2 Update CLAUDE.md testing instructions to use `-w60%` popup size
- [ ] 6.3 Verify light mode rendering — section boxes and button fills should respect `InitColors` appearance switching
- [ ] 6.4 Final visual test across all modes with both dark and light terminals
