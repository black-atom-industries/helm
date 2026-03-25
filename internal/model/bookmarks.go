package model

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/git"
	"github.com/black-atom-industries/helm/internal/tmux"
	"github.com/black-atom-industries/helm/internal/ui"
)

func (m *Model) handleBookmarksMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Cancel):
		// If filter is active, clear it first
		if m.bookmarkList.Filter() != "" {
			m.bookmarkList.SetFilter("")
			return m, nil
		}
		// Otherwise return to normal mode
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, keys.Up):
		m.bookmarkList.MoveCursor(-1)

	case key.Matches(msg, keys.Down):
		m.bookmarkList.MoveCursor(1)

	case key.Matches(msg, keys.Expand):
		// Expand bookmark if it has a session
		if selected, ok := m.bookmarkList.SelectedItem(); ok {
			sessionName := m.extractSessionName(selected.Path)
			if session := m.findSessionByName(sessionName); session != nil {
				m.bookmarkExpanded[selected.Path] = true
			}
		}

	case key.Matches(msg, keys.Collapse):
		// Collapse bookmark
		if selected, ok := m.bookmarkList.SelectedItem(); ok {
			m.bookmarkExpanded[selected.Path] = false
		}

	case key.Matches(msg, keys.Select):
		if selected, ok := m.bookmarkList.SelectedItem(); ok {
			return m.openBookmark(selected)
		}

	case key.Matches(msg, keys.PickDirectory):
		// C-p moves bookmark up in bookmarks mode
		return m.moveBookmark(-1)

	case key.Matches(msg, keys.Create):
		// C-n moves bookmark down in bookmarks mode
		return m.moveBookmark(1)

	case key.Matches(msg, keys.Kill):
		return m.removeBookmark()

	case key.Matches(msg, keys.AddBookmark):
		// Open project picker to add a new bookmark
		m.mode = ModePickDirectory
		m.returnToBookmarks = true
		m.projectList.Reset()
		m.projectList.SetItems(m.scanProjectDirectories())
		return m, tea.WindowSize()

	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case msg.Type == tea.KeyBackspace:
		filter := m.bookmarkList.Filter()
		if len(filter) > 0 {
			m.bookmarkList.SetFilter(filter[:len(filter)-1])
		}

	case msg.Type == tea.KeyRunes:
		m.bookmarkList.SetFilter(m.bookmarkList.Filter() + string(msg.Runes))
	}

	return m, nil
}

// addPathToBookmarks adds a path to bookmarks
func (m *Model) addPathToBookmarks(path string) (tea.Model, tea.Cmd) {
	// Check if already bookmarked
	for _, b := range m.config.Bookmarks {
		if b.Path == path {
			m.setError("Already bookmarked")
			return m, nil
		}
	}

	// Add bookmark
	m.config.Bookmarks = append(m.config.Bookmarks, config.Bookmark{
		Path: path,
	})

	// Save config
	if err := m.config.SaveBookmarks(); err != nil {
		m.setError("Failed to save config: %v", err)
		return m, nil
	}

	m.setMessage("Added bookmark: %s", filepath.Base(path))
	return m, nil
}

// addSelectedToBookmarks adds the currently selected session to bookmarks
func (m *Model) addSelectedToBookmarks() (tea.Model, tea.Cmd) {
	if len(m.items) == 0 || m.cursor >= len(m.items) {
		return m, nil
	}

	item := m.items[m.cursor]
	if item.Type != ItemTypeSession {
		m.setError("Select a session to bookmark")
		return m, nil
	}

	session := m.getSession(item)
	// Get session path from tmux
	path, err := git.GetSessionPath(session.Name)
	if err != nil || path == "" {
		// Fallback: assume it's in one of the project dirs
		for _, dir := range m.config.ProjectDirs {
			possiblePath := filepath.Join(dir, session.Name)
			if _, err := os.Stat(possiblePath); err == nil {
				path = possiblePath
				break
			}
		}
	}

	if path == "" {
		m.setError("Could not determine path for session")
		return m, nil
	}

	// Check if already bookmarked
	for _, b := range m.config.Bookmarks {
		if b.Path == path || filepath.Base(b.Path) == session.Name {
			m.setError("Session already bookmarked")
			return m, nil
		}
	}

	// Add bookmark
	m.config.Bookmarks = append(m.config.Bookmarks, config.Bookmark{
		Path: path,
	})

	// Save config
	if err := m.config.SaveBookmarks(); err != nil {
		m.setError("Failed to save config: %v", err)
		return m, nil
	}

	m.setMessage("Added bookmark: %s", session.Name)
	return m, nil
}

// openBookmark opens or switches to a bookmarked session
func (m *Model) openBookmark(bookmark config.Bookmark) (tea.Model, tea.Cmd) {
	sessionName := m.extractSessionName(bookmark.Path)

	// Create session if it doesn't exist
	if !tmux.SessionExists(sessionName) {
		if err := tmux.CreateSession(sessionName, bookmark.Path); err != nil {
			m.setError("Failed to create session: %v", err)
			return m, nil
		}

		// Apply layout if configured
		m.applyLayout(sessionName, bookmark.Path)
	}

	// Switch to the session
	if err := tmux.SwitchClient(sessionName); err != nil {
		m.setError("Failed to switch to session: %v", err)
		return m, nil
	}

	return m, tea.Quit
}

// moveBookmark moves the selected bookmark up or down
func (m *Model) moveBookmark(delta int) (tea.Model, tea.Cmd) {
	cursor := m.bookmarkList.Cursor()
	newPos := cursor + delta

	if newPos < 0 || newPos >= len(m.config.Bookmarks) {
		return m, nil
	}

	// Swap bookmarks
	m.config.Bookmarks[cursor], m.config.Bookmarks[newPos] = m.config.Bookmarks[newPos], m.config.Bookmarks[cursor]

	// Save config
	if err := m.config.SaveBookmarks(); err != nil {
		m.setError("Failed to save config: %v", err)
		return m, nil
	}

	// Update list and cursor
	m.bookmarkList.SetItems(m.config.Bookmarks)
	m.bookmarkList.SetCursor(newPos)

	return m, nil
}

// removeBookmark removes the selected bookmark
func (m *Model) removeBookmark() (tea.Model, tea.Cmd) {
	cursor := m.bookmarkList.Cursor()
	if cursor < 0 || cursor >= len(m.config.Bookmarks) {
		return m, nil
	}

	// Remove bookmark
	m.config.Bookmarks = append(m.config.Bookmarks[:cursor], m.config.Bookmarks[cursor+1:]...)

	// Save config
	if err := m.config.SaveBookmarks(); err != nil {
		m.setError("Failed to save config: %v", err)
		return m, nil
	}

	// Update list
	m.bookmarkList.SetItems(m.config.Bookmarks)

	m.setMessage("Bookmark removed")
	return m, nil
}

func (m Model) viewBookmarks() string {
	var header strings.Builder
	var b strings.Builder
	filter := m.bookmarkList.Filter()

	// Fixed header: title bar + prompt + border
	header.WriteString(ui.RenderTitleBar(config.AppName, m.mode.String(), m.width))
	header.WriteString("\n")
	header.WriteString(ui.RenderPrompt(filter, m.width))
	header.WriteString("\n")
	header.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")

	// Content area
	contentLines := 0

	if m.bookmarkList.Len() == 0 {
		if filter != "" {
			b.WriteString("  No bookmarks matching filter\n")
		} else {
			b.WriteString("  No bookmarks configured\n")
			b.WriteString("  Press C-a to add a bookmark\n")
			contentLines++
		}
		contentLines++
	} else {
		// Calculate max visible items (includes table header)
		contentH := m.contentHeight()
		maxItems := ui.DefaultVisibleItems
		if contentH > 0 {
			if available := contentH - 8; available > 0 { // header(3) + footer(3) + tableHeader(1) + dottedLine(1)
				maxItems = available
			}
		}
		m.bookmarkList.SetHeight(maxItems)

		visibleBookmarks := m.bookmarkList.VisibleItems()
		scrollOffset := m.bookmarkList.ScrollOffset()

		// Calculate layout for bookmarks - use shared maxNameWidth for consistency
		maxGitWidth := 0
		for _, bookmark := range visibleBookmarks {
			// Check if session has git status
			sessionName := m.extractSessionName(bookmark.Path)
			if _, ok := m.gitStatuses[sessionName]; ok && m.config.GitStatusEnabled {
				if ui.GitStatusColumnWidth > maxGitWidth {
					maxGitWidth = ui.GitStatusColumnWidth
				}
			}
		}

		layout := ui.RowLayout{
			NameWidth:      m.maxNameWidth, // Use shared width for stable layout
			GitStatusWidth: maxGitWidth,
		}

		// Table header row
		header := ui.RenderTableHeader(layout, ui.TableHeaderOpts{
			ShowExpandIcon: false,
			ShowTime:       false,
			ShowGit:        maxGitWidth > 0,
			NameLabel:      "BKMK",
		})
		b.WriteString(header)
		b.WriteString("\n")
		b.WriteString(ui.RenderDottedBorder(m.sessionListWidth()))
		b.WriteString("\n")
		contentLines += 2

		scrollbar := ui.ScrollbarChars(m.bookmarkList.Len(), m.bookmarkList.Height(), scrollOffset, len(visibleBookmarks))

		for i, bookmark := range visibleBookmarks {
			absoluteIdx := scrollOffset + i
			selected := m.bookmarkList.IsSelected(absoluteIdx)
			slot := absoluteIdx

			if i < len(scrollbar) {
				if selected {
					b.WriteString(ui.SpacerStyle(scrollbar[i]+" ", true))
				} else {
					b.WriteString(scrollbar[i])
					b.WriteString(" ")
				}
			}

			sessionName := m.extractSessionName(bookmark.Path)
			session := m.findSessionByName(sessionName)
			expanded := m.bookmarkExpanded[bookmark.Path]

			if session != nil {
				// Session exists - show full session data with expand icon
				lastActivity := session.LastActivity
				opts := ui.SessionRowOpts{
					RowOpts: ui.RowOpts{
						Num:            slot,
						Name:           sessionName,
						Selected:       selected,
						ShowExpandIcon: true,
						Expanded:       expanded,
						LastActivity:   &lastActivity,
						AnimFrame:      m.animationFrame,
					},
				}
				if status, ok := m.gitStatuses[sessionName]; ok {
					opts.GitStatus = &status
				}
				if m.gitStatusShowLoading && m.gitStatusPending[sessionName] {
					opts.GitStatusLoading = true
				}
				if status, ok := m.claudeStatuses[sessionName]; ok {
					opts.ClaudeStatus = &status
				}
				b.WriteString(ui.RenderSessionRow(sessionName, session.LastActivity, layout, opts, m.rowWidth()))
				b.WriteString("\n")
				contentLines++

				// Show windows if expanded
				if expanded {
					for _, window := range session.Windows {
						b.WriteString(ui.RenderWindowRow(window.Index, window.Name, ui.WindowRowOpts{Selected: false}, m.rowWidth()))
						b.WriteString("\n")
						contentLines++
					}
				}
			} else {
				// No session - show simple bookmark row
				opts := ui.RowOpts{
					Num:      slot,
					Name:     sessionName,
					Selected: selected,
				}
				b.WriteString(ui.RenderBookmarkRow(sessionName, layout, opts, m.rowWidth()))
				b.WriteString("\n")
				contentLines++
			}
		}
	}

	// Padding to push footer to bottom
	headerLines := ui.HeaderOverhead
	footerLines := 3 // border + notification + hints
	contentH := m.contentHeight()
	if contentH > 0 {
		padding := contentH - headerLines - contentLines - footerLines
		for i := 0; i < padding; i++ {
			b.WriteString("\n")
		}
	}

	hints := ui.UniversalHints
	return m.renderWithSidebar(header.String(), b.String(), ui.BookmarkActions, m.message, hints, m.messageIsError)
}
