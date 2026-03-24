package main

import (
	"os"
	"path/filepath"
	"sort"
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

// homeToTilde replaces the home directory prefix with "~" in the given path.
// Called once at load time to populate Session.DisplayDir.
func homeToTilde(dir string) string {
	home, _ := os.UserHomeDir()
	return strings.Replace(dir, home, "~", 1)
}

// baseName returns just the last path component of dir, or "~" if dir is the
// user's home directory. Called once at load time to populate Session.ShortDir.
func baseName(dir string) string {
	home, _ := os.UserHomeDir()
	if dir == home {
		return "~"
	}
	return filepath.Base(dir)
}

// filterSessions returns the subset of sessions matching query
// against each session's FilterValue (title + directory).
func filterSessions(sessions []Session, query string) []Session {
	if query == "" {
		return sessions
	}
	q := strings.ToLower(query)
	out := make([]Session, 0, len(sessions))
	for _, s := range sessions {
		if strings.Contains(strings.ToLower(s.FilterValue()), q) {
			out = append(out, s)
		}
	}
	return out
}

// buildWorkspaces returns a sorted, deduplicated list of workspace values from
// the given sessions. DisplayDir is pre-computed once here to avoid calling
// os.UserHomeDir on every render frame.
func buildWorkspaces(sessions []Session) []workspace {
	seen := make(map[string]struct{}, len(sessions))
	for _, s := range sessions {
		seen[s.Directory] = struct{}{}
	}
	ws := make([]workspace, 0, len(seen))
	for dir := range seen {
		ws = append(ws, workspace{Dir: dir, DisplayDir: homeToTilde(dir)})
	}
	sort.Slice(ws, func(i, j int) bool { return ws[i].Dir < ws[j].Dir })
	return ws
}

// removeSessionByID returns a new slice with the session matching id removed.
func removeSessionByID(sessions []Session, id string) []Session {
	out := make([]Session, 0, len(sessions))
	for _, s := range sessions {
		if s.ID != id {
			out = append(out, s)
		}
	}
	return out
}
