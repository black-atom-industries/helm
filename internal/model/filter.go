package model

import (
	"strings"

	"github.com/black-atom-industries/helm/internal/lib/fuzzy"
	"github.com/black-atom-industries/helm/internal/lib/filter"
)

// Filter is an alias for the shared filter.Filter type.
// Re-exported for convenience within the model package.
type Filter[T any] = filter.Filter[T]

// NewFilter creates a new Filter using the shared filter package.
func NewFilter[T any](items []T, matchFn func(T, string) bool) *Filter[T] {
	return filter.New(items, matchFn)
}

// matchesFilter checks if a name matches the current filter.
// Returns true if filter is empty or name matches via fuzzy matching.
func (m *Model) matchesFilter(name string) bool {
	if m.filter == "" {
		return true
	}
	return fuzzy.Match(name, strings.ToLower(m.filter))
}
