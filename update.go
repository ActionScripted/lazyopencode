package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		prev := m.cursor
		m.cursor = clamp(m.cursor-1, 0, len(m.filtered)-1)
		if m.cursor != prev {
			return m, m.loadMessagesForCursor()
		}

	case key.Matches(msg, m.keys.Down):
		prev := m.cursor
		m.cursor = clamp(m.cursor+1, 0, len(m.filtered)-1)
		if m.cursor != prev {
			return m, m.loadMessagesForCursor()
		}

	case key.Matches(msg, m.keys.Search):
		m.mode = ModeSearch
		m.search.Focus()
		return m, textinput.Blink

	case key.Matches(msg, m.keys.Tab):
		m.mode = ModeWorkspaces
		return m, nil

	case msg.String() == "d":
		if len(m.filtered) == 0 {
			return m, nil
		}
		m.pendingDeleteID = m.filtered[m.cursor].ID
		m.mode = ModeConfirmDelete
		return m, nil
	}

	return m, nil
}

func (m model) updateWorkspaces(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		m.workspaceCursor = clamp(m.workspaceCursor-1, 0, len(m.workspaces)-1)

	case key.Matches(msg, m.keys.Down):
		m.workspaceCursor = clamp(m.workspaceCursor+1, 0, len(m.workspaces)-1)

	case key.Matches(msg, m.keys.Tab):
		m.mode = ModeNormal
		return m, nil
	}

	return m, nil
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Back) {
		m.mode = ModeNormal
		m.search.Blur()
		return m, nil
	}

	prevID := ""
	if len(m.filtered) > 0 {
		prevID = m.filtered[m.cursor].ID
	}

	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	m.filtered = filterSessions(m.sessions, m.search.Value())
	m.cursor = clamp(m.cursor, 0, max(0, len(m.filtered)-1))

	// If the selected session changed due to filtering, reload messages.
	if len(m.filtered) > 0 && m.filtered[m.cursor].ID != prevID {
		return m, tea.Batch(cmd, m.loadMessagesForCursor())
	}

	return m, cmd
}

// loadMessagesForCursor fires a loadMessagesCmd for the currently selected
// session and clears the cached messages so the preview shows "loading…".
func (m *model) loadMessagesForCursor() tea.Cmd {
	if len(m.filtered) == 0 {
		return nil
	}
	id := m.filtered[m.cursor].ID
	m.messages = nil
	m.previewSessionID = id
	return m.loadMessagesCmd(id)
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

// updateConfirmDelete handles key events while ModeConfirmDelete is active.
// Y or d confirms the delete; n, esc, or q cancels.
func (m model) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "d":
		id := m.pendingDeleteID
		m.pendingDeleteID = ""
		m.mode = ModeNormal

		if id == "" {
			return m, nil
		}

		// Optimistic removal.
		m.sessions = removeSessionByID(m.sessions, id)
		m.filtered = removeSessionByID(m.filtered, id)
		m.workspaces = buildWorkspaces(m.sessions)
		m.cursor = clamp(m.cursor, 0, max(0, len(m.filtered)-1))

		var cmd tea.Cmd
		if len(m.filtered) > 0 {
			cmd = tea.Batch(m.loadMessagesForCursor(), m.deleteSessionCmd(id))
		} else {
			m.messages = nil
			cmd = m.deleteSessionCmd(id)
		}
		return m, cmd

	case "n", "esc", "q", "ctrl+c":
		m.pendingDeleteID = ""
		m.mode = ModeNormal
		return m, nil
	}

	return m, nil
}
