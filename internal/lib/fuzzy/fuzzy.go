// Package fuzzy provides fuzzy string matching utilities.
package fuzzy

import "strings"

// MatchPath checks if the pattern matches the text using fuzzy matching,
// with path segment awareness. Segments are separated by "/".
//
// If the pattern contains no "/", it matches against any segment of the
// text. This allows "imfusion" to match "imfusion/~brunner/agents" while
// still requiring fuzzy subsequence matching within each segment.
//
// If the pattern contains "/", segments are matched right-to-left. Each
// non-empty pattern segment must fuzzy-match its corresponding text segment.
// Empty pattern segments (from trailing "/") match any text segment. Extra
// text segments on the left are ignored (tail-matching).
//
// Matching within each segment uses Match, so standard fuzzy rules apply
// (case-insensitive, subsequence matching).
func MatchPath(text, pattern string) bool {
	if pattern == "" {
		return true
	}

	textSegments := strings.Split(text, "/")
	patternSegments := strings.Split(pattern, "/")

	// No "/" in pattern: match against any segment
	if len(patternSegments) == 1 {
		for _, segment := range textSegments {
			if Match(segment, pattern) {
				return true
			}
		}
		return false
	}

	// "/" in pattern: match segments right-to-left
	ti := len(textSegments) - 1
	for pi := len(patternSegments) - 1; pi >= 0; pi-- {
		if ti < 0 {
			return false // More pattern segments than text segments
		}

		if patternSegments[pi] == "" {
			// Empty segment (from trailing "/") matches any text segment
			ti--
			continue
		}

		if !Match(textSegments[ti], patternSegments[pi]) {
			return false
		}
		ti--
	}

	return true
}

// Match checks if the pattern matches the text using fuzzy matching.
// Each character in the pattern must appear in the text in order,
// but not necessarily consecutively. Matching is case-insensitive.
func Match(text, pattern string) bool {
	if pattern == "" {
		return true
	}

	patternRunes := []rune(pattern)
	patternIdx := 0

	for _, tr := range strings.ToLower(text) {
		if patternIdx >= len(patternRunes) {
			break
		}
		if tr == patternRunes[patternIdx] {
			patternIdx++
		}
	}

	return patternIdx == len(patternRunes)
}
