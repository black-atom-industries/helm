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

	"github.com/black-atom-industries/helm/internal/agent"
	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/git"
	"github.com/black-atom-industries/helm/internal/lib/filter"
	"github.com/black-atom-industries/helm/internal/lib/fuzzy"
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
	claudeStatuses    map[string]agent.Status
	piStatuses        map[string]agent.Status
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
	maxNameWidth      int                          // For column alignment
	maxGitStatusWidth int                          // For git status column alignment
	sessionFilter     *filter.Filter[tmux.Session] // Session filter (shared filter logic)

	// Directory picker state (uses ScrollList for cursor/scroll/filter)
	projectList        *ui.ScrollList[string]
	projectsLoading    bool   // True while the async project directory scan runs
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
	gitStatusPending     map[string]bool      // Sessions still being fetched (by name)
	gitStatusShowLoading bool                 // True after 500ms delay if still loading
	gitStatusFetched     map[string]time.Time // Last fetch per session, for the TTL cache
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

	// Create project list with filter function using segment-aware matching
	// Uses full path (relative to base dir) for matching — no depth truncation
	projectList := ui.NewScrollList(func(fullPath string, filter string) bool {
		return fuzzy.MatchPath(fullPath, filter)
	})

	// Create clone list with filter function using segment-aware matching
	cloneList := ui.NewScrollList(func(repo string, filter string) bool {
		return fuzzy.MatchPath(repo, filter)
	})

	// Create bookmark list with filter function using segment-aware matching
	bookmarkList := ui.NewScrollList(func(b config.Bookmark, filter string) bool {
		// Normalize path: strip leading slash to avoid empty first segment
		path := strings.TrimPrefix(b.Path, "/")
		return fuzzy.MatchPath(path, filter)
	})

	sessionFilter := filter.New([]tmux.Session{}, func(s tmux.Session, f string) bool {
		return fuzzy.Match(s.Name, f)
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
		sessionFilter:    sessionFilter,
		bookmarkExpanded: make(map[string]bool),
	}

	// Populate data for non-default initial views
	switch m.mode {
	case ModePickDirectory:
		m.projectsLoading = true // scan dispatched async in Init
	case ModeBookmarks:
		m.bookmarkList.SetItems(cfg.Bookmarks)
	}

	// Load cached sessions for instant startup
	if cached := m.loadSessionCache(); cached != nil {
		m.sessions = cached
		m.sessionFilter.SetItems(m.sessions)
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
	cmds := []tea.Cmd{m.loadSessions, animationTick(), statusPollTick()}
	if m.projectsLoading {
		cmds = append(cmds, m.scanProjectsCmd())
	}
	return tea.Batch(cmds...)
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

	m.preserveExpansion(sessions, selfSession)

	return sessionsMsg{sessions: sessions, selfSession: selfSession}
}

// expandedSession returns the currently expanded session, if any.
func (m *Model) expandedSession() *tmux.Session {
	if m.selfSession != nil && m.selfSession.Expanded {
		return m.selfSession
	}
	for i := range m.sessions {
		if m.sessions[i].Expanded {
			return &m.sessions[i]
		}
	}
	return nil
}

// preserveExpansion carries the expanded session (and its expanded window)
// across a session reload. Windows and panes are re-fetched so the expansion
// reflects current tmux state instead of the pre-reload snapshot.
func (m Model) preserveExpansion(sessions []tmux.Session, selfSession *tmux.Session) {
	old := m.expandedSession()
	if old == nil {
		return
	}

	var target *tmux.Session
	if selfSession != nil && selfSession.Name == old.Name {
		target = selfSession
	} else {
		for i := range sessions {
			if sessions[i].Name == old.Name {
				target = &sessions[i]
				break
			}
		}
	}
	if target == nil {
		return // expanded session no longer exists
	}

	windows, err := tmux.ListWindows(target.Name)
	if err != nil {
		return
	}
	for _, oldWindow := range old.Windows {
		if !oldWindow.Expanded {
			continue
		}
		for i := range windows {
			if windows[i].Index != oldWindow.Index {
				continue
			}
			if panes, err := tmux.ListPanes(target.Name, windows[i].Index); err == nil {
				windows[i].Panes = panes
				windows[i].Expanded = true
			}
		}
	}
	target.Windows = windows
	target.Expanded = true
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

type statusPollMsg struct{}

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

// statusPollTick returns a command that re-reads agent status files.
// Separate from animationTick so the 300ms animation never hits disk.
func statusPollTick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return statusPollMsg{}
	})
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sessionsMsg:
		m.sessions = msg.sessions
		m.sessionFilter.SetItems(m.sessions)
		m.selfSession = msg.selfSession
		m.sessionsLoaded = true
		m.saveSessionCache() // Cache for instant startup next time
		m.loadAgentStatuses()
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

	case statusPollMsg:
		return m, tea.Batch(m.pollAgentStatusesCmd(), statusPollTick())

	case agentStatusesMsg:
		m.claudeStatuses = msg.claude
		m.piStatuses = msg.pi
		return m, nil

	case projectsLoadedMsg:
		m.projectsLoading = false
		m.projectList.SetItems(msg.projects)
		return m, nil

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

// extractSessionName extracts a session name from a full path.
// Delegates to config.ExtractSessionName so the CLI and TUI agree on naming.
func (m *Model) extractSessionName(fullPath string) string {
	return config.ExtractSessionName(fullPath, m.config.ProjectDirs, m.config.ProjectDepth)
}

// extractDisplayPath extracts a display path from a full path.
// Delegates to config.ExtractDisplayPath for parity with extractSessionName.
func (m *Model) extractDisplayPath(fullPath string) string {
	return config.ExtractDisplayPath(fullPath, m.config.ProjectDirs, m.config.ProjectDepth)
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

// getSession returns the session for a given item (handles self session).
// Returns nil if the item's index no longer matches the session list.
func (m *Model) getSession(item Item) *tmux.Session {
	if item.IsSelf {
		return m.selfSession
	}
	if item.SessionIndex < 0 || item.SessionIndex >= len(m.sessions) {
		return nil
	}
	return &m.sessions[item.SessionIndex]
}

// windowAt returns the window for an item's indices, or nil if the indices
// no longer match the session's lazily loaded windows.
func (m *Model) windowAt(item Item) *tmux.Window {
	session := m.getSession(item)
	if session == nil || item.WindowIndex < 0 || item.WindowIndex >= len(session.Windows) {
		return nil
	}
	return &session.Windows[item.WindowIndex]
}

// paneAt returns the pane for an item's indices, or nil if out of range.
func (m *Model) paneAt(item Item) *tmux.Pane {
	window := m.windowAt(item)
	if window == nil || item.PaneIndex < 0 || item.PaneIndex >= len(window.Panes) {
		return nil
	}
	return &window.Panes[item.PaneIndex]
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
		m.setError("Layout script not found: %s", scriptPath)
		return
	}

	// Run layout script synchronously before switching to the session
	cmd := exec.Command(scriptPath, sessionName, workingDir)
	cmd.Env = append(os.Environ(),
		"TMUX_SESSION="+sessionName,
		"TMUX_WORKING_DIR="+workingDir,
	)
	if err := cmd.Run(); err != nil {
		m.setError("Layout %q failed: %v", m.config.Layout, err)
	}
}

// loadAgentStatuses refreshes the cached agent statuses for all sessions.
func (m *Model) loadAgentStatuses() {
	m.claudeStatuses = m.readAgentStatuses(agent.Claude, m.config.ClaudeStatusEnabled)
	m.piStatuses = m.readAgentStatuses(agent.Pi, m.config.PiStatusEnabled)
}

// agentStatusesMsg carries freshly polled agent statuses
type agentStatusesMsg struct {
	claude map[string]agent.Status
	pi     map[string]agent.Status
}

// pollAgentStatusesCmd is the periodic status refresh, run off the UI
// thread: re-reads status files, prunes files of sessions that no longer
// exist, and drops statuses whose agent process is gone — hooks don't fire
// on crash or SIGKILL, so a status file alone proves nothing.
func (m Model) pollAgentStatusesCmd() tea.Cmd {
	if !m.config.ClaudeStatusEnabled && !m.config.PiStatusEnabled {
		return nil
	}
	return func() tea.Msg {
		if m.sessionsLoaded {
			names := make([]string, 0, len(m.sessions)+1)
			for _, s := range m.allSessions() {
				names = append(names, s.Name)
			}
			for _, kind := range agent.Kinds {
				agent.CleanupStale(kind, m.config.CacheDir, names)
			}
		}

		claudeStatuses := m.readAgentStatuses(agent.Claude, m.config.ClaudeStatusEnabled)
		piStatuses := m.readAgentStatuses(agent.Pi, m.config.PiStatusEnabled)

		// Only spawn tmux/ps when something claims to be running
		if len(claudeStatuses)+len(piStatuses) > 0 {
			if panePIDs, err := tmux.PanePIDs(); err == nil {
				if live, err := agent.CheckLiveness(panePIDs); err == nil {
					dropDeadStatuses(agent.Claude, claudeStatuses, live, m.config.CacheDir)
					dropDeadStatuses(agent.Pi, piStatuses, live, m.config.CacheDir)
				}
			}
		}

		return agentStatusesMsg{claude: claudeStatuses, pi: piStatuses}
	}
}

// dropDeadStatuses removes statuses (and their files) for sessions where no
// matching agent process is running.
func dropDeadStatuses(kind agent.Kind, statuses map[string]agent.Status, live agent.Liveness, cacheDir string) {
	for name := range statuses {
		if !live.Alive(kind, name) {
			delete(statuses, name)
			agent.RemoveStatus(kind, name, cacheDir)
		}
	}
}

func (m *Model) readAgentStatuses(kind agent.Kind, enabled bool) map[string]agent.Status {
	statuses := make(map[string]agent.Status)
	if !enabled {
		return statuses
	}
	for _, s := range m.allSessions() {
		if status := agent.GetStatus(kind, s.Name, m.config.CacheDir); status.State != "" {
			statuses[s.Name] = status
		}
	}
	return statuses
}

// gitStatusTTL is how long a fetched git status stays fresh. Session
// reloads within this window reuse the cached result instead of spawning
// another round of git subprocesses.
const gitStatusTTL = 10 * time.Second

// fetchGitStatusesCmd returns commands that fetch git statuses in parallel
// Each session's status is fetched independently and updates the UI as soon as ready
func (m *Model) fetchGitStatusesCmd() tea.Cmd {
	if !m.config.GitStatusEnabled {
		return nil
	}

	// Only fetch sessions whose cached status has expired
	if m.gitStatusFetched == nil {
		m.gitStatusFetched = make(map[string]time.Time)
	}
	now := time.Now()
	var stale []tmux.Session
	for _, s := range m.allSessions() {
		if fetchedAt, ok := m.gitStatusFetched[s.Name]; ok && now.Sub(fetchedAt) < gitStatusTTL {
			continue
		}
		stale = append(stale, s)
	}
	if len(stale) == 0 {
		return nil
	}

	// Track which sessions we're waiting for
	m.gitStatusPending = make(map[string]bool)
	m.gitStatusShowLoading = false
	for _, s := range stale {
		m.gitStatusPending[s.Name] = true
		m.gitStatusFetched[s.Name] = now
	}

	// Create a command for each session - they run in parallel via tea.Batch
	cmds := make([]tea.Cmd, 0, len(stale)+1)

	// Add delayed loading indicator (500ms)
	cmds = append(cmds, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return gitStatusLoadingMsg{}
	}))

	for _, s := range stale {
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

func (m *Model) rebuildItems() {
	m.items = nil

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
		if !m.matchesFilter(session.Name) {
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

// sessionListWidth returns the width available for the session list.
// Now equivalent to contentWidth — actions moved to bottom bar.
func (m *Model) sessionListWidth() int {
	return m.contentWidth()
}

// rowWidth returns the width available for row content (accounts for scrollbar column).
func (m *Model) rowWidth() int {
	return m.contentWidth() - ui.ScrollbarColumnWidth
}

// statusLine returns a compact status string for the footer
func (m *Model) statusLine() string {
	total := len(m.sessions)
	if m.selfSession != nil {
		total++
	}
	return fmt.Sprintf("%d sessions", total)
}

// renderWithSidebar composes list content with a bottom action bar and simplified footer.
// listContent is the session/bookmark/project list string.
// actions is the mode-specific action set for the bottom bar.
// notification is the message to show in the footer.
// hints is a single-line keybind hint string.
// isError indicates notification is an error.
func (m *Model) renderWithSidebar(header, listContent string, actions []ui.Action, notification, hints string, isError bool) string {
	var b strings.Builder

	// Header (full width)
	b.WriteString(header)

	// List content (full width — no sidebar)
	b.WriteString(listContent)

	// Count content lines so far (header + list)
	content := b.String()
	contentLineCount := strings.Count(content, "\n")

	// Pad to push footer to bottom: target = contentHeight - dotted line - action bar - footer lines
	targetContentLines := m.contentHeight() - ui.ActionBarHeight - 5
	if targetContentLines > contentLineCount {
		padding := targetContentLines - contentLineCount
		for i := 0; i < padding; i++ {
			b.WriteString("\n")
		}
	}

	// Dotted separator above action bar
	b.WriteString(ui.RenderDottedBorder(m.contentWidth()))
	b.WriteString("\n")

	// Bottom action bar (2 rows of buttons)
	b.WriteString(ui.RenderButtonBar(actions, m.contentWidth()))
	b.WriteString("\n")

	// Footer at the very bottom
	b.WriteString(ui.RenderSimpleFooter(notification, hints, isError, m.width))

	return ui.AppStyle.Height(m.contentHeight()).Render(b.String())
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
