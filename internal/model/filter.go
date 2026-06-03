package model

import (
	"strings"

	"github.com/black-atom-industries/helm/internal/lib/fuzzy"
)

// Filter provides fuzzy filtering over a list of items.
// It encapsulates filter state and matching logic, making it testable
// independently of the main Model.
type Filter[T any] struct {
	items   []T
	filter  string
	matchFn func(T, string) bool
	results []T
}

// NewFilter creates a new Filter with the given items and match function.
// The matchFn receives an item and the lowercase filter string, returning
// true if the item matches.
func NewFilter[T any](items []T, matchFn func(T, string) bool) *Filter[T] {
	return &Filter[T]{
		items:   items,
		matchFn: matchFn,
		results: items, // initially all items match (empty filter)
	}
}

// SetItems updates the underlying item list and recomputes results.
func (f *Filter[T]) SetItems(items []T) {
	f.items = items
	f.recompute()
}

// SetFilter updates the filter text and recomputes results.
func (f *Filter[T]) SetFilter(filter string) {
	f.filter = filter
	f.recompute()
}

// Filter returns the current filter string.
func (f *Filter[T]) Filter() string {
	return f.filter
}

// Results returns the filtered items.
func (f *Filter[T]) Results() []T {
	return f.results
}

// Count returns the number of matching items.
func (f *Filter[T]) Count() int {
	return len(f.results)
}

// recompute applies the filter to all items.
func (f *Filter[T]) recompute() {
	if f.filter == "" {
		f.results = f.items
		return
	}
	filterLower := strings.ToLower(f.filter)
	var results []T
	for _, item := range f.items {
		if f.matchFn(item, filterLower) {
			results = append(results, item)
		}
	}
	f.results = results
}

// --- Match functions for common types ---

// MatchSessionName matches filter against session name using fuzzy matching.
func MatchSessionName(name string, filter string) bool {
	return fuzzy.Match(name, filter)
}

// MatchPath matches filter against a path using segment-aware fuzzy matching.
func MatchPath(path string, filter string) bool {
	return fuzzy.MatchPath(path, filter)
}
