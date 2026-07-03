package model

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/black-atom-industries/helm/internal/ui"
)

// helpAvailable reports whether "?" opens the help overlay in the current
// mode. Text-input modes need "?" as a literal character.
func (m *Model) helpAvailable() bool {
	switch m.mode {
	case ModeCreate, ModeCreatePath, ModeCloneURL, ModeHelp:
		return false
	default:
		return true
	}
}

// activeFilter returns the filter text being typed in the current mode.
func (m *Model) activeFilter() string {
	switch m.mode {
	case ModeNormal:
		return m.Filter()
	case ModeBookmarks:
		return m.bookmarkList.Filter()
	case ModePickDirectory:
		return m.projectList.Filter()
	case ModeCloneRepo:
		return m.cloneList.Filter()
	default:
		return ""
	}
}

// handleHelpMode closes the overlay on any close key, returning to the mode
// the user came from. All other keys are ignored.
func (m *Model) handleHelpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc", "enter", "q":
		m.mode = m.helpReturnMode
	}
	return m, nil
}

// helpActions returns the action set of the mode the overlay was opened from.
func (m *Model) helpActions() []ui.Action {
	switch m.helpReturnMode {
	case ModeBookmarks:
		return ui.BookmarkActions
	case ModePickDirectory, ModeConfirmRemoveFolder:
		return ui.ProjectActions
	case ModeCloneChoice, ModeCloneRepo:
		return ui.CloneActions
	case ModeConfirmKill:
		return ui.ConfirmKillActions
	default:
		return ui.SessionActions
	}
}

// viewHelp renders the centered keymap overlay.
func (m Model) viewHelp() string {
	overlay := ui.RenderHelpOverlay(m.helpActions())
	return ui.PlaceOverlay(m.width, m.contentHeight(), overlay)
}
