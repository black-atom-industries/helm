package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/black-atom-industries/helm/internal/ui/theme"
)

// HasTheme reports whether a Black Atom theme is active. When false,
// selection and buttons fall back to terminal-native reverse video.
var HasTheme bool

// ANSI 16 colors (terminal-adaptive)
type TermColors struct {
	Black         lipgloss.Color
	Red           lipgloss.Color
	Green         lipgloss.Color
	Yellow        lipgloss.Color
	Blue          lipgloss.Color
	Magenta       lipgloss.Color
	Cyan          lipgloss.Color
	White         lipgloss.Color
	BrightBlack   lipgloss.Color
	BrightRed     lipgloss.Color
	BrightGreen   lipgloss.Color
	BrightYellow  lipgloss.Color
	BrightBlue    lipgloss.Color
	BrightMagenta lipgloss.Color
	BrightCyan    lipgloss.Color
	BrightWhite   lipgloss.Color
}

var termColors = TermColors{
	Black:         lipgloss.Color("0"),
	Red:           lipgloss.Color("1"),
	Green:         lipgloss.Color("2"),
	Yellow:        lipgloss.Color("3"),
	Blue:          lipgloss.Color("4"),
	Magenta:       lipgloss.Color("5"),
	Cyan:          lipgloss.Color("6"),
	White:         lipgloss.Color("7"),
	BrightBlack:   lipgloss.Color("8"),
	BrightRed:     lipgloss.Color("9"),
	BrightGreen:   lipgloss.Color("10"),
	BrightYellow:  lipgloss.Color("11"),
	BrightBlue:    lipgloss.Color("12"),
	BrightMagenta: lipgloss.Color("13"),
	BrightCyan:    lipgloss.Color("14"),
	BrightWhite:   lipgloss.Color("15"),
}

// Hard coded hex colors (terminal-independent)
// Deliberate exceptions: the git status colors (semantic, should look the
// same everywhere) and the Claude brand orange (no ANSI equivalent).
type HardCodedPalette struct {
	ClaudeOrange lipgloss.Color // CC/Pi header brand color
	Blue         lipgloss.Color // Git file count
	Green        lipgloss.Color // Git additions
	Red          lipgloss.Color // Git deletions
}

// Hex colors (terminal-independent)
var hardCodedColor = struct {
	Light HardCodedPalette
	Dark  HardCodedPalette
}{
	Light: HardCodedPalette{
		ClaudeOrange: lipgloss.Color("#e2663c"),
		Blue:         lipgloss.Color("#0997ee"),
		Green:        lipgloss.Color("#67a52a"),
		Red:          lipgloss.Color("#e35f6d"),
	},
	Dark: HardCodedPalette{
		ClaudeOrange: lipgloss.Color("#f38b6a"),
		Blue:         lipgloss.Color("#5cb2fb"),
		Green:        lipgloss.Color("#89be61"),
		Red:          lipgloss.Color("#f4868c"),
	},
}

// FgColors defines all foreground (text) colors
type FgColors struct {
	Default   lipgloss.TerminalColor // Terminal default text
	Muted     lipgloss.TerminalColor // De-emphasized text
	Accent    lipgloss.TerminalColor // Primary accent
	Subtle    lipgloss.TerminalColor // Secondary/subtle text
	Error     lipgloss.TerminalColor // Error text
	Border    lipgloss.TerminalColor // Border characters
	Separator lipgloss.TerminalColor // Horizontal separator lines (dotted)

	// Theme-mode only (unused in reverse-video fallback)
	Selected lipgloss.TerminalColor // Text on selected row background

	// Title bar
	TitleBar lipgloss.TerminalColor // Text on title bar

	// Table
	TableHeader lipgloss.TerminalColor // Column headers
	SessionName lipgloss.TerminalColor // Unselected session names
	WindowName  lipgloss.TerminalColor // Unselected window names

	// Claude status
	ClaudeHeader  lipgloss.TerminalColor // "CC" label
	ClaudeWorking lipgloss.TerminalColor // Spinner
	ClaudeWaiting lipgloss.TerminalColor // "?" icon
	ClaudeUrgent  lipgloss.TerminalColor // "!" icon
	ClaudeIdle    lipgloss.TerminalColor // "Z" icon

	// Git status
	GitFiles lipgloss.TerminalColor // File count
	GitAdd   lipgloss.TerminalColor // Additions
	GitDel   lipgloss.TerminalColor // Deletions

	// Pi status
	PiHeader  lipgloss.TerminalColor // "Pi" label
	PiWorking lipgloss.TerminalColor // Spinner
	PiWaiting lipgloss.TerminalColor // "?" icon
	PiUrgent  lipgloss.TerminalColor // "!" icon
	PiIdle    lipgloss.TerminalColor // "Z" icon

	// Scrollbar
	ScrollbarTrack lipgloss.TerminalColor // Track (subtle background line)
	ScrollbarThumb lipgloss.TerminalColor // Thumb (visible position indicator)
}

// BgColors defines all background colors
// Without an active theme, selected rows and action buttons use reverse
// video (terminal-native) and the theme-mode fields stay unset.
type BgColors struct {
	Default  lipgloss.TerminalColor // Terminal default (none)
	TitleBar lipgloss.TerminalColor // Title bar background

	// Theme-mode only (unused in reverse-video fallback)
	Selected lipgloss.TerminalColor // Selected row background
	Negative lipgloss.TerminalColor // Warning/danger button background
}

// *****************************************************************************
// Dark theme
// *****************************************************************************

func darkFg() FgColors {
	tc := termColors
	hc := hardCodedColor.Dark
	return FgColors{
		Default:   lipgloss.NoColor{},
		Muted:     tc.BrightBlack,
		Accent:    tc.Blue,
		Subtle:    tc.White,
		Error:     tc.Red,
		Border:    tc.BrightBlack,
		Separator: tc.White,

		TitleBar: tc.BrightWhite,

		TableHeader: lipgloss.NoColor{},
		SessionName: lipgloss.NoColor{},
		WindowName:  lipgloss.NoColor{},

		ClaudeHeader:  hc.ClaudeOrange,
		ClaudeWorking: tc.Yellow,
		ClaudeWaiting: tc.Green,
		ClaudeUrgent:  tc.Red,
		ClaudeIdle:    tc.Blue,

		GitFiles: hc.Blue,
		GitAdd:   hc.Green,
		GitDel:   hc.Red,

		PiHeader:  hc.ClaudeOrange, // Same orange as CC
		PiWorking: tc.Yellow,
		PiWaiting: tc.Green,
		PiUrgent:  tc.Red,
		PiIdle:    tc.Blue,

		ScrollbarTrack: tc.BrightBlack,
		ScrollbarThumb: tc.White,
	}
}

func darkBg() BgColors {
	return BgColors{
		Default:  lipgloss.NoColor{},
		TitleBar: termColors.Black,
	}
}

// *****************************************************************************
// Light theme
// *****************************************************************************

func lightBg() BgColors {
	return BgColors{
		Default:  lipgloss.NoColor{},
		TitleBar: termColors.BrightWhite,
	}
}

func lightFg() FgColors {
	tc := termColors
	hc := hardCodedColor.Light
	return FgColors{
		Default:   lipgloss.NoColor{},
		Muted:     tc.BrightBlack,
		Accent:    tc.Blue,
		Subtle:    tc.BrightBlack,
		Error:     tc.Red,
		Border:    tc.BrightBlack,
		Separator: tc.BrightWhite,

		TitleBar: tc.Black,

		TableHeader: lipgloss.NoColor{},
		SessionName: lipgloss.NoColor{},
		WindowName:  lipgloss.NoColor{},

		ClaudeHeader:  hc.ClaudeOrange,
		ClaudeWorking: tc.Yellow,
		ClaudeWaiting: tc.Green,
		ClaudeUrgent:  tc.Red,
		ClaudeIdle:    tc.Blue,

		GitFiles: hc.Blue,
		GitAdd:   hc.Green,
		GitDel:   hc.Red,

		PiHeader:  hc.ClaudeOrange, // Same orange as CC
		PiWorking: tc.Yellow,
		PiWaiting: tc.Green,
		PiUrgent:  tc.Red,
		PiIdle:    tc.Blue,

		ScrollbarTrack: tc.BrightWhite,
		ScrollbarThumb: tc.White,
	}
}

// *****************************************************************************
// Black Atom theme mode
// *****************************************************************************

func themeFg(t theme.Theme) FgColors {
	// Brand orange has no theme token; pick by the theme's own appearance
	hc := hardCodedColor.Dark
	if t.Appearance == "light" {
		hc = hardCodedColor.Light
	}
	return FgColors{
		Default:   lipgloss.Color(t.FgDefault),
		Muted:     lipgloss.Color(t.FgSubtle),
		Accent:    lipgloss.Color(t.FgAccent),
		Subtle:    lipgloss.Color(t.FgDefault),
		Error:     lipgloss.Color(t.FgNegative),
		Border:    lipgloss.Color(t.FgSubtle),
		Separator: lipgloss.Color(t.FgDisabled),

		Selected: lipgloss.Color(t.FgDefault),

		TitleBar: lipgloss.Color(t.FgContrast),

		TableHeader: lipgloss.Color(t.FgDefault),
		SessionName: lipgloss.Color(t.FgDefault),
		WindowName:  lipgloss.Color(t.FgDefault),

		ClaudeHeader:  hc.ClaudeOrange,
		ClaudeWorking: lipgloss.Color(t.PaletteYellow),
		ClaudeWaiting: lipgloss.Color(t.PaletteGreen),
		ClaudeUrgent:  lipgloss.Color(t.PaletteRed),
		ClaudeIdle:    lipgloss.Color(t.PaletteBlue),

		GitFiles: lipgloss.Color(t.FgInfo),
		GitAdd:   lipgloss.Color(t.FgAdd),
		GitDel:   lipgloss.Color(t.FgDelete),

		PiHeader:  hc.ClaudeOrange, // Same orange as CC
		PiWorking: lipgloss.Color(t.PaletteYellow),
		PiWaiting: lipgloss.Color(t.PaletteGreen),
		PiUrgent:  lipgloss.Color(t.PaletteRed),
		PiIdle:    lipgloss.Color(t.PaletteBlue),

		ScrollbarTrack: lipgloss.Color(t.FgDisabled),
		ScrollbarThumb: lipgloss.Color(t.FgAccent),
	}
}

func themeBg(t theme.Theme) BgColors {
	return BgColors{
		Default:  lipgloss.NoColor{}, // Inherit terminal background
		TitleBar: lipgloss.Color(t.BgContrast),
		Selected: lipgloss.Color(t.BgSelection),
		Negative: lipgloss.Color(t.BgNegative),
	}
}

// Colors is the global color configuration
// Initialized with dark mode defaults, call InitColors to adapt for light mode
var Colors = struct {
	Fg FgColors
	Bg BgColors
}{
	Fg: darkFg(),
	Bg: darkBg(),
}

// InitColors sets the color palette and reinitializes all styles.
// Must be called before the TUI renders.
//
// themeKey selects a Black Atom theme (e.g. "black-atom-jpn-koyo-yoru");
// when set and registered it wins over appearance, which then only picks
// the ANSI fallback palette ("light" or "dark"). Returns false if themeKey
// was non-empty but unknown (callers may warn); the fallback is used then.
func InitColors(appearance string, themeKey string) bool {
	known := true
	HasTheme = false
	if themeKey != "" {
		if t, ok := theme.Get(themeKey); ok {
			Colors.Fg = themeFg(t)
			Colors.Bg = themeBg(t)
			HasTheme = true
		} else {
			known = false
		}
	}
	if !HasTheme {
		if appearance == "light" {
			Colors.Fg = lightFg()
			Colors.Bg = lightBg()
		} else {
			Colors.Fg = darkFg()
			Colors.Bg = darkBg()
		}
	}
	initStyles()
	return known
}
