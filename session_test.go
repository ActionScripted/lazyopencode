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

func TestRemoveSessionByID(t *testing.T) {
	sessions := []Session{
		makeSession("1", "A", "/a"),
		makeSession("2", "B", "/b"),
		makeSession("3", "C", "/c"),
	}
	tests := []struct {
		name    string
		input   []Session
		id      string
		wantIDs []string
	}{
		{"removes match", sessions, "2", []string{"1", "3"}},
		{"leaves others intact", sessions[:2], "1", []string{"2"}},
		{"noop on missing id", sessions[:2], "999", []string{"1", "2"}},
		{"empty input", nil, "1", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := removeSessionByID(tc.input, tc.id)
			if len(got) != len(tc.wantIDs) {
				t.Fatalf("got %d sessions, want %d", len(got), len(tc.wantIDs))
			}
			for i, s := range got {
				if s.ID != tc.wantIDs[i] {
					t.Errorf("got[%d].ID = %q, want %q", i, s.ID, tc.wantIDs[i])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// homeToTilde
// ---------------------------------------------------------------------------

func TestHomeToTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"replaces home prefix", filepath.Join(home, "projects", "foo"), "~/projects/foo"},
		{"non-home path unchanged", "/tmp/something", "/tmp/something"},
		{"exactly home", home, "~"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := homeToTilde(tc.in)
			if got != tc.want {
				t.Errorf("homeToTilde(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
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
