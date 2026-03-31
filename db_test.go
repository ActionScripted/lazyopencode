package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// createTestDB creates a temporary SQLite file seeded with the minimal schema
// that mirrors the opencode database. Cleanup is handled automatically by
// t.TempDir.
func createTestDB(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.db")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("createTestDB: open: %v", err)
	}

	schema := `
		CREATE TABLE session (
			id                TEXT PRIMARY KEY,
			title             TEXT,
			directory         TEXT,
			time_created      INTEGER,
			time_updated      INTEGER,
			summary_files     INTEGER,
			summary_additions INTEGER,
			summary_deletions INTEGER,
			parent_id         TEXT
		);
		CREATE TABLE message (
			id         TEXT PRIMARY KEY,
			session_id TEXT,
			data       TEXT
		);
		CREATE TABLE part (
			id         TEXT PRIMARY KEY,
			message_id TEXT,
			session_id TEXT,
			data       TEXT,
			time_created INTEGER
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("createTestDB: schema: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("createTestDB: close: %v", err)
	}

	return path
}

// insertSession adds a single row to the session table.
func insertSession(t *testing.T, db *sql.DB, id, title, directory string, createdMS, updatedMS int64, parentID *string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO session (id, title, directory, time_created, time_updated, parent_id)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id, title, directory, createdMS, updatedMS, parentID,
	)
	if err != nil {
		t.Fatalf("insertSession: %v", err)
	}
}

// openRW opens a test DB in read-write mode for seeding.
func openRW(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("openRW: %v", err)
	}
	return db
}

// ---------------------------------------------------------------------------
// openReadOnlyDB
// ---------------------------------------------------------------------------

func TestOpenReadOnlyDB_MissingFile(t *testing.T) {
	db, missing, err := openReadOnlyDB("/nonexistent/path/to/opencode.db")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if !missing {
		t.Fatal("expected missing=true for a non-existent file")
	}
	if db != nil {
		t.Fatal("expected nil db for missing file")
	}
}

func TestOpenReadOnlyDB_ValidFile(t *testing.T) {
	path := createTestDB(t)

	db, missing, err := openReadOnlyDB(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if missing {
		t.Fatal("expected missing=false for an existing file")
	}
	if db == nil {
		t.Fatal("expected non-nil db")
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// loadSessions
// ---------------------------------------------------------------------------

func TestLoadSessions_Empty(t *testing.T) {
	path := createTestDB(t)

	sessions, err := loadSessions(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestLoadSessions_MissingDB(t *testing.T) {
	sessions, err := loadSessions("/nonexistent/opencode.db")
	if err != nil {
		t.Fatalf("missing DB should not return an error, got: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected empty slice for missing DB, got %d sessions", len(sessions))
	}
}

func TestLoadSessions_Populated(t *testing.T) {
	path := createTestDB(t)

	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "projects", "myapp")

	// Use a known epoch: 2024-01-15 12:00:00 UTC → 1705320000000 ms
	createdMS := int64(1705320000000)
	updatedMS := int64(1705323600000)

	db := openRW(t, path)
	insertSession(t, db, "sess-1", "My Session", dir, createdMS, updatedMS, nil)
	_ = db.Close()

	sessions, err := loadSessions(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	s := sessions[0]

	if s.ID != "sess-1" {
		t.Errorf("ID: got %q, want %q", s.ID, "sess-1")
	}
	if s.Title != "My Session" {
		t.Errorf("Title: got %q, want %q", s.Title, "My Session")
	}
	if s.Directory != dir {
		t.Errorf("Directory: got %q, want %q", s.Directory, dir)
	}

	wantCreated := time.Unix(createdMS/1000, 0)
	if !s.CreatedAt.Equal(wantCreated) {
		t.Errorf("CreatedAt: got %v, want %v", s.CreatedAt, wantCreated)
	}
	wantUpdated := time.Unix(updatedMS/1000, 0)
	if !s.UpdatedAt.Equal(wantUpdated) {
		t.Errorf("UpdatedAt: got %v, want %v", s.UpdatedAt, wantUpdated)
	}

	if s.DisplayDir != "~/projects/myapp" {
		t.Errorf("DisplayDir: got %q, want %q", s.DisplayDir, "~/projects/myapp")
	}
	if s.ShortDir != "myapp" {
		t.Errorf("ShortDir: got %q, want %q", s.ShortDir, "myapp")
	}
}

func TestLoadSessions_ChildSessionsExcluded(t *testing.T) {
	path := createTestDB(t)

	db := openRW(t, path)
	parentID := "parent-1"
	insertSession(t, db, "parent-1", "Parent", "/tmp/a", 1000, 2000, nil)
	insertSession(t, db, "child-1", "Child", "/tmp/a", 1000, 2000, &parentID)
	_ = db.Close()

	sessions, err := loadSessions(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session (parent only), got %d", len(sessions))
	}
	if sessions[0].ID != "parent-1" {
		t.Errorf("expected parent session, got %q", sessions[0].ID)
	}
}

// ---------------------------------------------------------------------------
// loadMessages
// ---------------------------------------------------------------------------

func TestLoadMessages_EmptySession(t *testing.T) {
	path := createTestDB(t)

	messages, err := loadMessages(path, "no-such-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(messages))
	}
}

func TestLoadMessages_Populated(t *testing.T) {
	path := createTestDB(t)

	db := openRW(t, path)

	// Insert two messages (user then assistant) with a text part each.
	_, err := db.Exec(`INSERT INTO message (id, session_id, data) VALUES
		('msg-1', 'sess-1', '{"role":"user"}'),
		('msg-2', 'sess-1', '{"role":"assistant"}')`)
	if err != nil {
		t.Fatalf("insert messages: %v", err)
	}
	_, err = db.Exec(`INSERT INTO part (id, message_id, session_id, data, time_created) VALUES
		('p-1', 'msg-1', 'sess-1', '{"type":"text","text":"hello from user"}', 1000),
		('p-2', 'msg-2', 'sess-1', '{"type":"text","text":"hello from assistant"}', 2000)`)
	if err != nil {
		t.Fatalf("insert parts: %v", err)
	}
	_ = db.Close()

	messages, err := loadMessages(path, "sess-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "user" || messages[0].Text != "hello from user" {
		t.Errorf("message[0]: got {%q, %q}", messages[0].Role, messages[0].Text)
	}
	if messages[1].Role != "assistant" || messages[1].Text != "hello from assistant" {
		t.Errorf("message[1]: got {%q, %q}", messages[1].Role, messages[1].Text)
	}
}

func TestLoadMessages_BlankTextExcluded(t *testing.T) {
	path := createTestDB(t)

	db := openRW(t, path)
	_, err := db.Exec(`INSERT INTO message (id, session_id, data) VALUES
		('msg-1', 'sess-1', '{"role":"user"}'),
		('msg-2', 'sess-1', '{"role":"assistant"}')`)
	if err != nil {
		t.Fatalf("insert messages: %v", err)
	}
	// msg-1 has a blank-text part (should be excluded); msg-2 has valid text.
	_, err = db.Exec(`INSERT INTO part (id, message_id, session_id, data, time_created) VALUES
		('p-1', 'msg-1', 'sess-1', '{"type":"text","text":"   "}', 1000),
		('p-2', 'msg-2', 'sess-1', '{"type":"text","text":"good reply"}', 2000)`)
	if err != nil {
		t.Fatalf("insert parts: %v", err)
	}
	_ = db.Close()

	messages, err := loadMessages(path, "sess-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message (blank excluded), got %d", len(messages))
	}
	if messages[0].Text != "good reply" {
		t.Errorf("Text: got %q, want %q", messages[0].Text, "good reply")
	}
}

// ---------------------------------------------------------------------------
// loadStats
// ---------------------------------------------------------------------------

func TestLoadStats_NoStepFinish(t *testing.T) {
	path := createTestDB(t)

	db := openRW(t, path)
	_, err := db.Exec(`INSERT INTO message (id, session_id, data) VALUES
		('msg-1', 'sess-1', '{"role":"user"}')`)
	if err != nil {
		t.Fatalf("insert message: %v", err)
	}
	// Only a text part — no step-finish parts.
	_, err = db.Exec(`INSERT INTO part (id, message_id, session_id, data, time_created) VALUES
		('p-1', 'msg-1', 'sess-1', '{"type":"text","text":"hello"}', 1000)`)
	if err != nil {
		t.Fatalf("insert part: %v", err)
	}
	_ = db.Close()

	stats, err := loadStats(path, "sess-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.MsgCount != 1 {
		t.Errorf("MsgCount: got %d, want 1", stats.MsgCount)
	}
	if stats.OutputTokens != 0 {
		t.Errorf("OutputTokens: got %d, want 0", stats.OutputTokens)
	}
	if stats.ContextTokens != 0 {
		t.Errorf("ContextTokens: got %d, want 0", stats.ContextTokens)
	}
}

func TestLoadStats_WithTokens(t *testing.T) {
	path := createTestDB(t)

	db := openRW(t, path)
	_, err := db.Exec(`INSERT INTO message (id, session_id, data) VALUES
		('msg-1', 'sess-1', '{"role":"user"}'),
		('msg-2', 'sess-1', '{"role":"assistant"}')`)
	if err != nil {
		t.Fatalf("insert messages: %v", err)
	}
	// Two step-finish parts: input 400+600, output 100+250; context tokens 1000 and 2000.
	// The last one (time_created=2000) provides context_tokens=2000.
	_, err = db.Exec(`INSERT INTO part (id, message_id, session_id, data, time_created) VALUES
		('p-1', 'msg-1', 'sess-1', '{"type":"step-finish","tokens":{"input":400,"output":100,"total":1000}}', 1000),
		('p-2', 'msg-2', 'sess-1', '{"type":"step-finish","tokens":{"input":600,"output":250,"total":2000}}', 2000)`)
	if err != nil {
		t.Fatalf("insert parts: %v", err)
	}
	_ = db.Close()

	stats, err := loadStats(path, "sess-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.MsgCount != 2 {
		t.Errorf("MsgCount: got %d, want 2", stats.MsgCount)
	}
	if stats.InputTokens != 1000 {
		t.Errorf("InputTokens: got %d, want 1000 (400+600)", stats.InputTokens)
	}
	if stats.OutputTokens != 350 {
		t.Errorf("OutputTokens: got %d, want 350 (100+250)", stats.OutputTokens)
	}
	if stats.ContextTokens != 2000 {
		t.Errorf("ContextTokens: got %d, want 2000 (last step-finish)", stats.ContextTokens)
	}
}
