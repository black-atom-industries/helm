package ui

import (
	"strings"
	"testing"
	"time"
)

func TestFormatClaudeIcon(t *testing.T) {
	tests := []struct {
		name           string
		state          string
		animationFrame int
		waitDuration   time.Duration
		wantSpace      bool // true if should return " " (reserved space)
		contains       string
	}{
		{
			name:           "empty state returns space",
			state:          "",
			animationFrame: 0,
			waitDuration:   0,
			wantSpace:      true,
		},
		{
			name:           "new state returns space (no visual noise)",
			state:          "new",
			animationFrame: 0,
			waitDuration:   0,
			wantSpace:      true,
		},
		{
			name:           "working state frame 0",
			state:          "working",
			animationFrame: 0,
			waitDuration:   0,
			wantSpace:      false,
			contains:       "⠤",
		},
		{
			name:           "working state frame 1",
			state:          "working",
			animationFrame: 1,
			waitDuration:   0,
			wantSpace:      false,
			contains:       "⠆",
		},
		{
			name:           "working state frame 2",
			state:          "working",
			animationFrame: 2,
			waitDuration:   0,
			wantSpace:      false,
			contains:       "⠒",
		},
		{
			name:           "working state frame 3",
			state:          "working",
			animationFrame: 3,
			waitDuration:   0,
			wantSpace:      false,
			contains:       "⠰",
		},
		{
			name:           "working state frame wraps around",
			state:          "working",
			animationFrame: 4,
			waitDuration:   0,
			wantSpace:      false,
			contains:       "⠤", // Back to frame 0
		},
		{
			name:           "waiting state under threshold",
			state:          "waiting",
			animationFrame: 0,
			waitDuration:   4 * time.Minute,
			wantSpace:      false,
			contains:       "?",
		},
		{
			name:           "waiting state at threshold",
			state:          "waiting",
			animationFrame: 0,
			waitDuration:   5 * time.Minute,
			wantSpace:      false,
			contains:       "!",
		},
		{
			name:           "waiting state between thresholds",
			state:          "waiting",
			animationFrame: 0,
			waitDuration:   10 * time.Minute,
			wantSpace:      false,
			contains:       "!",
		},
		{
			name:           "waiting state at idle threshold",
			state:          "waiting",
			animationFrame: 0,
			waitDuration:   15 * time.Minute,
			wantSpace:      false,
			contains:       "Z",
		},
		{
			name:           "waiting state over idle threshold",
			state:          "waiting",
			animationFrame: 0,
			waitDuration:   20 * time.Minute,
			wantSpace:      false,
			contains:       "Z",
		},
		{
			name:           "unknown state returns space",
			state:          "unknown",
			animationFrame: 0,
			waitDuration:   0,
			wantSpace:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatClaudeIcon(tt.state, tt.animationFrame, tt.waitDuration)

			if tt.wantSpace && result != " " {
				t.Errorf("FormatClaudeIcon(%q, %d, %v) = %q, want space", tt.state, tt.animationFrame, tt.waitDuration, result)
			}

			if !tt.wantSpace {
				if result == " " {
					t.Errorf("FormatClaudeIcon(%q, %d, %v) returned space, want non-space", tt.state, tt.animationFrame, tt.waitDuration)
				}
				if tt.contains != "" && !strings.Contains(result, tt.contains) {
					t.Errorf("FormatClaudeIcon(%q, %d, %v) = %q, should contain %q", tt.state, tt.animationFrame, tt.waitDuration, result, tt.contains)
				}
			}
		})
	}
}

func TestScrollbarChars(t *testing.T) {
	tests := []struct {
		name         string
		totalItems   int
		visibleItems int
		scrollOffset int
		height       int
		wantLen      int
		allSpaces    bool
	}{
		{
			name:         "all items fit - no scrollbar",
			totalItems:   5,
			visibleItems: 10,
			scrollOffset: 0,
			height:       5,
			wantLen:      5,
			allSpaces:    true,
		},
		{
			name:         "exactly fits - no scrollbar",
			totalItems:   5,
			visibleItems: 5,
			scrollOffset: 0,
			height:       5,
			wantLen:      5,
			allSpaces:    true,
		},
		{
			name:         "needs scrollbar",
			totalItems:   20,
			visibleItems: 5,
			scrollOffset: 0,
			height:       5,
			wantLen:      5,
			allSpaces:    false,
		},
		{
			name:         "scrolled down",
			totalItems:   20,
			visibleItems: 5,
			scrollOffset: 10,
			height:       5,
			wantLen:      5,
			allSpaces:    false,
		},
		{
			name:         "zero height",
			totalItems:   20,
			visibleItems: 5,
			scrollOffset: 0,
			height:       0,
			wantLen:      0,
			allSpaces:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScrollbarChars(tt.totalItems, tt.visibleItems, tt.scrollOffset, tt.height)

			if len(result) != tt.wantLen {
				t.Errorf("len(ScrollbarChars) = %d, want %d", len(result), tt.wantLen)
			}

			if tt.allSpaces {
				for i, ch := range result {
					if ch != " " {
						t.Errorf("result[%d] = %q, want space (no scrollbar needed)", i, ch)
					}
				}
			} else {
				// Should have some non-space characters (the thumb)
				hasThumb := false
				for _, ch := range result {
					if ch != " " {
						hasThumb = true
						break
					}
				}
				if !hasThumb {
					t.Error("Scrollbar should have visible thumb characters")
				}
			}
		})
	}
}

func TestRenderBorder(t *testing.T) {
	tests := []struct {
		width   int
		wantLen int
	}{
		{width: 10, wantLen: 10},
		{width: 0, wantLen: 0},
		{width: 50, wantLen: 50},
	}

	for _, tt := range tests {
		result := RenderBorder(tt.width)
		// The result will have ANSI codes, but should contain the border chars
		if !strings.Contains(result, "─") && tt.width > 0 {
			t.Errorf("RenderBorder(%d) should contain border character", tt.width)
		}
	}
}
