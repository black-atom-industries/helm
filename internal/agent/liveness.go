package agent

import (
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Liveness reports which sessions have a live agent process,
// per agent kind name: kind.Name → session name → true.
type Liveness map[string]map[string]bool

// Alive returns true if the given kind has a live process in the session.
func (l Liveness) Alive(kind Kind, sessionName string) bool {
	return l[kind.Name][sessionName]
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
		return nil, err
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

// liveness walks the process tree beneath each session's pane PIDs and
// matches descendants against each kind's binary names.
func liveness(panePIDs map[string][]int, procs []process) Liveness {
	children := make(map[int][]int, len(procs))
	commands := make(map[int]string, len(procs))
	for _, p := range procs {
		children[p.ppid] = append(children[p.ppid], p.pid)
		commands[p.pid] = p.command
	}

	result := make(Liveness, len(Kinds))
	for _, kind := range Kinds {
		result[kind.Name] = make(map[string]bool)
	}

	for session, pids := range panePIDs {
		// BFS over each pane's descendants, including the pane shell itself
		queue := append([]int(nil), pids...)
		seen := make(map[int]bool, len(queue))
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
					result[kind.Name][session] = true
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
