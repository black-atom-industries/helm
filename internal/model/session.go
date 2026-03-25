package model

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/git"
	"github.com/black-atom-industries/helm/internal/tmux"
	"github.com/black-atom-industries/helm/internal/ui"
)

func (m *Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Cancel):
		// Escape: clear filter if active, otherwise quit
		if m.filter != "" {
			m.filter = ""
			m.rebuildItems()
			return m, nil
		}
		return m, tea.Quit

	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
			m.updateScrollOffset()
		}

	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.items)-1 {
			m.cursor++
			m.updateScrollOffset()
		}

	case key.Matches(msg, keys.Expand):
		m.expandCurrent()

	case key.Matches(msg, keys.Collapse):
		m.collapseCurrent()

	case key.Matches(msg, keys.Select):
		// If filter is active but no results, transition to path input mode
		if m.filter != "" && len(m.items) == 0 {
			name := strings.TrimSpace(m.filter)
			if name == "" {
				return m, nil
			}
			m.pendingSessionName = sanitizeSessionName(name)
			m.mode = ModeCreatePath
			m.filter = ""
			// Pre-fill with first ProjectDir + session name
			defaultPath := ""
			if len(m.config.ProjectDirs) > 0 {
				defaultPath = filepath.Join(m.config.ProjectDirs[0], m.pendingSessionName)
			} else {
				homeDir, _ := os.UserHomeDir()
				defaultPath = filepath.Join(homeDir, m.pendingSessionName)
			}
			m.pathInput.SetValue(defaultPath)
			m.pathInput.SetCursor(len(defaultPath))
			m.pathInput.Focus()
			m.updatePathCompletions()
			return m, textinput.Blink
		}
		return m.selectCurrent()

	case key.Matches(msg, keys.Kill):
		return m.confirmKill()

	case key.Matches(msg, keys.Create):
		m.mode = ModeCreate
		m.filter = "" // Clear any active filter
		// Reset input completely
		m.input.Reset()
		m.input.SetValue("")
		m.input.CharLimit = 50
		m.input.Focus()
		return m, textinput.Blink

	case key.Matches(msg, keys.PickDirectory):
		m.mode = ModePickDirectory
		m.returnToBookmarks = false // Coming from normal mode, not bookmarks
		m.projectList.Reset()
		m.projectList.SetItems(m.scanProjectDirectories())
		// Carry over the active filter
		if m.filter != "" {
			m.projectList.SetFilter(m.filter)
			m.filter = ""
		}
		// Request window size to get proper height for layout
		return m, tea.WindowSize()

	case key.Matches(msg, keys.OpenRemote):
		return m.openRemote()

	case key.Matches(msg, keys.DownloadRepo):
		if len(m.config.ProjectDirs) == 0 {
			m.setError("No project_dirs configured")
			return m, nil
		}
		m.cloneBasePath = m.config.ProjectDirs[0]
		m.cloneChoiceCursor = 0
		m.mode = ModeCloneChoice
		return m, nil

	case key.Matches(msg, keys.Lazygit):
		return m.openLazygit()

	case key.Matches(msg, keys.Bookmarks):
		m.mode = ModeBookmarks
		m.bookmarkList.Reset()
		m.bookmarkList.SetItems(m.config.Bookmarks)
		// Carry over the active filter
		if m.filter != "" {
			m.bookmarkList.SetFilter(m.filter)
			m.filter = ""
		}
		return m, tea.WindowSize()

	case key.Matches(msg, keys.AddBookmark):
		return m.addSelectedToBookmarks()

	// Number jumps (only when no filter active)
	case m.filter == "" && key.Matches(msg, keys.Jump0):
		return m.handleJump(0)
	case m.filter == "" && key.Matches(msg, keys.Jump1):
		return m.handleJump(1)
	case m.filter == "" && key.Matches(msg, keys.Jump2):
		return m.handleJump(2)
	case m.filter == "" && key.Matches(msg, keys.Jump3):
		return m.handleJump(3)
	case m.filter == "" && key.Matches(msg, keys.Jump4):
		return m.handleJump(4)
	case m.filter == "" && key.Matches(msg, keys.Jump5):
		return m.handleJump(5)
	case m.filter == "" && key.Matches(msg, keys.Jump6):
		return m.handleJump(6)
	case m.filter == "" && key.Matches(msg, keys.Jump7):
		return m.handleJump(7)
	case m.filter == "" && key.Matches(msg, keys.Jump8):
		return m.handleJump(8)
	case m.filter == "" && key.Matches(msg, keys.Jump9):
		return m.handleJump(9)

	case msg.Type == tea.KeyBackspace:
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.rebuildItems()
		}

	case msg.Type == tea.KeyRunes:
		// Add typed characters to filter
		m.filter += string(msg.Runes)
		m.rebuildItems()
	}

	return m, nil
}

func (m *Model) handleConfirmKillMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Kill):
		// Double C-x confirms the kill
		return m.killCurrent()
	case key.Matches(msg, keys.Cancel):
		m.mode = ModeNormal
		m.message = ""
		m.killTarget = ""
	}

	return m, nil
}

func (m *Model) handleCreateMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Cancel):
		m.mode = ModeNormal
		m.input.Blur()
		return m, nil

	case msg.Type == tea.KeyEnter:
		name := strings.TrimSpace(m.input.Value())
		if name == "" {
			m.setError("Session name cannot be empty")
			return m, nil
		}
		return m.createSession(name)
	}

	// Ignore ctrl key combinations - only pass regular typing to input
	if msg.Type == tea.KeyCtrlN || msg.Type == tea.KeyCtrlO ||
		msg.Type == tea.KeyCtrlJ || msg.Type == tea.KeyCtrlK ||
		msg.Type == tea.KeyCtrlH || msg.Type == tea.KeyCtrlL ||
		msg.Type == tea.KeyCtrlX || msg.Type == tea.KeyCtrlY ||
		msg.Type == tea.KeyCtrlP || msg.Type == tea.KeyCtrlD ||
		msg.Type == tea.KeyCtrlR {
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) handleCreatePathMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Cancel):
		m.mode = ModeNormal
		m.pendingSessionName = ""
		m.pathInput.Blur()
		return m, nil

	case msg.Type == tea.KeyTab:
		// Tab completion - use first completion if available
		if len(m.pathCompletions) > 0 {
			m.pathInput.SetValue(m.pathCompletions[0])
			m.pathInput.SetCursor(len(m.pathCompletions[0]))
			m.updatePathCompletions()
		}
		return m, nil

	case msg.Type == tea.KeyEnter:
		path := strings.TrimSpace(m.pathInput.Value())
		if path == "" {
			m.setError("Path cannot be empty")
			return m, nil
		}
		// Expand ~ to home directory
		if strings.HasPrefix(path, "~") {
			homeDir, _ := os.UserHomeDir()
			path = filepath.Join(homeDir, path[1:])
		}
		return m.createSessionAtPath(path)
	}

	// Ignore ctrl key combinations except for text editing
	if msg.Type == tea.KeyCtrlN || msg.Type == tea.KeyCtrlP ||
		msg.Type == tea.KeyCtrlJ || msg.Type == tea.KeyCtrlK ||
		msg.Type == tea.KeyCtrlH || msg.Type == tea.KeyCtrlL ||
		msg.Type == tea.KeyCtrlX || msg.Type == tea.KeyCtrlY ||
		msg.Type == tea.KeyCtrlB || msg.Type == tea.KeyCtrlR ||
		msg.Type == tea.KeyCtrlD ||
		msg.Type == tea.KeyCtrlG {
		return m, nil
	}

	var cmd tea.Cmd
	m.pathInput, cmd = m.pathInput.Update(msg)
	m.updatePathCompletions()
	return m, cmd
}

// updatePathCompletions updates the list of path completions based on current input
func (m *Model) updatePathCompletions() {
	path := m.pathInput.Value()
	if path == "" {
		m.pathCompletions = nil
		return
	}

	// Expand ~ to home directory for completion
	expandedPath := path
	if strings.HasPrefix(path, "~") {
		homeDir, _ := os.UserHomeDir()
		expandedPath = filepath.Join(homeDir, path[1:])
	}

	// Get the directory to scan and the prefix to match
	dir := filepath.Dir(expandedPath)
	prefix := filepath.Base(expandedPath)

	// If path ends with /, scan that directory
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, string(filepath.Separator)) {
		dir = expandedPath
		prefix = ""
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		m.pathCompletions = nil
		return
	}

	var completions []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip VCS/internal hidden directories but allow project dirs like .github-private
		if isInternalHiddenDir(entry.Name()) {
			continue
		}
		// Match prefix (case-insensitive)
		if prefix != "" && !strings.HasPrefix(strings.ToLower(entry.Name()), strings.ToLower(prefix)) {
			continue
		}
		fullPath := filepath.Join(dir, entry.Name())
		// Convert back to ~ notation if it was used
		if strings.HasPrefix(path, "~") {
			homeDir, _ := os.UserHomeDir()
			if strings.HasPrefix(fullPath, homeDir) {
				fullPath = "~" + fullPath[len(homeDir):]
			}
		}
		completions = append(completions, fullPath)
	}

	m.pathCompletions = completions
}

// createSessionAtPath creates a folder (if needed) and session at the given path
func (m *Model) createSessionAtPath(fullPath string) (tea.Model, tea.Cmd) {
	// Create folder if it doesn't exist
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			m.setError("Failed to create folder: %v", err)
			m.pendingSessionName = ""
			m.mode = ModeNormal
			m.pathInput.Blur()
			return m, nil
		}
	}

	// Extract session name from path
	sessionName := m.extractSessionName(fullPath)

	// Clear pending state
	m.pendingSessionName = ""
	m.pathInput.Blur()

	// Check if session already exists - if so, just switch to it
	if tmux.SessionExists(sessionName) {
		if err := tmux.SwitchClient(sessionName); err != nil {
			m.setError("Failed to switch: %v", err)
			return m, m.loadSessions
		}
		return m, tea.Quit
	}

	// Create the session
	if err := tmux.CreateSession(sessionName, fullPath); err != nil {
		m.setError("Error: %v", err)
		m.mode = ModeNormal
		return m, nil
	}

	// Apply layout if configured
	m.applyLayout(sessionName, fullPath)

	// Switch to the new session
	if err := tmux.SwitchClient(sessionName); err != nil {
		m.setError("Created but failed to switch: %v", err)
		return m, m.loadSessions
	}

	return m, tea.Quit
}

func (m *Model) createSession(name string) (tea.Model, tea.Cmd) {
	// Sanitize session name (spaces, dots, colons break tmux target syntax)
	name = sanitizeSessionName(name)
	workingDir := m.config.DefaultSessionDir
	if err := tmux.CreateSession(name, workingDir); err != nil {
		m.setError("Error: %v", err)
		m.mode = ModeNormal
		m.input.Blur()
		return m, nil
	}

	// Apply layout if configured
	m.applyLayout(name, workingDir)

	// Switch to the new session
	if err := tmux.SwitchClient(name); err != nil {
		m.setError("Created but failed to switch: %v", err)
		return m, m.loadSessions
	}

	return m, tea.Quit
}

func (m *Model) createSessionFromDir(fullPath string) (tea.Model, tea.Cmd) {
	// Extract session name from full path (last N components based on depth)
	name := m.extractSessionName(fullPath)

	// Check if session already exists - if so, just switch to it
	if tmux.SessionExists(name) {
		if err := tmux.SwitchClient(name); err != nil {
			m.setError("Failed to switch: %v", err)
			return m, m.loadSessions
		}
		return m, tea.Quit
	}

	if err := tmux.CreateSession(name, fullPath); err != nil {
		m.setError("Error: %v", err)
		m.mode = ModeNormal
		return m, nil
	}

	// Apply layout if configured
	m.applyLayout(name, fullPath)

	// Switch to the new session
	if err := tmux.SwitchClient(name); err != nil {
		m.setError("Created but failed to switch: %v", err)
		return m, m.loadSessions
	}

	return m, tea.Quit
}

// createSessionWithNewFolder creates a new folder at basePath/sessionName and starts a session there
func (m *Model) createSessionWithNewFolder(basePath, sessionName string) (tea.Model, tea.Cmd) {
	fullPath := filepath.Join(basePath, sessionName)

	// Create folder if it doesn't exist
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			m.setError("Failed to create folder: %v", err)
			m.pendingSessionName = ""
			m.mode = ModeNormal
			return m, nil
		}
	}

	// Clear pending state
	m.pendingSessionName = ""

	// Check if session already exists - if so, just switch to it
	if tmux.SessionExists(sessionName) {
		if err := tmux.SwitchClient(sessionName); err != nil {
			m.setError("Failed to switch: %v", err)
			return m, m.loadSessions
		}
		return m, tea.Quit
	}

	// Create the session
	if err := tmux.CreateSession(sessionName, fullPath); err != nil {
		m.setError("Error: %v", err)
		m.mode = ModeNormal
		return m, nil
	}

	// Apply layout if configured
	m.applyLayout(sessionName, fullPath)

	// Switch to the new session
	if err := tmux.SwitchClient(sessionName); err != nil {
		m.setError("Created but failed to switch: %v", err)
		return m, m.loadSessions
	}

	return m, tea.Quit
}

func (m *Model) handleJump(num int) (tea.Model, tea.Cmd) {
	// Check if we're inside an expanded session - numbers switch to windows
	if m.cursor >= 0 && m.cursor < len(m.items) {
		item := m.items[m.cursor]
		session := m.getSession(item)

		if session.Expanded {
			// Jump to window number within this session
			for _, w := range session.Windows {
				if w.Index == num {
					target := fmt.Sprintf("%s:%d", session.Name, w.Index)
					if err := tmux.SwitchClient(target); err != nil {
						m.setError("Error: %v", err)
						return m, nil
					}
					return m, tea.Quit
				}
			}
		}
	}

	// Session labels: 0, 1, 2... map to non-self session indices
	if num >= 0 && num < len(m.sessions) {
		session := m.sessions[num]
		if err := tmux.SwitchClient(session.Name); err != nil {
			m.setError("Error: %v", err)
			return m, nil
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) expandCurrent() {
	if !m.isCursorValid() {
		return
	}

	item := m.items[m.cursor]

	switch item.Type {
	case ItemTypeSession:
		// Collapse all other sessions first
		for i := range m.sessions {
			m.sessions[i].Expanded = false
		}
		if m.selfSession != nil {
			m.selfSession.Expanded = false
		}

		session := m.getSession(item)
		if len(session.Windows) == 0 {
			// Load windows lazily
			windows, err := tmux.ListWindows(session.Name)
			if err != nil {
				m.setError("Error loading windows: %v", err)
				return
			}
			session.Windows = windows
		}
		session.Expanded = true
		m.rebuildItems()

	case ItemTypeWindow:
		session := m.getSession(item)
		window := &session.Windows[item.WindowIndex]

		// Collapse other windows in this session first
		for i := range session.Windows {
			session.Windows[i].Expanded = false
		}

		if len(window.Panes) == 0 {
			// Load panes lazily
			panes, err := tmux.ListPanes(session.Name, window.Index)
			if err != nil {
				m.setError("Error loading panes: %v", err)
				return
			}
			window.Panes = panes
		}
		window.Expanded = true
		m.rebuildItems()

	case ItemTypePane:
		// Panes are leaf nodes, no-op
	}
}

func (m *Model) collapseCurrent() {
	if !m.isCursorValid() {
		return
	}

	item := m.items[m.cursor]

	switch item.Type {
	case ItemTypeSession:
		// Collapse the session
		m.getSession(item).Expanded = false
		m.rebuildItems()

	case ItemTypeWindow:
		// Collapse parent session, move cursor to session
		m.getSession(item).Expanded = false
		// Move cursor to the parent session
		for i, it := range m.items {
			if it.Type == ItemTypeSession && it.IsSelf == item.IsSelf && it.SessionIndex == item.SessionIndex {
				m.cursor = i
				break
			}
		}
		m.rebuildItems()

	case ItemTypePane:
		// Collapse parent window, move cursor to window
		windowIdx := item.WindowIndex
		m.getSession(item).Windows[windowIdx].Expanded = false
		// Move cursor to the window
		for i, it := range m.items {
			if it.Type == ItemTypeWindow && it.IsSelf == item.IsSelf && it.SessionIndex == item.SessionIndex && it.WindowIndex == windowIdx {
				m.cursor = i
				break
			}
		}
		m.rebuildItems()
	}
}

func (m *Model) selectCurrent() (tea.Model, tea.Cmd) {
	if !m.isCursorValid() {
		return m, nil
	}

	target := m.getTargetName(m.items[m.cursor])
	if err := tmux.SwitchClient(target); err != nil {
		m.setError("Error: %v", err)
		return m, nil
	}

	return m, tea.Quit
}

func (m *Model) openLazygit() (tea.Model, tea.Cmd) {
	if !m.isCursorValid() {
		return m, nil
	}

	item := m.items[m.cursor]
	if item.Type != ItemTypeSession {
		// For windows/panes, use the parent session
		item = Item{Type: ItemTypeSession, SessionIndex: item.SessionIndex}
	}

	session := m.getSession(item)
	path, err := git.GetSessionPath(session.Name)
	if err != nil || path == "" {
		m.setError("Could not get session path")
		return m, nil
	}

	// Schedule lazygit popup to open after helm closes, then reopen helm with same dimensions
	cmd := fmt.Sprintf("sleep 0.1 && tmux display-popup -w%s -h%s -d '%s' -E lazygit; tmux display-popup -w%d -h%d -B -E helm",
		m.config.LazygitPopup.Width, m.config.LazygitPopup.Height, path, m.width, m.height)
	_ = exec.Command("tmux", "run-shell", "-b", cmd).Start()

	return m, tea.Quit
}

func (m *Model) openRemote() (tea.Model, tea.Cmd) {
	if !m.isCursorValid() {
		return m, nil
	}

	item := m.items[m.cursor]
	if item.Type != ItemTypeSession {
		item = Item{Type: ItemTypeSession, SessionIndex: item.SessionIndex}
	}

	session := m.getSession(item)
	path, err := git.GetSessionPath(session.Name)
	if err != nil || path == "" {
		m.setError("Could not get session path")
		return m, clearMessageAfter(5 * time.Second)
	}

	remoteURL, err := git.GetRemoteURL(path)
	if err != nil {
		m.setError("No git remote found")
		return m, clearMessageAfter(5 * time.Second)
	}

	// Open in browser (macOS)
	if err := exec.Command("open", remoteURL).Start(); err != nil {
		m.setError("Failed to open browser: %v", err)
		return m, clearMessageAfter(5 * time.Second)
	}

	// Extract org/repo for the message
	parts := strings.Split(remoteURL, "/")
	displayName := remoteURL
	if len(parts) >= 2 {
		displayName = parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}

	m.setMessage("Opened: %s", displayName)
	return m, clearMessageAfter(5 * time.Second)
}

func (m *Model) confirmKill() (tea.Model, tea.Cmd) {
	if !m.isCursorValid() {
		return m, nil
	}

	item := m.items[m.cursor]
	m.killTarget = m.getTargetName(item)

	switch item.Type {
	case ItemTypeSession:
		m.message = fmt.Sprintf("Kill \"%s\"?", m.killTarget)
	case ItemTypeWindow:
		m.message = fmt.Sprintf("Kill window \"%s\"?", m.killTarget)
	case ItemTypePane:
		m.message = fmt.Sprintf("Kill pane \"%s\"?", m.killTarget)
	}

	m.mode = ModeConfirmKill
	return m, nil
}

func (m *Model) killCurrent() (tea.Model, tea.Cmd) {
	if !m.isCursorValid() {
		return m, nil
	}

	item := m.items[m.cursor]
	session := m.getSession(item)
	var err error

	// Killing the self session: switch to the last-used other session first
	if item.IsSelf && item.Type == ItemTypeSession {
		if len(m.sessions) > 0 {
			_ = tmux.SwitchClient(m.sessions[0].Name)
		}
		err = tmux.KillSession(session.Name)
		if err != nil {
			m.setError("Error: %v", err)
		}
		// Helm's tmux session is gone — exit
		return m, tea.Quit
	}

	switch item.Type {
	case ItemTypeSession:
		err = tmux.KillSession(session.Name)
		if err == nil {
			m.message = fmt.Sprintf("Killed \"%s\"", session.Name)
		}
	case ItemTypeWindow:
		window := session.Windows[item.WindowIndex]
		err = tmux.KillWindow(session.Name, window.Index)
		if err == nil {
			m.message = fmt.Sprintf("Killed window %d", window.Index)
		}
	case ItemTypePane:
		window := session.Windows[item.WindowIndex]
		pane := window.Panes[item.PaneIndex]
		err = tmux.KillPane(session.Name, window.Index, pane.Index)
		if err == nil {
			m.message = fmt.Sprintf("Killed pane %d", pane.Index)
		}
	}

	if err != nil {
		m.setError("Error: %v", err)
	}

	m.mode = ModeNormal
	m.killTarget = ""

	// Reload sessions and clear message after 5 seconds
	return m, tea.Batch(m.loadSessions, clearMessageAfter(5*time.Second))
}

// viewSessionList renders the main session list view
func (m Model) viewSessionList() string {
	var b strings.Builder

	// Fixed header: title bar + prompt + border
	b.WriteString(ui.RenderTitleBar(config.AppName, m.mode.String(), m.width))
	b.WriteString("\n")

	// Prompt line - always show filter (input goes in notification line for create mode)
	b.WriteString(ui.RenderPrompt(m.filter, m.width))
	b.WriteString("\n")

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")

	// Build layout for consistent column widths (needed for header)
	layout := ui.RowLayout{
		NameWidth:      m.maxNameWidth,
		GitStatusWidth: m.maxGitStatusWidth,
	}

	// --- Build session list content ---
	var listBuilder strings.Builder
	contentLines := 0
	listWidth := m.sessionListWidth()

	// Table header row (only show when sessions are loaded)
	if m.sessionsLoaded && len(m.items) > 0 {
		header := ui.RenderTableHeader(layout, ui.TableHeaderOpts{
			ShowExpandIcon: true,
			ShowTime:       true,
			ShowGit:        m.maxGitStatusWidth > 0,
			NameLabel:      "SESS",
		})
		listBuilder.WriteString(header)
		listBuilder.WriteString("\n")
		listBuilder.WriteString(ui.RenderDottedBorder(listWidth))
		listBuilder.WriteString("\n")
		contentLines += 2
	}

	// Session list (only visible items)
	maxVisible := m.sessionMaxVisibleItems()
	endIdx := m.scrollOffset + maxVisible
	if endIdx > len(m.items) {
		endIdx = len(m.items)
	}
	visibleCount := endIdx - m.scrollOffset

	// Get scrollbar characters for each line
	scrollbar := ui.ScrollbarChars(len(m.items), maxVisible, m.scrollOffset, visibleCount)

	// Calculate session numbers (count sessions before visible area)
	sessionNum := 0
	for i := 0; i < m.scrollOffset && i < len(m.items); i++ {
		if m.items[i].Type == ItemTypeSession {
			sessionNum++
		}
	}
	for i := m.scrollOffset; i < endIdx; i++ {
		item := m.items[i]
		selected := i == m.cursor
		lineIdx := i - m.scrollOffset

		// Scrollbar on the left (skip for pinned self session)
		if item.IsSelf {
			if selected {
				listBuilder.WriteString(ui.SpacerStyle("  ", true))
			} else {
				listBuilder.WriteString("  ")
			}
		} else if lineIdx < len(scrollbar) {
			if selected {
				listBuilder.WriteString(ui.SpacerStyle(scrollbar[lineIdx]+" ", true))
			} else {
				listBuilder.WriteString(scrollbar[lineIdx])
				listBuilder.WriteString(" ")
			}
		}

		switch item.Type {
		case ItemTypeSession:
			session := m.getSession(item)

			// Build options for this row
			lastActivity := session.LastActivity
			opts := ui.SessionRowOpts{
				RowOpts: ui.RowOpts{
					Num:            sessionNum,
					Name:           session.Name,
					Selected:       selected,
					ShowExpandIcon: true,
					Expanded:       session.Expanded,
					LastActivity:   &lastActivity,
					AnimFrame:      m.animationFrame,
					IsSelf:         item.IsSelf,
				},
			}
			if status, ok := m.gitStatuses[session.Name]; ok {
				opts.GitStatus = &status
			}
			if m.gitStatusShowLoading && m.gitStatusPending[session.Name] {
				opts.GitStatusLoading = true
			}
			if status, ok := m.claudeStatuses[session.Name]; ok {
				opts.ClaudeStatus = &status
			}

			listBuilder.WriteString(ui.RenderSessionRow(session.Name, session.LastActivity, layout, opts, m.rowWidth()))
			if !item.IsSelf {
				sessionNum++
			}

		case ItemTypeWindow:
			session := m.getSession(item)
			window := session.Windows[item.WindowIndex]
			listBuilder.WriteString(ui.RenderWindowRow(window.Index, window.Name, ui.WindowRowOpts{Selected: selected, Expanded: window.Expanded}, m.rowWidth()))

		case ItemTypePane:
			session := m.getSession(item)
			window := session.Windows[item.WindowIndex]
			pane := window.Panes[item.PaneIndex]
			listBuilder.WriteString(ui.RenderPaneRow(pane.Index, pane.Command, pane.Active, ui.PaneRowOpts{Selected: selected}, m.rowWidth()))
		}
		listBuilder.WriteString("\n")
		contentLines++

		// Separator between pinned self session and regular sessions
		if item.IsSelf && (i+1 >= endIdx || !m.items[i+1].IsSelf) {
			listBuilder.WriteString(ui.RenderDottedBorder(listWidth))
			listBuilder.WriteString("\n")
			contentLines++
		}
	}

	// Empty state (only show after sessions have loaded to avoid flash)
	if len(m.items) == 0 && m.sessionsLoaded {
		if m.filter != "" {
			listBuilder.WriteString("  No sessions matching filter\n")
		} else {
			listBuilder.WriteString("  No other sessions available\n")
		}
		contentLines++
	}

	// Pad session list to fill available height (push footer to bottom)
	// Header: 3 lines (title + prompt + border)
	// Footer: 3 lines (border + notification + hints)
	headerLines := ui.HeaderOverhead
	footerLines := 3 // border + notification + single-line hints
	contentH := m.contentHeight()
	if contentH > 0 {
		padding := contentH - headerLines - contentLines - footerLines
		for i := 0; i < padding; i++ {
			listBuilder.WriteString("\n")
		}
	}

	// --- Compose with sidebar and footer via shared helper ---
	header := b.String()
	listContent := listBuilder.String()

	var hints string
	var notification string
	switch m.mode {
	case ModeNormal:
		notification = m.message
		if notification == "" {
			notification = m.statusLine()
		}
		hints = "Type filter · C-j/k ↕ Nav · C-h/l ↔ Expand · Enter Switch"
	case ModeConfirmKill:
		notification = m.message
		hints = "C-x Confirm · Esc Cancel"
	case ModeCreate:
		notification = "New session: " + m.input.View()
		hints = "Enter Create · Esc Cancel"
	}

	return m.renderWithSidebar(header, listContent, ui.SessionActions, notification, hints, m.messageIsError)
}

// viewCreatePath renders the path input view for creating sessions at arbitrary paths
func (m Model) viewCreatePath() string {
	var b strings.Builder

	// Fixed header: title bar + prompt + border
	b.WriteString(ui.RenderTitleBar(config.AppName, m.mode.String(), m.width))
	b.WriteString("\n")

	// Path input line
	b.WriteString(ui.RenderPrompt(m.pathInput.View(), m.width))
	b.WriteString("\n")

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")

	// Content area - show completions
	contentLines := 0

	if len(m.pathCompletions) > 0 {
		b.WriteString("  Completions (Tab to complete):\n")
		contentLines++
		maxShow := 8
		if len(m.pathCompletions) < maxShow {
			maxShow = len(m.pathCompletions)
		}
		for i := 0; i < maxShow; i++ {
			b.WriteString("    " + m.pathCompletions[i] + "\n")
			contentLines++
		}
		if len(m.pathCompletions) > maxShow {
			fmt.Fprintf(&b, "    ... and %d more\n", len(m.pathCompletions)-maxShow)
			contentLines++
		}
	} else {
		b.WriteString("  Enter path for new session\n")
		contentLines++
		b.WriteString("  (folder will be created if it doesn't exist)\n")
		contentLines++
	}

	// Add padding to push footer to bottom
	headerLines := ui.HeaderOverhead
	footerLines := 3 // border + notification + hints
	contentH := m.contentHeight()
	if contentH > 0 {
		padding := contentH - headerLines - contentLines - footerLines
		for i := 0; i < padding; i++ {
			b.WriteString("\n")
		}
	}

	// Simplified footer
	hints := ui.HelpCreatePath()
	notification := m.message
	if notification == "" {
		notification = fmt.Sprintf("Create session: %s", m.pendingSessionName)
	}
	b.WriteString(ui.RenderSimpleFooter(notification, hints, m.messageIsError, m.width))

	return ui.AppStyle.Render(b.String())
}

// sessionMaxVisibleItems returns the actual number of session items that can be shown
// based on window height, accounting for fixed UI elements
func (m *Model) sessionMaxVisibleItems() int {
	contentH := m.contentHeight()
	if contentH > 0 {
		// Header: 3 (title + prompt + border)
		// Footer: 3 (border + notification + hints)
		// Total base: 6
		overhead := 6
		if m.sessionsLoaded && len(m.items) > 0 {
			overhead += ui.TableHeaderHeight + ui.TableDottedLineHeight // +2 = 8
		}
		// Self session row + separator are pinned, not scrollable
		if m.selfSession != nil {
			overhead += ui.SelfSessionOverhead
		}
		if available := contentH - overhead; available > 0 {
			return available
		}
	}
	return ui.DefaultVisibleItems
}
