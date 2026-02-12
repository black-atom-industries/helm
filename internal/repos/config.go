package repos

import (
	"os"
	"path/filepath"
	"sort"
)

// ListClonedRepos returns already-cloned repos in owner/repo format
// Scans the base path at depth 2 (owner/repo structure)
func ListClonedRepos(basePath string) ([]string, error) {
	var repos []string

	// Read owner directories
	owners, err := os.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	for _, owner := range owners {
		if !owner.IsDir() {
			continue
		}

		ownerPath := filepath.Join(basePath, owner.Name())
		repoEntries, err := os.ReadDir(ownerPath)
		if err != nil {
			continue
		}

		for _, repo := range repoEntries {
			if !repo.IsDir() {
				continue
			}

			// Check if it's a git repo
			gitPath := filepath.Join(ownerPath, repo.Name(), ".git")
			if _, err := os.Stat(gitPath); err == nil {
				repos = append(repos, owner.Name()+"/"+repo.Name())
			}
		}
	}

	sort.Strings(repos)
	return repos, nil
}

// FilterUncloned returns repos from available that are not in cloned
func FilterUncloned(available, cloned []string) []string {
	clonedSet := make(map[string]bool)
	for _, r := range cloned {
		clonedSet[r] = true
	}

	var uncloned []string
	for _, r := range available {
		if !clonedSet[r] {
			uncloned = append(uncloned, r)
		}
	}

	return uncloned
}
