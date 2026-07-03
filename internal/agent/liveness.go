package agent

import (
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Liveness reports where live agent processes run: per session for the
// status checks, and per pane shell PID for attributing an agent identity
// to individual panes in the UI.
type Liveness struct {
	sessions map[string]map[string]bool // kind name → session name → alive
	panes    map[int]string             // pane shell PID → kind name
}

// Alive returns true if the given kind has a live process in the session.
func (l Liveness) Alive(kind Kind, sessionName string) bool {
	return l.sessions[kind.Name][sessionName]
}

// PaneAgent returns the agent kind name running beneath the given pane
// shell PID, or "" if none.
func (l Liveness) PaneAgent(panePID int) string {
	return l.panes[panePID]
}

// PaneAgents returns the full pane PID → agent kind mapping.
func (l Liveness) PaneAgents() map[int]string {
	return l.panes
}

// process is one row of the process table.
type process struct {
	pid     int
	ppid    int
	command string
}

// CheckLiveness takes one snapshot of the process table and reports, for
// each agent kind, the sessions with a matching process running beneath any
// of their pane PIDs. Status hooks don't fire on crash or SIGKILL, so this
// is the ground truth behind a "working" status file.
func CheckLiveness(panePIDs map[string][]int) (Liveness, error) {
	procs, err := processSnapshot()
	if err != nil {
		return Liveness{}, err
	}
	return liveness(panePIDs, procs), nil
}

// processSnapshot reads the full process table in one ps call.
func processSnapshot() ([]process, error) {
	out, err := exec.Command("ps", "-axo", "pid=,ppid=,command=").Output()
	if err != nil {
		return nil, err
	}

	var procs []process
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err1 := strconv.Atoi(fields[0])
		ppid, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}
		procs = append(procs, process{pid: pid, ppid: ppid, command: strings.Join(fields[2:], " ")})
	}
	return procs, nil
}

// liveness walks the process tree beneath each pane's shell PID and matches
// descendants against each kind's binary names. Each pane is walked
// separately so an agent can be attributed to the exact pane it runs in.
func liveness(panePIDs map[string][]int, procs []process) Liveness {
	children := make(map[int][]int, len(procs))
	commands := make(map[int]string, len(procs))
	for _, p := range procs {
		children[p.ppid] = append(children[p.ppid], p.pid)
		commands[p.pid] = p.command
	}

	result := Liveness{
		sessions: make(map[string]map[string]bool, len(Kinds)),
		panes:    make(map[int]string),
	}
	for _, kind := range Kinds {
		result.sessions[kind.Name] = make(map[string]bool)
	}

	for session, pids := range panePIDs {
		for _, panePID := range pids {
			// BFS over this pane's descendants, including the shell itself
			queue := []int{panePID}
			seen := make(map[int]bool, 8)
			for len(queue) > 0 {
				pid := queue[0]
				queue = queue[1:]
				if seen[pid] {
					continue
				}
				seen[pid] = true
				queue = append(queue, children[pid]...)

				for _, kind := range Kinds {
					if commandMatches(commands[pid], kind.BinaryNames) {
						result.sessions[kind.Name][session] = true
						if _, taken := result.panes[panePID]; !taken {
							result.panes[panePID] = kind.Name
						}
					}
				}
			}
		}
	}
	return result
}

// interpreters are runtimes that may front an agent binary in the process
// table ("node /usr/local/bin/claude"). Only then is the script token
// considered — otherwise "grep claude" would count as a live agent.
var interpreters = []string{"node", "bun", "deno"}

// commandMatches checks a command line against agent binary names. The
// binary may be the executable itself ("claude --resume") or a script run
// by an interpreter ("node /usr/local/bin/claude").
func commandMatches(command string, binaryNames []string) bool {
	tokens := strings.Fields(command)
	if len(tokens) == 0 {
		return false
	}
	if slices.Contains(binaryNames, filepath.Base(tokens[0])) {
		return true
	}
	return len(tokens) > 1 &&
		slices.Contains(interpreters, filepath.Base(tokens[0])) &&
		slices.Contains(binaryNames, filepath.Base(tokens[1]))
}
