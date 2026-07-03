package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Border and padding overhead for the app container
const (
	// AppBorderOverhead is the total cells used by border + padding per axis
	AppBorderOverheadX = 4 // left border + left padding + right padding + right border
	AppBorderOverheadY = 2 // top border + bottom border (no vertical padding)

	// ScrollbarColumnWidth is the space used by scrollbar + separator
	ScrollbarColumnWidth = 2 // scrollbar char + space
)

// Styles — initialized with dark defaults, call InitColors() to reinitialize
// for light mode or a Black Atom theme
var (
	// Container styles
	AppStyle          lipgloss.Style
	HeaderStyle       lipgloss.Style
	FooterStyle       lipgloss.Style
	MessageStyle      lipgloss.Style
	ErrorMessageStyle lipgloss.Style

	// Session row styles
	SessionStyle         lipgloss.Style
	SessionSelectedStyle lipgloss.Style

	// Window row styles (indented)
	WindowStyle         lipgloss.Style
	WindowSelectedStyle lipgloss.Style

	// Pane row styles (further indented)
	PaneStyle         lipgloss.Style
	PaneSelectedStyle lipgloss.Style

	// Text styles
	IndexStyle               lipgloss.Style
	IndexSelectedStyle       lipgloss.Style
	SessionNameStyle         lipgloss.Style
	SessionNameSelectedStyle lipgloss.Style
	WindowNameStyle          lipgloss.Style
	WindowNameSelectedStyle  lipgloss.Style

	// Pre-rendered icon strings
	ExpandedIcon          string
	ExpandedIconSelected  string
	CollapsedIcon         string
	CollapsedIconSelected string

	// Time styles
	TimeStyle         lipgloss.Style
	TimeSelectedStyle lipgloss.Style

	// Claude status styles
	ClaudeNewStyle           lipgloss.Style
	ClaudeWorkingStyle       lipgloss.Style
	ClaudeWaitingStyle       lipgloss.Style
	ClaudeWaitingUrgentStyle lipgloss.Style
	ClaudeIdleStyle          lipgloss.Style

	// Pi status styles
	PiNewStyle           lipgloss.Style
	PiWorkingStyle       lipgloss.Style
	PiWaitingStyle       lipgloss.Style
	PiWaitingUrgentStyle lipgloss.Style
	PiIdleStyle          lipgloss.Style

	// Git status styles
	GitFilesStyle   lipgloss.Style
	GitAddStyle     lipgloss.Style
	GitDelStyle     lipgloss.Style
	GitLoadingStyle lipgloss.Style

	// Input styles
	InputPromptStyle lipgloss.Style

	// Help styles
	HelpKeyStyle     lipgloss.Style
	HelpDescStyle    lipgloss.Style
	HelpSepStyle     lipgloss.Style
	HelpSectionStyle lipgloss.Style
	HelpBoxStyle     lipgloss.Style

	// Filter style
	FilterStyle lipgloss.Style

	// Border style
	BorderStyle    lipgloss.Style
	SeparatorStyle lipgloss.Style

	// Statusline style
	StatuslineStyle lipgloss.Style

	// Title bar style
	TitleBarStyle lipgloss.Style

	// Prompt style
	PromptStyle lipgloss.Style

	// State line style
	StateStyle lipgloss.Style

	// Table header styles
	TableHeaderStyle     lipgloss.Style
	TableHeaderTextStyle lipgloss.Style

	// CC header label style
	CCHeaderStyle lipgloss.Style

	// Pi header label style
	PiHeaderStyle lipgloss.Style

	// Self session styles (pinned current session)
	SelfIndexStyle         lipgloss.Style
	SelfIndexSelectedStyle lipgloss.Style
	SelfNameStyle          lipgloss.Style
	SelfNameSelectedStyle  lipgloss.Style

	// Action button styles (bottom button bar)
	AgentIdentStyle  lipgloss.Style
	HintStyle        lipgloss.Style
	HintWarningStyle lipgloss.Style

	// SelectedCellStyle is the uniform reverse-video treatment for cells
	// on the selected row (spacers, icons, git status)
	SelectedCellStyle lipgloss.Style
)

func init() {
	initStyles()
}

// selectedBase is the base treatment for selected/highlighted cells.
// In theme mode it paints only the selection background and, when a base
// style is passed, preserves that style's foreground so per-cell colors
// survive selection (the row gains a noticeable background, nothing more).
// Without a theme it falls back to terminal-native reverse video.
func selectedBase(base ...lipgloss.Style) lipgloss.Style {
	if HasTheme {
		s := lipgloss.NewStyle().Background(Colors.Bg.Selected)
		if len(base) > 0 {
			s = base[0].Background(Colors.Bg.Selected)
		}
		return s
	}
	return lipgloss.NewStyle().Reverse(true)
}

// initStyles rebuilds all styles from the current Colors values.
// Called at init and again by InitColors when appearance or theme changes.
func initStyles() {
	AppStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Colors.Fg.Border).
		Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Colors.Fg.Accent).
		Padding(0, 1)

	FooterStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Subtle).
		Padding(0, 1)

	MessageStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Accent).
		Padding(0, 1)

	ErrorMessageStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Error).
		Padding(0, 1)

	SessionStyle = lipgloss.NewStyle().
		Padding(0, 1)

	SessionSelectedStyle = selectedBase(SessionStyle).
		Padding(0, 1).
		Bold(true)

	WindowStyle = lipgloss.NewStyle().
		Padding(0, 1).
		PaddingLeft(10)

	WindowSelectedStyle = selectedBase(WindowStyle).
		Padding(0, 1).
		PaddingLeft(10).
		Bold(true)

	PaneStyle = lipgloss.NewStyle().
		Padding(0, 1).
		PaddingLeft(14)

	PaneSelectedStyle = selectedBase(PaneStyle).
		Padding(0, 1).
		PaddingLeft(14).
		Bold(true)

	IndexStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Subtle).
		Width(3)

	IndexSelectedStyle = selectedBase(IndexStyle).
		Bold(true)

	SessionNameStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.SessionName)

	SessionNameSelectedStyle = selectedBase(SessionNameStyle).
		Bold(true)

	WindowNameStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.WindowName)

	WindowNameSelectedStyle = selectedBase(WindowNameStyle).
		Bold(true)

	ExpandedIcon = lipgloss.NewStyle().Foreground(Colors.Fg.Accent).Render("▼")
	ExpandedIconSelected = selectedBase(lipgloss.NewStyle().Foreground(Colors.Fg.Accent)).Bold(true).Render("▼")
	CollapsedIcon = lipgloss.NewStyle().Foreground(Colors.Fg.Muted).Render("▶")
	CollapsedIconSelected = selectedBase(lipgloss.NewStyle().Foreground(Colors.Fg.Muted)).Bold(true).Render("▶")

	TimeStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Muted)

	TimeSelectedStyle = selectedBase(TimeStyle).
		Bold(true)

	ClaudeNewStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Muted).
		Bold(true)

	ClaudeWorkingStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.ClaudeWorking).
		Bold(true)

	ClaudeWaitingStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.ClaudeWaiting).
		Bold(true)

	ClaudeWaitingUrgentStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.ClaudeUrgent).
		Bold(true)

	ClaudeIdleStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.ClaudeIdle).
		Bold(true)

	PiNewStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Muted).
		Bold(true)

	PiWorkingStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.PiWorking).
		Bold(true)

	PiWaitingStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.PiWaiting).
		Bold(true)

	PiWaitingUrgentStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.PiUrgent).
		Bold(true)

	PiIdleStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.PiIdle).
		Bold(true)

	GitFilesStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.GitFiles)

	GitAddStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.GitAdd)

	GitDelStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.GitDel)

	GitLoadingStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Muted)

	InputPromptStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Accent)

	HelpKeyStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Accent).
		Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Muted)

	HelpSepStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Muted)

	// Help overlay (? keymap box)
	HelpSectionStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Subtle).
		Bold(true)

	HelpBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Fg.Border).
		Padding(1, 3)

	FilterStyle = selectedBase().
		Bold(true)

	BorderStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Border)

	SeparatorStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Separator)

	StatuslineStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Muted).
		Padding(0, 1)

	TitleBarStyle = lipgloss.NewStyle().
		Background(Colors.Bg.TitleBar).
		Foreground(Colors.Fg.TitleBar).
		Bold(true)

	PromptStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Accent).
		Padding(0, 1)

	StateStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Muted).
		Padding(0, 1)

	TableHeaderStyle = lipgloss.NewStyle().
		Padding(0, 1)

	TableHeaderTextStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Colors.Fg.TableHeader)

	CCHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Colors.Fg.ClaudeHeader)

	PiHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Colors.Fg.PiHeader)

	SelfIndexStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Accent).
		Width(3)

	SelfIndexSelectedStyle = selectedBase(SelfIndexStyle).
		Bold(true)

	SelfNameStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Muted)

	SelfNameSelectedStyle = selectedBase(SelfNameStyle).
		Bold(true)

	// Agent identity marker ("● claude") on window/pane rows
	AgentIdentStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Accent)

	// Footer hint bar: subtle key/action pairs; destructive actions in
	// the error color so they stand out from the dim line
	HintStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Subtle)

	HintWarningStyle = lipgloss.NewStyle().
		Foreground(Colors.Fg.Error)

	SelectedCellStyle = selectedBase()
}

// RenderBorder returns a horizontal border line
func RenderBorder(width int) string {
	return SeparatorStyle.Render(strings.Repeat("─", width))
}

// RenderDottedBorder returns a subtle dotted horizontal line
func RenderDottedBorder(width int) string {
	return SeparatorStyle.Render(strings.Repeat("·", width))
}

// RenderTitleBar renders the inverted title bar with logo on left and view name on right
func RenderTitleBar(logo, viewName string, width int) string {
	// Account for padding in AppStyle (1 on each side)
	innerWidth := width - AppBorderOverheadX
	if innerWidth < 10 {
		innerWidth = 40 // fallback for initial render
	}

	// Calculate spacing between logo and view name
	spacing := innerWidth - len(logo) - len(viewName) - 2 // -2 for padding spaces
	if spacing < 1 {
		spacing = 1
	}

	content := " " + logo + strings.Repeat(" ", spacing) + viewName + " "
	return TitleBarStyle.Width(innerWidth).Render(content)
}

// RenderPrompt renders the prompt line with optional filter text and cursor
func RenderPrompt(filter string, width int) string {
	innerWidth := width - AppBorderOverheadX
	if innerWidth < 10 {
		innerWidth = 40 // fallback for initial render
	}
	// Add block cursor indicator
	prompt := "> " + filter + "\u2588" // █ (full block)
	return PromptStyle.Width(innerWidth).Render(prompt)
}

// RenderSimpleFooter renders a 3-line footer: border + notification + single-line hints.
// The hints line arrives pre-styled (see RenderHintBar), so it is only padded here —
// recoloring it would fight the per-pair styles.
func RenderSimpleFooter(notification, hints string, isError bool, width int) string {
	innerWidth := width - AppBorderOverheadX
	if innerWidth < 10 {
		innerWidth = 40
	}
	var b strings.Builder

	b.WriteString(RenderBorder(innerWidth))
	b.WriteString("\n")

	if notification != "" {
		if isError {
			b.WriteString(ErrorMessageStyle.Width(innerWidth).Render(notification))
		} else {
			b.WriteString(MessageStyle.Width(innerWidth).Render(notification))
		}
	} else {
		b.WriteString(strings.Repeat(" ", innerWidth))
	}
	b.WriteString("\n")

	// MaxWidth instead of Width: the hint bar must truncate on narrow
	// viewports, never wrap — the footer is a fixed 3 lines.
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).MaxWidth(innerWidth).Render(hints))

	return b.String()
}

// ClaudeSpinnerFrames is the 4-frame braille spinner for "working" state
// Uses bottom 4 dots (positions 2,3,5,6) for better vertical alignment
var ClaudeSpinnerFrames = []string{"⠤", "⠆", "⠒", "⠰"}

// ClaudeWaitThreshold is the duration after which "waiting" escalates from ? to !
const ClaudeWaitThreshold = 5 * time.Minute

// ClaudeIdleThreshold is the duration after which "waiting" escalates from ! to Z
const ClaudeIdleThreshold = 15 * time.Minute

// StatusIconChar returns the raw (unstyled) status icon character.
// animationFrame cycles 0-3 for the spinner, waitDuration determines ? vs ! vs Z
func StatusIconChar(state string, animationFrame int, waitDuration time.Duration) string {
	switch state {
	case "new":
		// Don't show icon for "new" - it's just noise
		return " "
	case "working":
		// Animated spinner
		frame := animationFrame % len(ClaudeSpinnerFrames)
		return ClaudeSpinnerFrames[frame]
	case "waiting":
		// Time-based progression: ? → ! → Z
		if waitDuration >= ClaudeIdleThreshold {
			return "Z"
		}
		if waitDuration >= ClaudeWaitThreshold {
			return "!"
		}
		return "?"
	default:
		return " "
	}
}

// FormatClaudeIcon formats the Claude status as a single character icon
// animationFrame cycles 0-3 for the spinner, waitDuration determines ? vs !
func FormatClaudeIcon(state string, animationFrame int, waitDuration time.Duration, selected bool) string {
	char := StatusIconChar(state, animationFrame, waitDuration)
	var style lipgloss.Style
	switch char {
	case " ":
		if selected {
			return selectedBase().Render(" ")
		}
		return " "
	case "Z":
		style = ClaudeIdleStyle
	case "!":
		style = ClaudeWaitingUrgentStyle
	case "?":
		style = ClaudeWaitingStyle
	default:
		style = ClaudeWorkingStyle
	}
	if selected {
		style = selectedBase(style)
	}
	return style.Render(char)
}

// FormatPiIcon formats the Pi status as a single character icon
// animationFrame cycles 0-3 for the spinner, waitDuration determines ? vs !
func FormatPiIcon(state string, animationFrame int, waitDuration time.Duration, selected bool) string {
	char := StatusIconChar(state, animationFrame, waitDuration)
	var style lipgloss.Style
	switch char {
	case " ":
		if selected {
			return selectedBase().Render(" ")
		}
		return " "
	case "Z":
		style = PiIdleStyle
	case "!":
		style = PiWaitingUrgentStyle
	case "?":
		style = PiWaitingStyle
	default:
		style = PiWorkingStyle
	}
	if selected {
		style = selectedBase(style)
	}
	return style.Render(char)
}

// GitStatusColumnWidth is the fixed width for the git status column
const GitStatusColumnWidth = 20 // fits "99 files +99 -99"

// FormatGitStatus formats git status for display
// Returns empty string for clean repos (no indicator shown)
// Format: 3 files +44 -7 (files blue, +additions green, -deletions red)
func FormatGitStatus(dirty, additions, deletions int, selected bool) string {
	if dirty == 0 && additions == 0 && deletions == 0 {
		return ""
	}

	// On selected rows keep the git colors and paint the selection
	// background behind them — only the background changes, not the colors.
	filesStyle := GitFilesStyle
	addStyle := GitAddStyle
	delStyle := GitDelStyle
	if selected {
		filesStyle = selectedBase(filesStyle)
		addStyle = selectedBase(addStyle)
		delStyle = selectedBase(delStyle)
	}

	var parts []string

	if dirty > 0 {
		label := "files"
		if dirty == 1 {
			label = "file"
		}
		parts = append(parts, filesStyle.Render(fmt.Sprintf("%d %s", dirty, label)))
	}
	if additions > 0 {
		parts = append(parts, addStyle.Render(fmt.Sprintf("+%d", additions)))
	}
	if deletions > 0 {
		parts = append(parts, delStyle.Render(fmt.Sprintf("-%d", deletions)))
	}

	if len(parts) == 0 {
		return ""
	}

	// Join with styled spaces when selected
	if selected {
		spacer := SelectedCellStyle.Render(" ")
		return strings.Join(parts, spacer)
	}
	return strings.Join(parts, " ")
}

// GitStatusWidth returns the visual width of a git status string (without ANSI codes)
func GitStatusWidth(dirty, additions, deletions int) int {
	if dirty == 0 && additions == 0 && deletions == 0 {
		return 0
	}

	var parts []string

	if dirty > 0 {
		label := "files"
		if dirty == 1 {
			label = "file"
		}
		parts = append(parts, fmt.Sprintf("%d %s", dirty, label))
	}
	if additions > 0 {
		parts = append(parts, fmt.Sprintf("+%d", additions))
	}
	if deletions > 0 {
		parts = append(parts, fmt.Sprintf("-%d", deletions))
	}

	if len(parts) == 0 {
		return 0
	}

	return len(strings.Join(parts, " "))
}

// ScrollbarChars returns scrollbar characters for each visible line
// totalItems: total number of items in the list
// visibleItems: number of items currently visible
// scrollOffset: current scroll position (first visible item index)
// height: number of lines to render scrollbar for
func ScrollbarChars(totalItems, visibleItems, scrollOffset, height int) []string {
	result := make([]string, height)

	// No scrollbar needed if all items fit
	if totalItems <= visibleItems || height <= 0 {
		for i := range result {
			result[i] = " "
		}
		return result
	}

	// Calculate thumb size (minimum 1 line)
	thumbSize := (visibleItems * height) / totalItems
	if thumbSize < 1 {
		thumbSize = 1
	}

	// Calculate thumb position
	scrollRange := totalItems - visibleItems
	trackRange := height - thumbSize
	thumbPos := 0
	if scrollRange > 0 && trackRange > 0 {
		thumbPos = (scrollOffset * trackRange) / scrollRange
	}

	// Build scrollbar using dedicated scrollbar tokens
	trackChar := lipgloss.NewStyle().Foreground(Colors.Fg.ScrollbarTrack).Render("│")
	thumbChar := lipgloss.NewStyle().Foreground(Colors.Fg.ScrollbarThumb).Render("┃")

	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			result[i] = thumbChar
		} else {
			result[i] = trackChar
		}
	}

	return result
}
