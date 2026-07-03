package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Session represents a tmux session
type Session struct {
	Name         string
	LastActivity time.Time
	Windows      []Window
	Expanded     bool
}

// Window represents a tmux window
type Window struct {
	Index    int
	Name     string
	Panes    []Pane
	Expanded bool
}

// Pane represents a tmux pane
type Pane struct {
	Index   int
	PID     int    // Pane shell process PID (for agent attribution)
	Command string // Current command running in the pane
	Active  bool   // Active pane in the window
}

// CurrentSession returns the name of the current tmux session
func CurrentSession() (string, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetSessionActivity returns the last activity time for a named session
func GetSessionActivity(name string) (time.Time, error) {
	out, err := exec.Command("tmux", "display-message", "-t", name, "-p", "#{session_activity}").Output()
	if err != nil {
		return time.Time{}, err
	}
	activityUnix, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(activityUnix, 0), nil
}

// ListSessions returns all tmux sessions sorted by activity (most recent first)
// Excludes the current session and popup sessions
func ListSessions(excludeCurrent string) ([]Session, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_activity} #{session_name}").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []Session{}, nil
	}

	var sessions []Session

	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		name := parts[1]

		// Skip current session and popup sessions
		if name == excludeCurrent || strings.HasPrefix(name, "_popup_") {
			continue
		}

		activityUnix, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}

		sessions = append(sessions, Session{
			Name:         name,
			LastActivity: time.Unix(activityUnix, 0),
		})
	}

	// Sort by activity (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.After(sessions[j].LastActivity)
	})

	return sessions, nil
}

// ListWindows returns all windows for a given session
func ListWindows(sessionName string) ([]Window, error) {
	out, err := exec.Command("tmux", "list-windows", "-t", sessionName, "-F", "#{window_index}:#{window_name}").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []Window{}, nil
	}

	var windows []Window
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		index, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		windows = append(windows, Window{
			Index: index,
			Name:  parts[1],
		})
	}

	return windows, nil
}

// KillSession kills a tmux session by name
func KillSession(name string) error {
	return exec.Command("tmux", "kill-session", "-t", name).Run()
}

// KillWindow kills a tmux window
func KillWindow(sessionName string, windowIndex int) error {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	return exec.Command("tmux", "kill-window", "-t", target).Run()
}

// SessionExists checks if a tmux session with the exact given name exists.
// Uses list-sessions to avoid tmux's implicit prefix matching when using
// the -t flag (e.g., "has-session -t foo" also matches "foobar").
func SessionExists(name string) bool {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == name {
			return true
		}
	}
	return false
}

// ClientSize returns the current tmux client's terminal dimensions.
// Falls back to 200x50 if the query fails (e.g., no attached client).
func ClientSize() (int, int) {
	out, err := exec.Command("tmux", "display-message", "-p", "#{client_width} #{client_height}").Output()
	if err != nil {
		return 200, 50
	}
	parts := strings.SplitN(strings.TrimSpace(string(out)), " ", 2)
	if len(parts) != 2 {
		return 200, 50
	}
	w, errW := strconv.Atoi(parts[0])
	h, errH := strconv.Atoi(parts[1])
	if errW != nil || errH != nil || w <= 0 || h <= 0 {
		return 200, 50
	}
	return w, h
}

// CreateSession creates a new tmux session.
// Passes the current client's terminal dimensions so that layout scripts
// can use percentage-based splits accurately on the detached session.
func CreateSession(name, dir string) error {
	w, h := ClientSize()
	return exec.Command("tmux", "new-session", "-d", "-s", name,
		"-x", strconv.Itoa(w), "-y", strconv.Itoa(h),
		"-c", dir,
	).Run()
}

// SwitchClient switches the tmux client to a session or window.
// If running inside tmux, uses switch-client. If outside, uses attach-session.
// For session-only targets (no : or .), resolves the exact session name to
// avoid tmux's implicit prefix matching (e.g., "foo" matching "foobar").
func SwitchClient(target string) error {
	// Resolve session-only targets to exact name to avoid prefix matching
	if !strings.Contains(target, ":") && !strings.Contains(target, ".") {
		if resolved := resolveExactSessionName(target); resolved != "" {
			target = resolved
		}
	}

	var cmd *exec.Cmd
	if os.Getenv("TMUX") != "" {
		cmd = exec.Command("tmux", "switch-client", "-t", target)
	} else {
		cmd = exec.Command("tmux", "attach-session", "-t", target)
		// Connect terminal for interactive attach
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// resolveExactSessionName finds the exact session name, avoiding tmux's
// prefix matching. Returns the exact name if found, empty string otherwise.
func resolveExactSessionName(name string) string {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == name {
			return line
		}
	}
	return ""
}

// SelectWindow selects a specific window in the current client
func SelectWindow(sessionName string, windowIndex int) error {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	return exec.Command("tmux", "switch-client", "-t", target).Run()
}

// ListPanes returns all panes for a given session and window
func ListPanes(sessionName string, windowIndex int) ([]Pane, error) {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	// Command is the last field — it may contain the separator itself
	out, err := exec.Command("tmux", "list-panes", "-t", target, "-F", "#{pane_index}:#{pane_pid}:#{pane_active}:#{pane_current_command}").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []Pane{}, nil
	}

	var panes []Pane
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 4)
		if len(parts) != 4 {
			continue
		}

		index, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		pid, _ := strconv.Atoi(parts[1])

		panes = append(panes, Pane{
			Index:   index,
			PID:     pid,
			Command: parts[3],
			Active:  parts[2] == "1",
		})
	}

	return panes, nil
}

// ListSessionPanes returns all panes of a session grouped by window index —
// one tmux call, so expanded sessions can attribute agents to collapsed
// windows without per-window pane fetches.
func ListSessionPanes(sessionName string) (map[int][]Pane, error) {
	out, err := exec.Command("tmux", "list-panes", "-s", "-t", sessionName, "-F", "#{window_index}:#{pane_index}:#{pane_pid}:#{pane_active}:#{pane_current_command}").Output()
	if err != nil {
		return nil, err
	}

	panes := make(map[int][]Pane)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, ":", 5)
		if len(parts) != 5 {
			continue
		}
		windowIndex, err1 := strconv.Atoi(parts[0])
		paneIndex, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue
		}
		pid, _ := strconv.Atoi(parts[2])

		panes[windowIndex] = append(panes[windowIndex], Pane{
			Index:   paneIndex,
			PID:     pid,
			Command: parts[4],
			Active:  parts[3] == "1",
		})
	}
	return panes, nil
}

// PanePIDs returns each pane's shell process PID across all sessions,
// grouped by session name. One tmux call for everything.
func PanePIDs() (map[string][]int, error) {
	out, err := exec.Command("tmux", "list-panes", "-a", "-F", "#{session_name}\t#{pane_pid}").Output()
	if err != nil {
		return nil, err
	}

	pids := make(map[string][]int)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		pid, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		pids[parts[0]] = append(pids[parts[0]], pid)
	}
	return pids, nil
}

// KillPane kills a tmux pane
func KillPane(sessionName string, windowIndex, paneIndex int) error {
	target := fmt.Sprintf("%s:%d.%d", sessionName, windowIndex, paneIndex)
	return exec.Command("tmux", "kill-pane", "-t", target).Run()
}

// SelectPane switches to a specific pane
func SelectPane(sessionName string, windowIndex, paneIndex int) error {
	target := fmt.Sprintf("%s:%d.%d", sessionName, windowIndex, paneIndex)
	return exec.Command("tmux", "switch-client", "-t", target).Run()
}
