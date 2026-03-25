package main

import (
	"os"
	"path/filepath"
	"testing"
)

// makeSession is a convenience constructor for tests that only care about a
// subset of Session fields.
func makeSession(id, title, dir string) Session {
	return Session{ID: id, Title: title, Directory: dir}
}

// ---------------------------------------------------------------------------
// filterSessions
// ---------------------------------------------------------------------------

func TestFilterSessions_EmptyQuery(t *testing.T) {
	sessions := []Session{
		makeSession("1", "Alpha", "/a"),
		makeSession("2", "Beta", "/b"),
	}
	got := filterSessions(sessions, "")
	if len(got) != len(sessions) {
		t.Fatalf("got %d sessions, want %d", len(got), len(sessions))
	}
}

func TestFilterSessions_MatchTitle(t *testing.T) {
	sessions := []Session{
		makeSession("1", "Fix login bug", "/a"),
		makeSession("2", "Refactor DB", "/b"),
	}
	got := filterSessions(sessions, "login")
	if len(got) != 1 || got[0].ID != "1" {
		t.Errorf("expected session 1, got %v", got)
	}
}

func TestFilterSessions_MatchTitleCaseInsensitive(t *testing.T) {
	sessions := []Session{
		makeSession("1", "Fix Login Bug", "/a"),
	}
	got := filterSessions(sessions, "LOGIN")
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %d", len(got))
	}
}

func TestFilterSessions_MatchDirectory(t *testing.T) {
	sessions := []Session{
		makeSession("1", "Session A", "/home/user/projects/myapp"),
		makeSession("2", "Session B", "/home/user/other"),
	}
	got := filterSessions(sessions, "myapp")
	if len(got) != 1 || got[0].ID != "1" {
		t.Errorf("expected session 1, got %v", got)
	}
}

func TestFilterSessions_NoMatch(t *testing.T) {
	sessions := []Session{
		makeSession("1", "Alpha", "/a"),
		makeSession("2", "Beta", "/b"),
	}
	got := filterSessions(sessions, "zzznomatch")
	if len(got) != 0 {
		t.Errorf("expected 0 matches, got %d", len(got))
	}
}

func TestFilterSessions_PartialMidString(t *testing.T) {
	sessions := []Session{
		makeSession("1", "Implement feature flag support", "/a"),
		makeSession("2", "Write tests", "/b"),
	}
	got := filterSessions(sessions, "ature fla")
	if len(got) != 1 || got[0].ID != "1" {
		t.Errorf("expected session 1, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// buildWorkspaces
// ---------------------------------------------------------------------------

func TestBuildWorkspaces_Empty(t *testing.T) {
	ws := buildWorkspaces(nil)
	if len(ws) != 0 {
		t.Errorf("expected empty slice, got %d workspaces", len(ws))
	}
}

func TestBuildWorkspaces_Deduplicates(t *testing.T) {
	sessions := []Session{
		makeSession("1", "A", "/tmp/proj"),
		makeSession("2", "B", "/tmp/proj"),
		makeSession("3", "C", "/tmp/proj"),
	}
	ws := buildWorkspaces(sessions)
	if len(ws) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(ws))
	}
	if ws[0].Dir != "/tmp/proj" {
		t.Errorf("Dir: got %q, want %q", ws[0].Dir, "/tmp/proj")
	}
}

func TestBuildWorkspaces_SortedByDir(t *testing.T) {
	sessions := []Session{
		makeSession("1", "A", "/tmp/zzz"),
		makeSession("2", "B", "/tmp/aaa"),
		makeSession("3", "C", "/tmp/mmm"),
	}
	ws := buildWorkspaces(sessions)
	if len(ws) != 3 {
		t.Fatalf("expected 3 workspaces, got %d", len(ws))
	}
	dirs := []string{ws[0].Dir, ws[1].Dir, ws[2].Dir}
	want := []string{"/tmp/aaa", "/tmp/mmm", "/tmp/zzz"}
	for i, d := range want {
		if dirs[i] != d {
			t.Errorf("ws[%d].Dir: got %q, want %q", i, dirs[i], d)
		}
	}
}

func TestBuildWorkspaces_DisplayDirSubstituted(t *testing.T) {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "projects", "lazyopencode")
	sessions := []Session{makeSession("1", "A", dir)}

	ws := buildWorkspaces(sessions)
	if len(ws) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(ws))
	}
	if ws[0].DisplayDir != "~/projects/lazyopencode" {
		t.Errorf("DisplayDir: got %q, want %q", ws[0].DisplayDir, "~/projects/lazyopencode")
	}
}

// ---------------------------------------------------------------------------
// removeSessionByID
// ---------------------------------------------------------------------------

func TestRemoveSessionByID_RemovesMatch(t *testing.T) {
	sessions := []Session{
		makeSession("1", "A", "/a"),
		makeSession("2", "B", "/b"),
		makeSession("3", "C", "/c"),
	}
	got := removeSessionByID(sessions, "2")
	if len(got) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(got))
	}
	for _, s := range got {
		if s.ID == "2" {
			t.Error("session 2 should have been removed")
		}
	}
}

func TestRemoveSessionByID_LeavesOthersIntact(t *testing.T) {
	sessions := []Session{
		makeSession("1", "A", "/a"),
		makeSession("2", "B", "/b"),
	}
	got := removeSessionByID(sessions, "1")
	if len(got) != 1 || got[0].ID != "2" {
		t.Errorf("expected only session 2, got %v", got)
	}
}

func TestRemoveSessionByID_NoopOnMissingID(t *testing.T) {
	sessions := []Session{
		makeSession("1", "A", "/a"),
		makeSession("2", "B", "/b"),
	}
	got := removeSessionByID(sessions, "999")
	if len(got) != 2 {
		t.Errorf("expected 2 sessions unchanged, got %d", len(got))
	}
}

func TestRemoveSessionByID_EmptyInput(t *testing.T) {
	got := removeSessionByID(nil, "1")
	if len(got) != 0 {
		t.Errorf("expected empty output, got %d sessions", len(got))
	}
}

// ---------------------------------------------------------------------------
// homeToTilde
// ---------------------------------------------------------------------------

func TestHomeToTilde_ReplacesHomePrefix(t *testing.T) {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "projects", "foo")
	got := homeToTilde(dir)
	want := "~/projects/foo"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHomeToTilde_NonHomePath(t *testing.T) {
	path := "/tmp/something"
	got := homeToTilde(path)
	if got != path {
		t.Errorf("got %q, want %q", got, path)
	}
}

func TestHomeToTilde_ExactlyHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := homeToTilde(home)
	if got != "~" {
		t.Errorf("got %q, want %q", got, "~")
	}
}

// ---------------------------------------------------------------------------
// baseName
// ---------------------------------------------------------------------------

func TestBaseName_LastComponent(t *testing.T) {
	got := baseName("/home/user/projects/myapp")
	if got != "myapp" {
		t.Errorf("got %q, want %q", got, "myapp")
	}
}

func TestBaseName_HomeDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := baseName(home)
	if got != "~" {
		t.Errorf("got %q, want %q", got, "~")
	}
}

// ---------------------------------------------------------------------------
// ShortDirectory / DisplayDirectory
// ---------------------------------------------------------------------------

func TestShortDirectory(t *testing.T) {
	s := Session{ShortDir: "myapp"}
	if got := s.ShortDirectory(); got != "myapp" {
		t.Errorf("got %q, want %q", got, "myapp")
	}
}

func TestDisplayDirectory(t *testing.T) {
	s := Session{DisplayDir: "~/projects/myapp"}
	if got := s.DisplayDirectory(); got != "~/projects/myapp" {
		t.Errorf("got %q, want %q", got, "~/projects/myapp")
	}
}
