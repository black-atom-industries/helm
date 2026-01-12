package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home := os.Getenv("HOME")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expands tilde",
			input:    "~/foo/bar",
			expected: filepath.Join(home, "foo/bar"),
		},
		{
			name:     "leaves absolute path unchanged",
			input:    "/usr/local/bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "leaves relative path unchanged",
			input:    "foo/bar",
			expected: "foo/bar",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles tilde only",
			input:    "~",
			expected: home,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Check that defaults are set
	if cfg.ProjectDepth != 2 {
		t.Errorf("ProjectDepth = %d, want 2", cfg.ProjectDepth)
	}

	if cfg.ClaudeStatusEnabled != false {
		t.Error("ClaudeStatusEnabled should be false by default")
	}

	if cfg.Layout != "" {
		t.Errorf("Layout = %q, want empty string", cfg.Layout)
	}
}

func TestPath(t *testing.T) {
	home := os.Getenv("HOME")
	expected := filepath.Join(home, ".config", "tsm", "config.toml")

	result := Path()
	if result != expected {
		t.Errorf("Path() = %q, want %q", result, expected)
	}
}
