package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// CheckGhCli verifies that gh CLI is installed and authenticated
func CheckGhCli() error {
	// Check if gh is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("GitHub CLI (gh) is not installed. Install from: https://cli.github.com/")
	}

	// Check if authenticated
	if err := exec.Command("gh", "auth", "status").Run(); err != nil {
		return fmt.Errorf("GitHub CLI is not authenticated. Run: gh auth login")
	}

	return nil
}

// FetchAvailableRepos returns repos the user has access to (owner/repo format)
func FetchAvailableRepos() ([]string, error) {
	out, err := exec.Command("gh", "api", "/user/repos", "--paginate", "--jq", ".[].full_name").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []string{}, nil
	}

	return lines, nil
}

// CloneRepo clones a repository to the specified destination path
func CloneRepo(ownerRepo, destPath string) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(destPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	// Construct SSH URL
	gitURL := fmt.Sprintf("git@github.com:%s.git", ownerRepo)

	// Clone the repository
	cmd := exec.Command("git", "clone", gitURL, destPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone %s: %w", ownerRepo, err)
	}

	return nil
}

// ParseGitURL extracts owner/repo from a git URL (SSH or HTTPS).
// Returns empty string if the URL cannot be parsed.
func ParseGitURL(url string) string {
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
