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

			status := GetStatus(Claude, sessionName, tmpDir)

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

			got := GetStatus(Claude, "sess", tmpDir)

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

	if got := GetStatus(Claude, "sess", tmpDir).State; got != "working" {
		t.Errorf("Claude state = %q, want %q", got, "working")
	}
	if got := GetStatus(Pi, "sess", tmpDir).State; got != "waiting" {
		t.Errorf("Pi state = %q, want %q", got, "waiting")
	}
}

func TestRemoveStatus(t *testing.T) {
	tmpDir := t.TempDir()
	writeFile(t, tmpDir, "sess.status", "working:"+fmt.Sprintf("%d", time.Now().Unix()))

	RemoveStatus(Claude, "sess", tmpDir)

	if _, err := os.Stat(filepath.Join(tmpDir, "sess.status")); !os.IsNotExist(err) {
		t.Error("sess.status should be deleted")
	}
	// Removing a nonexistent file must not panic or error
	RemoveStatus(Claude, "nonexistent", tmpDir)
}

func TestCleanupStale(t *testing.T) {
	tmpDir := t.TempDir()

	files := []string{"active.status", "stale1.status", "stale2.status", "notastatus.txt"}
	for _, f := range files {
		writeFile(t, tmpDir, f, "working:123")
	}

	CleanupStale(Claude, tmpDir, []string{"active"})

	if _, err := os.Stat(filepath.Join(tmpDir, "active.status")); os.IsNotExist(err) {
		t.Error("active.status should not be deleted")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "stale1.status")); !os.IsNotExist(err) {
		t.Error("stale1.status should be deleted")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "stale2.status")); !os.IsNotExist(err) {
		t.Error("stale2.status should be deleted")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "notastatus.txt")); os.IsNotExist(err) {
		t.Error("notastatus.txt should not be deleted")
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

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write %s: %v", name, err)
	}
}
