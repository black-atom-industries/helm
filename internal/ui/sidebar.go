package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Action defines a sidebar action button
type Action struct {
	Label   string // 3-char ALL CAPS label (e.g., "NEW", "KIL")
	Keybind string // Keybind hint (e.g., "C-n", "C-x")
	Warning bool   // Use warning/danger style instead of accent
}

// Mode-specific action sets

// SessionActions are the actions shown in ModeNormal (session list)
var SessionActions = []Action{
	{Label: "SWITCH", Keybind: "Enter"},
	{Label: "BOOKMARKS", Keybind: "C-b"},
	{Label: "PROJECTS", Keybind: "C-p"},
	{Label: "DOWNLOAD", Keybind: "C-d"},
	{Label: "NEW", Keybind: "C-n"},
	{Label: "LAZYGIT", Keybind: "C-g"},
	{Label: "REMOTE", Keybind: "C-r"},
	{Label: "KILL", Keybind: "C-x", Warning: true},
}

// BookmarkActions are the actions shown in ModeBookmarks
var BookmarkActions = []Action{
	{Label: "OPEN", Keybind: "Enter"},
	{Label: "ADD", Keybind: "C-a"},
	{Label: "MOVE UP", Keybind: "C-p"},
	{Label: "MOVE DOWN", Keybind: "C-n"},
	{Label: "REMOVE", Keybind: "C-x", Warning: true},
}

// ProjectActions are the actions shown in ModePickDirectory
var ProjectActions = []Action{
	{Label: "SELECT", Keybind: "Enter"},
	{Label: "BOOKMARK", Keybind: "C-a"},
	{Label: "REMOVE", Keybind: "C-x", Warning: true},
}

// CloneActions are the actions shown in ModeCloneRepo/ModeCloneChoice/ModeCloneURL
var CloneActions = []Action{
	{Label: "CLONE", Keybind: "Enter"},
}

// CreateActions are the actions shown in ModeCreate/ModeCreatePath
var CreateActions = []Action{
	{Label: "CREATE", Keybind: "Enter"},
}

// ConfirmKillActions are the actions shown in ModeConfirmKill
var ConfirmKillActions = []Action{
	{Label: "CONFIRM", Keybind: "C-x", Warning: true},
	{Label: "CANCEL", Keybind: "Esc"},
}

// UniversalHints is the footer hint line shown in all modes
const UniversalHints = "C-j/k ↕ Nav · C-h/l ↔ Expand · Type filter · Esc Back"

// RenderButton renders a single action button as "LABEL [keybind]" in one line.
func RenderButton(action Action) string {
	text := fmt.Sprintf(" %s [%s] ", action.Label, action.Keybind)
	style := ButtonStyle
	if action.Warning {
		style = ButtonWarningStyle
	}
	return style.Render(text)
}

// renderButtonRow renders all buttons side-by-side with 1-space gaps,
// padded to exact width. Returns exactly 1 line.
func renderButtonRow(actions []Action, width int) string {
	if len(actions) == 0 {
		return strings.Repeat(" ", width)
	}

	var parts []string
	for _, a := range actions {
		parts = append(parts, RenderButton(a))
	}
	row := strings.Join(parts, " ")

	visWidth := lipgloss.Width(row)
	if visWidth < width {
		row += strings.Repeat(" ", width-visWidth)
	} else if visWidth > width {
		row = lipgloss.NewStyle().MaxWidth(width).Render(row)
	}

	return row
}

// RenderButtonBar renders the full action bar as one or two rows of compact buttons,
// with a blank line between rows. Row 1 gets up to 5 actions, row 2 gets the rest.
// Returns 1-3 lines depending on whether a second row is needed.
func RenderButtonBar(actions []Action, width int) string {
	split := 5
	if len(actions) < split {
		split = len(actions)
	}
	row1 := renderButtonRow(actions[:split], width)
	if split < len(actions) {
		row2 := renderButtonRow(actions[split:], width)
		return row1 + "\n\n" + row2
	}
	return row1
}
