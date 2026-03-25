package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/tmux"
	"github.com/black-atom-industries/helm/internal/ui"
)

func (m *Model) handlePickDirectoryMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Cancel):
		// Clear filter first, then exit on second press
		if m.projectList.Filter() != "" {
			m.projectList.SetFilter("")
			return m, nil
		}
		// Clear pending session name when canceling
		m.pendingSessionName = ""
		// Return to bookmarks mode if we came from there
		if m.returnToBookmarks {
			m.returnToBookmarks = false
			m.mode = ModeBookmarks
			m.bookmarkList.SetItems(m.config.Bookmarks) // Refresh bookmarks
			return m, nil
		}
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, keys.Up):
		m.projectList.MoveCursor(-1)

	case key.Matches(msg, keys.Down):
		m.projectList.MoveCursor(1)

	case key.Matches(msg, keys.Select):
		if selected, ok := m.projectList.SelectedItem(); ok {
			if m.pendingSessionName != "" {
				return m.createSessionWithNewFolder(selected, m.pendingSessionName)
			}
			return m.createSessionFromDir(selected)
		}

	case key.Matches(msg, keys.Kill):
		return m.confirmRemoveFolder()

	case key.Matches(msg, keys.AddBookmark):
		// Add selected project to bookmarks
		if selected, ok := m.projectList.SelectedItem(); ok {
			result, cmd := m.addPathToBookmarks(selected)
			// Return to bookmarks mode if we came from there
			if m.returnToBookmarks {
				m.returnToBookmarks = false
				m.mode = ModeBookmarks
				m.bookmarkList.SetItems(m.config.Bookmarks) // Refresh bookmarks
			}
			return result, cmd
		}

	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case msg.Type == tea.KeyBackspace:
		filter := m.projectList.Filter()
		if len(filter) > 0 {
			m.projectList.SetFilter(filter[:len(filter)-1])
		}

	case msg.Type == tea.KeyRunes:
		m.projectList.SetFilter(m.projectList.Filter() + string(msg.Runes))
	}

	return m, nil
}

func (m *Model) handleConfirmRemoveFolderMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Kill):
		return m.removeFolder()
	case key.Matches(msg, keys.Cancel):
		m.mode = ModePickDirectory
		m.message = ""
		m.removeTarget = ""
	}

	return m, nil
}

func (m *Model) confirmRemoveFolder() (tea.Model, tea.Cmd) {
	selected, ok := m.projectList.SelectedItem()
	if !ok {
		return m, nil
	}

	m.removeTarget = selected
	displayPath := m.extractDisplayPath(m.removeTarget)
	m.message = fmt.Sprintf("Remove \"%s\" from disk?", displayPath)
	m.mode = ModeConfirmRemoveFolder
	return m, nil
}

func (m *Model) removeFolder() (tea.Model, tea.Cmd) {
	if m.removeTarget == "" {
		return m, nil
	}

	displayPath := m.extractDisplayPath(m.removeTarget)
	sessionName := m.extractSessionName(m.removeTarget)

	// Kill associated session if it exists
	if tmux.SessionExists(sessionName) {
		_ = tmux.KillSession(sessionName)
	}

	if err := os.RemoveAll(m.removeTarget); err != nil {
		m.setError("Failed to remove: %v", err)
		m.mode = ModePickDirectory
		m.removeTarget = ""
		return m, nil
	}

	m.message = fmt.Sprintf("Removed \"%s\"", displayPath)
	m.mode = ModePickDirectory
	m.removeTarget = ""

	// Rescan directories and re-apply current filter
	m.projectList.SetItems(m.scanProjectDirectories())

	return m, clearMessageAfter(5 * time.Second)
}

// viewPickDirectory renders the directory picker view
func (m Model) viewPickDirectory() string {
	var header strings.Builder
	var b strings.Builder

	// Fixed header: title bar + prompt + border
	header.WriteString(ui.RenderTitleBar(config.AppName, m.mode.String(), m.width))
	header.WriteString("\n")

	filter := m.projectList.Filter()
	header.WriteString(ui.RenderPrompt(filter, m.width))
	header.WriteString("\n")

	header.WriteString(ui.RenderBorder(m.borderWidth()))
	header.WriteString("\n")

	// Use shared helper for consistent visible item calculation
	maxItems := m.projectMaxVisibleItems()

	// Update ScrollList height for proper scrolling
	m.projectList.SetHeight(maxItems)

	// Get visible items from ScrollList
	visibleItems := m.projectList.VisibleItems()
	scrollOffset := m.projectList.ScrollOffset()
	totalItems := m.projectList.Len()

	// Get scrollbar characters for each line
	scrollbar := ui.ScrollbarChars(totalItems, maxItems, scrollOffset, len(visibleItems))

	contentLines := 0
	for i, fullPath := range visibleItems {
		displayPath := m.extractDisplayPath(fullPath)
		selected := m.projectList.IsSelected(scrollOffset + i)

		// Scrollbar on the left
		if i < len(scrollbar) {
			b.WriteString(scrollbar[i])
			b.WriteString(" ")
		}

		if selected {
			b.WriteString(ui.FilterStyle.Render(displayPath))
		} else {
			b.WriteString(displayPath)
		}
		b.WriteString("\n")
		contentLines++
	}

	// Empty state
	if totalItems == 0 {
		if filter != "" {
			b.WriteString("  No directories matching filter\n")
		} else {
			b.WriteString("  No directories found\n")
		}
		contentLines++
	}

	// Add padding to push footer to bottom
	// Fixed header: 3 lines (title + prompt + border)
	// Fixed footer: 5 lines (border + notification + state + hints(2))
	headerLines := ui.HeaderOverhead
	footerLines := 3 // border + notification + hints
	contentH := m.contentHeight()
	if contentH > 0 {
		padding := contentH - headerLines - contentLines - footerLines
		for i := 0; i < padding; i++ {
			b.WriteString("\n")
		}
	}

	return m.renderWithSidebar(header.String(), b.String(), ui.ProjectActions, m.message, ui.UniversalHints, m.messageIsError)
}

// projectMaxVisibleItems returns the actual number of items that can be shown
// based on window height, matching the View's calculation
func (m *Model) projectMaxVisibleItems() int {
	contentH := m.contentHeight()
	if contentH > 0 {
		if available := contentH - ui.BaseOverhead; available > 0 {
			return available
		}
	}
	return ui.DefaultVisibleItems
}

// scanProjectDirectories scans all configured project directories at the configured depth
// and returns full paths to each discovered directory
func (m *Model) scanProjectDirectories() []string {
	var dirs []string
	depth := m.config.ProjectDepth

	// Scan each configured base directory
	for _, baseDir := range m.config.ProjectDirs {
		m.walkAtDepth(baseDir, "", depth, &dirs)
	}

	return dirs
}

// walkAtDepth recursively walks directories and collects full paths at the target depth
func (m *Model) walkAtDepth(baseDir, currentPath string, remainingDepth int, dirs *[]string) {
	if remainingDepth == 0 {
		// We've reached the target depth - add the full path
		if currentPath != "" {
			fullPath := filepath.Join(baseDir, currentPath)
			*dirs = append(*dirs, fullPath)
		}
		return
	}

	// Read the current directory
	scanPath := filepath.Join(baseDir, currentPath)
	entries, err := os.ReadDir(scanPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip VCS/internal hidden directories but allow project dirs like .github-private
		if isInternalHiddenDir(entry.Name()) {
			continue
		}

		var nextPath string
		if currentPath == "" {
			nextPath = entry.Name()
		} else {
			nextPath = filepath.Join(currentPath, entry.Name())
		}

		m.walkAtDepth(baseDir, nextPath, remainingDepth-1, dirs)
	}
}

// isInternalHiddenDir returns true for VCS and internal metadata directories
// (e.g. .git, .hg, .svn) but false for project directories that happen to
// start with a dot (e.g. .github-private).
func isInternalHiddenDir(name string) bool {
	switch name {
	case ".git", ".hg", ".svn", ".DS_Store", ".Trash", ".cache", ".local", ".config":
		return true
	}
	return false
}
