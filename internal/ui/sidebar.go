package ui

import (
	"fmt"
	"strings"
)

// Action defines a mode action shown in the footer hint bar
type Action struct {
	Label   string // Action label (e.g., "NEW", "KILL")
	Keybind string // Keybind hint (e.g., "C-n", "C-x")
	Warning bool   // Use warning/danger style instead of subtle
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

// UniversalHints lists the navigation keys; shown in the ? help overlay.
const UniversalHints = "C-j/k ↕ Nav · C-h/l ↔ Expand · Type filter · Esc Back"

// RenderHintBar renders the mode's actions as a single lazygit-style hint
// line: "⏎ switch  C-b bookmarks … C-x kill  ? help". Each pair carries its
// own style (subtle; warning color for destructive actions), so the footer
// must not recolor the line. withHelp appends the "? help" hint — false in
// text-input modes where "?" is a literal character.
func RenderHintBar(actions []Action, withHelp bool) string {
	parts := make([]string, 0, len(actions)+1)
	for _, a := range actions {
		key := a.Keybind
		if key == "Enter" {
			key = "⏎"
		}
		pair := fmt.Sprintf("%s %s", key, strings.ToLower(a.Label))
		if a.Warning {
			parts = append(parts, HintWarningStyle.Render(pair))
		} else {
			parts = append(parts, HintStyle.Render(pair))
		}
	}
	if withHelp {
		parts = append(parts, HintStyle.Render("? help"))
	}
	return strings.Join(parts, "  ")
}
