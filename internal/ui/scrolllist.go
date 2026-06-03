package ui

import "github.com/black-atom-industries/helm/internal/lib/filter"

// ScrollList is a generic scrollable list with cursor, filtering, and scroll offset management.
// It eliminates duplicate scroll/cursor logic across different list modes.
type ScrollList[T any] struct {
	items        []T
	filter       *filter.Filter[T]
	cursor       int
	scrollOffset int
	height       int // Visible height (number of items that fit)
}

// NewScrollList creates a new ScrollList with a filter function.
// The filterFn should return true if the item matches the given filter string.
func NewScrollList[T any](filterFn func(T, string) bool) *ScrollList[T] {
	return &ScrollList[T]{
		filter: filter.New[T](nil, filterFn),
		height: 10, // Default fallback
	}
}

// SetItems replaces all items and re-applies the current filter
func (s *ScrollList[T]) SetItems(items []T) {
	s.items = items
	s.filter.SetItems(items)
	s.clampCursor()
	s.updateScrollOffset()
}

// SetHeight sets the visible height (number of items that fit on screen)
func (s *ScrollList[T]) SetHeight(height int) {
	if height > 0 {
		s.height = height
	}
	s.updateScrollOffset()
}

// Height returns the current visible height
func (s *ScrollList[T]) Height() int {
	return s.height
}

// SetFilter sets the filter text and re-filters the items
func (s *ScrollList[T]) SetFilter(filter string) {
	s.filter.SetFilter(filter)
	s.clampCursor()
	s.updateScrollOffset()
}

// Filter returns the current filter string
func (s *ScrollList[T]) Filter() string {
	return s.filter.Filter()
}

// Items returns all items (unfiltered)
func (s *ScrollList[T]) Items() []T {
	return s.items
}

// Filtered returns the filtered items
func (s *ScrollList[T]) Filtered() []T {
	return s.filter.Results()
}

// Len returns the number of filtered items
func (s *ScrollList[T]) Len() int {
	return s.filter.Count()
}

// Cursor returns the current cursor position
func (s *ScrollList[T]) Cursor() int {
	return s.cursor
}

// SetCursor sets the cursor position and updates scroll offset
func (s *ScrollList[T]) SetCursor(pos int) {
	s.cursor = pos
	s.clampCursor()
	s.updateScrollOffset()
}

// MoveCursor moves the cursor by delta and updates scroll offset
func (s *ScrollList[T]) MoveCursor(delta int) {
	s.cursor += delta
	s.clampCursor()
	s.updateScrollOffset()
}

// clampCursor ensures cursor is within valid bounds
func (s *ScrollList[T]) clampCursor() {
	filtered := s.filter.Results()
	if s.cursor >= len(filtered) {
		s.cursor = len(filtered) - 1
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
}

// ScrollOffset returns the current scroll offset
func (s *ScrollList[T]) ScrollOffset() int {
	return s.scrollOffset
}

// updateScrollOffset adjusts scroll offset to keep cursor visible
func (s *ScrollList[T]) updateScrollOffset() {
	filtered := s.filter.Results()
	// If cursor is above visible area, scroll up
	if s.cursor < s.scrollOffset {
		s.scrollOffset = s.cursor
	}
	// If cursor is below visible area, scroll down
	if s.cursor >= s.scrollOffset+s.height {
		s.scrollOffset = s.cursor - s.height + 1
	}
	// Ensure scroll offset is not negative
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
	// Ensure scroll offset doesn't exceed total items
	if len(filtered) > 0 && s.scrollOffset >= len(filtered) {
		s.scrollOffset = len(filtered) - 1
	}
}

// SelectedItem returns the currently selected item, or false if none
func (s *ScrollList[T]) SelectedItem() (T, bool) {
	var zero T
	filtered := s.filter.Results()
	if s.cursor < 0 || s.cursor >= len(filtered) {
		return zero, false
	}
	return filtered[s.cursor], true
}

// VisibleItems returns the slice of items currently visible on screen
func (s *ScrollList[T]) VisibleItems() []T {
	filtered := s.filter.Results()
	if len(filtered) == 0 {
		return nil
	}

	start := s.scrollOffset
	end := start + s.height
	if end > len(filtered) {
		end = len(filtered)
	}
	if start >= end {
		return nil
	}
	return filtered[start:end]
}

// VisibleRange returns the start and end indices of visible items in the filtered slice
func (s *ScrollList[T]) VisibleRange() (start, end int) {
	start = s.scrollOffset
	end = start + s.height
	filtered := s.filter.Results()
	if end > len(filtered) {
		end = len(filtered)
	}
	return start, end
}

// IsSelected returns true if the given index (in filtered list) is the cursor position
func (s *ScrollList[T]) IsSelected(index int) bool {
	return index == s.cursor
}

// Reset clears the filter and resets cursor to 0
func (s *ScrollList[T]) Reset() {
	s.filter.SetFilter("")
	s.cursor = 0
	s.scrollOffset = 0
}

// Clear removes all items
func (s *ScrollList[T]) Clear() {
	s.items = nil
	s.filter.SetItems(nil)
	s.cursor = 0
	s.scrollOffset = 0
	s.filter.SetFilter("")
}
