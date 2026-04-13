package main

import (
	"errors"
	"os"
	"slices"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ---------------------------------------------------------------------------
// deleteOneSession
// ---------------------------------------------------------------------------

func TestDeleteOneSession_Success(t *testing.T) {
	original := runCommand
	defer func() { runCommand = original }()

	var gotName string
	var gotArgs []string
	runCommand = func(name string, args ...string) error {
		gotName = name
		gotArgs = args
		return nil
	}

	if err := deleteOneSession("abc-123"); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if gotName != "opencode" {
		t.Errorf("command name: got %q, want %q", gotName, "opencode")
	}
	wantArgs := []string{"session", "delete", "abc-123"}
	if !slices.Equal(gotArgs, wantArgs) {
		t.Errorf("args: got %v, want %v", gotArgs, wantArgs)
	}
}

func TestDeleteOneSession_Failure(t *testing.T) {
	original := runCommand
	defer func() { runCommand = original }()

	sentinel := errors.New("opencode: session not found")
	runCommand = func(name string, args ...string) error {
		return sentinel
	}

	err := deleteOneSession("does-not-exist")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("error: got %v, want %v", err, sentinel)
	}
}

// ---------------------------------------------------------------------------
// openSessionCmd
// ---------------------------------------------------------------------------

func TestOpenSessionCmd_DemoMode(t *testing.T) {
	m := newModel("/tmp/fake.db", true /* demoMode */)
	cmd := m.openSessionCmd("sess-1", "/tmp/myproject")
	if cmd != nil {
		t.Error("expected nil cmd when demoMode=true")
	}
}

// ---------------------------------------------------------------------------
// loadSessionsCmd — invoke the returned closure directly
// ---------------------------------------------------------------------------

func TestLoadSessionsCmd_ReturnsSessionsLoadedMsg(t *testing.T) {
	path := createTestDB(t)

	db := openRW(t, path)
	insertSession(t, db, "s1", "My session", "/tmp/proj", 1000000, 2000000, nil)
	_ = db.Close()

	m := newModel(path, false)
	cmd := m.loadSessionsCmd()
	msg := cmd()

	loaded, ok := msg.(sessionsLoadedMsg)
	if !ok {
		t.Fatalf("expected sessionsLoadedMsg, got %T", msg)
	}
	if len(loaded.sessions) != 1 || loaded.sessions[0].ID != "s1" {
		t.Errorf("unexpected sessions: %v", loaded.sessions)
	}
}

func TestLoadSessionsCmd_MissingDBReturnsEmpty(t *testing.T) {
	m := newModel("/nonexistent/opencode.db", false)
	cmd := m.loadSessionsCmd()
	msg := cmd()

	loaded, ok := msg.(sessionsLoadedMsg)
	if !ok {
		t.Fatalf("expected sessionsLoadedMsg, got %T", msg)
	}
	if len(loaded.sessions) != 0 {
		t.Errorf("expected empty sessions for missing DB, got %d", len(loaded.sessions))
	}
}

// ---------------------------------------------------------------------------
// loadMessagesCmd — invoke the returned closure directly
// ---------------------------------------------------------------------------

func TestLoadMessagesCmd_ReturnsMessagesLoadedMsg(t *testing.T) {
	path := createTestDB(t)

	db := openRW(t, path)
	_, err := db.Exec(`INSERT INTO message (id, session_id, data) VALUES ('m1', 'sess-1', '{"role":"user"}')`)
	if err != nil {
		t.Fatalf("insert message: %v", err)
	}
	_, err = db.Exec(`INSERT INTO part (id, message_id, session_id, data, time_created) VALUES
		('p1', 'm1', 'sess-1', '{"type":"text","text":"hello"}', 1000)`)
	if err != nil {
		t.Fatalf("insert part: %v", err)
	}
	_ = db.Close()

	m := newModel(path, false)
	cmd := m.loadMessagesCmd("sess-1")
	msg := cmd()

	loaded, ok := msg.(messagesLoadedMsg)
	if !ok {
		t.Fatalf("expected messagesLoadedMsg, got %T", msg)
	}
	if loaded.sessionID != "sess-1" {
		t.Errorf("sessionID: got %q, want %q", loaded.sessionID, "sess-1")
	}
	if len(loaded.messages) != 1 || loaded.messages[0].Text != "hello" {
		t.Errorf("unexpected messages: %v", loaded.messages)
	}
}

func TestLoadMessagesCmd_ErrorReturnsEmpty(t *testing.T) {
	// A totally corrupt/non-SQLite file triggers the error path, which returns
	// an empty messagesLoadedMsg rather than crashing.
	path := t.TempDir() + "/bad.db"
	if err := createCorruptDB(t, path); err != nil {
		t.Fatalf("createCorruptDB: %v", err)
	}

	m := newModel(path, false)
	cmd := m.loadMessagesCmd("any-session")
	msg := cmd()

	loaded, ok := msg.(messagesLoadedMsg)
	if !ok {
		t.Fatalf("expected messagesLoadedMsg, got %T", msg)
	}
	if len(loaded.messages) != 0 {
		t.Errorf("expected empty messages on error, got %d", len(loaded.messages))
	}
}

// ---------------------------------------------------------------------------
// loadStatsCmd — invoke the returned closure directly
// ---------------------------------------------------------------------------

func TestLoadStatsCmd_ReturnsStatsLoadedMsg(t *testing.T) {
	path := createTestDB(t)

	db := openRW(t, path)
	_, err := db.Exec(`INSERT INTO message (id, session_id, data) VALUES ('m1', 'sess-1', '{"role":"user"}')`)
	if err != nil {
		t.Fatalf("insert message: %v", err)
	}
	_, err = db.Exec(`INSERT INTO part (id, message_id, session_id, data, time_created) VALUES
		('p1', 'm1', 'sess-1', '{"type":"step-finish","tokens":{"output":42,"total":500}}', 1000)`)
	if err != nil {
		t.Fatalf("insert part: %v", err)
	}
	_ = db.Close()

	m := newModel(path, false)
	cmd := m.loadStatsCmd("sess-1")
	msg := cmd()

	loaded, ok := msg.(statsLoadedMsg)
	if !ok {
		t.Fatalf("expected statsLoadedMsg, got %T", msg)
	}
	if loaded.sessionID != "sess-1" {
		t.Errorf("sessionID: got %q, want %q", loaded.sessionID, "sess-1")
	}
	if loaded.stats.OutputTokens != 42 {
		t.Errorf("OutputTokens: got %d, want 42", loaded.stats.OutputTokens)
	}
}

func TestLoadStatsCmd_ErrorReturnsZeroStats(t *testing.T) {
	path := t.TempDir() + "/bad.db"
	if err := createCorruptDB(t, path); err != nil {
		t.Fatalf("createCorruptDB: %v", err)
	}

	m := newModel(path, false)
	cmd := m.loadStatsCmd("any-session")
	msg := cmd()

	loaded, ok := msg.(statsLoadedMsg)
	if !ok {
		t.Fatalf("expected statsLoadedMsg, got %T", msg)
	}
	if loaded.stats.MsgCount != 0 || loaded.stats.InputTokens != 0 || loaded.stats.OutputTokens != 0 || loaded.stats.ContextTokens != 0 || len(loaded.stats.Models) != 0 {
		t.Errorf("expected zero stats on error, got %+v", loaded.stats)
	}
}

// ---------------------------------------------------------------------------
// deleteSessionCmd — invoke the returned closure directly
// ---------------------------------------------------------------------------

func TestDeleteSessionCmd_DemoModeReturnsNil(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	if cmd := m.deleteSessionCmd("sess-1"); cmd != nil {
		t.Error("expected nil cmd in demoMode")
	}
}

func TestDeleteSessionCmd_SuccessReturnsSessionDeletedMsg(t *testing.T) {
	original := runCommand
	defer func() { runCommand = original }()
	runCommand = func(name string, args ...string) error { return nil }

	m := newModel("/tmp/fake.db", false)
	msg := m.deleteSessionCmd("sess-1")()

	if _, ok := msg.(sessionDeletedMsg); !ok {
		t.Errorf("expected sessionDeletedMsg, got %T", msg)
	}
}

func TestDeleteSessionCmd_FailureReturnsOpErrMsg(t *testing.T) {
	original := runCommand
	defer func() { runCommand = original }()
	runCommand = func(name string, args ...string) error {
		return errors.New("delete failed")
	}

	m := newModel("/tmp/fake.db", false)
	msg := m.deleteSessionCmd("sess-1")()

	if _, ok := msg.(opErrMsg); !ok {
		t.Errorf("expected opErrMsg, got %T", msg)
	}
}

// ---------------------------------------------------------------------------
// deleteSessionsCmd — invoke the returned closure directly
// ---------------------------------------------------------------------------

func TestDeleteSessionsCmd_DemoModeReturnsNil(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	if cmd := m.deleteSessionsCmd([]string{"s1", "s2"}); cmd != nil {
		t.Error("expected nil cmd in demoMode")
	}
}

func TestDeleteSessionsCmd_SuccessReturnsSessionDeletedMsg(t *testing.T) {
	original := runCommand
	defer func() { runCommand = original }()

	var deleted []string
	runCommand = func(name string, args ...string) error {
		deleted = append(deleted, args[2]) // args[2] is the session id
		return nil
	}

	m := newModel("/tmp/fake.db", false)
	msg := m.deleteSessionsCmd([]string{"s1", "s2", "s3"})()

	if _, ok := msg.(sessionDeletedMsg); !ok {
		t.Errorf("expected sessionDeletedMsg, got %T", msg)
	}
	if len(deleted) != 3 {
		t.Errorf("expected 3 deletes, got %d", len(deleted))
	}
}

func TestDeleteSessionsCmd_FailureReturnsSessionsDeleteErrMsg(t *testing.T) {
	original := runCommand
	defer func() { runCommand = original }()
	runCommand = func(name string, args ...string) error {
		return errors.New("delete failed")
	}

	m := newModel("/tmp/fake.db", false)
	msg := m.deleteSessionsCmd([]string{"s1", "s2"})()

	if _, ok := msg.(sessionsDeleteErrMsg); !ok {
		t.Errorf("expected sessionsDeleteErrMsg, got %T", msg)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// createCorruptDB writes a file that is not a valid SQLite database, which
// causes openReadOnlyDB to return an error (treated as non-fatal by callers).
func createCorruptDB(t *testing.T, path string) error {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("this is not a sqlite database")
	return err
}

// ---------------------------------------------------------------------------
// Update message routing
// ---------------------------------------------------------------------------

func TestUpdate_SessionsLoadedMsg(t *testing.T) {
	m := newModel("/tmp/fake.db", true)

	sessions := []session{
		makeSession("s1", "session One", "/tmp/dir-a"),
		makeSession("s2", "session Two", "/tmp/dir-b"),
	}
	msg := sessionsLoadedMsg{sessions: sessions}

	result, _ := m.Update(msg)
	rm := mustModel(t, result)

	if len(rm.sessions) != 2 {
		t.Errorf("sessions: got %d, want 2", len(rm.sessions))
	}
	if len(rm.filtered) != 2 {
		t.Errorf("filtered: got %d, want 2", len(rm.filtered))
	}
	if len(rm.workspaces) != 2 {
		t.Errorf("workspaces: got %d, want 2 (one per dir)", len(rm.workspaces))
	}
	if rm.cursor != 0 {
		t.Errorf("cursor: got %d, want 0", rm.cursor)
	}
}

func TestUpdate_DbErrMsg(t *testing.T) {
	m := newModel("/tmp/fake.db", true)

	result, _ := m.Update(dbErrMsg{err: errors.New("db is toast")})
	rm := mustModel(t, result)

	if rm.mode != modeError {
		t.Errorf("mode: got %v, want modeError", rm.mode)
	}
	if rm.err == nil {
		t.Fatal("expected err to be set, got nil")
	}
	if rm.err.Error() != "db is toast" {
		t.Errorf("err.Error(): got %q, want %q", rm.err.Error(), "db is toast")
	}
}

func TestUpdate_OpErrMsg(t *testing.T) {
	m := newModel("/tmp/fake.db", true)

	result, _ := m.Update(opErrMsg{err: errors.New("clipboard failed")})
	rm := mustModel(t, result)

	if rm.notice != "clipboard failed" {
		t.Errorf("notice: got %q, want %q", rm.notice, "clipboard failed")
	}
	if rm.mode != modeNormal {
		t.Errorf("mode: got %v, want modeNormal", rm.mode)
	}
}

func TestUpdate_KeyMsgClearsNotice(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	m.notice = "some error"

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	rm := mustModel(t, result)

	if rm.notice != "" {
		t.Errorf("notice: got %q, want empty string", rm.notice)
	}
}

func TestUpdate_SessionsDeleteErrMsg(t *testing.T) {
	m := newModel("/tmp/fake.db", true)

	result, cmd := m.Update(sessionsDeleteErrMsg{err: errors.New("delete failed")})
	rm := mustModel(t, result)

	if rm.notice != "delete failed" {
		t.Errorf("notice: got %q, want %q", rm.notice, "delete failed")
	}
	if cmd == nil {
		t.Error("expected a non-nil reload cmd, got nil")
	}
}

func TestUpdate_StatsLoadedMsg_Accepted(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	m.filtered = []session{{ID: "sess-1", Title: "A", Directory: "/tmp/a"}}
	m.cursor = 0
	m.previewSessionID = "sess-1"

	result, _ := m.Update(statsLoadedMsg{sessionID: "sess-1", stats: sessionStats{MsgCount: 42}})
	rm := mustModel(t, result)

	if rm.stats == nil {
		t.Fatal("expected stats to be set, got nil")
	}
	if rm.stats.MsgCount != 42 {
		t.Errorf("stats.MsgCount: got %d, want 42", rm.stats.MsgCount)
	}
}

func TestUpdate_StatsLoadedMsg_Stale(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	m.filtered = []session{{ID: "sess-1", Title: "A", Directory: "/tmp/a"}}
	m.cursor = 0
	m.previewSessionID = "sess-1"

	result, _ := m.Update(statsLoadedMsg{sessionID: "stale-id", stats: sessionStats{MsgCount: 99}})
	rm := mustModel(t, result)

	if rm.stats != nil {
		t.Errorf("expected stats to be nil (stale load discarded), got %+v", rm.stats)
	}
}

func TestUpdate_MessagesLoadedMsg_Accepted(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	m.filtered = []session{{ID: "sess-1", Title: "A", Directory: "/tmp/a"}}
	m.cursor = 0

	result, _ := m.Update(messagesLoadedMsg{
		sessionID: "sess-1",
		messages:  []message{{Role: "user", Text: "hello"}},
	})
	rm := mustModel(t, result)

	if len(rm.messages) != 1 {
		t.Fatalf("messages: got %d, want 1", len(rm.messages))
	}
	if rm.messages[0].Text != "hello" {
		t.Errorf("messages[0].Text: got %q, want %q", rm.messages[0].Text, "hello")
	}
}

func TestUpdate_MessagesLoadedMsg_Stale(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	m.filtered = []session{{ID: "sess-1", Title: "A", Directory: "/tmp/a"}}
	m.cursor = 0

	result, _ := m.Update(messagesLoadedMsg{
		sessionID: "wrong-id",
		messages:  []message{{Role: "user", Text: "hello"}},
	})
	rm := mustModel(t, result)

	if rm.messages != nil {
		t.Errorf("expected messages to be nil (stale load discarded), got %v", rm.messages)
	}
}

// ---------------------------------------------------------------------------
// deleteSessionsCmd partial failure
// ---------------------------------------------------------------------------

func TestDeleteSessionsCmd_PartialFailure(t *testing.T) {
	original := runCommand
	defer func() { runCommand = original }()

	runCommand = func(name string, args ...string) error {
		if args[2] == "s2" {
			return errors.New("opencode: session s2 not found")
		}
		return nil
	}

	m := newModel("/tmp/fake.db", false)
	msg := m.deleteSessionsCmd([]string{"s1", "s2", "s3"})()

	errMsg, ok := msg.(sessionsDeleteErrMsg)
	if !ok {
		t.Fatalf("expected sessionsDeleteErrMsg, got %T", msg)
	}
	if errMsg.err == nil {
		t.Error("expected err to be non-nil")
	}
}

// ---------------------------------------------------------------------------
// loadGlobalStatsCmd
// ---------------------------------------------------------------------------

func TestLoadGlobalStatsCmd_DemoMode(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	cmd := m.loadGlobalStatsCmd()
	msg := cmd()

	if _, ok := msg.(globalStatsLoadedMsg); !ok {
		t.Fatalf("expected globalStatsLoadedMsg, got %T", msg)
	}
}

func TestLoadGlobalStatsCmd_WithDB(t *testing.T) {
	path := createTestDB(t)
	m := newModel(path, false)
	cmd := m.loadGlobalStatsCmd()
	msg := cmd()

	if _, ok := msg.(globalStatsLoadedMsg); !ok {
		t.Fatalf("expected globalStatsLoadedMsg, got %T", msg)
	}
}
