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
	OpenRemote    key.Binding
	DownloadRepo  key.Binding
	Lazygit       key.Binding
	Bookmarks     key.Binding
	AddBookmark   key.Binding
	Quit          key.Binding
	Help          key.Binding
	Cancel        key.Binding
	Confirm       key.Binding
	Jump0         key.Binding
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
	OpenRemote: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("C-r", "Remote"),
	),
	DownloadRepo: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("C-d", "Download"),
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
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "Help"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "Cancel"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("C-y", "Confirm"),
	),
	Jump0: key.NewBinding(key.WithKeys("0")),
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
