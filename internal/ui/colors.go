package ui

import "github.com/charmbracelet/lipgloss"

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
type HardCodedPalette struct {
	ClaudeOrange lipgloss.Color
	Blue         lipgloss.Color
	Green        lipgloss.Color
	Red          lipgloss.Color
	Yellow       lipgloss.Color
	SelectedBg   lipgloss.Color // Accent background for selected rows
	ButtonAccent lipgloss.Color // Button accent background
	ButtonWarn   lipgloss.Color // Button warning/danger background
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
		Yellow:       lipgloss.Color("#b98700"),
		SelectedBg:   lipgloss.Color("#f0d8c0"),
		ButtonAccent: lipgloss.Color("#e2663c"),
		ButtonWarn:   lipgloss.Color("#e35f6d"),
	},
	Dark: HardCodedPalette{
		ClaudeOrange: lipgloss.Color("#f38b6a"),
		Blue:         lipgloss.Color("#5cb2fb"),
		Green:        lipgloss.Color("#89be61"),
		Red:          lipgloss.Color("#f4868c"),
		Yellow:       lipgloss.Color("#d5a335"),
		SelectedBg:   lipgloss.Color("#5c3a1e"),
		ButtonAccent: lipgloss.Color("#c07040"),
		ButtonWarn:   lipgloss.Color("#b03030"),
	},
}

// FgColors defines all foreground (text) colors
type FgColors struct {
	Default   lipgloss.TerminalColor // Terminal default text
	Selected  lipgloss.TerminalColor // Selected/highlighted items
	Muted     lipgloss.TerminalColor // De-emphasized text
	Accent    lipgloss.TerminalColor // Primary accent
	Subtle    lipgloss.TerminalColor // Secondary/subtle text
	Error     lipgloss.TerminalColor // Error text
	Border    lipgloss.TerminalColor // Border characters
	Separator lipgloss.TerminalColor // Horizontal separator lines (dotted)

	// Title bar
	TitleBar lipgloss.TerminalColor // Text on title bar

	// Table
	TableHeader         lipgloss.TerminalColor // Column headers
	SessionName         lipgloss.TerminalColor // Unselected session names
	SessionNameSelected lipgloss.TerminalColor // Selected session name
	WindowName          lipgloss.TerminalColor // Unselected window names

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

	// Scrollbar
	ScrollbarTrack lipgloss.TerminalColor // Track (subtle background line)
	ScrollbarThumb lipgloss.TerminalColor // Thumb (visible position indicator)

	// Section box
	SectionLabel lipgloss.TerminalColor // Section header label text

	// Sidebar buttons
	ButtonLabel   lipgloss.TerminalColor // Button label text (on accent bg)
	ButtonKeybind lipgloss.TerminalColor // Button keybind hint (on accent bg)
}

// BgColors defines all background colors
type BgColors struct {
	Default      lipgloss.TerminalColor // Terminal default (none)
	TitleBar     lipgloss.TerminalColor // Title bar background
	Selected     lipgloss.TerminalColor // Selected row background (accent)
	ButtonAccent lipgloss.TerminalColor // Action button background
	ButtonWarn   lipgloss.TerminalColor // Warning/danger button background
}

// *****************************************************************************
// Dark theme
// *****************************************************************************

func darkFg() FgColors {
	tc := termColors
	hc := hardCodedColor.Dark
	return FgColors{
		Default:   lipgloss.NoColor{},
		Selected:  tc.BrightWhite,
		Muted:     tc.BrightBlack,
		Accent:    tc.Blue,
		Subtle:    tc.White,
		Error:     tc.Red,
		Border:    tc.BrightBlack,
		Separator: tc.White,

		TitleBar: tc.BrightWhite,

		TableHeader:         lipgloss.NoColor{},
		SessionName:         lipgloss.NoColor{},
		SessionNameSelected: tc.BrightWhite,
		WindowName:          lipgloss.NoColor{},

		ClaudeHeader:  hc.ClaudeOrange,
		ClaudeWorking: hc.Yellow,
		ClaudeWaiting: hc.Green,
		ClaudeUrgent:  hc.Red,
		ClaudeIdle:    hc.Blue,

		GitFiles: hc.Blue,
		GitAdd:   hc.Green,
		GitDel:   hc.Red,

		ScrollbarTrack: tc.BrightBlack,
		ScrollbarThumb: tc.White,

		SectionLabel: hc.ClaudeOrange,

		ButtonLabel:   tc.BrightWhite,
		ButtonKeybind: tc.White,
	}
}

func darkBg() BgColors {
	hc := hardCodedColor.Dark
	return BgColors{
		Default:      lipgloss.NoColor{},
		TitleBar:     termColors.Black,
		Selected:     hc.SelectedBg,
		ButtonAccent: hc.ButtonAccent,
		ButtonWarn:   hc.ButtonWarn,
	}
}

// *****************************************************************************
// Light theme
// *****************************************************************************

func lightBg() BgColors {
	hc := hardCodedColor.Light
	return BgColors{
		Default:      lipgloss.NoColor{},
		TitleBar:     termColors.BrightWhite,
		Selected:     hc.SelectedBg,
		ButtonAccent: hc.ButtonAccent,
		ButtonWarn:   hc.ButtonWarn,
	}
}

func lightFg() FgColors {
	tc := termColors
	hc := hardCodedColor.Light
	return FgColors{
		Default:   lipgloss.NoColor{},
		Selected:  tc.Black,
		Muted:     tc.BrightBlack,
		Accent:    tc.Blue,
		Subtle:    tc.BrightBlack,
		Error:     tc.Red,
		Border:    tc.BrightBlack,
		Separator: tc.BrightWhite,

		TitleBar: tc.Black,

		TableHeader:         lipgloss.NoColor{},
		SessionName:         lipgloss.NoColor{},
		SessionNameSelected: tc.Black,
		WindowName:          lipgloss.NoColor{},

		ClaudeHeader:  hc.ClaudeOrange,
		ClaudeWorking: hc.Yellow,
		ClaudeWaiting: hc.Green,
		ClaudeUrgent:  hc.Red,
		ClaudeIdle:    hc.Blue,

		GitFiles: hc.Blue,
		GitAdd:   hc.Green,
		GitDel:   hc.Red,

		ScrollbarTrack: tc.BrightWhite,
		ScrollbarThumb: tc.White,

		SectionLabel: hc.ClaudeOrange,

		ButtonLabel:   tc.BrightWhite,
		ButtonKeybind: tc.White,
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

// InitColors sets the color palette based on appearance mode
// and reinitializes all styles. Must be called before the TUI renders.
func InitColors(appearance string) {
	if appearance == "light" {
		Colors.Fg = lightFg()
		Colors.Bg = lightBg()
	} else {
		Colors.Fg = darkFg()
		Colors.Bg = darkBg()
	}
	initStyles()
}
