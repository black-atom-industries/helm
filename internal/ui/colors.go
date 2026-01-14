package ui

import "github.com/charmbracelet/lipgloss"

// Terminal-adaptive colors (ANSI 0-15)
// These adapt to the user's terminal theme
var (
	// ANSI 0-7: Standard colors
	// ANSI 8-15: Bright variants
	ColorPrimary   = lipgloss.Color("4") // Blue
	ColorSecondary = lipgloss.Color("7") // White/light gray
	ColorSuccess   = lipgloss.Color("2") // Green
	ColorWarning   = lipgloss.Color("3") // Yellow
	ColorError     = lipgloss.Color("1") // Red
	ColorDim       = lipgloss.Color("8") // Bright black (dark gray)
	ColorMagenta   = lipgloss.Color("5") // Magenta
)

// Hardcoded hex colors (terminal-independent)
// These remain consistent regardless of terminal theme
var (
	// Git status colors (One Dark inspired)
	HexGitFiles = lipgloss.Color("#61AFEF") // Blue
	HexGitAdd   = lipgloss.Color("#98C379") // Green
	HexGitDel   = lipgloss.Color("#E06C75") // Red

	// Claude Code colors
	HexClaudeOrange = lipgloss.Color("#DA7756") // Orange/Gold - CC header and accents
)
