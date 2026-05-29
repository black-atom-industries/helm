package giturl

import "testing"

func TestResolveOwnerRepo_plain_owner_repo(t *testing.T) {
	got, err := ResolveOwnerRepo("black-atom-industries/helm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "black-atom-industries/helm" {
		t.Errorf("got %q, want %q", got, "black-atom-industries/helm")
	}
}

func TestResolveOwnerRepo_ssh_url(t *testing.T) {
	got, err := ResolveOwnerRepo("git@github.com:black-atom-industries/helm.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "black-atom-industries/helm" {
		t.Errorf("got %q, want %q", got, "black-atom-industries/helm")
	}
}

func TestResolveOwnerRepo_https_url(t *testing.T) {
	got, err := ResolveOwnerRepo("https://github.com/black-atom-industries/helm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "black-atom-industries/helm" {
		t.Errorf("got %q, want %q", got, "black-atom-industries/helm")
	}
}

func TestResolveOwnerRepo_rejects_bare_name(t *testing.T) {
	_, err := ResolveOwnerRepo("helm")
	if err == nil {
		t.Fatal("expected error for bare name without owner, got nil")
	}
}

func TestResolveOwnerRepo_rejects_unparseable_url(t *testing.T) {
	_, err := ResolveOwnerRepo("git@example.com:")
	if err == nil {
		t.Fatal("expected error for unparseable URL, got nil")
	}
}

func TestResolveOwnerRepo_ssh_protocol_url(t *testing.T) {
	got, err := ResolveOwnerRepo("ssh://git@codeberg.org/ziglings/exercises.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ziglings/exercises" {
		t.Errorf("got %q, want %q", got, "ziglings/exercises")
	}
}

func TestResolveOwnerRepo_codeberg_https(t *testing.T) {
	got, err := ResolveOwnerRepo("https://codeberg.org/ziglings/exercises.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ziglings/exercises" {
		t.Errorf("got %q, want %q", got, "ziglings/exercises")
	}
}

func TestParseGitURL_ssh_protocol_no_git_suffix(t *testing.T) {
	got := ParseGitURL("ssh://git@codeberg.org/ziglings/exercises")
	if got != "ziglings/exercises" {
		t.Errorf("got %q, want %q", got, "ziglings/exercises")
	}
}

func TestParseGitURL_github_ssh(t *testing.T) {
	got := ParseGitURL("git@github.com:black-atom-industries/helm.git")
	if got != "black-atom-industries/helm" {
		t.Errorf("got %q, want %q", got, "black-atom-industries/helm")
	}
}

func TestParseGitURL_github_https(t *testing.T) {
	got := ParseGitURL("https://github.com/black-atom-industries/helm")
	if got != "black-atom-industries/helm" {
		t.Errorf("got %q, want %q", got, "black-atom-industries/helm")
	}
}
