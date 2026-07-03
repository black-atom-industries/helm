package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// navRows are the universal navigation keys shown in the help overlay.
var navRows = [][2]string{
	{"C-j/k ↑↓", "move"},
	{"C-h/l ←→", "collapse / expand"},
	{"1-9", "jump to session"},
	{"type", "fuzzy filter"},
	{"esc", "clear filter / back"},
}

// generalRows are the app-level keys shown in the help overlay.
var generalRows = [][2]string{
	{"C-c", "quit"},
	{"?", "close help"},
}

// RenderHelpOverlay renders the full keymap as a bordered box for the help
// mode: universal navigation, the current mode's actions, and general keys.
func RenderHelpOverlay(actions []Action) string {
	var b strings.Builder

	writeSection := func(title string, rows [][2]string) {
		b.WriteString(HelpSectionStyle.Render(title))
		b.WriteString("\n")
		for _, row := range rows {
			fmt.Fprintf(&b, "%s %s\n",
				HelpKeyStyle.Render(fmt.Sprintf("%-11s", row[0])),
				HelpDescStyle.Render(row[1]))
		}
	}

	actionRows := make([][2]string, 0, len(actions))
	for _, a := range actions {
		actionRows = append(actionRows, [2]string{a.Keybind, strings.ToLower(a.Label)})
	}

	writeSection("NAVIGATE", navRows)
	b.WriteString("\n")
	writeSection("ACTIONS", actionRows)
	b.WriteString("\n")
	writeSection("GENERAL", generalRows)

	return HelpBoxStyle.Render(strings.TrimRight(b.String(), "\n"))
}

// PlaceOverlay centers content in the given viewport dimensions.
func PlaceOverlay(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
