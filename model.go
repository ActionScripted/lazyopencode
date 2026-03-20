package main

import (
	"os/exec"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// sessionsLoadedMsg carries sessions fetched from the DB.
type sessionsLoadedMsg struct {
	sessions []Session
}

// messagesLoadedMsg carries messages for a specific session.
type messagesLoadedMsg struct {
	sessionID string
	messages  []Message
}

// statsLoadedMsg carries aggregated stats for a specific session.
type statsLoadedMsg struct {
	sessionID string
	stats     SessionStats
}

// sessionDeletedMsg signals that a session was successfully deleted.
type sessionDeletedMsg struct{}

// sessionOpenedMsg signals that an opencode session subprocess has exited.
type sessionOpenedMsg struct{}

// errMsg carries a non-fatal error to display in the UI.
type errMsg struct {
	err error
}

type model struct {
	mode                   Mode
	keys                   KeyMap
	sessions               []Session
	filtered               []Session
	cursor                 int
	search                 textinput.Model
	width                  int
	height                 int
	dbPath                 string
	err                    error
	messages               []Message     // messages for currently selected session; nil = loading
	stats                  *SessionStats // stats for currently selected session; nil = loading
	previewSessionID       string        // session ID whose messages are loaded
	workspaces             []string      // sorted unique workspace directories
	workspaceCursor        int           // cursor into workspaces slice
	pendingDeleteID        string        // session ID awaiting delete confirmation
	pendingDeleteWorkspace string        // workspace directory awaiting delete confirmation
}

func newModel(dbPath string) model {
	ti := textinput.New()
	ti.Placeholder = "search sessions…"
	ti.Prompt = ""

	return model{
		mode:   ModeNormal,
		keys:   DefaultKeyMap(),
		dbPath: dbPath,
		search: ti,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.loadSessionsCmd())
}

func (m model) loadSessionsCmd() tea.Cmd {
	return func() tea.Msg {
		sessions, err := loadSessions(m.dbPath)
		if err != nil {
			return errMsg{err: err}
		}
		return sessionsLoadedMsg{sessions: sessions}
	}
}

func (m model) loadMessagesCmd(sessionID string) tea.Cmd {
	return func() tea.Msg {
		messages, err := loadMessages(m.dbPath, sessionID)
		if err != nil {
			// Non-fatal: just return empty rather than crashing the preview.
			return messagesLoadedMsg{sessionID: sessionID, messages: []Message{}}
		}
		return messagesLoadedMsg{sessionID: sessionID, messages: messages}
	}
}

func (m model) loadStatsCmd(sessionID string) tea.Cmd {
	return func() tea.Msg {
		stats, err := loadStats(m.dbPath, sessionID)
		if err != nil {
			// Non-fatal: show zero stats rather than an error.
			return statsLoadedMsg{sessionID: sessionID, stats: SessionStats{}}
		}
		return statsLoadedMsg{sessionID: sessionID, stats: stats}
	}
}

// openSessionCmd suspends lazyoc, hands the terminal to opencode running in
// the session's directory, then resumes and reloads sessions when it exits.
func (m model) openSessionCmd(id, dir string) tea.Cmd {
	c := exec.Command("opencode", "--session", id)
	c.Dir = dir
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return sessionOpenedMsg{}
	})
}

func (m model) deleteSessionCmd(sessionID string) tea.Cmd {
	return func() tea.Msg {
		if err := exec.Command("opencode", "session", "delete", sessionID).Run(); err != nil {
			return errMsg{err: err}
		}
		return sessionDeletedMsg{}
	}
}

// deleteSessionsCmd deletes multiple sessions sequentially in a single
// goroutine and returns one sessionDeletedMsg when all are done, avoiding
// an N-reload storm on workspace delete.
func (m model) deleteSessionsCmd(ids []string) tea.Cmd {
	return func() tea.Msg {
		for _, id := range ids {
			if err := exec.Command("opencode", "session", "delete", id).Run(); err != nil {
				return errMsg{err: err}
			}
		}
		return sessionDeletedMsg{}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case sessionsLoadedMsg:
		m.sessions = msg.sessions
		m.filtered = filterSessions(m.sessions, m.search.Value())
		m.workspaces = buildWorkspaces(m.sessions)
		m.cursor = 0
		m.workspaceCursor = 0
		if len(m.filtered) > 0 {
			id := m.filtered[0].ID
			m.previewSessionID = id
			m.messages = nil
			m.stats = nil
			return m, tea.Batch(m.loadMessagesCmd(id), m.loadStatsCmd(id))
		}
		return m, nil

	case messagesLoadedMsg:
		// Discard if the user has already moved to a different session.
		if len(m.filtered) > 0 && msg.sessionID == m.filtered[m.cursor].ID {
			m.messages = msg.messages
			m.previewSessionID = msg.sessionID
		}
		return m, nil

	case statsLoadedMsg:
		// Discard if the user has already moved to a different session.
		if len(m.filtered) > 0 && msg.sessionID == m.filtered[m.cursor].ID {
			stats := msg.stats
			m.stats = &stats
		}
		return m, nil

	case sessionDeletedMsg:
		// Optimistic removal already applied at confirm time; nothing to reload.
		return m, nil

	case sessionOpenedMsg:
		// Reload sessions to pick up any changes made during the opencode session.
		return m, m.loadSessionsCmd()

	case errMsg:
		m.err = msg.err
		return m, m.loadSessionsCmd()

	case tea.KeyMsg:
		switch m.mode {
		case ModeNormal:
			return m.updateNormal(msg)
		case ModeSearch:
			return m.updateSearch(msg)
		case ModeWorkspaces:
			return m.updateWorkspaces(msg)
		case ModeConfirmDelete:
			return m.updateConfirmDelete(msg)
		case ModeConfirmDeleteWorkspace:
			return m.updateConfirmDeleteWorkspace(msg)
		}
	}
	return m, nil
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

// buildWorkspaces returns a sorted, deduplicated list of directories from
// the given sessions. Called once when sessions are loaded.
func buildWorkspaces(sessions []Session) []string {
	seen := make(map[string]struct{}, len(sessions))
	for _, s := range sessions {
		seen[s.Directory] = struct{}{}
	}
	ws := make([]string, 0, len(seen))
	for dir := range seen {
		ws = append(ws, dir)
	}
	sort.Strings(ws)
	return ws
}

// clamp constrains v to [lo, hi].
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
