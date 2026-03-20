package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "lazyopencode: could not resolve home directory: %v\n", err)
		os.Exit(1)
	}
	dbPath := filepath.Join(home, ".local", "share", "opencode", "opencode.db")

	p := tea.NewProgram(
		newModel(dbPath),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "lazyopencode: %v\n", err)
		os.Exit(1)
	}
}
