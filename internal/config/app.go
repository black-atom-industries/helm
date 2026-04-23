package config

import "path/filepath"

// App-level constants (not user-configurable)
const (
	AppName = "BLACK ATOM HELM"

	// Directory and file names
	AppDirName        = "helm"
	ConfigFileName    = "config.yml"
	BookmarksFileName = "bookmarks.yml"
	StatusFileExt     = ".status"
)

// ConfigDirName returns the relative path for config files under ~/.config/
func ConfigDirName() string {
	return filepath.Join("black-atom", "helm")
}
