package model

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/black-atom-industries/helm/internal/claude"
	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/git"
	"github.com/black-atom-industries/helm/internal/tmux"
	"github.com/black-atom-industries/helm/internal/ui"
)

// Mode represents the current UI mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeConfirmKill
	ModeCreate
	ModePickDirectory
	ModeConfirmRemoveFolder
	ModeCloneChoice // Sub-menu: URL or My Repos
	ModeCloneRepo
	ModeCloneURL // Text input for arbitrary repo URL
	ModeBookmarks
	ModeCreatePath // Path input for creating session at arbitrary path
)

// String returns the display name for the mode (used in title bar)
func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "SESSIONS"
	case ModeBookmarks:
		return "BOOKMARKS"
	case ModePickDirectory:
		return "PROJECTS"
	case ModeCloneChoice, ModeCloneRepo, ModeCloneURL:
		return "DOWNLOADS"
	case ModeCreate:
		return "NEW"
	case ModeCreatePath:
		return "PATH"
	case ModeConfirmKill:
		return "KILL"
	case ModeConfirmRemoveFolder:
		return "REMOVE"
	default:
		return "SESSIONS"
	}
}

// ItemType represents the type of item in the flattened list
type ItemType int

const (
	ItemTypeSession ItemType = iota
	ItemTypeWindow
	ItemTypePane
)

// Item represents a session, window, or pane in the flattened list
type Item struct {
	Type         ItemType
	SessionIndex int  // Index in the sessions slice
	WindowIndex  int  // Index in the session's windows slice (for windows and panes)
	PaneIndex    int  // Index in the window's panes slice (for panes only)
	IsSelf       bool // True if this item belongs to the current/self session
}

// Model is the main application state
type Model struct {
	sessions          []tmux.Session
	selfSession       *tmux.Session // The current/self session (pinned at top)
	claudeStatuses    map[string]claude.Status
	gitStatuses       map[string]git.Status
	currentSession    string
	cursor            int
	items             []Item // Flattened list of visible items
	mode              Mode
	message           string
	messageIsError    bool
	input             textinput.Model
	killTarget        string // Name of session/window being killed
	removeTarget      string // Full path of folder being removed
	config            config.Config
	maxNameWidth      int    // For column alignment
	maxGitStatusWidth int    // For git status column alignment
	filter            string // Current filter text for fuzzy matching

	// Directory picker state (uses ScrollList for cursor/scroll/filter)
	projectList        *ui.ScrollList[string]
	returnToBookmarks  bool   // True if we should return to bookmarks mode after project picker
	pendingSessionName string // Session name pending directory selection (for create-from-filter flow)

	// Path input state (for ModeCreatePath)
	pathInput       textinput.Model // Text input for path entry
	pathCompletions []string        // Available path completions

	// Scroll state
	scrollOffset int // Scroll offset for session list

	// Window size
	width  int
	height int

	// Animation state
	animationFrame int

	// Clone mode state
	cloneChoiceCursor   int // 0 = Enter URL, 1 = My repos
	cloneList           *ui.ScrollList[string]
	cloneBasePath       string // From config.ProjectDirs
	cloneLoading        bool   // True while fetching repos
	cloneError          string // Error message if fetch/clone fails
	cloneCloning        bool   // True while cloning
	cloneCloningRepo    string // Repo being cloned
	cloneSuccess        bool   // True when clone completed, awaiting confirmation
	cloneSuccessPath    string // Path of cloned repo (for layout)
	cloneSuccessSession string // Session name to switch to
	clonePendingFilter  string // Filter to apply once repos are loaded

	// Bookmarks mode state (uses ScrollList for cursor/scroll/filter)
	bookmarkList     *ui.ScrollList[config.Bookmark]
	bookmarkExpanded map[string]bool // Tracks which bookmarks are expanded (by path)

	// Loading state
	sessionsLoaded bool // True after sessions have been loaded at least once

	// Git status loading state
	gitStatusPending     map[string]bool // Sessions still being fetched (by name)
	gitStatusShowLoading bool            // True after 500ms delay if still loading
}

// New creates a new Model
// ParseInitialView maps a CLI flag value to a Mode.
// Returns ModeNormal for empty or unrecognized values.
func ParseInitialView(view string) Mode {
	switch strings.ToLower(view) {
	case "bookmarks":
		return ModeBookmarks
	case "projects":
		return ModePickDirectory
	case "clone":
		return ModeCloneChoice
	default:
		return ModeNormal
	}
}

func New(currentSession string, cfg config.Config, initialView string) Model {
	ti := textinput.New()
	ti.Prompt = "" // We handle the prompt in RenderPrompt
	ti.CharLimit = 50

	// Path input for ModeCreatePath
	pathInput := textinput.New()
	pathInput.Prompt = ""
	pathInput.CharLimit = 256

	// Create project list with filter function that matches on directory basename
	projectList := ui.NewScrollList(func(fullPath string, filter string) bool {
		name := filepath.Base(fullPath)
		return strings.Contains(strings.ToLower(name), filter)
	})

	// Create clone list with filter function that matches on repo name
	cloneList := ui.NewScrollList(func(repo string, filter string) bool {
		return strings.Contains(strings.ToLower(repo), filter)
	})

	// Create bookmark list with filter function that matches on path basename
	bookmarkList := ui.NewScrollList(func(b config.Bookmark, filter string) bool {
		name := filepath.Base(b.Path)
		return strings.Contains(strings.ToLower(name), filter) ||
			strings.Contains(strings.ToLower(b.Path), filter)
	})

	m := Model{
		currentSession:   currentSession,
		mode:             ParseInitialView(initialView),
		input:            ti,
		pathInput:        pathInput,
		config:           cfg,
		projectList:      projectList,
		cloneList:        cloneList,
		bookmarkList:     bookmarkList,
		bookmarkExpanded: make(map[string]bool),
	}

	// Populate data for non-default initial views
	switch m.mode {
	case ModePickDirectory:
		m.projectList.SetItems(m.scanProjectDirectories())
	case ModeBookmarks:
		m.bookmarkList.SetItems(cfg.Bookmarks)
	}

	// Load cached sessions for instant startup
	if cached := m.loadSessionCache(); cached != nil {
		m.sessions = cached
		m.sessionsLoaded = true
		m.calculateColumnWidths()
		// Reserve git status column to prevent layout shift when statuses load
		if cfg.GitStatusEnabled {
			m.maxGitStatusWidth = ui.GitStatusColumnWidth
		}
		m.rebuildItems()
	}

	return m
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadSessions, animationTick())
}

// loadSessions fetches sessions from tmux
func (m Model) loadSessions() tea.Msg {
	sessions, err := tmux.ListSessions(m.currentSession)
	if err != nil {
		return errMsg{err}
	}

	// Fetch self session activity
	var selfSession *tmux.Session
	if m.currentSession != "" {
		activity, err := tmux.GetSessionActivity(m.currentSession)
		if err == nil {
			selfSession = &tmux.Session{
				Name:         m.currentSession,
				LastActivity: activity,
			}
		}
	}

	return sessionsMsg{sessions: sessions, selfSession: selfSession}
}

type sessionsMsg struct {
	sessions    []tmux.Session
	selfSession *tmux.Session
}

type errMsg struct {
	err error
}

type clearMessageMsg struct{}

type animationTickMsg struct{}

// gitStatusSingleMsg is sent when a single session's git status is ready
type gitStatusSingleMsg struct {
	sessionName string
	status      git.Status
	hasStatus   bool // true if status should be shown (repo with changes)
}

// gitStatusLoadingMsg is sent after 500ms to show loading indicator
type gitStatusLoadingMsg struct{}

// clearMessageAfter returns a command that clears the message after a delay
func clearMessageAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

// animationTick returns a command that ticks the animation
func animationTick() tea.Cmd {
	return tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
		return animationTickMsg{}
	})
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sessionsMsg:
		m.sessions = msg.sessions
		m.selfSession = msg.selfSession
		m.sessionsLoaded = true
		m.saveSessionCache() // Cache for instant startup next time
		m.loadClaudeStatuses()
		// Initialize git statuses map (will be populated async)
		if m.gitStatuses == nil {
			m.gitStatuses = make(map[string]git.Status)
		}
		// Reserve git status column width to prevent layout shift
		if m.config.GitStatusEnabled {
			m.maxGitStatusWidth = ui.GitStatusColumnWidth
		}
		m.calculateColumnWidths()
		m.rebuildItems()
		// Place cursor on the first regular (non-self) session
		if m.selfSession != nil {
			for i, item := range m.items {
				if !item.IsSelf {
					m.cursor = i
					break
				}
			}
			m.updateScrollOffset()
		}
		if len(m.items) == 0 {
			m.message = "No sessions. Press C-n to create one."
		}
		// Fetch git statuses asynchronously to avoid blocking UI
		return m, m.fetchGitStatusesCmd()

	case errMsg:
		m.setError("Error: %v", msg.err)
		return m, nil

	case clearMessageMsg:
		m.message = ""
		m.messageIsError = false
		return m, nil

	case animationTickMsg:
		m.animationFrame = (m.animationFrame + 1) % 3
		return m, animationTick()

	case cloneReposLoadedMsg:
		m.cloneLoading = false
		m.cloneList.SetItems(msg.repos)
		if m.clonePendingFilter != "" {
			m.cloneList.SetFilter(m.clonePendingFilter)
			m.clonePendingFilter = ""
		}
		if len(msg.repos) == 0 {
			m.cloneError = "All repositories are already cloned!"
		}
		return m, nil

	case cloneErrorMsg:
		m.cloneLoading = false
		m.cloneCloning = false
		m.cloneError = msg.err.Error()
		return m, nil

	case cloneSuccessMsg:
		// Store success state and await confirmation
		m.cloneCloning = false
		m.cloneSuccess = true
		m.cloneSuccessPath = msg.repoPath
		m.cloneSuccessSession = msg.sessionName
		return m, nil

	case gitStatusSingleMsg:
		// Single git status loaded - update incrementally
		if msg.hasStatus {
			m.gitStatuses[msg.sessionName] = msg.status
		}
		delete(m.gitStatusPending, msg.sessionName)
		if len(m.gitStatusPending) == 0 {
			m.gitStatusShowLoading = false
		}
		return m, nil

	case gitStatusLoadingMsg:
		// 500ms elapsed - show loading indicator if still fetching
		if len(m.gitStatusPending) > 0 {
			m.gitStatusShowLoading = true
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Handle text input updates in create mode
	if m.mode == ModeCreate {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	// Handle text input updates in create path mode
	if m.mode == ModeCreatePath {
		var cmd tea.Cmd
		m.pathInput, cmd = m.pathInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeNormal:
		return m.handleNormalMode(msg)
	case ModeConfirmKill:
		return m.handleConfirmKillMode(msg)
	case ModeCreate:
		return m.handleCreateMode(msg)
	case ModeCreatePath:
		return m.handleCreatePathMode(msg)
	case ModePickDirectory:
		return m.handlePickDirectoryMode(msg)
	case ModeConfirmRemoveFolder:
		return m.handleConfirmRemoveFolderMode(msg)
	case ModeCloneChoice:
		return m.handleCloneChoiceMode(msg)
	case ModeCloneRepo:
		return m.handleCloneRepoMode(msg)
	case ModeCloneURL:
		return m.handleCloneURLMode(msg)
	case ModeBookmarks:
		return m.handleBookmarksMode(msg)
	}
	return m, nil
}

// extractSessionName extracts a session name from a full path
// Uses the last N path components based on ProjectDepth config
func (m *Model) extractSessionName(fullPath string) string {
	parts := strings.Split(fullPath, string(filepath.Separator))
	depth := m.config.ProjectDepth
	if depth > len(parts) {
		depth = len(parts)
	}
	relPath := strings.Join(parts[len(parts)-depth:], "/")
	return sanitizeSessionName(relPath)
}

// extractDisplayPath extracts a display path from a full path
// Uses the last N path components based on ProjectDepth config
func (m *Model) extractDisplayPath(fullPath string) string {
	parts := strings.Split(fullPath, string(filepath.Separator))
	depth := m.config.ProjectDepth
	if depth > len(parts) {
		depth = len(parts)
	}
	return strings.Join(parts[len(parts)-depth:], "/")
}

// allSessions returns self session + other sessions as a combined slice
func (m *Model) allSessions() []tmux.Session {
	var all []tmux.Session
	if m.selfSession != nil {
		all = append(all, *m.selfSession)
	}
	all = append(all, m.sessions...)
	return all
}

// getSession returns the session for a given item (handles self session)
func (m *Model) getSession(item Item) *tmux.Session {
	if item.IsSelf {
		return m.selfSession
	}
	return &m.sessions[item.SessionIndex]
}

// findSessionByName finds a session by its name, returns nil if not found
func (m *Model) findSessionByName(name string) *tmux.Session {
	for i := range m.sessions {
		if m.sessions[i].Name == name {
			return &m.sessions[i]
		}
	}
	return nil
}

func (m *Model) applyLayout(sessionName, workingDir string) {
	if !m.config.EnableLayouts || m.config.Layout == "" {
		return
	}

	scriptPath := fmt.Sprintf("%s/%s.sh", m.config.LayoutDir, m.config.Layout)
	if _, err := os.Stat(scriptPath); err != nil {
		return
	}

	// Run layout script synchronously before switching to the session
	cmd := exec.Command(scriptPath, sessionName, workingDir)
	cmd.Env = append(os.Environ(),
		"TMUX_SESSION="+sessionName,
		"TMUX_WORKING_DIR="+workingDir,
	)
	_ = cmd.Run()
}

func (m *Model) loadClaudeStatuses() {
	m.claudeStatuses = make(map[string]claude.Status)
	if !m.config.ClaudeStatusEnabled {
		return
	}
	if m.selfSession != nil {
		status := claude.GetStatus(m.selfSession.Name, m.config.CacheDir)
		if status.State != "" {
			m.claudeStatuses[m.selfSession.Name] = status
		}
	}
	for _, s := range m.sessions {
		status := claude.GetStatus(s.Name, m.config.CacheDir)
		if status.State != "" {
			m.claudeStatuses[s.Name] = status
		}
	}
}

// fetchGitStatusesCmd returns commands that fetch git statuses in parallel
// Each session's status is fetched independently and updates the UI as soon as ready
func (m *Model) fetchGitStatusesCmd() tea.Cmd {
	allSessions := m.allSessions()
	if !m.config.GitStatusEnabled || len(allSessions) == 0 {
		return nil
	}

	// Track which sessions we're waiting for
	m.gitStatusPending = make(map[string]bool)
	m.gitStatusShowLoading = false
	for _, s := range allSessions {
		m.gitStatusPending[s.Name] = true
	}

	// Create a command for each session - they run in parallel via tea.Batch
	cmds := make([]tea.Cmd, 0, len(allSessions)+1)

	// Add delayed loading indicator (500ms)
	cmds = append(cmds, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return gitStatusLoadingMsg{}
	}))

	for _, s := range allSessions {
		sessionName := s.Name // capture for closure
		cmds = append(cmds, func() tea.Msg {
			path, err := git.GetSessionPath(sessionName)
			if err != nil || path == "" {
				return gitStatusSingleMsg{sessionName: sessionName, hasStatus: false}
			}
			status := git.GetStatus(path)
			if status.IsRepo && !status.IsClean() {
				return gitStatusSingleMsg{sessionName: sessionName, status: status, hasStatus: true}
			}
			return gitStatusSingleMsg{sessionName: sessionName, hasStatus: false}
		})
	}

	return tea.Batch(cmds...)
}

func (m *Model) calculateColumnWidths() {
	// Don't reset - preserve cached width to prevent layout shift
	if m.selfSession != nil {
		if len(m.selfSession.Name) > m.maxNameWidth {
			m.maxNameWidth = len(m.selfSession.Name)
		}
	}
	for _, s := range m.sessions {
		if len(s.Name) > m.maxNameWidth {
			m.maxNameWidth = len(s.Name)
		}
	}
}

// stateText returns the state line text based on current mode and context.
// Now only used by the STATUS sidebar section — no longer in footer.
// Kept for potential future use.
//
//nolint:unused
func (m *Model) stateText() string {
	switch m.mode {
	case ModeNormal:
		total := len(m.sessions)
		if m.selfSession != nil {
			total++
		}
		visible := len(m.items)
		if m.filter != "" {
			return fmt.Sprintf("Showing %d/%d sessions", visible, total)
		}
		return fmt.Sprintf("%d sessions", total)
	case ModeBookmarks:
		total := len(m.config.Bookmarks)
		if total == 0 {
			return "No bookmarks"
		}
		return fmt.Sprintf("%d bookmarks", total)
	case ModePickDirectory:
		if m.pendingSessionName != "" {
			return fmt.Sprintf("Select location for: %s", m.pendingSessionName)
		}
		total := len(m.projectList.Items())
		visible := m.projectList.Len()
		if m.projectList.Filter() != "" {
			return fmt.Sprintf("Showing %d/%d projects", visible, total)
		}
		return fmt.Sprintf("%d projects", total)
	case ModeCloneChoice:
		return "Clone repository"
	case ModeCloneURL:
		if m.cloneSuccess {
			return "Clone successful"
		}
		if m.cloneCloning {
			return fmt.Sprintf("Cloning %s...", m.cloneCloningRepo)
		}
		return "Enter repo URL"
	case ModeCloneRepo:
		if m.cloneSuccess {
			return "Clone successful"
		}
		if m.cloneLoading {
			return "Loading repositories..."
		}
		if m.cloneCloning {
			return fmt.Sprintf("Cloning %s...", m.cloneCloningRepo)
		}
		total := len(m.cloneList.Items())
		visible := m.cloneList.Len()
		if m.cloneList.Filter() != "" {
			return fmt.Sprintf("Showing %d/%d repositories", visible, total)
		}
		return fmt.Sprintf("%d repositories", total)
	case ModeCreate:
		return "Enter session name"
	case ModeConfirmKill:
		return fmt.Sprintf("Kill session: %s?", m.killTarget)
	case ModeConfirmRemoveFolder:
		return fmt.Sprintf("Remove folder: %s?", filepath.Base(m.removeTarget))
	default:
		return ""
	}
}

func (m *Model) rebuildItems() {
	m.items = nil
	filterLower := strings.ToLower(m.filter)

	// Pin self session at top (always visible, not affected by filter)
	if m.selfSession != nil {
		m.items = append(m.items, Item{
			Type:   ItemTypeSession,
			IsSelf: true,
		})

		if m.selfSession.Expanded {
			for j, window := range m.selfSession.Windows {
				m.items = append(m.items, Item{
					Type:        ItemTypeWindow,
					IsSelf:      true,
					WindowIndex: j,
				})

				if window.Expanded {
					for k := range window.Panes {
						m.items = append(m.items, Item{
							Type:        ItemTypePane,
							IsSelf:      true,
							WindowIndex: j,
							PaneIndex:   k,
						})
					}
				}
			}
		}
	}

	for i, session := range m.sessions {
		// Apply fuzzy filter if active
		if m.filter != "" && !fuzzyMatch(session.Name, filterLower) {
			continue
		}

		m.items = append(m.items, Item{
			Type:         ItemTypeSession,
			SessionIndex: i,
		})

		if session.Expanded {
			for j, window := range session.Windows {
				m.items = append(m.items, Item{
					Type:         ItemTypeWindow,
					SessionIndex: i,
					WindowIndex:  j,
				})

				// Add panes if window is expanded
				if window.Expanded {
					for k := range window.Panes {
						m.items = append(m.items, Item{
							Type:         ItemTypePane,
							SessionIndex: i,
							WindowIndex:  j,
							PaneIndex:    k,
						})
					}
				}
			}
		}
	}

	// Ensure cursor is in bounds
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.updateScrollOffset()
}

// updateScrollOffset adjusts scroll offset to keep cursor visible in session list
func (m *Model) updateScrollOffset() {
	maxVisible := m.sessionMaxVisibleItems()
	// If cursor is above visible area, scroll up
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	// If cursor is below visible area, scroll down
	if m.cursor >= m.scrollOffset+maxVisible {
		m.scrollOffset = m.cursor - maxVisible + 1
	}
	// Ensure scroll offset is not negative
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// contentWidth returns the available width inside the app border/padding
func (m *Model) contentWidth() int {
	if m.width > 0 {
		return m.width - ui.AppBorderOverheadX
	}
	return 56 // Default fallback (60 - 4)
}

// contentHeight returns the available height inside the app border/padding
func (m *Model) contentHeight() int {
	if m.height > 0 {
		return m.height - ui.AppBorderOverheadY
	}
	return 0
}

// borderWidth returns the width to use for internal borders
func (m *Model) borderWidth() int {
	return m.contentWidth()
}

// sidebarWidth returns the total width consumed by the sidebar (box + gap), or 0 if window is too narrow
func (m *Model) sidebarWidth() int {
	if m.contentWidth() < 40 {
		return 0 // Skip sidebar in very narrow windows
	}
	return ui.SidebarTotalWidth()
}

// sessionListWidth returns the width available for the session list (content minus sidebar)
func (m *Model) sessionListWidth() int {
	return m.contentWidth() - m.sidebarWidth()
}

// rowWidth returns the width available for row content (accounts for scrollbar column and sidebar)
func (m *Model) rowWidth() int {
	return m.sessionListWidth() - ui.ScrollbarColumnWidth
}

// statusLine returns a compact status string for the footer
func (m *Model) statusLine() string {
	total := len(m.sessions)
	if m.selfSession != nil {
		total++
	}
	return fmt.Sprintf("%d sessions", total)
}

// renderWithSidebar joins list content with the sidebar and appends a simplified footer.
// listContent is the session/bookmark/project list string.
// actions is the mode-specific action set for the sidebar.
// notification is the message to show in the footer.
// hints is a single-line keybind hint string.
// isError indicates notification is an error.
func (m *Model) renderWithSidebar(header, listContent string, actions []ui.Action, notification, hints string, isError bool) string {
	var b strings.Builder

	// Header (full width)
	b.WriteString(header)

	// Join list + sidebar line-by-line for exact width control
	if m.sidebarWidth() > 0 {
		sidebarStr := ui.RenderSidebar(actions, 0)
		listLines := strings.Split(strings.TrimRight(listContent, "\n"), "\n")
		sidebarLines := strings.Split(strings.TrimRight(sidebarStr, "\n"), "\n")

		listW := m.sessionListWidth()
		maxLines := len(listLines)
		if len(sidebarLines) > maxLines {
			maxLines = len(sidebarLines)
		}

		for i := 0; i < maxLines; i++ {
			// Left: session list line, forced to exact visual width
			left := ""
			if i < len(listLines) {
				left = listLines[i]
			}
			leftVisWidth := lipgloss.Width(left)
			if leftVisWidth < listW {
				left += strings.Repeat(" ", listW-leftVisWidth)
			} else if leftVisWidth > listW {
				// Truncate lines wider than list width to prevent sidebar shift
				left = lipgloss.NewStyle().MaxWidth(listW).Render(left)
			}

			// Gap
			gap := strings.Repeat(" ", ui.SidebarGap)

			// Right: sidebar line
			right := ""
			if i < len(sidebarLines) {
				right = sidebarLines[i]
			}

			b.WriteString(left + gap + right + "\n")
		}
	} else {
		b.WriteString(listContent)
	}

	// Count content lines so far (header + joined list/sidebar)
	content := b.String()
	contentLineCount := strings.Count(content, "\n")

	// Pad to push footer to bottom: target = contentHeight - footer lines (3)
	targetContentLines := m.contentHeight() - 3
	if targetContentLines > contentLineCount {
		padding := targetContentLines - contentLineCount
		for i := 0; i < padding; i++ {
			b.WriteString("\n")
		}
	}

	// Footer at the very bottom
	b.WriteString(ui.RenderSimpleFooter(notification, hints, isError, m.width))

	return ui.AppStyle.Height(m.contentHeight()).Render(b.String())
}

// fuzzyMatch checks if the pattern matches the text (case-insensitive, substring match)
func fuzzyMatch(text, pattern string) bool {
	textLower := strings.ToLower(text)
	return strings.Contains(textLower, pattern)
}

// isCursorValid returns true if cursor points to a valid item
func (m *Model) isCursorValid() bool {
	return m.cursor >= 0 && m.cursor < len(m.items)
}

// getTargetName returns the tmux target name for the given item
func (m *Model) getTargetName(item Item) string {
	session := m.getSession(item)
	switch item.Type {
	case ItemTypeSession:
		return session.Name
	case ItemTypeWindow:
		window := session.Windows[item.WindowIndex]
		return fmt.Sprintf("%s:%d", session.Name, window.Index)
	case ItemTypePane:
		window := session.Windows[item.WindowIndex]
		pane := window.Panes[item.PaneIndex]
		return fmt.Sprintf("%s:%d.%d", session.Name, window.Index, pane.Index)
	default:
		return session.Name
	}
}

// setError sets an error message on the model
func (m *Model) setError(format string, args ...any) {
	m.message = fmt.Sprintf(format, args...)
	m.messageIsError = true
}

// setMessage sets a non-error message on the model
func (m *Model) setMessage(format string, args ...any) {
	m.message = fmt.Sprintf(format, args...)
	m.messageIsError = false
}

// sanitizeSessionName converts a path to a valid tmux session name
// Dots and colons have special meaning in tmux target syntax (window.pane, session:window)
// Spaces cause issues with shell commands
func sanitizeSessionName(name string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		".", "-",
		":", "-",
		" ", "-",
	)
	return replacer.Replace(name)
}

// View implements tea.Model
func (m Model) View() string {
	if m.mode == ModePickDirectory || m.mode == ModeConfirmRemoveFolder {
		return m.viewPickDirectory()
	}
	if m.mode == ModeCloneChoice {
		return m.viewCloneChoice()
	}
	if m.mode == ModeCloneRepo {
		return m.viewCloneRepo()
	}
	if m.mode == ModeCloneURL {
		return m.viewCloneURL()
	}
	if m.mode == ModeBookmarks {
		return m.viewBookmarks()
	}
	if m.mode == ModeCreatePath {
		return m.viewCreatePath()
	}
	return m.viewSessionList()
}

// sessionCachePath returns the path to the session cache file
func (m *Model) sessionCachePath() string {
	return filepath.Join(m.config.CacheDir, "sessions.json")
}

// cachedSession is a simplified session for caching (excludes UI state)
type cachedSession struct {
	Name         string    `json:"name"`
	LastActivity time.Time `json:"last_activity"`
}

// sessionCache wraps cached sessions with layout metadata for stable column widths
type sessionCache struct {
	Sessions     []cachedSession `json:"sessions"`
	MaxNameWidth int             `json:"max_name_width"` // Persisted to prevent layout shift
}

// loadSessionCache loads cached sessions from disk
// Returns nil if cache doesn't exist or is invalid
func (m *Model) loadSessionCache() []tmux.Session {
	data, err := os.ReadFile(m.sessionCachePath())
	if err != nil {
		return nil
	}

	// Try new format first (with layout metadata)
	var cache sessionCache
	if err := json.Unmarshal(data, &cache); err == nil && len(cache.Sessions) > 0 {
		m.maxNameWidth = cache.MaxNameWidth
		sessions := make([]tmux.Session, len(cache.Sessions))
		for i, c := range cache.Sessions {
			sessions[i] = tmux.Session{
				Name:         c.Name,
				LastActivity: c.LastActivity,
			}
		}
		return sessions
	}

	// Fallback: try old format (array of sessions) for backwards compatibility
	var cached []cachedSession
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil
	}

	sessions := make([]tmux.Session, len(cached))
	for i, c := range cached {
		sessions[i] = tmux.Session{
			Name:         c.Name,
			LastActivity: c.LastActivity,
		}
	}
	return sessions
}

// saveSessionCache saves sessions to disk for instant startup
func (m *Model) saveSessionCache() {
	cached := make([]cachedSession, len(m.sessions))
	for i, s := range m.sessions {
		cached[i] = cachedSession{
			Name:         s.Name,
			LastActivity: s.LastActivity,
		}
	}

	cache := sessionCache{
		Sessions:     cached,
		MaxNameWidth: m.maxNameWidth,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(m.config.CacheDir, 0755); err != nil {
		return
	}

	_ = os.WriteFile(m.sessionCachePath(), data, 0644)
}
