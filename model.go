package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// sessionsLoadedMsg carries sessions fetched from the DB.
type sessionsLoadedMsg struct {
	sessions []session
}

// messagesLoadedMsg carries messages for a specific session.
type messagesLoadedMsg struct {
	sessionID string
	messages  []message
}

// statsLoadedMsg carries aggregated stats for a specific session.
type statsLoadedMsg struct {
	sessionID string
	stats     sessionStats
}

// sessionDeletedMsg signals that a session was successfully deleted.
type sessionDeletedMsg struct{}

// sessionsDeleteErrMsg signals that one or more session deletions failed.
// The app reloads from the DB to re-sync state after a partial failure.
type sessionsDeleteErrMsg struct {
	err error
}

// sessionOpenedMsg signals that an opencode session subprocess has exited.
type sessionOpenedMsg struct{}

// shellExitedMsg signals that a shell subprocess launched via g has exited.
type shellExitedMsg struct{}

// yankDoneMsg signals that the yank command completed (success or silent failure).
type yankDoneMsg struct{}

// dbErrMsg carries a fatal database error that transitions the app to modeError.
type dbErrMsg struct {
	err error
}

// opErrMsg carries a transient operation error (e.g. delete failure) to display
// briefly in the hint bar without terminating the session.
type opErrMsg struct {
	err error
}

// globalStatsLoadedMsg carries aggregate stats for the stats dashboard.
type globalStatsLoadedMsg struct {
	stats globalStats
}

// resolveHome returns the current user's home directory, or "" on failure.
// Used wherever a home dir string is needed for display (homeToTilde,
// baseName, buildWorkspaces). Callers degrade gracefully to raw paths when
// home is unavailable — acceptable since this is display-only.
func resolveHome() string {
	home, _ := os.UserHomeDir() // display-only; raw path is acceptable fallback
	return home
}

type model struct {
	mode                   mode
	keys                   keyMap
	sessions               []session
	filtered               []session
	cursor                 int
	search                 textinput.Model
	width                  int
	height                 int
	dbPath                 string
	home                   string // user home directory, resolved once at startup
	demoMode               bool
	err                    error
	notice                 string        // transient operation error; cleared on next keypress
	messages               []message     // messages for currently selected session; nil = loading
	stats                  *sessionStats // stats for currently selected session; nil = loading
	previewSessionID       string        // session ID whose messages are loaded
	workspaces             []workspace   // sorted unique workspace directories
	workspaceCursor        int           // cursor into workspaces slice
	pendingDeleteID        string        // session ID awaiting delete confirmation
	pendingDeleteWorkspace string        // workspace directory awaiting delete confirmation
	globalStats            *globalStats  // cached aggregate stats; nil until first load
	statsScrollOffset      int           // scroll position (in lines) for the stats view
	pathColW               int           // cached path column width for renderList; recomputed whenever m.filtered changes
}

func newModel(dbPath string, demoMode bool) model {
	ti := textinput.New()
	ti.Placeholder = "search sessions…"
	ti.Prompt = ""

	return model{
		mode:     modeNormal,
		keys:     defaultKeyMap(),
		dbPath:   dbPath,
		home:     resolveHome(),
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
		sessions, err := loadSessions(m.dbPath, m.home)
		if err != nil {
			return dbErrMsg{err: err}
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
			return messagesLoadedMsg{sessionID: sessionID, messages: []message{}}
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
			return statsLoadedMsg{sessionID: sessionID, stats: sessionStats{}}
		}
		return statsLoadedMsg{sessionID: sessionID, stats: stats}
	}
}

func (m model) loadGlobalStatsCmd() tea.Cmd {
	if m.demoMode {
		return func() tea.Msg { return globalStatsLoadedMsg{stats: demoGlobalStats()} }
	}
	return func() tea.Msg {
		stats, err := loadGlobalStats(m.dbPath, m.home)
		if err != nil {
			return globalStatsLoadedMsg{} // zero value on error — dashboard shows zeros
		}
		return globalStatsLoadedMsg{stats: stats}
	}
}

// recomputePathColW recomputes m.pathColW from the current m.filtered slice.
// Must be called whenever m.filtered is reassigned. The cap is intentionally
// NOT applied here because it depends on the pane width, which is only known
// at render time; renderList reads m.pathColW and caps it there.
func (m model) recomputePathColW() model {
	w := 0
	for _, s := range m.filtered {
		if n := lipgloss.Width(s.ShortDirectory()); n > w {
			w = n
		}
	}
	m.pathColW = w
	return m
}

// loadMessagesForCursor returns an updated model with a batched command to
// reload messages and stats for the selected session. Stale content is kept
// visible until the new data arrives to prevent layout jumps and flash.
func (m model) loadMessagesForCursor() (model, tea.Cmd) {
	if len(m.filtered) == 0 {
		return m, nil
	}
	id := m.filtered[m.cursor].ID
	m.previewSessionID = id
	return m, tea.Batch(m.loadMessagesCmd(id), m.loadStatsCmd(id))
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
		if err != nil {
			return opErrMsg{err: err}
		}
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
	notice := fmt.Sprintf("\nopening shell in %s — type 'exit' to return to lazyopencode\n", homeToTilde(dir, m.home))
	return tea.Sequence(tea.Println(notice), tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return opErrMsg{err: err}
		}
		return shellExitedMsg{}
	}))
}

// yankCmd copies text to the system clipboard using pbcopy (macOS) or
// xclip (Linux). Non-fatal — a missing clipboard tool surfaces in the hint
// bar via opErrMsg rather than crashing the app.
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
		if err := c.Run(); err != nil {
			return opErrMsg{err: err}
		}
		return yankDoneMsg{}
	}
}

func (m model) deleteSessionCmd(sessionID string) tea.Cmd {
	if m.demoMode {
		return nil
	}
	return func() tea.Msg {
		if err := deleteOneSession(sessionID); err != nil {
			return opErrMsg{err: err}
		}
		return sessionDeletedMsg{}
	}
}

// deleteSessionsCmd deletes multiple sessions sequentially in a single
// goroutine. All deletions are attempted even if one fails — partial failures
// are joined and returned as sessionsDeleteErrMsg so the caller can reload
// from the DB to re-sync state. On full success it returns sessionDeletedMsg.
func (m model) deleteSessionsCmd(ids []string) tea.Cmd {
	if m.demoMode {
		return nil
	}
	return func() tea.Msg {
		var errs []error
		for _, id := range ids {
			if err := deleteOneSession(id); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return sessionsDeleteErrMsg{err: errors.Join(errs...)}
		}
		return sessionDeletedMsg{}
	}
}

// runCommand is the exec back-end used by deleteOneSession. It is a
// package-level variable so tests can swap it for a stub without spawning a
// real process.
var runCommand = func(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

// deleteOneSession shells out to `opencode session delete <id>`.
// Keeping deletion delegated to the owning process ensures lazyopencode
// remains read-only at the DB layer.
func deleteOneSession(id string) error {
	return runCommand("opencode", "session", "delete", id)
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
		m.workspaces = buildWorkspaces(m.sessions, m.home)
		m = m.recomputePathColW()
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

	case sessionsDeleteErrMsg:
		// One or more deletions failed. Surface the error and reload from the DB
		// so the in-memory state re-syncs with reality.
		m.notice = msg.err.Error()
		return m, m.loadSessionsCmd()

	case sessionOpenedMsg:
		// Reload sessions to pick up any changes made during the opencode session.
		return m, m.loadSessionsCmd()

	case shellExitedMsg:
		// Shell exited — nothing to reload, just resume.
		return m, nil

	case yankDoneMsg:
		// Nothing to do — clipboard write is fire-and-forget.
		return m, nil

	case dbErrMsg:
		m.err = msg.err
		m.mode = modeError
		return m, nil

	case opErrMsg:
		m.notice = msg.err.Error()
		return m, nil

	case globalStatsLoadedMsg:
		gs := msg.stats
		m.globalStats = &gs
		return m, nil

	case tea.KeyMsg:
		m.notice = "" // clear any transient notice on the next keypress
		switch m.mode {
		case modeNormal:
			return m.updateNormal(msg)
		case modeSearch:
			return m.updateSearch(msg)
		case modeWorkspaces:
			return m.updateWorkspaces(msg)
		case modeConfirmDelete:
			return m.updateConfirmDelete(msg)
		case modeConfirmDeleteWorkspace:
			return m.updateConfirmDeleteWorkspace(msg)
		case modeYank:
			return m.updateYank(msg)
		case modeGoto:
			return m.updateGoto(msg)
		case modeStats:
			return m.updateStats(msg)
		case modeError:
			return m.updateError(msg)
		}
	}
	return m, nil
}
