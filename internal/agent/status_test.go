package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetStatus(t *testing.T) {
	tmpDir := t.TempDir()

	// Use current timestamp to avoid stale detection
	currentTimestamp := time.Now().Unix()

	tests := []struct {
		name        string
		filename    string
		content     string
		wantState   string
		wantTimeSet bool
	}{
		{
			name:        "valid working status",
			filename:    "test-session.status",
			content:     "working:" + fmt.Sprintf("%d", currentTimestamp),
			wantState:   "working",
			wantTimeSet: true,
		},
		{
			name:        "valid waiting status",
			filename:    "test-session.status",
			content:     "waiting:" + fmt.Sprintf("%d", currentTimestamp),
			wantState:   "waiting",
			wantTimeSet: true,
		},
		{
			name:        "stale waiting status returns empty",
			filename:    "test-session.status",
			content:     "waiting:" + fmt.Sprintf("%d", currentTimestamp-int64(WaitingStaleThreshold.Seconds())-1),
			wantState:   "",
			wantTimeSet: false,
		},
		{
			name:        "stale working status returns empty",
			filename:    "test-session.status",
			content:     "working:" + fmt.Sprintf("%d", currentTimestamp-int64(StaleThreshold.Seconds())-1),
			wantState:   "",
			wantTimeSet: false,
		},
		{
			name:        "valid new status",
			filename:    "test-session.status",
			content:     "new:" + fmt.Sprintf("%d", currentTimestamp),
			wantState:   "new",
			wantTimeSet: true,
		},
		{
			name:        "missing file returns empty",
			filename:    "nonexistent.status",
			content:     "",
			wantState:   "",
			wantTimeSet: false,
		},
		{
			name:        "malformed content - no colon",
			filename:    "test-session.status",
			content:     "working",
			wantState:   "",
			wantTimeSet: false,
		},
		{
			name:        "malformed content - invalid timestamp",
			filename:    "test-session.status",
			content:     "working:notanumber",
			wantState:   "",
			wantTimeSet: false,
		},
		{
			name:        "empty file",
			filename:    "test-session.status",
			content:     "",
			wantState:   "",
			wantTimeSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.content != "" || tt.name == "empty file" {
				filePath := filepath.Join(tmpDir, tt.filename)
				if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				defer func() { _ = os.Remove(filePath) }()
			}

			sessionName := "test-session"
			if tt.name == "missing file returns empty" {
				sessionName = "nonexistent"
			}

			status := primaryStatus(Claude, sessionName, tmpDir)

			if status.State != tt.wantState {
				t.Errorf("State = %q, want %q", status.State, tt.wantState)
			}

			if tt.wantTimeSet && status.Timestamp.IsZero() {
				t.Error("Timestamp should be set")
			}

			if !tt.wantTimeSet && !status.Timestamp.IsZero() {
				t.Error("Timestamp should be zero")
			}
		})
	}
}

func TestGetStatusJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	ts := time.Now().Unix()

	tests := []struct {
		name     string
		content  string
		want     Status
		wantZero bool
	}{
		{
			name:    "full JSON status",
			content: fmt.Sprintf(`{"state":"working","ts":%d,"tool":"Bash","session_id":"abc","transcript":"/t.jsonl","cwd":"/repo"}`, ts),
			want:    Status{State: "working", Tool: "Bash", SessionID: "abc", Transcript: "/t.jsonl", Cwd: "/repo"},
		},
		{
			name:    "minimal JSON status (jq fallback)",
			content: fmt.Sprintf(`{"state":"waiting","ts":%d}`, ts),
			want:    Status{State: "waiting"},
		},
		{
			name:     "stale JSON working status",
			content:  fmt.Sprintf(`{"state":"working","ts":%d}`, ts-int64(StaleThreshold.Seconds())-1),
			wantZero: true,
		},
		{
			name:     "malformed JSON",
			content:  `{"state":"working"`,
			wantZero: true,
		},
		{
			name:     "JSON without timestamp",
			content:  `{"state":"working"}`,
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeFile(t, tmpDir, "sess.status", tt.content)

			got := primaryStatus(Claude, "sess", tmpDir)

			if tt.wantZero {
				if got.State != "" {
					t.Errorf("State = %q, want empty", got.State)
				}
				return
			}
			if got.State != tt.want.State || got.Tool != tt.want.Tool ||
				got.SessionID != tt.want.SessionID || got.Transcript != tt.want.Transcript ||
				got.Cwd != tt.want.Cwd {
				t.Errorf("got %+v, want %+v (timestamp ignored)", got, tt.want)
			}
			if got.Timestamp.IsZero() {
				t.Error("Timestamp should be set")
			}
		})
	}
}

func TestGetStatusPerKind(t *testing.T) {
	tmpDir := t.TempDir()
	ts := fmt.Sprintf("%d", time.Now().Unix())

	writeFile(t, tmpDir, "sess.status", "working:"+ts)
	writeFile(t, tmpDir, "sess.pi-status", "waiting:"+ts)

	if got := primaryStatus(Claude, "sess", tmpDir).State; got != "working" {
		t.Errorf("Claude state = %q, want %q", got, "working")
	}
	if got := primaryStatus(Pi, "sess", tmpDir).State; got != "waiting" {
		t.Errorf("Pi state = %q, want %q", got, "waiting")
	}
}

func TestGetStatusesMultiInstance(t *testing.T) {
	tmpDir := t.TempDir()
	now := time.Now().Unix()

	// Legacy file + two per-instance files in one session
	writeFile(t, tmpDir, "sess.status", fmt.Sprintf("new:%d", now))
	writeFile(t, tmpDir, "sess.uuid-1.status", fmt.Sprintf(`{"state":"working","ts":%d,"tool":"Bash"}`, now))
	writeFile(t, tmpDir, "sess.uuid-2.status", fmt.Sprintf(`{"state":"waiting","ts":%d}`, now))
	// Different session — must not bleed in
	writeFile(t, tmpDir, "other.uuid-3.status", fmt.Sprintf(`{"state":"working","ts":%d}`, now))

	statuses := GetStatuses(Claude, "sess", tmpDir)

	if len(statuses) != 3 {
		t.Fatalf("len = %d, want 3 (%+v)", len(statuses), statuses)
	}
	// Sorted most-active first: working > waiting > new
	if statuses[0].State != "working" || statuses[1].State != "waiting" || statuses[2].State != "new" {
		t.Errorf("order = %s/%s/%s, want working/waiting/new",
			statuses[0].State, statuses[1].State, statuses[2].State)
	}
	if statuses[0].Tool != "Bash" {
		t.Errorf("primary tool = %q, want Bash", statuses[0].Tool)
	}
}

func TestRemoveStatuses(t *testing.T) {
	tmpDir := t.TempDir()
	ts := fmt.Sprintf("%d", time.Now().Unix())
	writeFile(t, tmpDir, "sess.status", "working:"+ts)
	writeFile(t, tmpDir, "sess.uuid-1.status", "working:"+ts)
	writeFile(t, tmpDir, "sess.pi-status", "working:"+ts)

	RemoveStatuses(Claude, "sess", tmpDir)

	if _, err := os.Stat(filepath.Join(tmpDir, "sess.status")); !os.IsNotExist(err) {
		t.Error("sess.status should be deleted")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "sess.uuid-1.status")); !os.IsNotExist(err) {
		t.Error("sess.uuid-1.status should be deleted")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "sess.pi-status")); os.IsNotExist(err) {
		t.Error("sess.pi-status must survive Claude removal")
	}
	// Removing a nonexistent session must not panic or error
	RemoveStatuses(Claude, "nonexistent", tmpDir)
}

func TestCleanupStale(t *testing.T) {
	tmpDir := t.TempDir()

	keep := []string{"active.status", "active.uuid-1.status", "notastatus.txt"}
	remove := []string{"stale1.status", "stale2.status", "gone.uuid-2.status"}
	for _, f := range append(append([]string{}, keep...), remove...) {
		writeFile(t, tmpDir, f, "working:123")
	}

	CleanupStale(Claude, tmpDir, []string{"active"})

	for _, f := range keep {
		if _, err := os.Stat(filepath.Join(tmpDir, f)); os.IsNotExist(err) {
			t.Errorf("%s should not be deleted", f)
		}
	}
	for _, f := range remove {
		if _, err := os.Stat(filepath.Join(tmpDir, f)); !os.IsNotExist(err) {
			t.Errorf("%s should be deleted", f)
		}
	}
}

// Claude's ".status" extension is a suffix of Pi's ".pi-status" — cleanup
// for one kind must never touch the other kind's files.
func TestCleanupStaleDoesNotCrossKinds(t *testing.T) {
	tmpDir := t.TempDir()

	writeFile(t, tmpDir, "sess.status", "working:123")
	writeFile(t, tmpDir, "sess.pi-status", "working:123")

	// "sess" is inactive for Claude — only the Claude file may go
	CleanupStale(Claude, tmpDir, nil)

	if _, err := os.Stat(filepath.Join(tmpDir, "sess.status")); !os.IsNotExist(err) {
		t.Error("sess.status should be deleted")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "sess.pi-status")); os.IsNotExist(err) {
		t.Error("sess.pi-status must not be deleted by Claude cleanup")
	}

	// And Pi cleanup removes its own file
	CleanupStale(Pi, tmpDir, nil)
	if _, err := os.Stat(filepath.Join(tmpDir, "sess.pi-status")); !os.IsNotExist(err) {
		t.Error("sess.pi-status should be deleted by Pi cleanup")
	}
}

// primaryStatus returns the most-active status instance, or zero Status.
func primaryStatus(kind Kind, sessionName, cacheDir string) Status {
	statuses := GetStatuses(kind, sessionName, cacheDir)
	if len(statuses) == 0 {
		return Status{}
	}
	return statuses[0]
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write %s: %v", name, err)
	}
}
