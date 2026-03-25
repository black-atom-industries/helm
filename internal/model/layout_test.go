package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/tmux"
	"github.com/black-atom-industries/helm/internal/ui"
)

// testModel creates a Model with known dimensions and minimal state for layout testing.
func testModel(width, height int, mode Mode) Model {
	cfg := config.DefaultConfig()
	cfg.Bookmarks = []config.Bookmark{
		{Path: "/home/user/projects/alpha"},
		{Path: "/home/user/projects/beta"},
		{Path: "/home/user/projects/gamma"},
	}

	m := New("test-session", cfg, "")
	m.mode = mode
	m.width = width
	m.height = height
	m.sessionsLoaded = true
	m.sessions = []tmux.Session{
		{Name: "session-one", Windows: []tmux.Window{{Index: 1, Name: "editor"}}},
		{Name: "session-two", Windows: []tmux.Window{{Index: 1, Name: "main"}}},
		{Name: "session-three", Windows: []tmux.Window{{Index: 1, Name: "shell"}}},
	}
	m.calculateColumnWidths()
	if cfg.GitStatusEnabled {
		m.maxGitStatusWidth = ui.GitStatusColumnWidth
	}
	m.rebuildItems()

	// Initialize bookmark list for bookmarks mode
	m.bookmarkList.SetItems(cfg.Bookmarks)

	return m
}

// viewLines splits View() output into lines, trimming the trailing newline.
func viewLines(m Model) []string {
	output := m.View()
	return strings.Split(strings.TrimRight(output, "\n"), "\n")
}

func TestViewLinesHaveConsistentWidth(t *testing.T) {
	modes := []struct {
		name string
		mode Mode
	}{
		{"sessions", ModeNormal},
		{"bookmarks", ModeBookmarks},
		{"projects", ModePickDirectory},
		{"clone_choice", ModeCloneChoice},
		{"create", ModeCreate},
		{"confirm_kill", ModeConfirmKill},
	}

	for _, tt := range modes {
		t.Run(tt.name, func(t *testing.T) {
			m := testModel(120, 35, tt.mode)
			lines := viewLines(m)

			if len(lines) == 0 {
				t.Fatal("View() produced no output")
			}

			// All lines must have the same visual width
			expectedWidth := lipgloss.Width(lines[0])
			for i, line := range lines {
				w := lipgloss.Width(line)
				if w != expectedWidth {
					t.Errorf("line %d width = %d, want %d (first line width)\n  line: %q",
						i, w, expectedWidth, line)
				}
			}
		})
	}
}

func TestViewLineCountNotShorterThanHeight(t *testing.T) {
	modes := []struct {
		name string
		mode Mode
	}{
		{"sessions", ModeNormal},
		{"bookmarks", ModeBookmarks},
		{"projects", ModePickDirectory},
		{"clone_choice", ModeCloneChoice},
	}

	for _, tt := range modes {
		t.Run(tt.name, func(t *testing.T) {
			m := testModel(120, 35, tt.mode)
			lines := viewLines(m)

			// View must fill at least the terminal height.
			// It may slightly exceed due to lipgloss Height() being a minimum not a maximum,
			// but it must never be shorter (which would leave the footer floating mid-screen).
			if len(lines) < m.height {
				t.Errorf("View() line count = %d, want >= %d (terminal height)",
					len(lines), m.height)
			}
		})
	}
}

func TestViewWidthMatchesTerminalWidth(t *testing.T) {
	widths := []int{80, 120, 150, 200}

	for _, w := range widths {
		t.Run(fmt.Sprintf("width_%d", w), func(t *testing.T) {
			m := testModel(w, 35, ModeNormal)
			lines := viewLines(m)

			if len(lines) == 0 {
				t.Fatal("View() produced no output")
			}

			// The rendered view should fill exactly the terminal width
			gotWidth := lipgloss.Width(lines[0])
			if gotWidth != w {
				t.Errorf("View() width = %d, want %d (terminal width)", gotWidth, w)
			}
		})
	}
}
