package ui

import (
	"fmt"
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
	{Label: "PRJ", Keybind: "C-p"},
	{Label: "BKM", Keybind: "C-b"},
	{Label: "KIL", Keybind: "C-x", Warning: true},
	{Label: "RMT", Keybind: "C-r"},
	{Label: "DWL", Keybind: "C-d"},
	{Label: "GIT", Keybind: "C-g"},
}

// BookmarkActions are the actions shown in ModeBookmarks
var BookmarkActions = []Action{
	{Label: "OPN", Keybind: "Ent"},
	{Label: "ADD", Keybind: "C-a"},
	{Label: "MV↑", Keybind: "C-p"},
	{Label: "MV↓", Keybind: "C-n"},
	{Label: "RMV", Keybind: "C-x", Warning: true},
	{Label: "BCK", Keybind: "Esc"},
}

// ProjectActions are the actions shown in ModePickDirectory
var ProjectActions = []Action{
	{Label: "SEL", Keybind: "Ent"},
	{Label: "BKM", Keybind: "C-a"},
	{Label: "RMV", Keybind: "C-x", Warning: true},
	{Label: "BCK", Keybind: "Esc"},
}

// CloneActions are the actions shown in ModeCloneRepo/ModeCloneChoice/ModeCloneURL
var CloneActions = []Action{
	{Label: "CLN", Keybind: "Ent"},
	{Label: "BCK", Keybind: "Esc"},
}

// CreateActions are the actions shown in ModeCreate/ModeCreatePath
var CreateActions = []Action{
	{Label: "CRT", Keybind: "Ent"},
	{Label: "BCK", Keybind: "Esc"},
}

// ButtonInnerWidth is the character width of button content (padded label/keybind)
const ButtonInnerWidth = 7 // " NEW  " or " C-n  " — generous padding

// RenderButton renders a 2-line action button with color-filled background.
// Line 1: 3-char ALL CAPS label (padded)
// Line 2: Keybind hint in dimmer color (padded)
func RenderButton(action Action, width int) string {
	labelStyle := ButtonStyle
	kbStyle := ButtonKeybindStyle
	if action.Warning {
		labelStyle = ButtonWarningStyle
		kbStyle = ButtonWarnKbStyle
	}

	label := fmt.Sprintf(" %-*s", width-1, action.Label)
	kb := fmt.Sprintf(" %-*s", width-1, action.Keybind)

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
	var lines []string
	for _, action := range actions {
		btn := RenderButton(action, ButtonInnerWidth)
		for _, line := range strings.Split(btn, "\n") {
			lines = append(lines, " "+line)
		}
	}

	content := strings.Join(lines, "\n")
	return RenderSectionBox("ACTIONS", content, boxWidth, SectionBoxOpts{})
}
