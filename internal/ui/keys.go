package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application
type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	Expand        key.Binding
	Collapse      key.Binding
	Select        key.Binding
	Kill          key.Binding
	Create        key.Binding
	PickDirectory key.Binding
	CloneRepo     key.Binding
	Lazygit       key.Binding
	Bookmarks     key.Binding
	AddBookmark   key.Binding
	Quit          key.Binding
	Cancel        key.Binding
	Confirm       key.Binding
	Jump1         key.Binding
	Jump2         key.Binding
	Jump3         key.Binding
	Jump4         key.Binding
	Jump5         key.Binding
	Jump6         key.Binding
	Jump7         key.Binding
	Jump8         key.Binding
	Jump9         key.Binding
}

// DefaultKeyMap returns the default key bindings
// Navigation uses Ctrl+key or arrows, letters are reserved for filtering
var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("ctrl+k", "up"),
		key.WithHelp("↑", "Up"),
	),
	Down: key.NewBinding(
		key.WithKeys("ctrl+j", "down"),
		key.WithHelp("↓", "Down"),
	),
	Expand: key.NewBinding(
		key.WithKeys("ctrl+l", "right"),
		key.WithHelp("→", "Expand"),
	),
	Collapse: key.NewBinding(
		key.WithKeys("ctrl+h", "left"),
		key.WithHelp("←", "Collapse"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "Switch"),
	),
	Kill: key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("C-x", "Kill"),
	),
	Create: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("C-n", "New"),
	),
	PickDirectory: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("C-p", "Projects"),
	),
	CloneRepo: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("C-r", "Clone repo"),
	),
	Lazygit: key.NewBinding(
		key.WithKeys("ctrl+g"),
		key.WithHelp("C-g", "Lazygit"),
	),
	Bookmarks: key.NewBinding(
		key.WithKeys("ctrl+b"),
		key.WithHelp("C-b", "Bookmarks"),
	),
	AddBookmark: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("C-a", "Add bookmark"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("C-c", "Quit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "Cancel"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("C-y", "Confirm"),
	),
	Jump1: key.NewBinding(key.WithKeys("1")),
	Jump2: key.NewBinding(key.WithKeys("2")),
	Jump3: key.NewBinding(key.WithKeys("3")),
	Jump4: key.NewBinding(key.WithKeys("4")),
	Jump5: key.NewBinding(key.WithKeys("5")),
	Jump6: key.NewBinding(key.WithKeys("6")),
	Jump7: key.NewBinding(key.WithKeys("7")),
	Jump8: key.NewBinding(key.WithKeys("8")),
	Jump9: key.NewBinding(key.WithKeys("9")),
}

// helpItem formats a single help item (key + description)
func helpItem(key, desc string) string {
	return HelpKeyStyle.Render(key) + " " + HelpDescStyle.Render(desc)
}

// helpSep returns the separator between help items
func helpSep() string {
	return HelpSepStyle.Render(" · ")
}

// HelpNormal returns the help text for normal mode (two lines)
func HelpNormal() string {
	line1 := helpItem("Type", "filter") + helpSep() +
		helpItem("C-j/k | ↑↓", "Nav") + helpSep() +
		helpItem("C-h/l | ←→", "Expand") + helpSep() +
		helpItem("C-x", "Kill")
	line2 := helpItem("C-n", "New") + helpSep() +
		helpItem("C-p", "Projects") + helpSep() +
		helpItem("C-b", "Bookmarks") + helpSep() +
		helpItem("C-a", "Bookmark") + helpSep() +
		helpItem("C-r", "Clone") + helpSep() +
		helpItem("C-g", "Lazygit")
	return line1 + "\n" + line2
}

// HelpFiltering returns the help text when filter is active
func HelpFiltering() string {
	return helpItem("Esc", "Clear") + helpSep() +
		helpItem("Enter", "Select") + helpSep() +
		helpItem("C-c", "Quit")
}

// HelpConfirmKill returns the help text for kill confirmation mode
func HelpConfirmKill() string {
	return helpItem("C-x", "Confirm") + helpSep() +
		helpItem("Esc", "Cancel")
}

// HelpCreate returns the help text for create mode
func HelpCreate() string {
	return helpItem("Enter", "Create") + helpSep() +
		helpItem("Esc", "Cancel")
}

// HelpPickDirectory returns the help text for directory picker mode
func HelpPickDirectory() string {
	return helpItem("C-j/k | ↑↓", "Nav") + helpSep() +
		helpItem("Enter", "Select") + helpSep() +
		helpItem("C-a", "Bookmark") + helpSep() +
		helpItem("C-x", "Remove") + helpSep() +
		helpItem("Esc", "Back")
}

// HelpAddBookmark returns the help text when adding a bookmark from project picker
func HelpAddBookmark() string {
	return helpItem("C-j/k | ↑↓", "Nav") + helpSep() +
		helpItem("C-a", "Add bookmark") + helpSep() +
		helpItem("Esc", "Back")
}

// HelpConfirmRemoveFolder returns the help text for folder removal confirmation
func HelpConfirmRemoveFolder() string {
	return helpItem("C-x", "Confirm") + helpSep() +
		helpItem("Esc", "Cancel")
}

// HelpCloneRepo returns the help text for clone repo mode
func HelpCloneRepo() string {
	return helpItem("C-j/k | ↑↓", "Nav") + helpSep() +
		helpItem("Enter", "Clone") + helpSep() +
		helpItem("Esc", "Back/Cancel")
}

// HelpCloneRepoLoading returns the help text while loading repos
func HelpCloneRepoLoading() string {
	return helpItem("Esc", "Cancel")
}

// HelpCloneSuccess returns the help text after successful clone
func HelpCloneSuccess() string {
	return helpItem("Enter", "Switch to session") + helpSep() +
		helpItem("Esc", "Back to sessions")
}

// HelpBookmarks returns the help text for bookmarks mode
func HelpBookmarks() string {
	return helpItem("C-j/k | ↑↓", "Nav") + helpSep() +
		helpItem("Enter", "Open") + helpSep() +
		helpItem("C-p/n", "Move") + helpSep() +
		helpItem("C-a", "Add") + helpSep() +
		helpItem("C-x", "Remove") + helpSep() +
		helpItem("Esc", "Back")
}
