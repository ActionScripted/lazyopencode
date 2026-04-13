package main

import (
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// session represents a single opencode session.
type session struct {
	ID         string
	Title      string
	Directory  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DisplayDir string // Directory with home replaced by "~"; computed at load time
	ShortDir   string // Last path component (or "~"); computed at load time
	FilterKey  string // lowercase Title + " " + Directory; computed at load time for fast search

	// Summary fields from the session row; zero for pure-chat sessions.
	SummaryFiles     int
	SummaryAdditions int
	SummaryDeletions int
}

// sessionStats holds per-session aggregates fetched asynchronously from the
// part table. Loaded in parallel with messages whenever the cursor moves.
type sessionStats struct {
	MsgCount      int
	InputTokens   int      // sum of tokens.input across all step-finish parts
	OutputTokens  int      // sum of tokens.output across all step-finish parts
	ContextTokens int      // tokens.total from the last step-finish part (0 if none)
	Models        []string // distinct model IDs used in the session, ordered by first use
	Cost          float64  // sum of $.cost from all assistant messages in the session
}

// modelStat holds aggregate usage for a single AI model.
type modelStat struct {
	Name         string
	Sessions     int
	Prompts      int // count of distinct user prompts responded to by this model
	InputTokens  int
	OutputTokens int
	DurationMS   int64   // sum of (time_updated - time_created) for sessions using this model
	Cost         float64 // sum of $.cost from opencode's message data
}

// projectStat holds aggregate stats for a project directory.
type projectStat struct {
	Dir          string
	DisplayDir   string
	Sessions     int
	Prompts      int // count of distinct user prompts across all sessions in this project
	InputTokens  int
	OutputTokens int
	DurationMS   int64
	Cost         float64     // sum of $.cost from opencode's message data
	Models       []modelStat // per-model breakdown within this project, ordered by session count desc
}

// globalStats holds aggregate stats across all sessions, used by the stats dashboard.
type globalStats struct {
	// All-time totals
	TotalSessions   int
	TotalPrompts    int // count of user messages across all sessions
	TotalInput      int
	TotalOutput     int
	TotalCacheRead  int
	TotalCacheWrite int
	TotalFiles      int
	TotalAdditions  int
	TotalDeletions  int
	TotalDurationMS int64   // sum of (time_updated - time_created) across all sessions
	TotalCost       float64 // sum of $.cost from opencode's message data, all time
	// Last-7-days totals
	RecentSessions   int
	RecentPrompts    int // count of user messages in the last 7 days
	RecentInput      int
	RecentOutput     int
	RecentCacheRead  int
	RecentCacheWrite int
	RecentFiles      int
	RecentAdditions  int
	RecentDeletions  int
	RecentDurationMS int64
	RecentCost       float64 // sum of $.cost from opencode's message data, last 7 days
	// Breakdowns
	Models   []modelStat   // all-time, ordered by session count desc
	Projects []projectStat // top 10, ordered by session count desc
}

// message represents a single chat message in a session.
type message struct {
	Role string // "user" or "assistant"
	Text string
}

// FilterValue returns the string used for search matching.
// Implements the bubbles/list.Item interface for future compatibility.
func (s session) FilterValue() string {
	return s.Title + " " + s.Directory
}

// workspace represents a unique project directory that owns one or more sessions.
// DisplayDir is pre-computed at load time to avoid repeated os.UserHomeDir calls.
type workspace struct {
	Dir        string // raw absolute path — used for session matching and DB queries
	DisplayDir string // "~"-substituted path — used for display only
}

// DisplayDirectory returns the pre-computed display path (home replaced by "~").
func (s session) DisplayDirectory() string {
	return s.DisplayDir
}

// ShortDirectory returns the pre-computed short path (last component, or "~").
func (s session) ShortDirectory() string {
	return s.ShortDir
}

// homeToTilde replaces the home directory prefix with "~" in the given path.
// home must be the result of os.UserHomeDir, resolved once by the caller.
// Called once at load time to populate session.DisplayDir.
func homeToTilde(dir, home string) string {
	if home == "" {
		return dir
	}
	trimmed := strings.TrimPrefix(dir, home)
	if trimmed == dir {
		return dir // home was not a prefix
	}
	return "~" + trimmed
}

// baseName returns just the last path component of dir, or "~" if dir equals
// home. home must be the result of os.UserHomeDir, resolved once by the caller.
// Called once at load time to populate session.ShortDir.
func baseName(dir, home string) string {
	if dir == home {
		return "~"
	}
	return filepath.Base(dir)
}

// filterSessions returns the subset of sessions matching query
// against each session's FilterKey (pre-lowercased title + directory).
// Always returns an independent slice — callers may safely append to or modify
// the result without aliasing m.sessions.
func filterSessions(sessions []session, query string) []session {
	if query == "" {
		return append([]session(nil), sessions...)
	}
	q := strings.ToLower(query)
	out := make([]session, 0, len(sessions))
	for _, s := range sessions {
		if strings.Contains(s.FilterKey, q) {
			out = append(out, s)
		}
	}
	return out
}

// buildWorkspaces returns a sorted, deduplicated list of workspace values from
// the given sessions. home must be the result of os.UserHomeDir, resolved once
// by the caller. DisplayDir is pre-computed once here to avoid calling
// os.UserHomeDir on every render frame.
func buildWorkspaces(sessions []session, home string) []workspace {
	seen := make(map[string]struct{}, len(sessions))
	for _, s := range sessions {
		seen[s.Directory] = struct{}{}
	}
	ws := make([]workspace, 0, len(seen))
	for dir := range seen {
		ws = append(ws, workspace{Dir: dir, DisplayDir: homeToTilde(dir, home)})
	}
	sort.Slice(ws, func(i, j int) bool { return ws[i].Dir < ws[j].Dir })
	return ws
}

// removeSessionByID returns a new slice with the session matching id removed.
func removeSessionByID(sessions []session, id string) []session {
	out := make([]session, 0, len(sessions))
	for _, s := range sessions {
		if s.ID != id {
			out = append(out, s)
		}
	}
	return out
}
