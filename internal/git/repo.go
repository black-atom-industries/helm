package git

import (
	"os/exec"
	"strconv"
	"strings"
)

// RepoState represents the sync state of a git repository
type RepoState string

const (
	StateClean       RepoState = "clean"
	StateDirty       RepoState = "dirty"
	StateAhead       RepoState = "ahead"
	StateBehind      RepoState = "behind"
	StateDiverged    RepoState = "diverged"
	StateDirtyAhead  RepoState = "dirty+ahead"
	StateDirtyBehind RepoState = "dirty+behind"
	StateNoUpstream  RepoState = "no-upstream"
)

// SyncStatus holds the full sync state of a repository
type SyncStatus struct {
	State  RepoState
	Ahead  int
	Behind int
	Dirty  int
}

// GetSyncStatus returns the sync state of a git repository at dir.
func GetSyncStatus(dir string) SyncStatus {
	dirty := getDirtyCount(dir)

	// Check if there's an upstream tracking branch
	if _, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "@{u}").Output(); err != nil {
		return SyncStatus{State: StateNoUpstream, Dirty: dirty}
	}

	ahead := revListCount(dir, "@{u}..")
	behind := revListCount(dir, "..@{u}")

	state := resolveState(dirty, ahead, behind)
	return SyncStatus{
		State:  state,
		Ahead:  ahead,
		Behind: behind,
		Dirty:  dirty,
	}
}

// Fetch runs git fetch --all --quiet in the given directory.
func Fetch(dir string) error {
	cmd := exec.Command("git", "-C", dir, "fetch", "--all", "--quiet")
	return cmd.Run()
}

// Pull runs git pull --ff-only in the given directory.
// Returns an error if the pull cannot fast-forward.
func Pull(dir string) error {
	cmd := exec.Command("git", "-C", dir, "pull", "--ff-only")
	return cmd.Run()
}

// Push runs git push in the given directory.
func Push(dir string) error {
	cmd := exec.Command("git", "-C", dir, "push")
	return cmd.Run()
}

// GetBranch returns the current branch name for the repo at dir.
func GetBranch(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// revListCount runs git rev-list --count with the given revspec and returns the count.
func revListCount(dir, revspec string) int {
	out, err := exec.Command("git", "-C", dir, "rev-list", "--count", revspec).Output()
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return n
}

// resolveState determines the RepoState from dirty, ahead, and behind counts.
func resolveState(dirty, ahead, behind int) RepoState {
	switch {
	case dirty > 0 && behind > 0:
		return StateDirtyBehind
	case dirty > 0 && ahead > 0:
		return StateDirtyAhead
	case dirty > 0:
		return StateDirty
	case ahead > 0 && behind > 0:
		return StateDiverged
	case ahead > 0:
		return StateAhead
	case behind > 0:
		return StateBehind
	default:
		return StateClean
	}
}
