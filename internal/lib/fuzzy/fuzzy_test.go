package fuzzy

import "testing"

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		pattern string
		want    bool
	}{
		{
			name:    "exact match",
			text:    "hello",
			pattern: "hello",
			want:    true,
		},
		{
			name:    "case insensitive",
			text:    "Hello",
			pattern: "hello",
			want:    true,
		},
		{
			name:    "substring match",
			text:    "hello-world",
			pattern: "world",
			want:    true,
		},
		{
			name:    "no match",
			text:    "hello",
			pattern: "xyz",
			want:    false,
		},
		{
			name:    "empty pattern matches all",
			text:    "hello",
			pattern: "",
			want:    true,
		},
		{
			name:    "empty text with pattern",
			text:    "",
			pattern: "hello",
			want:    false,
		},
		{
			name:    "both empty",
			text:    "",
			pattern: "",
			want:    true,
		},
		{
			name:    "fuzzy match with separator",
			text:    "nikbrunner/imf-notes",
			pattern: "imfnotes",
			want:    true,
		},
		{
			name:    "fuzzy match abbreviated",
			text:    "hello-world",
			pattern: "hw",
			want:    true,
		},
		{
			name:    "fuzzy match out of order",
			text:    "acb",
			pattern: "abc",
			want:    false,
		},
		{
			name:    "fuzzy match case insensitive with gaps",
			text:    "Black-Atom-Industries",
			pattern: "bai",
			want:    true,
		},
		{
			name:    "fuzzy match with repeated characters",
			text:    "banana",
			pattern: "bnn",
			want:    true,
		},
		{
			name:    "fuzzy match unicode text",
			text:    "日本語テキスト",
			pattern: "日本",
			want:    true,
		},
		{
			name:    "fuzzy match unicode pattern with gaps",
			text:    "日本語テキスト",
			pattern: "日テ",
			want:    true,
		},
		{
			name:    "fuzzy match pattern longer than text",
			text:    "hi",
			pattern: "hello",
			want:    false,
		},
		{
			name:    "fuzzy match single character",
			text:    "hello",
			pattern: "e",
			want:    true,
		},
		{
			name:    "fuzzy match single character not found",
			text:    "hello",
			pattern: "z",
			want:    false,
		},
		{
			name:    "fuzzy match with numbers",
			text:    "helm-2.0-release",
			pattern: "h20r",
			want:    true,
		},
		{
			name:    "fuzzy match with spaces",
			text:    "my session name",
			pattern: "msn",
			want:    true,
		},
		{
			name:    "fuzzy match underscores",
			text:    "some_test_file.go",
			pattern: "stfg",
			want:    true,
		},
		{
			name:    "fuzzy match mixed case",
			text:    "GitHubActions",
			pattern: "gha",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Match(tt.text, tt.pattern)
			if got != tt.want {
				t.Errorf("Match(%q, %q) = %v, want %v", tt.text, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		pattern string
		want    bool
	}{
		// No "/" in pattern — match against last segment only
		{
			name:    "last segment exact match",
			text:    "black-atom-industries/core",
			pattern: "core",
			want:    true,
		},
		{
			name:    "last segment no match",
			text:    "black-atom-industries/ai",
			pattern: "core",
			want:    false,
		},
		{
			name:    "org chars scattered in last segment no match",
			text:    "black-atom-industries/docs",
			pattern: "core",
			want:    false,
		},
		{
			name:    "fuzzy within last segment",
			text:    "nikbrunner/imf-notes",
			pattern: "imfnotes",
			want:    true,
		},
		{
			name:    "abbreviated within last segment",
			text:    "hello-world",
			pattern: "hw",
			want:    true,
		},
		{
			name:    "no segments text matches normally",
			text:    "core",
			pattern: "core",
			want:    true,
		},
		{
			name:    "single segment no match",
			text:    "ai",
			pattern: "core",
			want:    false,
		},
		{
			name:    "empty pattern matches all",
			text:    "black-atom-industries/core",
			pattern: "",
			want:    true,
		},
		{
			name:    "empty text with pattern",
			text:    "",
			pattern: "core",
			want:    false,
		},

		// "/" in pattern — match segments right-to-left
		{
			name:    "slash pattern matches org+repo",
			text:    "black-atom-industries/core",
			pattern: "black/core",
			want:    true,
		},
		{
			name:    "slash pattern no match on repo",
			text:    "black-atom-industries/ai",
			pattern: "black/core",
			want:    false,
		},
		{
			name:    "slash pattern fuzzy org match",
			text:    "black-atom-industries/core",
			pattern: "indu/core",
			want:    true,
		},
		{
			name:    "more pattern segments than text",
			text:    "core",
			pattern: "black/core",
			want:    false,
		},
		{
			name:    "fewer pattern segments than text (tail match)",
			text:    "some/deep/black-atom-industries/core",
			pattern: "industries/core",
			want:    true,
		},
		{
			name:    "trailing slash matches all repos in org",
			text:    "black-atom-industries/core",
			pattern: "black-atom-industries/",
			want:    true,
		},
		{
			name:    "leading slash no match wrong repo",
			text:    "black-atom-industries/ai",
			pattern: "/core",
			want:    false,
		},
		{
			name:    "leading slash match right repo",
			text:    "black-atom-industries/core",
			pattern: "/core",
			want:    true,
		},
		{
			name:    "triple segment match tail",
			text:    "a/b/c",
			pattern: "b/c",
			want:    true,
		},
		{
			name:    "triple segment no match wrong tail",
			text:    "a/b/c",
			pattern: "a/c",
			want:    false,
		},
		{
			name:    "deeper path than pattern segments",
			text:    "w/x/deep/path/here",
			pattern: "path/here",
			want:    true,
		},
		{
			name:    "empty pattern segments for trailing slash",
			text:    "black-atom-industries/ui",
			pattern: "industries/",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchPath(tt.text, tt.pattern)
			if got != tt.want {
				t.Errorf("MatchPath(%q, %q) = %v, want %v", tt.text, tt.pattern, got, tt.want)
			}
		})
	}
}

func BenchmarkMatch(b *testing.B) {
	text := "nikbrunner/imf-notes"
	pattern := "imfnotes"

	for i := 0; i < b.N; i++ {
		Match(text, pattern)
	}
}

func BenchmarkMatchLongText(b *testing.B) {
	text := "github.com/black-atom-industries/helm/internal/lib/fuzzy/fuzzy.go"
	pattern := "gibailifugo"

	for i := 0; i < b.N; i++ {
		Match(text, pattern)
	}
}
