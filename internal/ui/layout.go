package ui

// Layout height constants for UI sections.
// These constants define the fixed overhead for each part of the UI,
// making it easy to calculate available space for content.
const (
	// Header section box (title bar inside bordered box)
	// Top border(1) + gap(1) + title(1) + gap(1) + bottom border(1)
	HeaderBoxHeight = 5

	// Prompt line (between header and sessions box)
	PromptHeight = 1

	// Footer: notification line + hints
	NotificationHeight = 1
	HintsBoxHeight     = 5 // top border + gap + content + gap + bottom border (new layout)

	// FooterOverhead — legacy value (5) used by existing view functions.
	// Old: border(1) + notification(1) + state(1) + hints(2) = 5
	// Will change when views are migrated to section box layout.
	FooterOverhead = 5

	// Session list section box overhead (vertical)
	// Top border(1) + gap(1) + ... content ... + gap(1) + bottom border(1)
	SessionBoxOverhead = 4

	// Optional content elements inside session box
	TableHeaderHeight     = 1
	TableDottedLineHeight = 1

	// Pinned self session (row + separator)
	SelfSessionOverhead = 2

	// Sidebar
	SidebarGap = 1 // Gap between session list and sidebar

	// Computed totals for visible items calculation
	// Legacy values — used by existing views until migrated
	BaseOverhead            = HeaderOverhead + FooterOverhead                          // 3 + 5 = 8
	WithTableHeaderOverhead = BaseOverhead + TableHeaderHeight + TableDottedLineHeight // 8 + 1 + 1 = 10

	// Legacy constants — used by view functions until they are migrated to section box layout.
	// These will be removed in Phase 5 cleanup.
	HeaderOverhead     = 3 // TitleBar(1) + Prompt(1) + TopBorder(1)
	TopBorderHeight    = 1
	TitleBarHeight     = 1
	BottomBorderHeight = 1
	StateLineHeight    = 1
	HintsHeight        = 2

	// Fallback values
	DefaultVisibleItems = 10
)

// SidebarTotalWidth returns the total width consumed by the sidebar + gap
func SidebarTotalWidth() int {
	return SidebarBoxWidth() + SidebarGap
}
