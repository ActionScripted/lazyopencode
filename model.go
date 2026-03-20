package main

import (
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

// errMsg carries a non-fatal error to display in the UI.
type errMsg struct {
	err error
}

type model struct {
	mode             Mode
	keys             KeyMap
	sessions         []Session
	filtered         []Session
	cursor           int
	search           textinput.Model
	width            int
	height           int
	dbPath           string
	err              error
	messages         []Message // messages for currently selected session; nil = loading
	previewSessionID string    // session ID whose messages are loaded
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case sessionsLoadedMsg:
		m.sessions = msg.sessions
		m.filtered = filterSessions(m.sessions, m.search.Value())
		m.cursor = 0
		if len(m.filtered) > 0 {
			id := m.filtered[0].ID
			m.previewSessionID = id
			m.messages = nil
			return m, m.loadMessagesCmd(id)
		}
		return m, nil

	case messagesLoadedMsg:
		// Discard if the user has already moved to a different session.
		if len(m.filtered) > 0 && msg.sessionID == m.filtered[m.cursor].ID {
			m.messages = msg.messages
			m.previewSessionID = msg.sessionID
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case ModeNormal:
			return m.updateNormal(msg)
		case ModeSearch:
			return m.updateSearch(msg)
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
