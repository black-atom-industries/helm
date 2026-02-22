package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetRemoteURL returns the normalized HTTPS URL for the origin remote.
// Returns an error if the directory is not a git repo or has no remote.
func GetRemoteURL(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no git remote found")
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return "", fmt.Errorf("no git remote found")
	}

	return NormalizeRemoteURL(raw), nil
}

// NormalizeRemoteURL converts a git remote URL to an HTTPS URL.
// Handles SSH (git@host:org/repo.git) and HTTPS formats.
func NormalizeRemoteURL(raw string) string {
	// Strip trailing .git
	url := strings.TrimSuffix(raw, ".git")

	// Convert SSH format: git@github.com:org/repo -> https://github.com/org/repo
	if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@")
		url = strings.Replace(url, ":", "/", 1)
		url = "https://" + url
	}

	return url
}
