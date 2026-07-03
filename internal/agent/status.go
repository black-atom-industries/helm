// Package agent tracks per-session status of LLM agent clients
// (Claude Code, Pi). Each client writes "state:timestamp" status files
// into the cache dir via its hook script; the TUI reads them here.
package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/black-atom-industries/helm/internal/config"
)

// StaleThreshold is how long before a "working" status is considered stale.
// If the agent hasn't updated the status file in this time, assume it's not running.
const StaleThreshold = 2 * time.Minute

// WaitingStaleThreshold is how long before a "waiting" status is considered stale.
// Safety net — the TUI handles visual progression (? → ! → Z) before this kicks in.
const WaitingStaleThreshold = 30 * time.Minute

// Kind describes one supported agent client.
type Kind struct {
	Name        string   // display name, e.g. "claude"
	FileExt     string   // status-file extension in the cache dir
	BinaryNames []string // process names used for liveness checks
}

var (
	// Claude is the Claude Code client.
	Claude = Kind{Name: "claude", FileExt: config.StatusFileExt, BinaryNames: []string{"claude"}}
	// Pi is the Pi client.
	Pi = Kind{Name: "pi", FileExt: config.PiStatusFileExt, BinaryNames: []string{"pi"}}

	// Kinds lists all supported agent clients.
	Kinds = []Kind{Claude, Pi}
)

// ownsFile reports whether a status file belongs to this kind. A plain
// suffix check is not enough: ".pi-status" also ends in ".status", so the
// longest matching extension across all kinds wins.
func (k Kind) ownsFile(name string) bool {
	if !strings.HasSuffix(name, k.FileExt) {
		return false
	}
	for _, other := range Kinds {
		if len(other.FileExt) > len(k.FileExt) && strings.HasSuffix(name, other.FileExt) {
			return false
		}
	}
	return true
}

// Status represents an agent's status for a session
type Status struct {
	State     string    // "new", "working", "waiting", or ""
	Timestamp time.Time // When the status was last updated

	// Extra context, only present in the JSON status format
	Tool       string // Tool in use when the status was written
	SessionID  string // Agent session id
	Transcript string // Path to the session transcript
	Cwd        string // Agent working directory
}

// IsStale returns true if the status hasn't been updated within the appropriate threshold.
func (s Status) IsStale() bool {
	if s.State == "" {
		return false // No status to be stale
	}
	if s.State == "waiting" {
		return time.Since(s.Timestamp) > WaitingStaleThreshold
	}
	return time.Since(s.Timestamp) > StaleThreshold
}

// GetStatuses reads all status instances for a session: the legacy
// <session><ext> file plus per-instance <session>.<id><ext> files written by
// newer hooks, so multiple agents of the same kind in one tmux session each
// keep their own status. Stale instances are filtered out. The result is
// sorted most-active first (working > waiting > new, then most recent).
func GetStatuses(kind Kind, sessionName, cacheDir string) []Status {
	var statuses []Status

	legacy := sessionName + kind.FileExt
	if s := readStatusFile(filepath.Join(cacheDir, legacy)); s.State != "" {
		statuses = append(statuses, s)
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return statuses
	}
	prefix := sessionName + "."
	for _, entry := range entries {
		name := entry.Name()
		if name == legacy || !kind.ownsFile(name) || !strings.HasPrefix(name, prefix) {
			continue
		}
		if s := readStatusFile(filepath.Join(cacheDir, name)); s.State != "" {
			statuses = append(statuses, s)
		}
	}

	slices.SortFunc(statuses, func(a, b Status) int {
		if pa, pb := statePriority(a.State), statePriority(b.State); pa != pb {
			return pb - pa
		}
		return b.Timestamp.Compare(a.Timestamp)
	})
	return statuses
}

// statePriority orders states by how "active" they are.
func statePriority(state string) int {
	switch state {
	case "working":
		return 3
	case "waiting":
		return 2
	case "new":
		return 1
	default:
		return 0
	}
}

// readStatusFile reads and parses one status file.
// Returns empty Status if missing, malformed, or stale.
func readStatusFile(path string) Status {
	content, err := os.ReadFile(path)
	if err != nil {
		return Status{}
	}

	raw := strings.TrimSpace(string(content))
	var status Status
	if strings.HasPrefix(raw, "{") {
		status = parseJSONStatus(raw)
	} else {
		status = parseLegacyStatus(raw)
	}

	// If status is stale, treat it as no status
	if status.State == "" || status.IsStale() {
		return Status{}
	}

	return status
}

// parseJSONStatus parses the JSON status format written by newer hooks.
func parseJSONStatus(raw string) Status {
	var file struct {
		State      string `json:"state"`
		Ts         int64  `json:"ts"`
		Tool       string `json:"tool"`
		SessionID  string `json:"session_id"`
		Transcript string `json:"transcript"`
		Cwd        string `json:"cwd"`
	}
	if err := json.Unmarshal([]byte(raw), &file); err != nil || file.Ts == 0 {
		return Status{}
	}
	return Status{
		State:      file.State,
		Timestamp:  time.Unix(file.Ts, 0),
		Tool:       file.Tool,
		SessionID:  file.SessionID,
		Transcript: file.Transcript,
		Cwd:        file.Cwd,
	}
}

// parseLegacyStatus parses the original "state:timestamp" format.
func parseLegacyStatus(raw string) Status {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return Status{}
	}
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return Status{}
	}
	return Status{
		State:     parts[0],
		Timestamp: time.Unix(timestamp, 0),
	}
}

// RemoveStatuses deletes all of a session's status files for a kind —
// legacy and per-instance — e.g. after a liveness check found no running
// agent process behind them.
func RemoveStatuses(kind Kind, sessionName, cacheDir string) {
	_ = os.Remove(filepath.Join(cacheDir, sessionName+kind.FileExt))

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return
	}
	prefix := sessionName + "."
	for _, entry := range entries {
		name := entry.Name()
		if kind.ownsFile(name) && strings.HasPrefix(name, prefix) {
			_ = os.Remove(filepath.Join(cacheDir, name))
		}
	}
}

// CleanupStale removes status files for sessions that no longer exist
func CleanupStale(kind Kind, cacheDir string, activeSessions []string) {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return
	}

	activeSet := make(map[string]bool)
	for _, s := range activeSessions {
		activeSet[s] = true
	}

	for _, entry := range entries {
		if !kind.ownsFile(entry.Name()) {
			continue
		}

		base := strings.TrimSuffix(entry.Name(), kind.FileExt)
		if activeSet[base] {
			continue // legacy <session> file
		}
		// Per-instance <session>.<id> file
		if i := strings.LastIndex(base, "."); i > 0 && activeSet[base[:i]] {
			continue
		}
		_ = os.Remove(filepath.Join(cacheDir, entry.Name()))
	}
}
