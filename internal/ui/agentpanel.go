package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/black-atom-industries/helm/internal/agent"
)

// Agents panel layout constants.
const (
	// AgentPanelRatio is the panel's share of the content area, in percent —
	// the list gets the rest (65/35 split).
	AgentPanelRatio = 35
	// AgentPanelWidth is the panel's minimum content width in columns.
	AgentPanelWidth = 28
	// MinAgentPanelWidth / MinAgentPanelHeight are the viewport thresholds
	// below which the panel hides and the UI falls back to list-only.
	// Width is the real constraint (the list must keep room next to the
	// panel); the panel itself needs little height.
	MinAgentPanelWidth  = 100
	MinAgentPanelHeight = 15
)

// AgentEntry is one live agent instance shown in the panel.
type AgentEntry struct {
	Kind   string // "claude", "pi"
	Status agent.Status
}

// RenderAgentPanel renders the right-hand AGENTS panel for the selected
// session: one block per live agent instance (state dot, kind, elapsed,
// tool) plus a shared cwd line. width is the panel's content width (it
// grows with whatever the list doesn't need). The block is padded to height
// lines, each carrying the left rule separator.
func RenderAgentPanel(entries []AgentEntry, cwd string, width, height int) string {
	if width < AgentPanelWidth {
		width = AgentPanelWidth
	}
	lines := []string{HelpSectionStyle.Render("AGENTS"), ""}

	if len(entries) == 0 {
		lines = append(lines, HelpDescStyle.Render("none"))
	}
	for _, e := range entries {
		style := agentStateStyle(e.Status.State)
		elapsed := compactDuration(time.Since(e.Status.Timestamp))
		lines = append(lines, fmt.Sprintf("%s %-7s %s",
			style.Render("●"),
			e.Kind,
			style.Render(e.Status.State+" · "+elapsed)))
		if e.Status.Tool != "" {
			label := "tool"
			if e.Status.State != "working" {
				label = "last"
			}
			lines = append(lines, "  "+HelpDescStyle.Render(label)+" "+
				HelpKeyStyle.Render(truncateTo(e.Status.Tool, width-8)))
		}
		lines = append(lines, "")
	}

	if cwd != "" {
		lines = append(lines, HelpDescStyle.Render(truncateTo("cwd "+tildePath(cwd), width)))
	}

	// Pad to the list height so the rule runs the full column
	for len(lines) < height {
		lines = append(lines, "")
	}

	rule := BorderStyle.Render("│")
	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(rule)
		b.WriteString(" ")
		b.WriteString(line)
	}
	return b.String()
}

// agentStateStyle maps an agent state to its display style.
func agentStateStyle(state string) lipgloss.Style {
	switch state {
	case "working":
		return ClaudeWorkingStyle
	case "waiting":
		return ClaudeWaitingStyle
	default:
		return HelpDescStyle
	}
}

// compactDuration renders a duration as a short "7s / 2m / 3h / 1d" token.
func compactDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// truncateTo shortens s to max runes with an ellipsis.
func truncateTo(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max || max < 1 {
		return s
	}
	return string(runes[:max-1]) + "…"
}

// TruncateLines hard-truncates every line of s to width columns,
// ANSI-aware. Unlike lipgloss Width/MaxWidth, this never wraps — a wrapped
// line would break the line counting the layout relies on.
func TruncateLines(s string, width int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "")
	}
	return strings.Join(lines, "\n")
}

// tildePath shortens the home directory prefix to ~.
func tildePath(path string) string {
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
