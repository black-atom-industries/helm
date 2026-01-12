package repos

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Config represents the repos configuration from ~/.config/repos/config.json
type Config struct {
	ReposBasePath string `json:"repos_base_path"`
}

// DefaultBasePath returns the default repos base path
func DefaultBasePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "repos")
}

// LoadConfig reads the repos config file
// Returns default config if file doesn't exist
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return &Config{ReposBasePath: DefaultBasePath()}, nil
	}

	configPath := filepath.Join(home, ".config", "repos", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config doesn't exist, use defaults
		return &Config{ReposBasePath: DefaultBasePath()}, nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{ReposBasePath: DefaultBasePath()}, nil
	}

	// Expand ~ in path
	if strings.HasPrefix(cfg.ReposBasePath, "~") {
		cfg.ReposBasePath = filepath.Join(home, cfg.ReposBasePath[1:])
	}

	// Use default if empty
	if cfg.ReposBasePath == "" {
		cfg.ReposBasePath = DefaultBasePath()
	}

	return &cfg, nil
}

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
