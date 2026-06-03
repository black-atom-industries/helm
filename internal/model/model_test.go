package model

import (
	"testing"

	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/lib/fuzzy"
	"github.com/black-atom-industries/helm/internal/tmux"
)

func TestIsCursorValid(t *testing.T) {
	m := Model{
		items: []Item{
			{Type: ItemTypeSession, SessionIndex: 0},
			{Type: ItemTypeSession, SessionIndex: 1},
			{Type: ItemTypeSession, SessionIndex: 2},
		},
	}

	tests := []struct {
		name   string
		cursor int
		want   bool
	}{
		{name: "valid first", cursor: 0, want: true},
		{name: "valid middle", cursor: 1, want: true},
		{name: "valid last", cursor: 2, want: true},
		{name: "negative", cursor: -1, want: false},
		{name: "out of bounds", cursor: 3, want: false},
		{name: "way out of bounds", cursor: 100, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.cursor = tt.cursor
			got := m.isCursorValid()
			if got != tt.want {
				t.Errorf("isCursorValid() with cursor=%d = %v, want %v", tt.cursor, got, tt.want)
			}
		})
	}
}

func TestGetTargetName(t *testing.T) {
	selfSession := &tmux.Session{
		Name: "self-session",
		Windows: []tmux.Window{
			{Index: 1, Name: "editor"},
		},
	}
	m := Model{
		selfSession: selfSession,
		sessions: []tmux.Session{
			{
				Name: "session1",
				Windows: []tmux.Window{
					{Index: 1, Name: "window1", Panes: []tmux.Pane{{Index: 0, Command: "zsh"}, {Index: 1, Command: "nvim"}}},
					{Index: 2, Name: "window2"},
				},
			},
			{
				Name: "session2",
				Windows: []tmux.Window{
					{Index: 1, Name: "main"},
				},
			},
		},
	}

	tests := []struct {
		name string
		item Item
		want string
	}{
		{
			name: "self session",
			item: Item{Type: ItemTypeSession, IsSelf: true},
			want: "self-session",
		},
		{
			name: "self session window",
			item: Item{Type: ItemTypeWindow, IsSelf: true, WindowIndex: 0},
			want: "self-session:1",
		},
		{
			name: "session item",
			item: Item{Type: ItemTypeSession, SessionIndex: 0},
			want: "session1",
		},
		{
			name: "second session",
			item: Item{Type: ItemTypeSession, SessionIndex: 1},
			want: "session2",
		},
		{
			name: "window item",
			item: Item{Type: ItemTypeWindow, SessionIndex: 0, WindowIndex: 0},
			want: "session1:1",
		},
		{
			name: "second window",
			item: Item{Type: ItemTypeWindow, SessionIndex: 0, WindowIndex: 1},
			want: "session1:2",
		},
		{
			name: "pane item",
			item: Item{Type: ItemTypePane, SessionIndex: 0, WindowIndex: 0, PaneIndex: 0},
			want: "session1:1.0",
		},
		{
			name: "second pane",
			item: Item{Type: ItemTypePane, SessionIndex: 0, WindowIndex: 0, PaneIndex: 1},
			want: "session1:1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.getTargetName(tt.item)
			if got != tt.want {
				t.Errorf("getTargetName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetError(t *testing.T) {
	m := Model{}

	m.setError("test error: %d", 42)

	if m.message != "test error: 42" {
		t.Errorf("message = %q, want %q", m.message, "test error: 42")
	}

	if !m.messageIsError {
		t.Error("messageIsError should be true")
	}
}

func TestNew(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New("current-session", cfg, "")

	if m.currentSession != "current-session" {
		t.Errorf("currentSession = %q, want %q", m.currentSession, "current-session")
	}

	if m.mode != ModeNormal {
		t.Errorf("mode = %v, want ModeNormal", m.mode)
	}
}

func TestSanitizeSessionName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple name unchanged",
			input: "my-session",
			want:  "my-session",
		},
		{
			name:  "slashes to dashes",
			input: "owner/repo",
			want:  "owner-repo",
		},
		{
			name:  "dots to dashes",
			input: "nbr.haus",
			want:  "nbr-haus",
		},
		{
			name:  "colons to dashes",
			input: "session:window",
			want:  "session-window",
		},
		{
			name:  "mixed special chars",
			input: "owner/repo.name:tag",
			want:  "owner-repo-name-tag",
		},
		{
			name:  "spaces to dashes",
			input: "my session name",
			want:  "my-session-name",
		},
		{
			name:  "real world example",
			input: "nikbrunner/nbr.haus",
			want:  "nikbrunner-nbr-haus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeSessionName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeSessionName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Layout height calculation tests

func TestContentWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
		want  int
	}{
		{
			name:  "zero width returns default",
			width: 0,
			want:  56, // Default fallback (60 - 4)
		},
		{
			name:  "normal width subtracts border overhead",
			width: 80,
			want:  76, // 80 - 4 (AppBorderOverheadX)
		},
		{
			name:  "small width",
			width: 40,
			want:  36, // 40 - 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{width: tt.width}
			got := m.contentWidth()
			if got != tt.want {
				t.Errorf("contentWidth() with width=%d = %d, want %d", tt.width, got, tt.want)
			}
		})
	}
}

func TestContentHeight(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{
			name:   "zero height returns zero",
			height: 0,
			want:   0,
		},
		{
			name:   "normal height subtracts border overhead",
			height: 30,
			want:   28, // 30 - 2 (AppBorderOverheadY)
		},
		{
			name:   "small height",
			height: 10,
			want:   8, // 10 - 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{height: tt.height}
			got := m.contentHeight()
			if got != tt.want {
				t.Errorf("contentHeight() with height=%d = %d, want %d", tt.height, got, tt.want)
			}
		})
	}
}

func TestSessionMaxVisibleItems(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{
			name:   "zero height returns fallback",
			height: 0,
			want:   10,
		},
		{
			name:   "small window",
			height: 12, // contentHeight = 10, overhead = 6 + ActionBarHeight(3) = 9, available = 1
			want:   1,
		},
		{
			name:   "large window uses all space",
			height: 50, // contentHeight = 48, overhead = 9, available = 39
			want:   39,
		},
		{
			name:   "very small window",
			height: 10, // contentHeight = 8, overhead = 9, available = -1 → fallback
			want:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			m := Model{
				height: tt.height,
				config: cfg,
			}
			got := m.sessionMaxVisibleItems()
			if got != tt.want {
				t.Errorf("sessionMaxVisibleItems() with height=%d = %d, want %d",
					tt.height, got, tt.want)
			}
		})
	}
}

func TestProjectMaxVisibleItems(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{
			name:   "zero height returns fallback",
			height: 0,
			want:   10,
		},
		{
			name:   "small window",
			height: 12, // contentHeight = 10, overhead = 6 + ActionBarHeight(3) = 9, available = 1
			want:   1,
		},
		{
			name:   "large window uses all space",
			height: 50, // contentHeight = 48, overhead = 9, available = 39
			want:   39,
		},
		{
			name:   "medium window",
			height: 17, // contentHeight = 15, overhead = 9, available = 6
			want:   6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			m := Model{
				height: tt.height,
				config: cfg,
			}
			got := m.projectMaxVisibleItems()
			if got != tt.want {
				t.Errorf("projectMaxVisibleItems() with height=%d = %d, want %d",
					tt.height, got, tt.want)
			}
		})
	}
}

// TestFilterMatching verifies that the fuzzy filter works correctly with
// real session names from the user's environment.
func TestFilterMatching(t *testing.T) {
	// Real session names from the user's tmux
	sessions := []string{
		"black-atom-industries-core",
		"black-atom-industries-helm",
		"imfusion-brunner-agents",
		"nikbrunner-dots",
		"nikbrunner-flux-nvim",
		"nikbrunner-imf-notes",
		"nikbrunner-notes",
	}

	tests := []struct {
		name    string
		filter  string
		wantAll []string // sessions that should match
		noneOf  []string // sessions that should NOT match
	}{
		{
			name:    "imfusion matches imfusion-brunner-agents",
			filter:  "imfusion",
			wantAll: []string{"imfusion-brunner-agents"},
			noneOf:  []string{"black-atom-industries-helm", "nikbrunner-dots"},
		},
		{
			name:    "black matches both black-atom repos",
			filter:  "black",
			wantAll: []string{"black-atom-industries-core", "black-atom-industries-helm"},
			noneOf:  []string{"imfusion-brunner-agents", "nikbrunner-dots"},
		},
		{
			name:    "helm matches only helm",
			filter:  "helm",
			wantAll: []string{"black-atom-industries-helm"},
			noneOf:  []string{"black-atom-industries-core", "imfusion-brunner-agents"},
		},
		{
			name:    "brunner matches imfusion-brunner-agents and nikbrunner repos",
			filter:  "brunner",
			wantAll: []string{"imfusion-brunner-agents", "nikbrunner-dots", "nikbrunner-flux-nvim", "nikbrunner-imf-notes", "nikbrunner-notes"},
			noneOf:  []string{"black-atom-industries-helm"},
		},
		{
			name:    "nik matches all nikbrunner repos",
			filter:  "nik",
			wantAll: []string{"nikbrunner-dots", "nikbrunner-flux-nvim", "nikbrunner-imf-notes", "nikbrunner-notes"},
			noneOf:  []string{"black-atom-industries-helm", "imfusion-brunner-agents"},
		},
		{
			name:    "agents matches imfusion-brunner-agents",
			filter:  "agents",
			wantAll: []string{"imfusion-brunner-agents"},
			noneOf:  []string{"black-atom-industries-helm"},
		},
		{
			name:    "dots matches nikbrunner-dots",
			filter:  "dots",
			wantAll: []string{"nikbrunner-dots"},
			noneOf:  []string{"black-atom-industries-helm", "imfusion-brunner-agents"},
		},
		{
			name:    "xyz matches nothing",
			filter:  "xyz",
			wantAll: []string{},
			noneOf:  sessions,
		},
		{
			name:    "empty filter matches all",
			filter:  "",
			wantAll: sessions,
			noneOf:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterLower := tt.filter
			var matched []string
			for _, name := range sessions {
				if tt.filter == "" || fuzzy.Match(name, filterLower) {
					matched = append(matched, name)
				}
			}

			// Check wanted sessions matched
			matchedSet := make(map[string]bool)
			for _, m := range matched {
				matchedSet[m] = true
			}
			for _, want := range tt.wantAll {
				if !matchedSet[want] {
					t.Errorf("filter %q should match %q but didn't. Matched: %v", tt.filter, want, matched)
				}
			}

			// Check unwanted sessions didn't match
			for _, none := range tt.noneOf {
				if matchedSet[none] {
					t.Errorf("filter %q should NOT match %q but did. Matched: %v", tt.filter, none, matched)
				}
			}
		})
	}
}
