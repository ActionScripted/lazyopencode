package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Session represents a single opencode session.
type Session struct {
	ID        string
	Title     string
	Directory string
	UpdatedAt time.Time
}

// Message represents a single chat message in a session.
type Message struct {
	Role string // "user" or "assistant"
	Text string
}

// FilterValue returns the string used for search matching.
// Implements the bubbles/list.Item interface for future compatibility.
func (s Session) FilterValue() string {
	return s.Title + " " + s.Directory
}

// DisplayDirectory replaces the home directory with ~ for display.
func (s Session) DisplayDirectory() string {
	home, _ := os.UserHomeDir()
	return strings.Replace(s.Directory, home, "~", 1)
}

// ShortDirectory returns just the last path component (project folder name).
// Returns "~" if the directory is the user's home directory.
func (s Session) ShortDirectory() string {
	home, _ := os.UserHomeDir()
	if s.Directory == home {
		return "~"
	}
	return filepath.Base(s.Directory)
}
