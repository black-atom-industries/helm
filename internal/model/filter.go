package model

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/black-atom-industries/helm/internal/lib/filter"
	"github.com/black-atom-industries/helm/internal/lib/fuzzy"
	"github.com/black-atom-industries/helm/internal/tmux"
)

// Filter is an alias for the shared filter.Filter type.
// Re-exported for convenience within the model package.
type Filter[T any] = filter.Filter[T]

// NewFilter creates a new Filter using the shared filter package.
func NewFilter[T any](items []T, matchFn func(T, string) bool) *Filter[T] {
	return filter.New(items, matchFn)
}

// SessionFilter returns the session filter (for external access).
func (m *Model) SessionFilter() *filter.Filter[tmux.Session] {
	return m.sessionFilter
}

// SetFilter updates the session filter and rebuilds items.
func (m *Model) SetFilter(f string) {
	m.sessionFilter.SetFilter(f)
	m.rebuildItems()
}

// Filter returns the current session filter string.
func (m *Model) Filter() string {
	return m.sessionFilter.Filter()
}

// HandleFilterKey processes filter-related key events (space, runes, backspace).
// Returns true if the key was handled, false if the caller should handle it.
func (m *Model) HandleFilterKey(msg tea.KeyMsg) bool {
	f := m.sessionFilter.Filter()
	switch msg.Type {
	case tea.KeySpace:
		m.SetFilter(f + " ")
		return true
	case tea.KeyRunes:
		m.SetFilter(f + string(msg.Runes))
		return true
	case tea.KeyBackspace:
		if len(f) > 0 {
			m.SetFilter(f[:len(f)-1])
		}
		return true
	}
	return false
}

// matchesFilter checks if a name matches the current filter.
// Returns true if filter is empty or name matches via fuzzy matching.
func (m *Model) matchesFilter(name string) bool {
	if m.sessionFilter.Filter() == "" {
		return true
	}
	return fuzzy.Match(name, strings.ToLower(m.sessionFilter.Filter()))
}
