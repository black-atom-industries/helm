package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SectionBoxOpts controls section box rendering behavior
type SectionBoxOpts struct {
	OmitLeftBorder bool // Omit left │ border (e.g., for session list where scrollbar occupies left edge)
}

// RenderSectionBox renders content inside a bordered container with a label on the top border line.
// Format:
//
//	 LABEL ───────┐
//	│             │
//	│  content    │
//	│             │
//	└─────────────┘
//
// The label sits on the top border. Gap lines after top border and before bottom border.
func RenderSectionBox(label string, content string, width int, opts SectionBoxOpts) string {
	if width < 4 {
		return content
	}

	borderFg := SectionBorderStyle
	labelStyled := SectionLabelStyle.Render(label)
	labelVisualWidth := lipgloss.Width(labelStyled)

	rightBorder := borderFg.Render("│")
	leftBorder := borderFg.Render("│")

	// Inner width available for content (between borders)
	innerWidth := width - 2 // subtract left + right border
	if opts.OmitLeftBorder {
		innerWidth = width - 1 // only right border
	}

	var b strings.Builder

	// Top border: " LABEL ────┐"
	topRuleWidth := width - labelVisualWidth - 1 - 1 // -1 space after label, -1 for ┐
	if !opts.OmitLeftBorder {
		topRuleWidth-- // leading space before label
		b.WriteString(" ")
	}
	if topRuleWidth < 1 {
		topRuleWidth = 1
	}
	b.WriteString(labelStyled)
	b.WriteString(" ")
	b.WriteString(borderFg.Render(strings.Repeat("─", topRuleWidth) + "┐"))
	b.WriteString("\n")

	// Helper: write a line with borders, forcing content to innerWidth
	writeLine := func(line string) {
		// Use lipgloss to force exact width — handles ANSI codes correctly
		padded := lipgloss.NewStyle().Width(innerWidth).Render(line)
		if !opts.OmitLeftBorder {
			b.WriteString(leftBorder)
		}
		b.WriteString(padded)
		b.WriteString(rightBorder)
		b.WriteString("\n")
	}

	// Gap line after top border
	writeLine("")

	// Content lines
	for _, line := range strings.Split(content, "\n") {
		writeLine(line)
	}

	// Gap line before bottom border
	writeLine("")

	// Bottom border
	if opts.OmitLeftBorder {
		b.WriteString(borderFg.Render(strings.Repeat("─", width-1) + "┘"))
	} else {
		b.WriteString(borderFg.Render("└" + strings.Repeat("─", width-2) + "┘"))
	}

	return b.String()
}
