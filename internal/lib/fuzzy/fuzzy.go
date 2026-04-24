// Package fuzzy provides fuzzy string matching utilities.
package fuzzy

import "strings"

// Match checks if the pattern matches the text using fuzzy matching.
// Each character in the pattern must appear in the text in order,
// but not necessarily consecutively. Matching is case-insensitive.
func Match(text, pattern string) bool {
	if pattern == "" {
		return true
	}

	textRunes := []rune(strings.ToLower(text))
	patternRunes := []rune(pattern)
	patternIdx := 0

	for _, tr := range textRunes {
		if patternIdx >= len(patternRunes) {
			break
		}
		if tr == patternRunes[patternIdx] {
			patternIdx++
		}
	}

	return patternIdx == len(patternRunes)
}
