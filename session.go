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
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DisplayDir string // Directory with home replaced by "~"; computed at load time
	ShortDir   string // Last path component (or "~"); computed at load time

	// Summary fields from the session row; zero for pure-chat sessions.
	SummaryFiles     int
	SummaryAdditions int
	SummaryDeletions int
}

// SessionStats holds per-session aggregates fetched asynchronously from the
// part table. Loaded in parallel with messages whenever the cursor moves.
type SessionStats struct {
	MsgCount      int
	OutputTokens  int // sum of tokens.output across all step-finish parts
	ContextTokens int // tokens.total from the last step-finish part (0 if none)
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

// workspace represents a unique project directory that owns one or more sessions.
// DisplayDir is pre-computed at load time to avoid repeated os.UserHomeDir calls.
type workspace struct {
	Dir        string // raw absolute path — used for session matching and DB queries
	DisplayDir string // "~"-substituted path — used for display only
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
