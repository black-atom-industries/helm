package ui

import (
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
	{Label: "EXPAND", Keybind: "C-h/l"},
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
	{Label: "EXPAND", Keybind: "C-h/l"},
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
const UniversalHints = "C-j/k ↕ Nav · Type filter · Esc Back"

// ButtonInnerWidth is the character width of button content
const ButtonInnerWidth = 12 // fits "BOOKMARKS" + centering padding

// centerText centers text within the given width, padding with spaces
func centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	totalPad := width - len(text)
	left := totalPad / 2
	right := totalPad - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

// RenderButtonLabel returns the styled label line for a single button.
func RenderButtonLabel(action Action, width int) string {
	style := ButtonStyle
	if action.Warning {
		style = ButtonWarningStyle
	}
	return style.Render(centerText(action.Label, width))
}

// RenderButtonKeybind returns the styled keybind line for a single button.
func RenderButtonKeybind(action Action, width int) string {
	style := ButtonKeybindStyle
	if action.Warning {
		style = ButtonWarnKbStyle
	}
	return style.Render(centerText(action.Keybind, width))
}

// renderButtonRow renders one row of buttons side-by-side with 1-space gaps,
// left-aligned, padded/truncated to exact width. Returns exactly 2 lines.
func renderButtonRow(actions []Action, width int) string {
	if len(actions) == 0 {
		padding := strings.Repeat(" ", width)
		return padding + "\n" + padding
	}

	// Build label row and keybind row
	var labelParts []string
	var kbParts []string
	for _, a := range actions {
		labelParts = append(labelParts, RenderButtonLabel(a, ButtonInnerWidth))
		kbParts = append(kbParts, RenderButtonKeybind(a, ButtonInnerWidth))
	}
	labelRow := strings.Join(labelParts, " ")
	kbRow := strings.Join(kbParts, " ")

	// Pad or truncate each row to exact width
	labelVisWidth := lipgloss.Width(labelRow)
	if labelVisWidth < width {
		labelRow += strings.Repeat(" ", width-labelVisWidth)
	} else if labelVisWidth > width {
		labelRow = lipgloss.NewStyle().MaxWidth(width).Render(labelRow)
	}

	kbVisWidth := lipgloss.Width(kbRow)
	if kbVisWidth < width {
		kbRow += strings.Repeat(" ", width-kbVisWidth)
	} else if kbVisWidth > width {
		kbRow = lipgloss.NewStyle().MaxWidth(width).Render(kbRow)
	}

	return labelRow + "\n" + kbRow
}

// RenderButtonBar renders the full action bar as two rows of buttons.
// Row 1 gets up to 5 actions, row 2 gets the rest.
// Always returns exactly 5 lines (ActionBarHeight).
func RenderButtonBar(actions []Action, width int) string {
	split := 5
	if len(actions) < split {
		split = len(actions)
	}
	row1 := renderButtonRow(actions[:split], width)
	row2 := renderButtonRow(actions[split:], width)
	return row1 + "\n\n" + row2
}
