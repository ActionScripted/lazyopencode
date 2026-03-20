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
