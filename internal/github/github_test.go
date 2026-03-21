package github

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
