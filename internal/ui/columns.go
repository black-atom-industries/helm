package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/nikbrunner/tsm/internal/claude"
	"github.com/nikbrunner/tsm/internal/git"
)

// RowLayout holds calculated column widths for consistent alignment across rows
type RowLayout struct {
	NameWidth      int
	GitStatusWidth int
}

// RowOpts contains options for rendering a generic row
// Required: Num, Name, Selected
// Optional: all others (use zero values to omit)
type RowOpts struct {
	// Required
	Num      int
	Name     string
	Selected bool

	// Optional - set to enable
	ShowLastIcon   bool           // Show ★ for first item
	IsFirst        bool           // Is this the first item (for ★)
	ShowExpandIcon bool           // Show ▸/▾ expand indicator
	Expanded       bool           // Expansion state
	LastActivity   *time.Time     // Show time ago if set
	GitStatus      *git.Status    // Show git status if set
	ClaudeStatus   *claude.Status // Show claude status if set
	AnimFrame      int            // Animation frame for claude status
}

// WindowRowOpts contains per-row options for rendering a window
type WindowRowOpts struct {
	Selected bool
}

// PaneRowOpts contains per-row options for rendering a pane (future use)
type PaneRowOpts struct {
	Selected bool
}

// Column component functions - each returns a styled string

// RenderIndex renders the session number (1-9)
func RenderIndex(num int, selected bool) string {
	label := fmt.Sprintf("%d", num)
	if selected {
		return IndexSelectedStyle.Render(label)
	}
	return IndexStyle.Render(label)
}

// RenderLastIcon renders the "last session" indicator
func RenderLastIcon(isFirst, selected bool) string {
	if isFirst {
		if selected {
			return LastIconSelected
		}
		return LastIcon
	}
	return " " // Fixed width placeholder
}

// RenderExpandIcon renders the expand/collapse indicator
func RenderExpandIcon(expanded, selected bool) string {
	if expanded {
		if selected {
			return ExpandedIconSelected
		}
		return ExpandedIcon
	}
	if selected {
		return CollapsedIconSelected
	}
	return CollapsedIcon
}

// RenderName renders a name with padding to a fixed width
func RenderName(name string, width int, selected bool, style lipgloss.Style) string {
	padded := fmt.Sprintf("%-*s", width, name)
	if selected {
		return style.Render(padded)
	}
	return padded
}

// RenderSessionName renders the session name
func RenderSessionName(name string, width int, selected bool) string {
	return RenderName(name, width, selected, SessionNameSelectedStyle)
}

// RenderWindowName renders the window name with index
func RenderWindowName(index int, name string, selected bool) string {
	text := fmt.Sprintf("%d: %s", index, name)
	if selected {
		return WindowNameSelectedStyle.Render(text)
	}
	return text
}

// RenderTimeAgo renders the time since last activity
func RenderTimeAgo(t time.Time, selected bool) string {
	timeAgo := FormatTimeAgo(t)
	padded := fmt.Sprintf("%-8s", timeAgo)
	if selected {
		return TimeSelectedStyle.Render(padded)
	}
	return TimeStyle.Render(padded)
}

// FormatTimeAgo formats a time as a human-readable "X ago" string
func FormatTimeAgo(t time.Time) string {
	d := time.Since(t)

	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

// RenderGitStatusColumn renders the git status with padding to a fixed width
func RenderGitStatusColumn(status *git.Status, maxWidth int) string {
	if maxWidth == 0 {
		return ""
	}

	if status == nil {
		return strings.Repeat(" ", maxWidth)
	}

	formatted := FormatGitStatus(status.Dirty, status.Ahead, status.Behind)
	actualWidth := GitStatusWidth(status.Dirty, status.Ahead, status.Behind)

	if actualWidth < maxWidth {
		return formatted + strings.Repeat(" ", maxWidth-actualWidth)
	}
	return formatted
}

// RenderClaudeStatusColumn renders the Claude status badge
func RenderClaudeStatusColumn(status *claude.Status, animFrame int) string {
	if status == nil {
		return ""
	}
	return FormatClaudeStatus(status.State, animFrame)
}

// SessionRowOpts wraps RowOpts with session-specific settings
type SessionRowOpts struct {
	RowOpts
}

// RenderSessionRow composes all columns into a complete session row
func RenderSessionRow(name string, lastActivity time.Time, layout RowLayout, opts SessionRowOpts) string {
	cols := []string{
		RenderIndex(opts.Num, opts.Selected),
		" ",
	}

	// Last icon (optional)
	if opts.ShowLastIcon {
		cols = append(cols, RenderLastIcon(opts.IsFirst, opts.Selected), " ")
	}

	// Expand icon (optional)
	if opts.ShowExpandIcon {
		cols = append(cols, RenderExpandIcon(opts.Expanded, opts.Selected), " ")
	}

	// Name (always shown)
	cols = append(cols, RenderSessionName(name, layout.NameWidth, opts.Selected))

	// Time ago (optional)
	if opts.LastActivity != nil {
		cols = append(cols, "  ", RenderTimeAgo(*opts.LastActivity, opts.Selected))
	}

	// Git status (optional column)
	if layout.GitStatusWidth > 0 && opts.GitStatus != nil {
		cols = append(cols, " ", RenderGitStatusColumn(opts.GitStatus, layout.GitStatusWidth))
	} else if layout.GitStatusWidth > 0 {
		// Pad for alignment even when no git status
		cols = append(cols, " ", strings.Repeat(" ", layout.GitStatusWidth))
	}

	// Claude status (optional column)
	if opts.ClaudeStatus != nil {
		claudeStr := RenderClaudeStatusColumn(opts.ClaudeStatus, opts.AnimFrame)
		if claudeStr != "" {
			cols = append(cols, " ", claudeStr)
		}
	}

	content := strings.Join(cols, "")
	return SessionStyle.Render(content)
}

// RenderBookmarkRow composes a bookmark row (simpler than session row)
func RenderBookmarkRow(name string, layout RowLayout, opts RowOpts) string {
	cols := []string{
		RenderIndex(opts.Num, opts.Selected),
		" ",
		RenderSessionName(name, layout.NameWidth, opts.Selected),
	}

	// Git status (optional)
	if layout.GitStatusWidth > 0 && opts.GitStatus != nil {
		cols = append(cols, " ", RenderGitStatusColumn(opts.GitStatus, layout.GitStatusWidth))
	}

	content := strings.Join(cols, "")
	return SessionStyle.Render(content)
}

// RenderWindowRow composes a window row
func RenderWindowRow(index int, name string, opts WindowRowOpts) string {
	content := RenderWindowName(index, name, opts.Selected)
	if opts.Selected {
		return WindowSelectedStyle.Render(content)
	}
	return WindowStyle.Render(content)
}

// RenderPaneRow composes a pane row (future use for tsm-xdn)
// Panes will be indented further than windows
func RenderPaneRow(index int, title string, opts PaneRowOpts) string {
	text := fmt.Sprintf("%d: %s", index, title)
	if opts.Selected {
		// TODO: Add PaneSelectedStyle when implementing tsm-xdn
		return WindowSelectedStyle.PaddingLeft(14).Render(text)
	}
	// TODO: Add PaneStyle when implementing tsm-xdn
	return WindowStyle.PaddingLeft(14).Render(text)
}

// ItemDepth represents the hierarchy level of an item
// This design allows easy extension for panes (tsm-xdn)
type ItemDepth int

const (
	DepthSession ItemDepth = 0
	DepthWindow  ItemDepth = 1
	DepthPane    ItemDepth = 2 // Future use
)

// IndentForDepth returns the left padding for a given depth level
func IndentForDepth(depth ItemDepth) int {
	switch depth {
	case DepthSession:
		return 0
	case DepthWindow:
		return 10 // Matches current WindowStyle.PaddingLeft
	case DepthPane:
		return 14 // Further indented for panes
	default:
		return 0
	}
}
