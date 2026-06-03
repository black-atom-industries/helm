package filter

import (
	"testing"
)

func TestFilter_BasicMatching(t *testing.T) {
	items := []string{"apple", "banana", "avocado", "blueberry"}
	f := New(items, MatchSessionName)

	// Empty filter returns all
	if f.Count() != 4 {
		t.Errorf("empty filter: got %d items, want 4", f.Count())
	}

	// Filter by subsequence (a appears in apple, banana, avocado)
	f.SetFilter("a")
	if f.Count() != 3 {
		t.Errorf("filter 'a': got %d items, want 3 (apple, banana, avocado)", f.Count())
	}

	// Filter by substring
	f.SetFilter("ban")
	if f.Count() != 1 {
		t.Errorf("filter 'ban': got %d items, want 1 (banana)", f.Count())
	}

	// No match
	f.SetFilter("xyz")
	if f.Count() != 0 {
		t.Errorf("filter 'xyz': got %d items, want 0", f.Count())
	}
}

func TestFilter_SetItems(t *testing.T) {
	f := New([]string{"a", "b"}, MatchSessionName)

	f.SetFilter("a")
	if f.Count() != 1 {
		t.Errorf("before SetItems: got %d, want 1 (only 'a' matches)", f.Count())
	}

	// Update items — 'a' matches alpha and beta (subsequence)
	f.SetItems([]string{"alpha", "beta"})
	if f.Count() != 2 {
		t.Errorf("after SetItems with filter 'a': got %d, want 2 (alpha, beta)", f.Count())
	}
}

func TestFilter_SessionNames(t *testing.T) {
	// Real session names from the user's environment
	sessions := []string{
		"black-atom-industries-core",
		"black-atom-industries-helm",
		"imfusion-brunner-agents",
		"nikbrunner-dots",
		"nikbrunner-flux-nvim",
		"nikbrunner-imf-notes",
		"nikbrunner-notes",
	}

	f := New(sessions, MatchSessionName)

	tests := []struct {
		filter  string
		wantAll []string
		noneOf  []string
	}{
		{"imfusion", []string{"imfusion-brunner-agents"}, []string{"black-atom-industries-helm"}},
		{"black", []string{"black-atom-industries-core", "black-atom-industries-helm"}, nil},
		{"helm", []string{"black-atom-industries-helm"}, []string{"black-atom-industries-core"}},
		{"nik", []string{"nikbrunner-dots", "nikbrunner-flux-nvim", "nikbrunner-imf-notes", "nikbrunner-notes"}, nil},
		{"agents", []string{"imfusion-brunner-agents"}, nil},
		{"xyz", nil, sessions},
		{"", sessions, nil},
	}

	for _, tt := range tests {
		t.Run("filter_"+tt.filter, func(t *testing.T) {
			f.SetFilter(tt.filter)
			results := f.Results()

			resultSet := make(map[string]bool)
			for _, r := range results {
				resultSet[r] = true
			}

			for _, want := range tt.wantAll {
				if !resultSet[want] {
					t.Errorf("filter %q should match %q but didn't. Got: %v", tt.filter, want, results)
				}
			}

			for _, none := range tt.noneOf {
				if resultSet[none] {
					t.Errorf("filter %q should NOT match %q but did. Got: %v", tt.filter, none, results)
				}
			}
		})
	}
}

func TestFilter_PathMatching(t *testing.T) {
	paths := []string{
		"imfusion/~brunner/agents",
		"imfusion/websdk/web-ui",
		"imfusion/websdk/web-viewer",
		"black-atom-industries/helm",
	}

	f := New(paths, MatchPath)

	tests := []struct {
		filter  string
		wantAll []string
		noneOf  []string
	}{
		{"imfusion", []string{"imfusion/~brunner/agents", "imfusion/websdk/web-ui", "imfusion/websdk/web-viewer"}, nil},
		{"brunner", []string{"imfusion/~brunner/agents"}, []string{"imfusion/websdk/web-ui"}},
		{"websdk", []string{"imfusion/websdk/web-ui", "imfusion/websdk/web-viewer"}, nil},
		{"helm", []string{"black-atom-industries/helm"}, nil},
		{"agents", []string{"imfusion/~brunner/agents"}, nil},
	}

	for _, tt := range tests {
		t.Run("filter_"+tt.filter, func(t *testing.T) {
			f.SetFilter(tt.filter)
			results := f.Results()

			resultSet := make(map[string]bool)
			for _, r := range results {
				resultSet[r] = true
			}

			for _, want := range tt.wantAll {
				if !resultSet[want] {
					t.Errorf("filter %q should match %q but didn't. Got: %v", tt.filter, want, results)
				}
			}

			for _, none := range tt.noneOf {
				if resultSet[none] {
					t.Errorf("filter %q should NOT match %q but did. Got: %v", tt.filter, none, results)
				}
			}
		})
	}
}
