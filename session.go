package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Session represents a single opencode session.
type Session struct {
	ID         string
	Title      string
	Directory  string
	UpdatedAt  time.Time
	DisplayDir string // Directory with home replaced by "~"; computed at load time
	ShortDir   string // Last path component (or "~"); computed at load time
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

// DisplayDirectory returns the pre-computed display path (home replaced by "~").
func (s Session) DisplayDirectory() string {
	return s.DisplayDir
}

// ShortDirectory returns the pre-computed short path (last component, or "~").
func (s Session) ShortDirectory() string {
	return s.ShortDir
}

// displayDir replaces the home directory prefix with "~" in the given path.
// Called once at load time to populate Session.DisplayDir.
func displayDir(dir string) string {
	home, _ := os.UserHomeDir()
	return strings.Replace(dir, home, "~", 1)
}

// shortDir returns just the last path component of dir, or "~" if dir is the
// user's home directory. Called once at load time to populate Session.ShortDir.
func shortDir(dir string) string {
	home, _ := os.UserHomeDir()
	if dir == home {
		return "~"
	}
	return filepath.Base(dir)
}
