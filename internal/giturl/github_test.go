package giturl

import "testing"

// --- ParseGitURL tests ---

func TestParseGitURL_scp_with_git_suffix(t *testing.T) {
	got, err := ParseGitURL("git@github.com:black-atom-industries/helm.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "github.com" {
		t.Errorf("Host = %q, want %q", got.Host, "github.com")
	}
	if got.Path != "black-atom-industries/helm" {
		t.Errorf("Path = %q, want %q", got.Path, "black-atom-industries/helm")
	}
	if got.Port != 0 {
		t.Errorf("Port = %d, want 0", got.Port)
	}
}

func TestParseGitURL_scp_without_git_suffix(t *testing.T) {
	got, err := ParseGitURL("git@github.com:black-atom-industries/helm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "github.com" {
		t.Errorf("Host = %q, want %q", got.Host, "github.com")
	}
	if got.Path != "black-atom-industries/helm" {
		t.Errorf("Path = %q, want %q", got.Path, "black-atom-industries/helm")
	}
}

func TestParseGitURL_ssh_protocol_with_git_suffix(t *testing.T) {
	got, err := ParseGitURL("ssh://git@codeberg.org/ziglings/exercises.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "codeberg.org" {
		t.Errorf("Host = %q, want %q", got.Host, "codeberg.org")
	}
	if got.Path != "ziglings/exercises" {
		t.Errorf("Path = %q, want %q", got.Path, "ziglings/exercises")
	}
	if got.Port != 0 {
		t.Errorf("Port = %d, want 0", got.Port)
	}
}

func TestParseGitURL_ssh_protocol_without_git_suffix(t *testing.T) {
	got, err := ParseGitURL("ssh://git@codeberg.org/ziglings/exercises")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "codeberg.org" {
		t.Errorf("Host = %q, want %q", got.Host, "codeberg.org")
	}
	if got.Path != "ziglings/exercises" {
		t.Errorf("Path = %q, want %q", got.Path, "ziglings/exercises")
	}
}

func TestParseGitURL_ssh_protocol_with_port(t *testing.T) {
	got, err := ParseGitURL("ssh://git@git.corp.example.com:7999/~alice/my-project.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "git.corp.example.com" {
		t.Errorf("Host = %q, want %q", got.Host, "git.corp.example.com")
	}
	if got.Port != 7999 {
		t.Errorf("Port = %d, want 7999", got.Port)
	}
	if got.Path != "~alice/my-project" {
		t.Errorf("Path = %q, want %q", got.Path, "~alice/my-project")
	}
}

func TestParseGitURL_https_with_git_suffix(t *testing.T) {
	got, err := ParseGitURL("https://github.com/black-atom-industries/helm.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "github.com" {
		t.Errorf("Host = %q, want %q", got.Host, "github.com")
	}
	if got.Path != "black-atom-industries/helm" {
		t.Errorf("Path = %q, want %q", got.Path, "black-atom-industries/helm")
	}
}

func TestParseGitURL_https_without_git_suffix(t *testing.T) {
	got, err := ParseGitURL("https://github.com/black-atom-industries/helm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "github.com" {
		t.Errorf("Host = %q, want %q", got.Host, "github.com")
	}
	if got.Path != "black-atom-industries/helm" {
		t.Errorf("Path = %q, want %q", got.Path, "black-atom-industries/helm")
	}
}

func TestParseGitURL_rejects_empty_string(t *testing.T) {
	_, err := ParseGitURL("")
	if err == nil {
		t.Fatal("expected error for empty string, got nil")
	}
}

func TestParseGitURL_rejects_bare_name(t *testing.T) {
	_, err := ParseGitURL("helm")
	if err == nil {
		t.Fatal("expected error for bare name, got nil")
	}
}

func TestParseGitURL_rejects_incomplete_ssh(t *testing.T) {
	_, err := ParseGitURL("git@example.com:")
	if err == nil {
		t.Fatal("expected error for incomplete SSH URL, got nil")
	}
}

// --- ResolveOwnerRepo tests (backward compat) ---

func TestResolveOwnerRepo_plain_owner_repo(t *testing.T) {
	got, err := ResolveOwnerRepo("black-atom-industries/helm", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "black-atom-industries/helm" {
		t.Errorf("got %q, want %q", got, "black-atom-industries/helm")
	}
}

func TestResolveOwnerRepo_ssh_url(t *testing.T) {
	got, err := ResolveOwnerRepo("git@github.com:black-atom-industries/helm.git", map[string]string{"github.com": ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "black-atom-industries/helm" {
		t.Errorf("got %q, want %q", got, "black-atom-industries/helm")
	}
}

func TestResolveOwnerRepo_https_url(t *testing.T) {
	got, err := ResolveOwnerRepo("https://github.com/black-atom-industries/helm", map[string]string{"github.com": ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "black-atom-industries/helm" {
		t.Errorf("got %q, want %q", got, "black-atom-industries/helm")
	}
}

func TestResolveOwnerRepo_rejects_bare_name(t *testing.T) {
	_, err := ResolveOwnerRepo("helm", nil)
	if err == nil {
		t.Fatal("expected error for bare name without owner, got nil")
	}
}

func TestResolveOwnerRepo_rejects_unparseable_url(t *testing.T) {
	_, err := ResolveOwnerRepo("git@example.com:", nil)
	if err == nil {
		t.Fatal("expected error for unparseable URL, got nil")
	}
}

func TestResolveOwnerRepo_ssh_protocol_url(t *testing.T) {
	got, err := ResolveOwnerRepo("ssh://git@codeberg.org/ziglings/exercises.git", map[string]string{"codeberg.org": ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ziglings/exercises" {
		t.Errorf("got %q, want %q", got, "ziglings/exercises")
	}
}

func TestResolveOwnerRepo_codeberg_https(t *testing.T) {
	got, err := ResolveOwnerRepo("https://codeberg.org/ziglings/exercises.git", map[string]string{"codeberg.org": ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ziglings/exercises" {
		t.Errorf("got %q, want %q", got, "ziglings/exercises")
	}
}

func TestResolveOwnerRepo_with_providers_config(t *testing.T) {
	// When GitHub is in config with empty alias, resolves to owner/repo
	got, err := ResolveOwnerRepo("git@github.com:owner/repo.git", map[string]string{"github.com": ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "owner/repo" {
		t.Errorf("got %q, want %q", got, "owner/repo")
	}

	// When self-hosted host is in config with alias, resolves to alias/path
	got, err = ResolveOwnerRepo("ssh://git@git.corp.example.com:7999/~alice/my-project.git", map[string]string{"git.corp.example.com": "corp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "corp/alice/my-project" {
		t.Errorf("got %q, want %q", got, "corp/alice/my-project")
	}
}

// --- ResolveRepoDir tests ---

func TestResolveRepoDir_github(t *testing.T) {
	u, _ := ParseGitURL("git@github.com:owner/repo.git")
	got := ResolveRepoDir(u, map[string]string{"github.com": ""})
	if got != "owner/repo" {
		t.Errorf("got %q, want %q", got, "owner/repo")
	}
}

func TestResolveRepoDir_codeberg(t *testing.T) {
	u, _ := ParseGitURL("ssh://git@codeberg.org/ziglings/exercises.git")
	got := ResolveRepoDir(u, map[string]string{"codeberg.org": ""})
	if got != "ziglings/exercises" {
		t.Errorf("got %q, want %q", got, "ziglings/exercises")
	}
}

func TestResolveRepoDir_personal_repo_with_alias(t *testing.T) {
	u, _ := ParseGitURL("ssh://git@git.corp.example.com:7999/~alice/my-project.git")
	got := ResolveRepoDir(u, map[string]string{"git.corp.example.com": "corp"})
	if got != "corp/alice/my-project" {
		t.Errorf("got %q, want %q", got, "corp/alice/my-project")
	}
}

func TestResolveRepoDir_team_repo_with_alias(t *testing.T) {
	u, _ := ParseGitURL("ssh://git@git.corp.example.com:7999/websdk/web-ui.git")
	got := ResolveRepoDir(u, map[string]string{"git.corp.example.com": "corp"})
	if got != "corp/websdk/web-ui" {
		t.Errorf("got %q, want %q", got, "corp/websdk/web-ui")
	}
}

func TestResolveRepoDir_unknown_host_uses_prefix(t *testing.T) {
	u, _ := ParseGitURL("git@gitlab.com:mygroup/myproject.git")
	got := ResolveRepoDir(u, map[string]string{"github.com": ""})
	if got != "gitlab.com/mygroup/myproject" {
		t.Errorf("got %q, want %q", got, "gitlab.com/mygroup/myproject")
	}
}

func TestResolveRepoDir_tilde_preserved_in_middle(t *testing.T) {
	u := GitURL{Host: "example.com", Path: "a/~user/b"}
	got := ResolveRepoDir(u, map[string]string{"example.com": "x"})
	if got != "x/a/~user/b" {
		t.Errorf("got %q, want %q", got, "x/a/~user/b")
	}
}

func TestResolveRepoDir_strips_leading_tilde(t *testing.T) {
	u := GitURL{Host: "bitbucket.corp.com", Path: "~alice/repo"}
	got := ResolveRepoDir(u, map[string]string{"bitbucket.corp.com": "bb"})
	if got != "bb/alice/repo" {
		t.Errorf("got %q, want %q", got, "bb/alice/repo")
	}
}

func TestResolveRepoDir_leading_slash_stripped(t *testing.T) {
	u := GitURL{Host: "example.com", Path: "/owner/repo"}
	got := ResolveRepoDir(u, map[string]string{"example.com": ""})
	if got != "owner/repo" {
		t.Errorf("got %q, want %q", got, "owner/repo")
	}
}

func TestResolveRepoDir_nil_providers(t *testing.T) {
	u := GitURL{Host: "gitlab.com", Path: "group/project"}
	got := ResolveRepoDir(u, nil)
	if got != "gitlab.com/group/project" {
		t.Errorf("got %q, want %q", got, "gitlab.com/group/project")
	}
}
