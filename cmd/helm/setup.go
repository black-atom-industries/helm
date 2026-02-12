package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/github"
	"github.com/black-atom-industries/helm/internal/repos"
)

// setupResult tracks the outcome of a single clone operation
type setupResult struct {
	repo   string
	status string // "cloned", "skipped", "failed"
	err    error
}

// runSetup executes the helm setup subcommand.
// Clones all repositories from ensure_cloned config with parallel execution.
func runSetup() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.EnsureCloned) == 0 {
		fmt.Println("No ensure_cloned entries in config.")
		fmt.Printf("Add repositories to %s under the ensure_cloned key.\n", config.Path())
		return nil
	}

	// Determine clone target directory
	cloneDir, err := resolveCloneDir(cfg.ProjectDirs)
	if err != nil {
		return err
	}

	fmt.Printf("Clone target: %s\n", cloneDir)

	// Expand all entries to concrete URLs
	urls, postCloneMap, err := expandEntries(cfg.EnsureCloned)
	if err != nil {
		return err
	}

	if len(urls) == 0 {
		fmt.Println("No repositories to clone after expansion.")
		return nil
	}

	// Get already cloned repos to skip
	cloned, _ := repos.ListClonedRepos(cloneDir)
	clonedSet := make(map[string]bool)
	for _, r := range cloned {
		clonedSet[r] = true
	}

	fmt.Printf("Found %d repositories to process\n\n", len(urls))

	// Clone in parallel (max 4 concurrent)
	const maxParallel = 4
	sem := make(chan struct{}, maxParallel)
	var mu sync.Mutex
	var results []setupResult

	var wg sync.WaitGroup
	for _, url := range urls {
		ownerRepo := parseGitURL(url)
		if ownerRepo == "" {
			mu.Lock()
			results = append(results, setupResult{repo: url, status: "failed", err: fmt.Errorf("could not parse URL")})
			mu.Unlock()
			continue
		}

		if clonedSet[ownerRepo] {
			mu.Lock()
			results = append(results, setupResult{repo: ownerRepo, status: "skipped"})
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(gitURL, repo string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			destPath := filepath.Join(cloneDir, repo)
			result := setupResult{repo: repo}

			if err := github.CloneRepo(repo, destPath); err != nil {
				result.status = "failed"
				result.err = err
			} else {
				result.status = "cloned"
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(url, ownerRepo)
	}
	wg.Wait()

	// Print results
	var clonedCount, skippedCount, failedCount int
	for _, r := range results {
		switch r.status {
		case "cloned":
			fmt.Printf("  ✓ %s\n", r.repo)
			clonedCount++
		case "skipped":
			skippedCount++
		case "failed":
			fmt.Printf("  ✗ %s: %v\n", r.repo, r.err)
			failedCount++
		}
	}

	fmt.Printf("\nDone: %d cloned, %d skipped, %d failed\n", clonedCount, skippedCount, failedCount)

	// Run post_clone commands for successfully cloned repos
	if clonedCount > 0 && len(postCloneMap) > 0 {
		fmt.Println("\nRunning post_clone commands...")
		for _, r := range results {
			if r.status != "cloned" {
				continue
			}
			cmd, ok := postCloneMap[r.repo]
			if !ok {
				continue
			}
			repoPath := filepath.Join(cloneDir, r.repo)
			fmt.Printf("  %s: %s\n", r.repo, cmd)
			if err := runPostClone(repoPath, cmd); err != nil {
				fmt.Printf("    ✗ post_clone failed: %v\n", err)
			} else {
				fmt.Printf("    ✓ post_clone done\n")
			}
		}
	}

	return nil
}

// resolveCloneDir picks the clone target from project_dirs.
// If there's one entry, uses it. If multiple, prompts the user.
func resolveCloneDir(dirs []string) (string, error) {
	if len(dirs) == 0 {
		return "", fmt.Errorf("no project_dirs configured in %s", config.Path())
	}

	if len(dirs) == 1 {
		return dirs[0], nil
	}

	fmt.Println("Multiple project directories configured. Select clone target:")
	for i, d := range dirs {
		fmt.Printf("  [%d] %s\n", i+1, d)
	}
	fmt.Print("Choice: ")

	var choice int
	if _, err := fmt.Scan(&choice); err != nil || choice < 1 || choice > len(dirs) {
		return "", fmt.Errorf("invalid selection")
	}

	return dirs[choice-1], nil
}

// expandEntries resolves all ensure_cloned entries to concrete git URLs.
// Returns the URL list and a map of owner/repo -> post_clone command.
func expandEntries(entries []config.EnsureClonedEntry) ([]string, map[string]string, error) {
	var urls []string
	postCloneMap := make(map[string]string)

	for _, entry := range entries {
		url := entry.URL
		if url == "" {
			continue
		}

		// Check for wildcard pattern
		if isWildcard(url) {
			expanded, err := expandWildcard(url)
			if err != nil {
				fmt.Printf("  ⚠ Failed to expand %s: %v\n", url, err)
				continue
			}
			urls = append(urls, expanded...)
			continue
		}

		urls = append(urls, url)

		// Track post_clone command by owner/repo
		if entry.PostClone != "" {
			ownerRepo := parseGitURL(url)
			if ownerRepo != "" {
				postCloneMap[ownerRepo] = entry.PostClone
			}
		}
	}

	return urls, postCloneMap, nil
}

// isWildcard checks if a URL contains an org/* wildcard pattern
func isWildcard(url string) bool {
	return strings.HasSuffix(url, "/*") || strings.HasSuffix(url, "/*.git")
}

// expandWildcard expands an org/* pattern to individual repo URLs via gh CLI
func expandWildcard(url string) ([]string, error) {
	if err := github.CheckGhCli(); err != nil {
		return nil, err
	}

	// Extract org/user from the URL
	orgUser := extractOrgFromURL(url)
	if orgUser == "" {
		return nil, fmt.Errorf("could not extract org/user from: %s", url)
	}

	// Fetch repos via gh CLI
	out, err := exec.Command("gh", "repo", "list", orgUser, "--limit", "1000", "--json", "name", "--jq", ".[].name").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list repos for %s: %w", orgUser, err)
	}

	names := strings.Split(strings.TrimSpace(string(out)), "\n")

	// Reconstruct URLs using the same base
	base := strings.TrimSuffix(url, "/*")
	base = strings.TrimSuffix(base, "/*.git")

	var urls []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		urls = append(urls, base+"/"+name+".git")
	}

	return urls, nil
}

// extractOrgFromURL extracts the org/user from a git URL wildcard pattern
func extractOrgFromURL(url string) string {
	// SSH format: git@github.com:org/*
	sshRe := regexp.MustCompile(`git@[^:]+:([^/]+)/\*`)
	if m := sshRe.FindStringSubmatch(url); len(m) == 2 {
		return m[1]
	}

	// HTTPS format: https://github.com/org/*
	httpsRe := regexp.MustCompile(`https?://[^/]+/([^/]+)/\*`)
	if m := httpsRe.FindStringSubmatch(url); len(m) == 2 {
		return m[1]
	}

	return ""
}

// parseGitURL extracts owner/repo from a git URL
func parseGitURL(url string) string {
	// SSH: git@github.com:owner/repo.git
	sshRe := regexp.MustCompile(`git@[^:]+:(.+?)(?:\.git)?$`)
	if m := sshRe.FindStringSubmatch(url); len(m) == 2 {
		return m[1]
	}

	// HTTPS: https://github.com/owner/repo.git
	httpsRe := regexp.MustCompile(`https?://[^/]+/(.+?)(?:\.git)?$`)
	if m := httpsRe.FindStringSubmatch(url); len(m) == 2 {
		return m[1]
	}

	return ""
}

// runPostClone executes a post_clone command in the repo directory
func runPostClone(repoPath, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
