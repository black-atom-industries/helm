// Package theme holds the Black Atom theme registry.
//
// helm is a Black Atom adapter: the sibling black-atom-*.go files are
// generated from collection.template.go via `make themes` (which runs
// `deno task generate` against jsr:@black-atom/core). Each generated file
// self-registers in init(), so the full theme set is embedded in the
// binary at compile time.
package theme

import (
	"sort"
	"strings"
)

// Theme is the subset of Black Atom theme tokens helm consumes.
// All color values are hex strings like "#1a1b26".
type Theme struct {
	Key        string // e.g. "black-atom-jpn-koyo-yoru"
	Appearance string // "dark" or "light"

	// UI backgrounds
	BgSelection string // Selected row background
	BgContrast  string // Title bar background
	BgNegative  string // Warning/danger button background

	// UI foregrounds
	FgDefault  string // Default text
	FgSubtle   string // De-emphasized text, borders
	FgAccent   string // Primary accent
	FgDisabled string // Separators, scrollbar track
	FgContrast string // Text on contrast/negative backgrounds
	FgNegative string // Errors
	FgInfo     string // Git file count
	FgAdd      string // Git additions
	FgDelete   string // Git deletions

	// ANSI palette (agent status icons + accent ramp)
	PaletteYellow      string
	PaletteGreen       string
	PaletteRed         string
	PaletteBlue        string
	PaletteMagenta     string // Secondary accent (a30 where accent-driven)
	PaletteDarkMagenta string // Secondary accent (a40 where accent-driven)
}

var registry = map[string]Theme{}

// Register adds a theme to the registry. Keys not starting with
// "black-atom-" are ignored — this drops the unrendered entry that
// collection.template.go (itself valid Go) registers at init.
func Register(t Theme) {
	if !strings.HasPrefix(t.Key, "black-atom-") {
		return
	}
	registry[t.Key] = t
}

// Get returns the theme for key, if registered.
func Get(key string) (Theme, bool) {
	t, ok := registry[key]
	return t, ok
}

// Keys returns all registered theme keys, sorted.
func Keys() []string {
	keys := make([]string, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
