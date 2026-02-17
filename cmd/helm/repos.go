package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/git"
	"github.com/black-atom-industries/helm/internal/github"
)

// repoStatus holds status info for a single repo (used in JSON output)
type repoStatus struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Branch string `json:"branch"`
	Ahead  int    `json:"ahead"`
	Behind int    `json:"behind"`
	Dirty  int    `json:"dirty"`
}

func runRepos(args []string) error {
	if len(args) == 0 {
		printReposUsage()
		return nil
	}

	switch args[0] {
	case "status":
		return runReposStatus(args[1:])
	case "pull":
		return runReposPull(args[1:])
	case "push":
		return runReposPush(args[1:])
	case "dirty":
		return runReposDirty(args[1:])
	case "rebuild":
		return runReposRebuild(args[1:])
	default:
		fmt.Printf("Unknown repos command: %s\n", args[0])
		printReposUsage()
		return nil
	}
}

func printReposUsage() {
	fmt.Println("Usage: helm repos <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  status [--json]                Show sync state of all repos")
	fmt.Println("  pull   [--json]                Fetch and pull (ff-only) all clean repos")
	fmt.Println("  push   [--json]                Push all ahead repos")
	fmt.Println("  dirty  [--walk]                 Print paths of dirty repos (--walk runs configured command)")
	fmt.Println("  rebuild [--all | --repos r,r]  Re-run post_clone hooks")
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func getFlagValue(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(a, flag+"=") {
			return strings.TrimPrefix(a, flag+"=")
		}
	}
	return ""
}

// collectRepoStatuses gathers sync status for all repos in parallel.
func collectRepoStatuses(repos []config.RepoInfo) []repoStatus {
	results := make([]repoStatus, len(repos))
	var wg sync.WaitGroup

	const maxParallel = 8
	sem := make(chan struct{}, maxParallel)

	for i, repo := range repos {
		wg.Add(1)
		go func(idx int, r config.RepoInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			sync := git.GetSyncStatus(r.Path)
			branch, _ := git.GetBranch(r.Path)

			results[idx] = repoStatus{
				Name:   r.Name,
				State:  string(sync.State),
				Branch: branch,
				Ahead:  sync.Ahead,
				Behind: sync.Behind,
				Dirty:  sync.Dirty,
			}
		}(i, repo)
	}

	wg.Wait()
	return results
}

// --- status ---

func runReposStatus(args []string) error {
	jsonOut := hasFlag(args, "--json")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	repos, err := config.ListAllRepos(cfg.ProjectDirs)
	if err != nil {
		return fmt.Errorf("failed to list repos: %w", err)
	}

	if len(repos) == 0 {
		if jsonOut {
			fmt.Println(`{"repos":[]}`)
		} else {
			fmt.Println("No repos found in project_dirs.")
		}
		return nil
	}

	statuses := collectRepoStatuses(repos)

	if jsonOut {
		out := struct {
			Repos []repoStatus `json:"repos"`
		}{Repos: statuses}
		data, _ := json.Marshal(out)
		fmt.Println(string(data))
		return nil
	}

	// Human output
	stateSymbol := map[string]string{
		"clean":        "✓",
		"dirty":        "~",
		"ahead":        "↑",
		"behind":       "↓",
		"diverged":     "↕",
		"dirty+ahead":  "~↑",
		"dirty+behind": "~↓",
		"no-upstream":  "⊘",
	}

	for _, s := range statuses {
		sym := stateSymbol[s.State]
		if sym == "" {
			sym = "?"
		}
		detail := ""
		if s.Ahead > 0 || s.Behind > 0 || s.Dirty > 0 {
			parts := []string{}
			if s.Dirty > 0 {
				parts = append(parts, fmt.Sprintf("%d dirty", s.Dirty))
			}
			if s.Ahead > 0 {
				parts = append(parts, fmt.Sprintf("↑%d", s.Ahead))
			}
			if s.Behind > 0 {
				parts = append(parts, fmt.Sprintf("↓%d", s.Behind))
			}
			detail = " (" + strings.Join(parts, ", ") + ")"
		}
		fmt.Printf("  %s %-40s %s%s\n", sym, s.Name, s.Branch, detail)
	}

	return nil
}

// --- pull ---

type pullResult struct {
	Name   string `json:"name"`
	Action string `json:"action"` // "pulled", "skipped", "failed"
	Reason string `json:"reason,omitempty"`
	Error  string `json:"error,omitempty"`
}

func runReposPull(args []string) error {
	jsonOut := hasFlag(args, "--json")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	repos, err := config.ListAllRepos(cfg.ProjectDirs)
	if err != nil {
		return fmt.Errorf("failed to list repos: %w", err)
	}

	if len(repos) == 0 {
		if jsonOut {
			fmt.Println(`{"pulled":[],"skipped":[],"failed":[],"summary":{"pulled":0,"skipped":0,"failed":0}}`)
		} else {
			fmt.Println("No repos found.")
		}
		return nil
	}

	// Phase 1: Fetch all in parallel
	if !jsonOut {
		fmt.Printf("Fetching %d repos...\n", len(repos))
	}

	const maxNetwork = 4
	sem := make(chan struct{}, maxNetwork)
	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		go func(r config.RepoInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			_ = git.Fetch(r.Path)
		}(repo)
	}
	wg.Wait()

	// Phase 2: Check status and pull
	statuses := collectRepoStatuses(repos)

	var results []pullResult
	var mu sync.Mutex

	var wg2 sync.WaitGroup
	for i, s := range statuses {
		switch git.RepoState(s.State) {
		case git.StateBehind:
			// Safe to pull
			wg2.Add(1)
			go func(r config.RepoInfo, name string) {
				defer wg2.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				res := pullResult{Name: name}
				if err := git.Pull(r.Path); err != nil {
					res.Action = "failed"
					res.Error = err.Error()
				} else {
					res.Action = "pulled"
				}
				mu.Lock()
				results = append(results, res)
				mu.Unlock()
			}(repos[i], s.Name)
		case git.StateClean:
			results = append(results, pullResult{Name: s.Name, Action: "skipped", Reason: "clean"})
		case git.StateAhead:
			results = append(results, pullResult{Name: s.Name, Action: "skipped", Reason: "ahead"})
		default:
			results = append(results, pullResult{Name: s.Name, Action: "skipped", Reason: s.State})
		}
	}
	wg2.Wait()

	if jsonOut {
		var pulled, skipped, failed []pullResult
		for _, r := range results {
			switch r.Action {
			case "pulled":
				pulled = append(pulled, r)
			case "skipped":
				skipped = append(skipped, r)
			case "failed":
				failed = append(failed, r)
			}
		}
		out := struct {
			Pulled  []pullResult `json:"pulled"`
			Skipped []pullResult `json:"skipped"`
			Failed  []pullResult `json:"failed"`
			Summary struct {
				Pulled  int `json:"pulled"`
				Skipped int `json:"skipped"`
				Failed  int `json:"failed"`
			} `json:"summary"`
		}{
			Pulled:  orEmpty(pulled),
			Skipped: orEmpty(skipped),
			Failed:  orEmpty(failed),
		}
		out.Summary.Pulled = len(pulled)
		out.Summary.Skipped = len(skipped)
		out.Summary.Failed = len(failed)
		data, _ := json.Marshal(out)
		fmt.Println(string(data))
		return nil
	}

	// Human output
	var pulledCount, skippedCount, failedCount int
	for _, r := range results {
		switch r.Action {
		case "pulled":
			fmt.Printf("  ✓ %s\n", r.Name)
			pulledCount++
		case "failed":
			fmt.Printf("  ✗ %s: %s\n", r.Name, r.Error)
			failedCount++
		case "skipped":
			skippedCount++
		}
	}
	fmt.Printf("\nDone: %d pulled, %d skipped, %d failed\n", pulledCount, skippedCount, failedCount)
	return nil
}

// --- push ---

type pushResult struct {
	Name  string `json:"name"`
	Error string `json:"error,omitempty"`
}

func runReposPush(args []string) error {
	jsonOut := hasFlag(args, "--json")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	repos, err := config.ListAllRepos(cfg.ProjectDirs)
	if err != nil {
		return fmt.Errorf("failed to list repos: %w", err)
	}

	if len(repos) == 0 {
		if jsonOut {
			fmt.Println(`{"pushed":[],"failed":[],"summary":{"pushed":0,"failed":0}}`)
		} else {
			fmt.Println("No repos found.")
		}
		return nil
	}

	// Check status (local only, no fetch)
	statuses := collectRepoStatuses(repos)

	// Find repos that are clean+ahead
	type pushCandidate struct {
		info config.RepoInfo
		name string
	}
	var candidates []pushCandidate
	for i, s := range statuses {
		st := git.RepoState(s.State)
		if st == git.StateAhead || st == git.StateDirtyAhead {
			candidates = append(candidates, pushCandidate{info: repos[i], name: s.Name})
		}
	}

	if len(candidates) == 0 {
		if jsonOut {
			fmt.Println(`{"pushed":[],"failed":[],"summary":{"pushed":0,"failed":0}}`)
		} else {
			fmt.Println("No repos to push (none in 'ahead' state).")
		}
		return nil
	}

	const maxNetwork = 4
	sem := make(chan struct{}, maxNetwork)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var pushed, failed []pushResult

	for _, c := range candidates {
		wg.Add(1)
		go func(r config.RepoInfo, name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := git.Push(r.Path); err != nil {
				mu.Lock()
				failed = append(failed, pushResult{Name: name, Error: err.Error()})
				mu.Unlock()
			} else {
				mu.Lock()
				pushed = append(pushed, pushResult{Name: name})
				mu.Unlock()
			}
		}(c.info, c.name)
	}
	wg.Wait()

	if jsonOut {
		out := struct {
			Pushed  []pushResult `json:"pushed"`
			Failed  []pushResult `json:"failed"`
			Summary struct {
				Pushed int `json:"pushed"`
				Failed int `json:"failed"`
			} `json:"summary"`
		}{
			Pushed: orEmpty(pushed),
			Failed: orEmpty(failed),
		}
		out.Summary.Pushed = len(pushed)
		out.Summary.Failed = len(failed)
		data, _ := json.Marshal(out)
		fmt.Println(string(data))
		return nil
	}

	for _, r := range pushed {
		fmt.Printf("  ✓ %s\n", r.Name)
	}
	for _, r := range failed {
		fmt.Printf("  ✗ %s: %s\n", r.Name, r.Error)
	}
	fmt.Printf("\nDone: %d pushed, %d failed\n", len(pushed), len(failed))
	return nil
}

// --- dirty ---

func runReposDirty(args []string) error {
	walk := hasFlag(args, "--walk")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	repos, err := config.ListAllRepos(cfg.ProjectDirs)
	if err != nil {
		return fmt.Errorf("failed to list repos: %w", err)
	}

	statuses := collectRepoStatuses(repos)

	var dirtyPaths []string
	for i, s := range statuses {
		st := git.RepoState(s.State)
		if st == git.StateDirty || st == git.StateDirtyAhead || st == git.StateDirtyBehind {
			dirtyPaths = append(dirtyPaths, repos[i].Path)
		}
	}

	if !walk {
		for _, p := range dirtyPaths {
			fmt.Println(p)
		}
		return nil
	}

	// --walk: run configured command on each dirty repo
	if cfg.DirtyWalkthroughCommand == "" {
		return fmt.Errorf("no dirty_walkthrough_command configured in %s", config.Path())
	}

	if len(dirtyPaths) == 0 {
		fmt.Println("No dirty repos.")
		return nil
	}

	for _, p := range dirtyPaths {
		cmdStr := strings.ReplaceAll(cfg.DirtyWalkthroughCommand, "{}", p)
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: command failed for %s: %v\n", p, err)
		}
	}

	return nil
}

// --- rebuild ---

func runReposRebuild(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.EnsureCloned) == 0 {
		fmt.Println("No ensure_cloned entries in config.")
		return nil
	}

	// Build owner/repo -> post_clone map
	postCloneMap := make(map[string]string)
	for _, entry := range cfg.EnsureCloned {
		if entry.PostClone == "" {
			continue
		}
		ownerRepo := github.ParseGitURL(entry.URL)
		if ownerRepo != "" {
			postCloneMap[ownerRepo] = entry.PostClone
		}
	}

	if len(postCloneMap) == 0 {
		fmt.Println("No post_clone hooks configured.")
		return nil
	}

	// Determine which repos to rebuild
	all := hasFlag(args, "--all")
	reposFlag := getFlagValue(args, "--repos")

	var targets []string
	if all {
		for name := range postCloneMap {
			targets = append(targets, name)
		}
	} else if reposFlag != "" {
		targets = strings.Split(reposFlag, ",")
	} else {
		fmt.Println("Specify --all or --repos owner/repo1,owner/repo2")
		return nil
	}

	// Resolve project dir for paths
	cloneDir, err := resolveCloneDir(cfg.ProjectDirs)
	if err != nil {
		return err
	}

	for _, name := range targets {
		name = strings.TrimSpace(name)
		cmd, ok := postCloneMap[name]
		if !ok {
			fmt.Printf("  ⊘ %s: no post_clone hook\n", name)
			continue
		}

		repoPath := filepath.Join(cloneDir, name)
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			fmt.Printf("  ⊘ %s: not cloned\n", name)
			continue
		}

		fmt.Printf("  → %s: %s\n", name, cmd)
		shellCmd := exec.Command("sh", "-c", cmd)
		shellCmd.Dir = repoPath
		shellCmd.Stdout = os.Stdout
		shellCmd.Stderr = os.Stderr
		if err := shellCmd.Run(); err != nil {
			fmt.Printf("    ✗ failed: %v\n", err)
		} else {
			fmt.Printf("    ✓ done\n")
		}
	}

	return nil
}

// orEmpty returns a non-nil empty slice if s is nil (for clean JSON output).
func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
