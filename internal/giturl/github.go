package giturl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// GitURL represents a parsed git URL with structured fields.
type GitURL struct {
	Scheme string // "ssh", "https", "git" (SCP-like: git@host:path)
	Host   string // hostname (no port)
	Port   int    // 0 if default
	Path   string // repo path relative to host (without .git)
}

var (
	// SCP-like: git@github.com:owner/repo.git
	scpRe = regexp.MustCompile(`^git@([^:]+):(.+?)(?:\.git)?$`)

	// SSH protocol: ssh://git@host:port/path/repo.git
	sshProtoRe = regexp.MustCompile(`^ssh://[^@]+@([^:/]+)(?::(\d+))?/(.+?)(?:\.git)?$`)

	// HTTPS: https://host/path/repo.git
	httpsRe = regexp.MustCompile(`^https?://([^/]+)/(.+?)(?:\.git)?$`)
)

// ParseGitURL parses any supported git URL into a GitURL struct.
// Supported formats:
//
//	SCP-like:    git@github.com:owner/repo.git
//	SSH protocol: ssh://git@host:port/path/repo.git
//	HTTPS:       https://host/path/repo.git
func ParseGitURL(url string) (GitURL, error) {
	if url == "" {
		return GitURL{}, fmt.Errorf("empty URL")
	}

	// SCP-like: git@host:path.git
	if m := scpRe.FindStringSubmatch(url); m != nil {
		return GitURL{
			Scheme: "git",
			Host:   m[1],
			Path:   m[2],
		}, nil
	}

	// SSH protocol: ssh://git@host:port/path.git
	if m := sshProtoRe.FindStringSubmatch(url); m != nil {
		port := 0
		if m[2] != "" {
			port, _ = strconv.Atoi(m[2])
		}
		return GitURL{
			Scheme: "ssh",
			Host:   m[1],
			Port:   port,
			Path:   m[3],
		}, nil
	}

	// HTTPS: https://host/path.git
	if m := httpsRe.FindStringSubmatch(url); m != nil {
		return GitURL{
			Scheme: "https",
			Host:   m[1],
			Path:   m[2],
		}, nil
	}

	return GitURL{}, fmt.Errorf("could not parse git URL: %s", url)
}

// cleanPath strips a leading slash from a path but preserves ~ prefixes
// (which distinguish personal repos from team/org repos).
func cleanPath(path string) string {
	return strings.TrimPrefix(path, "/")
}

// ResolveRepoDir returns the directory name for a parsed git URL.
// The providers map is from config.GitProviders (host → alias).
//
// Rules:
//  1. Look up GitURL.Host in providers map
//  2. If found with alias: "alias/{cleaned_path}"
//  3. If found with empty string: "{cleaned_path}"
//  4. If not found: "{host}/{cleaned_path}"
func ResolveRepoDir(gitURL GitURL, providers map[string]string) string {
	cleaned := cleanPath(gitURL.Path)
	if providers == nil {
		return gitURL.Host + "/" + cleaned
	}
	alias, ok := providers[gitURL.Host]
	if !ok {
		return gitURL.Host + "/" + cleaned
	}
	if alias == "" {
		return cleaned
	}
	return alias + "/" + cleaned
}

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

// CloneRepo clones a repository to the specified destination path.
// gitURL should be a full clone URL (ssh://, git@, https://, or owner/repo).
// If given owner/repo, it defaults to git@github.com:owner/repo.git.
func CloneRepo(gitURL, destPath string) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(destPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	// If it's just owner/repo (no host), default to GitHub SSH
	if !strings.Contains(gitURL, "@") && !strings.Contains(gitURL, "://") && strings.Contains(gitURL, "/") {
		gitURL = fmt.Sprintf("git@github.com:%s.git", gitURL)
	}

	// Clone the repository
	cmd := exec.Command("git", "clone", gitURL, destPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone %s: %w\n%s", gitURL, err, string(out))
	}

	return nil
}

// ResolveOwnerRepo normalizes a repo identifier to owner/repo format.
// Accepts: "owner/repo", SSH URLs, SSH protocol URLs, or HTTPS URLs.
// The providers map (from config.GitProviders) is used to resolve directory names
// for non-GitHub hosts (host-qualified paths like "gitlab.com/group/repo").
func ResolveOwnerRepo(input string, providers map[string]string) (string, error) {
	// If it looks like a URL, parse it
	if strings.Contains(input, "@") || strings.Contains(input, "://") {
		parsed, err := ParseGitURL(input)
		if err != nil {
			return "", fmt.Errorf("could not parse repo from URL: %s", input)
		}
		return ResolveRepoDir(parsed, providers), nil
	}

	// Must contain a slash for owner/repo format
	if !strings.Contains(input, "/") {
		return "", fmt.Errorf("invalid repo format: %s (expected owner/repo)", input)
	}

	return input, nil
}
