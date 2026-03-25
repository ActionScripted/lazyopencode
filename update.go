package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

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

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		prev := m.cursor
		m.cursor = clamp(m.cursor-1, 0, len(m.filtered)-1)
		if m.cursor != prev {
			m, cmd := m.loadMessagesForCursor()
			return m, cmd
		}

	case key.Matches(msg, m.keys.Down):
		prev := m.cursor
		m.cursor = clamp(m.cursor+1, 0, len(m.filtered)-1)
		if m.cursor != prev {
			m, cmd := m.loadMessagesForCursor()
			return m, cmd
		}

	case key.Matches(msg, m.keys.Back):
		m.search.SetValue("")
		m.filtered = filterSessions(m.sessions, "")
		m.cursor = clamp(m.cursor, 0, max(0, len(m.filtered)-1))
		m, cmd := m.loadMessagesForCursor()
		return m, cmd

	case key.Matches(msg, m.keys.Search):
		m.mode = ModeSearch
		m.search.Focus()
		return m, textinput.Blink

	case key.Matches(msg, m.keys.Workspaces):
		m.mode = ModeWorkspaces
		return m, nil

	case key.Matches(msg, m.keys.Open):
		if len(m.filtered) == 0 {
			return m, nil
		}
		s := m.filtered[m.cursor]
		return m, m.openSessionCmd(s.ID, s.Directory)

	case key.Matches(msg, m.keys.Yank):
		if len(m.filtered) == 0 {
			return m, nil
		}
		m.mode = ModeYank
		return m, nil

	case key.Matches(msg, m.keys.GotoPrefix):
		if len(m.filtered) == 0 {
			return m, nil
		}
		m.mode = ModeGoto
		return m, nil

	case key.Matches(msg, m.keys.Delete):
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
		return m, nil

	case key.Matches(msg, m.keys.Down):
		m.workspaceCursor = clamp(m.workspaceCursor+1, 0, len(m.workspaces)-1)
		return m, nil

	case key.Matches(msg, m.keys.Workspaces), key.Matches(msg, m.keys.Back):
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, m.keys.Delete):
		if len(m.workspaces) == 0 {
			return m, nil
		}
		m.pendingDeleteWorkspace = m.workspaces[m.workspaceCursor].Dir
		m.mode = ModeConfirmDeleteWorkspace
		return m, nil
	}

	return m, nil
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Open) {
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
		m, loadCmd := m.loadMessagesForCursor()
		return m, tea.Batch(cmd, loadCmd)
	}

	return m, cmd
}

// updateConfirmDeleteWorkspace handles key events while
// ModeConfirmDeleteWorkspace is active. y or d confirms deletion of all
// sessions in the pending workspace; n, esc, q, or ctrl+c cancels.
func (m model) updateConfirmDeleteWorkspace(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Delete), key.Matches(msg, m.keys.Confirm):
		ws := m.pendingDeleteWorkspace
		m.pendingDeleteWorkspace = ""
		m.mode = ModeWorkspaces

		if ws == "" {
			return m, nil
		}

		// Collect IDs belonging to the workspace.
		var ids []string
		for _, s := range m.sessions {
			if s.Directory == ws {
				ids = append(ids, s.ID)
			}
		}

		// Optimistic removal.
		for _, id := range ids {
			m.sessions = removeSessionByID(m.sessions, id)
			m.filtered = removeSessionByID(m.filtered, id)
		}
		m.workspaces = buildWorkspaces(m.sessions)
		m.workspaceCursor = clamp(m.workspaceCursor, 0, max(0, len(m.workspaces)-1))
		m.cursor = clamp(m.cursor, 0, max(0, len(m.filtered)-1))

		// Delete all sessions in one command to avoid an N-reload storm.
		return m, m.deleteSessionsCmd(ids)

	case key.Matches(msg, m.keys.Cancel):
		m.pendingDeleteWorkspace = ""
		m.mode = ModeWorkspaces
		return m, nil
	}

	return m, nil
}

func (m model) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Delete), key.Matches(msg, m.keys.Confirm):
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
			m, loadCmd := m.loadMessagesForCursor()
			return m, tea.Batch(loadCmd, m.deleteSessionCmd(id))
		}
		m.messages = nil
		cmd = m.deleteSessionCmd(id)
		return m, cmd

	case key.Matches(msg, m.keys.Cancel):
		m.pendingDeleteID = ""
		m.mode = ModeNormal
		return m, nil
	}

	return m, nil
}

// updateGoto handles key events while ModeGoto is active.
// s opens a shell in the session's directory; w jumps to the workspace view.
// esc/q/n cancels back to normal mode.
func (m model) updateGoto(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.filtered) == 0 {
		m.mode = ModeNormal
		return m, nil
	}
	s := m.filtered[m.cursor]

	switch {
	case key.Matches(msg, m.keys.GotoShell):
		m.mode = ModeNormal
		return m, m.openShellCmd(s.Directory)

	case key.Matches(msg, m.keys.GotoWorkspace):
		m.mode = ModeWorkspaces
		for i, ws := range m.workspaces {
			if ws.Dir == s.Directory {
				m.workspaceCursor = i
				break
			}
		}
		return m, nil

	case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Quit):
		m.mode = ModeNormal
		return m, nil
	}

	return m, nil
}

// updateError handles key events while ModeError is active.
// Only q and ctrl+c are accepted — the app is in a hard error state
// and requires a restart to recover.
func (m model) updateError(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}
	return m, nil
}

// d yanks the session's display directory; s yanks the session ID.
// esc/q cancels back to normal mode.
func (m model) updateYank(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.filtered) == 0 {
		m.mode = ModeNormal
		return m, nil
	}
	s := m.filtered[m.cursor]

	switch {
	case key.Matches(msg, m.keys.YankDirectory):
		m.mode = ModeNormal
		return m, m.yankCmd(s.DisplayDirectory())

	case key.Matches(msg, m.keys.YankSession):
		m.mode = ModeNormal
		return m, m.yankCmd(s.ID)

	case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Quit):
		m.mode = ModeNormal
		return m, nil
	}

	return m, nil
}
