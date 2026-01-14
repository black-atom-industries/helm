package git

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Status represents git repository status for a session
type Status struct {
	IsRepo    bool
	Dirty     int // Count of uncommitted changes (staged + unstaged + untracked)
	Additions int // Lines added
	Deletions int // Lines deleted
}

// IsClean returns true if there are no changes to show
func (s Status) IsClean() bool {
	return !s.IsRepo || (s.Dirty == 0 && s.Additions == 0 && s.Deletions == 0)
}

// GetSessionPath returns the current working directory of a tmux session's active pane
func GetSessionPath(sessionName string) (string, error) {
	out, err := exec.Command("tmux", "display-message", "-t", sessionName, "-p", "#{pane_current_path}").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetStatus returns the git status for a directory
// Returns Status{IsRepo: false} if the directory is not a git repository
func GetStatus(dir string) Status {
	// Check if this is a git repo by looking for .git
	gitDir := filepath.Join(dir, ".git")
	if !isDir(gitDir) && !isFile(gitDir) {
		return Status{IsRepo: false}
	}

	status := Status{IsRepo: true}

	// Get dirty count: staged + unstaged + untracked
	status.Dirty = getDirtyCount(dir)

	// Get line additions/deletions
	status.Additions, status.Deletions = getLineStats(dir)

	return status
}

// getDirtyCount returns the number of dirty files (modified, staged, untracked)
func getDirtyCount(dir string) int {
	cmd := exec.Command("git", "-C", dir, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}

// getLineStats returns lines added and deleted in working directory
func getLineStats(dir string) (additions, deletions int) {
	// Get stats for all uncommitted changes (staged + unstaged)
	cmd := exec.Command("git", "-C", dir, "diff", "--numstat", "HEAD")
	out, _ := cmd.Output()

	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[0] != "-" {
			add, _ := strconv.Atoi(parts[0])
			del, _ := strconv.Atoi(parts[1])
			additions += add
			deletions += del
		}
	}
	return additions, deletions
}

// isDir checks if path is a directory
func isDir(path string) bool {
	cmd := exec.Command("test", "-d", path)
	return cmd.Run() == nil
}

// isFile checks if path is a file (for git worktrees where .git is a file)
func isFile(path string) bool {
	cmd := exec.Command("test", "-f", path)
	return cmd.Run() == nil
}
