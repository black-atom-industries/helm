package model

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/github"
	"github.com/black-atom-industries/helm/internal/tmux"
	"github.com/black-atom-industries/helm/internal/ui"
)

// Clone-specific message types

type cloneReposLoadedMsg struct {
	repos []string
}

type cloneErrorMsg struct {
	err error
}

type cloneSuccessMsg struct {
	repoPath    string
	sessionName string
}

// handleCloneChoiceMode handles input in the clone choice sub-menu (URL vs My Repos)
func (m *Model) handleCloneChoiceMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Cancel):
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, keys.Up), key.Matches(msg, keys.Down):
		m.cloneChoiceCursor = 1 - m.cloneChoiceCursor

	case key.Matches(msg, keys.Select):
		if m.cloneChoiceCursor == 0 {
			// Enter URL
			m.input.Reset()
			m.input.Placeholder = "e.g. black-atom-industries/helm or git@github.com:black-atom-industries/helm.git"
			m.input.CharLimit = 256
			m.input.Width = m.width - 6 // account for padding
			m.input.Focus()
			m.cloneError = ""
			m.mode = ModeCloneURL
			return m, nil
		}
		// My repos — enter existing clone flow
		m.clonePendingFilter = m.filter
		m.filter = ""
		m.cloneList.Reset()
		m.cloneList.Clear()
		m.cloneError = ""
		m.cloneLoading = true
		m.cloneCloning = false
		m.mode = ModeCloneRepo
		return m, m.fetchAvailableReposCmd()

	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) handleCloneURLMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	// Handle confirmation state (reused from clone flow)
	if m.cloneSuccess {
		switch {
		case key.Matches(msg, keys.Select):
			m.applyLayout(m.cloneSuccessSession, m.cloneSuccessPath)
			if err := tmux.SwitchClient(m.cloneSuccessSession); err != nil {
				m.setError("Created but failed to switch: %v", err)
				m.mode = ModeNormal
				m.cloneSuccess = false
				return m, m.loadSessions
			}
			return m, tea.Quit
		case key.Matches(msg, keys.Cancel):
			m.mode = ModeNormal
			m.cloneSuccess = false
			return m, m.loadSessions
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, keys.Cancel):
		if m.cloneCloning {
			m.mode = ModeNormal
			m.cloneCloning = false
			m.cloneError = ""
			return m, nil
		}
		if m.cloneError != "" {
			m.cloneError = ""
			return m, nil
		}
		m.input.Blur()
		m.mode = ModeCloneChoice
		return m, nil

	case msg.Type == tea.KeyEnter:
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			return m, nil
		}
		ownerRepo, err := github.ResolveOwnerRepo(value)
		if err != nil {
			m.cloneError = err.Error()
			return m, nil
		}
		m.input.Blur()
		return m.cloneSelectedRepo(ownerRepo)

	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	}

	// Ignore ctrl key combinations — only pass regular typing to input
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

func (m *Model) handleCloneRepoMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	// Handle confirmation state
	if m.cloneSuccess {
		switch {
		case key.Matches(msg, keys.Select):
			// Apply layout and switch to the session
			m.applyLayout(m.cloneSuccessSession, m.cloneSuccessPath)
			if err := tmux.SwitchClient(m.cloneSuccessSession); err != nil {
				m.setError("Created but failed to switch: %v", err)
				m.mode = ModeNormal
				m.cloneSuccess = false
				return m, m.loadSessions
			}
			return m, tea.Quit

		case key.Matches(msg, keys.Cancel):
			// Go back to session list without switching
			m.mode = ModeNormal
			m.cloneSuccess = false
			return m, m.loadSessions
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, keys.Cancel):
		// If loading or cloning, just cancel and go back
		if m.cloneLoading || m.cloneCloning {
			m.mode = ModeNormal
			m.cloneLoading = false
			m.cloneCloning = false
			m.cloneError = ""
			return m, nil
		}
		// Clear filter first, then exit on second press
		if m.cloneList.Filter() != "" {
			m.cloneList.SetFilter("")
			return m, nil
		}
		// If there's an error, clear it and go back
		if m.cloneError != "" {
			m.mode = ModeNormal
			m.cloneError = ""
			return m, nil
		}
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, keys.Up):
		m.cloneList.MoveCursor(-1)

	case key.Matches(msg, keys.Down):
		m.cloneList.MoveCursor(1)

	case key.Matches(msg, keys.Select):
		if selected, ok := m.cloneList.SelectedItem(); ok && !m.cloneLoading && !m.cloneCloning && m.cloneError == "" {
			return m.cloneSelectedRepo(selected)
		}

	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case msg.Type == tea.KeyBackspace:
		filter := m.cloneList.Filter()
		if len(filter) > 0 && !m.cloneLoading && !m.cloneCloning {
			m.cloneList.SetFilter(filter[:len(filter)-1])
		}

	case msg.Type == tea.KeyRunes:
		if !m.cloneLoading && !m.cloneCloning && m.cloneError == "" {
			m.cloneList.SetFilter(m.cloneList.Filter() + string(msg.Runes))
		}
	}

	return m, nil
}

func (m *Model) cloneSelectedRepo(selected string) (tea.Model, tea.Cmd) {
	m.cloneCloning = true
	m.cloneCloningRepo = selected

	destPath := filepath.Join(m.cloneBasePath, selected)
	sessionName := sanitizeSessionName(selected)

	return m, func() tea.Msg {
		if err := github.CloneRepo(selected, destPath); err != nil {
			return cloneErrorMsg{err: err}
		}

		// Create tmux session
		if err := tmux.CreateSession(sessionName, destPath); err != nil {
			return cloneErrorMsg{err: fmt.Errorf("cloned but failed to create session: %w", err)}
		}

		return cloneSuccessMsg{
			repoPath:    destPath,
			sessionName: sessionName,
		}
	}
}

// fetchAvailableReposCmd fetches repos from GitHub
func (m *Model) fetchAvailableReposCmd() tea.Cmd {
	basePath := m.cloneBasePath
	return func() tea.Msg {
		// Check gh CLI
		if err := github.CheckGhCli(); err != nil {
			return cloneErrorMsg{err: err}
		}

		// Fetch available repos
		available, err := github.FetchAvailableRepos()
		if err != nil {
			return cloneErrorMsg{err: err}
		}

		// Get already cloned
		cloned, _ := config.ListClonedRepos(basePath)

		// Filter out cloned
		uncloned := config.FilterUncloned(available, cloned)

		return cloneReposLoadedMsg{repos: uncloned}
	}
}

// cloneMaxVisibleItems returns the actual number of items that can be shown
// in the clone repo view based on window height
func (m *Model) cloneMaxVisibleItems() int {
	contentH := m.contentHeight()
	if contentH > 0 {
		if available := contentH - 6; available > 0 { // header(3) + footer(3)
			return available
		}
	}
	return ui.DefaultVisibleItems
}

func (m Model) viewCloneChoice() string {
	var header strings.Builder
	var b strings.Builder

	header.WriteString(ui.RenderTitleBar(config.AppName, m.mode.String(), m.width))
	header.WriteString("\n")
	header.WriteString(ui.RenderPrompt("", m.width))
	header.WriteString("\n")
	header.WriteString(ui.RenderBorder(m.borderWidth()))
	header.WriteString("\n")

	options := []string{"Enter URL", "My repos"}

	for i, opt := range options {
		if i == m.cloneChoiceCursor {
			b.WriteString(ui.FilterStyle.Render("  "+opt) + "\n")
		} else {
			b.WriteString("  " + opt + "\n")
		}
	}

	// Padding is handled by renderWithSidebar
	return m.renderWithSidebar(header.String(), b.String(), ui.CloneActions, m.message, ui.UniversalHints, m.messageIsError)
}

func (m Model) viewCloneURL() string {
	var header strings.Builder
	var b strings.Builder

	header.WriteString(ui.RenderTitleBar(config.AppName, m.mode.String(), m.width))
	header.WriteString("\n")
	header.WriteString(ui.RenderPrompt("", m.width))
	header.WriteString("\n")
	header.WriteString(ui.RenderBorder(m.borderWidth()))
	header.WriteString("\n")

	if m.cloneSuccess {
		fmt.Fprintf(&b, "  Cloned: %s\n", m.cloneCloningRepo)
		fmt.Fprintf(&b, "  Session: %s\n", m.cloneSuccessSession)
		b.WriteString("\n")
		b.WriteString("  Switch to the new session?\n")
	} else if m.cloneCloning {
		fmt.Fprintf(&b, "  Cloning %s...\n", m.cloneCloningRepo)
	} else if m.cloneError != "" {
		b.WriteString(ui.ErrorMessageStyle.Render("  "+m.cloneError) + "\n")
		b.WriteString("  " + m.input.View() + "\n")
	} else {
		b.WriteString("  " + m.input.View() + "\n")
	}

	// Padding is handled by renderWithSidebar
	return m.renderWithSidebar(header.String(), b.String(), ui.CloneActions, m.message, ui.UniversalHints, m.messageIsError)
}

func (m Model) viewCloneRepo() string {
	var header strings.Builder
	var b strings.Builder

	header.WriteString(ui.RenderTitleBar(config.AppName, m.mode.String(), m.width))
	header.WriteString("\n")

	cloneFilter := m.cloneList.Filter()
	header.WriteString(ui.RenderPrompt(cloneFilter, m.width))
	header.WriteString("\n")

	header.WriteString(ui.RenderBorder(m.borderWidth()))
	header.WriteString("\n")

	// Content area
	if m.cloneSuccess {
		fmt.Fprintf(&b, "  Cloned: %s\n", m.cloneCloningRepo)
		fmt.Fprintf(&b, "  Session: %s\n", m.cloneSuccessSession)
		b.WriteString("\n")
		b.WriteString("  Switch to the new session?\n")
	} else if m.cloneLoading {
		b.WriteString("  Fetching available repositories...\n")
	} else if m.cloneCloning {
		fmt.Fprintf(&b, "  Cloning %s...\n", m.cloneCloningRepo)
	} else if m.cloneError != "" {
		b.WriteString(ui.ErrorMessageStyle.Render("  "+m.cloneError) + "\n")
	} else if m.cloneList.Len() == 0 {
		if cloneFilter != "" {
			b.WriteString("  No repositories matching filter\n")
		} else {
			b.WriteString("  No repositories available to clone\n")
		}
	} else {
		// Repository list - calculate max visible items
		maxItems := m.cloneMaxVisibleItems()
		m.cloneList.SetHeight(maxItems)

		visibleRepos := m.cloneList.VisibleItems()
		scrollOffset := m.cloneList.ScrollOffset()

		scrollbar := ui.ScrollbarChars(m.cloneList.Len(), m.cloneList.Height(), scrollOffset, len(visibleRepos))

		for i, repo := range visibleRepos {
			absoluteIdx := scrollOffset + i
			selected := m.cloneList.IsSelected(absoluteIdx)

			if i < len(scrollbar) {
				b.WriteString(scrollbar[i])
				b.WriteString(" ")
			}

			if selected {
				b.WriteString(ui.FilterStyle.Render(repo))
			} else {
				b.WriteString(repo)
			}
			b.WriteString("\n")
		}
	}

	// Padding is handled by renderWithSidebar
	return m.renderWithSidebar(header.String(), b.String(), ui.CloneActions, m.message, ui.UniversalHints, m.messageIsError)
}
