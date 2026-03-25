package ui

import (
	"strings"
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
	{Label: "NEW", Keybind: "C-n"},
	{Label: "PROJECTS", Keybind: "C-p"},
	{Label: "BOOKMARKS", Keybind: "C-b"},
	{Label: "KILL", Keybind: "C-x", Warning: true},
	{Label: "REMOTE", Keybind: "C-r"},
	{Label: "DOWNLOAD", Keybind: "C-d"},
	{Label: "LAZYGIT", Keybind: "C-g"},
}

// BookmarkActions are the actions shown in ModeBookmarks
var BookmarkActions = []Action{
	{Label: "OPEN", Keybind: "Enter"},
	{Label: "ADD", Keybind: "C-a"},
	{Label: "MOVE UP", Keybind: "C-p"},
	{Label: "MOVE DOWN", Keybind: "C-n"},
	{Label: "REMOVE", Keybind: "C-x", Warning: true},
	{Label: "BACK", Keybind: "Esc"},
}

// ProjectActions are the actions shown in ModePickDirectory
var ProjectActions = []Action{
	{Label: "SELECT", Keybind: "Enter"},
	{Label: "BOOKMARK", Keybind: "C-a"},
	{Label: "REMOVE", Keybind: "C-x", Warning: true},
	{Label: "BACK", Keybind: "Esc"},
}

// CloneActions are the actions shown in ModeCloneRepo/ModeCloneChoice/ModeCloneURL
var CloneActions = []Action{
	{Label: "CLONE", Keybind: "Enter"},
	{Label: "BACK", Keybind: "Esc"},
}

// CreateActions are the actions shown in ModeCreate/ModeCreatePath
var CreateActions = []Action{
	{Label: "CREATE", Keybind: "Enter"},
	{Label: "BACK", Keybind: "Esc"},
}

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

// RenderButton renders a 2-line action button with color-filled background.
// Line 1: ALL CAPS label (centered)
// Line 2: Keybind hint in dimmer color (centered)
func RenderButton(action Action, width int) string {
	labelStyle := ButtonStyle
	kbStyle := ButtonKeybindStyle
	if action.Warning {
		labelStyle = ButtonWarningStyle
		kbStyle = ButtonWarnKbStyle
	}

	// Center the label and keybind within the button width
	label := centerText(action.Label, width)
	kb := centerText(action.Keybind, width)

	return labelStyle.Render(label) + "\n" + kbStyle.Render(kb)
}

// RenderButtonColumn renders buttons in a single vertical column.
// Each button is 2 lines (label + keybind), separated by a blank line.
func RenderButtonColumn(actions []Action, colWidth int) string {
	var rows []string
	for _, action := range actions {
		rows = append(rows, RenderButton(action, colWidth))
	}
	return strings.Join(rows, "\n\n")
}

// SidebarBoxWidth returns the total width of the sidebar section box.
// Single column: left border + left pad + button + right pad + right border
func SidebarBoxWidth() int {
	return ButtonInnerWidth + 4 // │ + " " + button(7) + " " + │
}

// RenderSidebar renders the sidebar: just the ACTIONS section box.
// Status info is now in the footer instead.
func RenderSidebar(actions []Action, height int) string {
	boxWidth := SidebarBoxWidth()

	// Build button column — each line gets " " prefix for left padding inside box
	// Blank line between each button for vertical breathing room
	var lines []string
	for i, action := range actions {
		btn := RenderButton(action, ButtonInnerWidth)
		for _, line := range strings.Split(btn, "\n") {
			lines = append(lines, " "+line)
		}
		if i < len(actions)-1 {
			lines = append(lines, "") // gap between buttons
		}
	}

	content := strings.Join(lines, "\n")
	return RenderSectionBox("ACTIONS", content, boxWidth, SectionBoxOpts{})
}
