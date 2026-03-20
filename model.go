package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
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

// shellExitedMsg signals that a shell subprocess launched via g has exited.
type shellExitedMsg struct{}

// yankDoneMsg signals that the yank command completed (success or silent failure).
type yankDoneMsg struct{}

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
	demoMode               bool
	err                    error
	messages               []Message     // messages for currently selected session; nil = loading
	stats                  *SessionStats // stats for currently selected session; nil = loading
	previewSessionID       string        // session ID whose messages are loaded
	workspaces             []workspace   // sorted unique workspace directories
	workspaceCursor        int           // cursor into workspaces slice
	pendingDeleteID        string        // session ID awaiting delete confirmation
	pendingDeleteWorkspace string        // workspace directory awaiting delete confirmation
}

func newModel(dbPath string, demoMode bool) model {
	ti := textinput.New()
	ti.Placeholder = "search sessions…"
	ti.Prompt = ""

	return model{
		mode:     ModeNormal,
		keys:     DefaultKeyMap(),
		dbPath:   dbPath,
		demoMode: demoMode,
		search:   ti,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.loadSessionsCmd())
}

func (m model) loadSessionsCmd() tea.Cmd {
	if m.demoMode {
		return func() tea.Msg { return sessionsLoadedMsg{sessions: demoSessions()} }
	}
	return func() tea.Msg {
		sessions, err := loadSessions(m.dbPath)
		if err != nil {
			return errMsg{err: err}
		}
		return sessionsLoadedMsg{sessions: sessions}
	}
}

func (m model) loadMessagesCmd(sessionID string) tea.Cmd {
	if m.demoMode {
		return func() tea.Msg {
			return messagesLoadedMsg{sessionID: sessionID, messages: demoMessages(sessionID)}
		}
	}
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
	if m.demoMode {
		return func() tea.Msg {
			return statsLoadedMsg{sessionID: sessionID, stats: demoStats(sessionID)}
		}
	}
	return func() tea.Msg {
		stats, err := loadStats(m.dbPath, sessionID)
		if err != nil {
			// Non-fatal: show zero stats rather than an error.
			return statsLoadedMsg{sessionID: sessionID, stats: SessionStats{}}
		}
		return statsLoadedMsg{sessionID: sessionID, stats: stats}
	}
}

// openSessionCmd suspends lazyopencode, hands the terminal to opencode running in
// the session's directory, then resumes and reloads sessions when it exits.
func (m model) openSessionCmd(id, dir string) tea.Cmd {
	if m.demoMode {
		return nil
	}
	c := exec.Command("opencode", "--session", id)
	c.Dir = dir
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return sessionOpenedMsg{}
	})
}

// openShellCmd prints a suspend notice then hands the terminal to $SHELL
// in the given directory. lazyopencode resumes automatically when the shell exits.
func (m model) openShellCmd(dir string) tea.Cmd {
	if m.demoMode {
		return nil
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	c := exec.Command(shell)
	c.Dir = dir
	printCmd := func() tea.Msg {
		_, _ = fmt.Fprintf(os.Stdout, "\nopening shell in %s — type 'exit' to return to lazyopencode\n\n", displayDir(dir))
		return nil
	}
	return tea.Sequence(printCmd, tea.ExecProcess(c, func(err error) tea.Msg {
		return shellExitedMsg{}
	}))
}

// yankCmd copies text to the system clipboard using pbcopy (macOS) or
// xclip (Linux). Fails silently — a missing clipboard tool is not fatal.
func (m model) yankCmd(text string) tea.Cmd {
	return func() tea.Msg {
		var c *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			c = exec.Command("pbcopy")
		default:
			c = exec.Command("xclip", "-selection", "clipboard")
		}
		c.Stdin = strings.NewReader(text)
		_ = c.Run() // silent on failure
		return yankDoneMsg{}
	}
}

func (m model) deleteSessionCmd(sessionID string) tea.Cmd {
	if m.demoMode {
		return nil
	}
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
	if m.demoMode {
		return nil
	}
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
		// Preserve the selected workspace across reloads by remembering the raw
		// directory before rebuilding, then re-finding it in the new slice.
		prevWS := ""
		if m.workspaceCursor < len(m.workspaces) {
			prevWS = m.workspaces[m.workspaceCursor].Dir
		}
		m.sessions = msg.sessions
		m.filtered = filterSessions(m.sessions, m.search.Value())
		m.workspaces = buildWorkspaces(m.sessions)
		m.cursor = 0
		m.workspaceCursor = 0
		for i, ws := range m.workspaces {
			if ws.Dir == prevWS {
				m.workspaceCursor = i
				break
			}
		}
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

	case shellExitedMsg:
		// Shell exited — nothing to reload, just resume.
		return m, nil

	case yankDoneMsg:
		// Nothing to do — clipboard write is fire-and-forget.
		return m, nil

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
		case ModeYank:
			return m.updateYank(msg)
		case ModeGotoMenu:
			return m.updateGotoMenu(msg)
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
		ws = append(ws, workspace{Dir: dir, DisplayDir: displayDir(dir)})
	}
	sort.Slice(ws, func(i, j int) bool { return ws[i].Dir < ws[j].Dir })
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
