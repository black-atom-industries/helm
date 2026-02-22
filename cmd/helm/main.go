package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/model"
	"github.com/black-atom-industries/helm/internal/tmux"
	"github.com/black-atom-industries/helm/internal/ui"
)

func main() {
	// Ensure HOME is set (required for config paths)
	if os.Getenv("HOME") == "" {
		fmt.Println("Error: HOME environment variable not set")
		os.Exit(1)
	}

	// Handle subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			if err := config.Init(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created config file at %s\n", config.Path())
			return
		case "bookmark":
			if len(os.Args) < 3 {
				fmt.Println("Usage: helm bookmark <N>")
				fmt.Println("Opens bookmark at slot N (0-9)")
				os.Exit(1)
			}
			if err := runBookmark(os.Args[2]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "tmux-bindings":
			if err := printTmuxBindings(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "setup":
			if err := runSetup(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "repos":
			if err := runRepos(os.Args[2:]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			fmt.Println("Usage: helm [init | setup | repos | bookmark <N> | tmux-bindings]")
			os.Exit(1)
		}
	}

	// Check if running inside tmux
	if os.Getenv("TMUX") == "" {
		fmt.Println("Error: helm must be run from within tmux")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize colors based on appearance config
	ui.InitColors(string(cfg.Appearance))

	// Get current session to exclude from list
	currentSession, err := tmux.CurrentSession()
	if err != nil {
		fmt.Printf("Error getting current session: %v\n", err)
		os.Exit(1)
	}

	// Initialize and run the TUI
	m := model.New(currentSession, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// runBookmark opens the bookmark at slot N (0-9)
func runBookmark(slotStr string) error {
	slot, err := strconv.Atoi(slotStr)
	if err != nil || slot < 0 || slot > 9 {
		return fmt.Errorf("invalid slot: %s (must be 0-9)", slotStr)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if slot >= len(cfg.Bookmarks) {
		return fmt.Errorf("no bookmark at slot %d", slot)
	}

	bookmark := cfg.Bookmarks[slot]
	sessionName := extractSessionName(bookmark.Path, cfg.ProjectDepth)

	// Create session if it doesn't exist
	if !tmux.SessionExists(sessionName) {
		if err := tmux.CreateSession(sessionName, bookmark.Path); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Apply layout if configured
		if cfg.Layout != "" && cfg.LayoutDir != "" {
			layoutPath := filepath.Join(cfg.LayoutDir, cfg.Layout+".sh")
			if _, err := os.Stat(layoutPath); err == nil {
				cmd := exec.Command(layoutPath, sessionName, bookmark.Path)
				cmd.Env = append(os.Environ(),
					"TMUX_SESSION="+sessionName,
					"TMUX_WORKING_DIR="+bookmark.Path,
				)
				_ = cmd.Run()
			}
		}
	}

	// Switch to the session
	return tmux.SwitchClient(sessionName)
}

// printTmuxBindings outputs tmux bind commands for configured bookmarks
// Uses Alt+Shift+number keybindings (M-) through M-()
func printTmuxBindings() error {
	// Shifted number keys: 0=) 1=! 2=@ 3=# 4=$ 5=% 6=^ 7=& 8=* 9=(
	shiftedKeys := []string{")", "!", "@", "#", "$", "%", "^", "&", "*", "("}

	fmt.Println("# helm bookmark bindings (Alt+Shift+0-9)")
	fmt.Println("# Add to your tmux.conf or source with: run-shell \"helm tmux-bindings | tmux source-stdin\"")

	// Always output all 10 slots
	for i := 0; i < 10; i++ {
		fmt.Printf("bind -n M-%s run-shell \"helm bookmark %d\"\n", shiftedKeys[i], i)
	}

	return nil
}

// extractSessionName extracts a session name from a full path
// Uses the last N path components based on depth and sanitizes for tmux
func extractSessionName(fullPath string, depth int) string {
	parts := strings.Split(fullPath, string(filepath.Separator))
	if depth > len(parts) {
		depth = len(parts)
	}
	relPath := strings.Join(parts[len(parts)-depth:], "/")
	return sanitizeSessionName(relPath)
}

// sanitizeSessionName converts a path to a valid tmux session name
func sanitizeSessionName(name string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		".", "-",
		":", "-",
		" ", "-",
	)
	return replacer.Replace(name)
}
