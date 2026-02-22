package git

import "testing"

func TestNormalizeRemoteURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "HTTPS URL unchanged",
			input: "https://github.com/org/repo.git",
			want:  "https://github.com/org/repo",
		},
		{
			name:  "HTTPS URL without .git",
			input: "https://github.com/org/repo",
			want:  "https://github.com/org/repo",
		},
		{
			name:  "SSH URL converted to HTTPS",
			input: "git@github.com:org/repo.git",
			want:  "https://github.com/org/repo",
		},
		{
			name:  "SSH URL without .git",
			input: "git@github.com:org/repo",
			want:  "https://github.com/org/repo",
		},
		{
			name:  "SSH with different host",
			input: "git@gitlab.com:org/repo.git",
			want:  "https://gitlab.com/org/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeRemoteURL(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeRemoteURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
