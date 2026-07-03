package agent

import "testing"

func TestLiveness(t *testing.T) {
	// Process tree:
	//   100 (zsh, pane of "work")   → 200 (claude --resume)
	//   101 (zsh, pane of "work")   → 201 (vim)
	//   110 (zsh, pane of "idle")
	//   120 (zsh, pane of "scripted") → 220 (node /usr/local/bin/claude)
	//   130 (zsh, pane of "pi-sess")  → 230 (pi) → 231 (some child)
	procs := []process{
		{pid: 100, ppid: 1, command: "-zsh"},
		{pid: 200, ppid: 100, command: "claude --resume"},
		{pid: 101, ppid: 1, command: "-zsh"},
		{pid: 201, ppid: 101, command: "vim main.go"},
		{pid: 110, ppid: 1, command: "-zsh"},
		{pid: 120, ppid: 1, command: "-zsh"},
		{pid: 220, ppid: 120, command: "node /usr/local/bin/claude"},
		{pid: 130, ppid: 1, command: "-zsh"},
		{pid: 230, ppid: 130, command: "/Users/x/.local/bin/pi"},
		{pid: 231, ppid: 230, command: "git status"},
	}
	panePIDs := map[string][]int{
		"work":     {100, 101},
		"idle":     {110},
		"scripted": {120},
		"pi-sess":  {130},
	}

	live := liveness(panePIDs, procs)

	tests := []struct {
		kind    Kind
		session string
		want    bool
	}{
		{Claude, "work", true},     // direct child binary
		{Claude, "idle", false},    // shell only
		{Claude, "scripted", true}, // interpreter + script path
		{Claude, "pi-sess", false}, // pi is not claude
		{Pi, "pi-sess", true},      // nested under pane shell
		{Pi, "work", false},        // claude is not pi
		{Claude, "unknown", false}, // session without panes
	}
	for _, tt := range tests {
		if got := live.Alive(tt.kind, tt.session); got != tt.want {
			t.Errorf("Alive(%s, %q) = %v, want %v", tt.kind.Name, tt.session, got, tt.want)
		}
	}
}

func TestCommandMatches(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		{"claude", true},
		{"claude --resume", true},
		{"/opt/homebrew/bin/claude", true},
		{"node /usr/local/bin/claude --flag", true},
		{"vim claude.md", false}, // claude.md != claude
		{"grep claude", false},   // grep is not an interpreter
		{"claudette", false},     // no substring matching
		{"", false},
	}
	for _, tt := range tests {
		if got := commandMatches(tt.command, Claude.BinaryNames); got != tt.want {
			t.Errorf("commandMatches(%q) = %v, want %v", tt.command, got, tt.want)
		}
	}
}
