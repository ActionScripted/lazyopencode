package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// keyMsg builds a tea.KeyMsg from a string. Single characters are treated as
// rune keys; the following special names are also recognised:
//
//	"esc"    → tea.KeyEsc
//	"enter"  → tea.KeyEnter
//	"up"     → tea.KeyUp
//	"down"   → tea.KeyDown
func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

// modelWithWorkspaces returns a model seeded with sessions spread across the
// given directories. workspaceCursor starts at 0, mode is ModeWorkspaces.
func modelWithWorkspaces(dirs ...string) model {
	var sessions []Session
	for i, d := range dirs {
		sessions = append(sessions, Session{
			ID:        string(rune('a' + i)),
			Title:     d,
			Directory: d,
		})
	}
	m := newModel("/tmp/fake.db", true)
	m.sessions = sessions
	m.filtered = sessions
	m.workspaces = buildWorkspaces(sessions)
	m.mode = ModeWorkspaces
	return m
}

// modelWithSessions returns a ModeNormal model seeded with the given sessions.
func modelWithSessions(sessions ...Session) model {
	m := newModel("/tmp/fake.db", true)
	m.sessions = sessions
	m.filtered = sessions
	m.workspaces = buildWorkspaces(sessions)
	m.mode = ModeNormal
	return m
}

// ---------------------------------------------------------------------------
// clamp
// ---------------------------------------------------------------------------

func TestClamp(t *testing.T) {
	tests := []struct {
		v, lo, hi, want int
	}{
		{v: -1, lo: 0, hi: 5, want: 0}, // below floor
		{v: 10, lo: 0, hi: 5, want: 5}, // above ceiling
		{v: 3, lo: 0, hi: 5, want: 3},  // within range
		{v: 0, lo: 0, hi: 0, want: 0},  // lo == hi
		{v: 0, lo: 0, hi: 5, want: 0},  // exactly at lo
		{v: 5, lo: 0, hi: 5, want: 5},  // exactly at hi
	}
	for _, tc := range tests {
		got := clamp(tc.v, tc.lo, tc.hi)
		if got != tc.want {
			t.Errorf("clamp(%d, %d, %d) = %d, want %d", tc.v, tc.lo, tc.hi, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// updateWorkspaces
// ---------------------------------------------------------------------------

func TestUpdateWorkspaces_CursorMovement(t *testing.T) {
	tests := []struct {
		name       string
		dirs       []string
		startAt    int
		key        string
		wantCursor int
	}{
		{"down moves cursor", []string{"/tmp/a", "/tmp/b", "/tmp/c"}, 0, "j", 1},
		{"up moves cursor", []string{"/tmp/a", "/tmp/b", "/tmp/c"}, 2, "k", 1},
		{"up clamps at top", []string{"/tmp/a", "/tmp/b"}, 0, "k", 0},
		{"down clamps at bottom", []string{"/tmp/a", "/tmp/b"}, 1, "j", 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := modelWithWorkspaces(tc.dirs...)
			m.workspaceCursor = tc.startAt
			result, _ := m.updateWorkspaces(keyMsg(tc.key))
			got := result.(model).workspaceCursor
			if got != tc.wantCursor {
				t.Errorf("cursor after %q: got %d, want %d", tc.key, got, tc.wantCursor)
			}
		})
	}
}

func TestUpdateWorkspaces_WReturnsNormalMode(t *testing.T) {
	m := modelWithWorkspaces("/tmp/a")

	result, _ := m.updateWorkspaces(keyMsg("w"))
	got := result.(model).mode
	if got != ModeNormal {
		t.Errorf("mode after w: got %v, want ModeNormal", got)
	}
}

func TestUpdateWorkspaces_EscReturnsNormalMode(t *testing.T) {
	m := modelWithWorkspaces("/tmp/a")

	result, _ := m.updateWorkspaces(keyMsg("esc"))
	got := result.(model).mode
	if got != ModeNormal {
		t.Errorf("mode after esc: got %v, want ModeNormal", got)
	}
}

func TestUpdateWorkspaces_DeleteSetsPendingAndMode(t *testing.T) {
	m := modelWithWorkspaces("/tmp/a", "/tmp/b")
	m.workspaceCursor = 1

	result, _ := m.updateWorkspaces(keyMsg("d"))
	rm := result.(model)

	if rm.mode != ModeConfirmDeleteWorkspace {
		t.Errorf("mode: got %v, want ModeConfirmDeleteWorkspace", rm.mode)
	}
	if rm.pendingDeleteWorkspace != "/tmp/b" {
		t.Errorf("pendingDeleteWorkspace: got %q, want %q", rm.pendingDeleteWorkspace, "/tmp/b")
	}
}

func TestUpdateWorkspaces_DeleteWithEmptyWorkspacesNoOp(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	m.mode = ModeWorkspaces
	// workspaces is empty

	result, _ := m.updateWorkspaces(keyMsg("d"))
	rm := result.(model)

	if rm.mode != ModeWorkspaces {
		t.Errorf("mode should stay ModeWorkspaces, got %v", rm.mode)
	}
	if rm.pendingDeleteWorkspace != "" {
		t.Errorf("pendingDeleteWorkspace should be empty, got %q", rm.pendingDeleteWorkspace)
	}
}

// ---------------------------------------------------------------------------
// updateNormal
// ---------------------------------------------------------------------------

func TestUpdateNormal_DownMovesCursor(t *testing.T) {
	m := modelWithSessions(
		makeSession("1", "A", "/a"),
		makeSession("2", "B", "/b"),
	)
	result, _ := m.updateNormal(keyMsg("j"))
	got := result.(model).cursor
	if got != 1 {
		t.Errorf("cursor after j: got %d, want 1", got)
	}
}

func TestUpdateNormal_UpClampsAtTop(t *testing.T) {
	m := modelWithSessions(makeSession("1", "A", "/a"))
	result, _ := m.updateNormal(keyMsg("k"))
	got := result.(model).cursor
	if got != 0 {
		t.Errorf("cursor should stay 0, got %d", got)
	}
}

func TestUpdateNormal_DownClampsAtBottom(t *testing.T) {
	m := modelWithSessions(makeSession("1", "A", "/a"))
	result, _ := m.updateNormal(keyMsg("j"))
	got := result.(model).cursor
	if got != 0 {
		t.Errorf("cursor should stay 0, got %d", got)
	}
}

func TestUpdateNormal_SearchEntersSearchMode(t *testing.T) {
	m := modelWithSessions(makeSession("1", "A", "/a"))
	result, _ := m.updateNormal(keyMsg("/"))
	if result.(model).mode != ModeSearch {
		t.Errorf("expected ModeSearch, got %v", result.(model).mode)
	}
}

func TestUpdateNormal_WEntersWorkspacesMode(t *testing.T) {
	m := modelWithSessions(makeSession("1", "A", "/a"))
	result, _ := m.updateNormal(keyMsg("w"))
	if result.(model).mode != ModeWorkspaces {
		t.Errorf("expected ModeWorkspaces, got %v", result.(model).mode)
	}
}

func TestUpdateNormal_YEntersYankMode(t *testing.T) {
	m := modelWithSessions(makeSession("1", "A", "/a"))
	result, _ := m.updateNormal(keyMsg("y"))
	if result.(model).mode != ModeYank {
		t.Errorf("expected ModeYank, got %v", result.(model).mode)
	}
}

func TestUpdateNormal_YNoopWhenEmpty(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	m.mode = ModeNormal
	result, _ := m.updateNormal(keyMsg("y"))
	if result.(model).mode != ModeNormal {
		t.Errorf("mode should stay ModeNormal, got %v", result.(model).mode)
	}
}

func TestUpdateNormal_GEntersGotoMode(t *testing.T) {
	m := modelWithSessions(makeSession("1", "A", "/a"))
	result, _ := m.updateNormal(keyMsg("g"))
	if result.(model).mode != ModeGoto {
		t.Errorf("expected ModeGoto, got %v", result.(model).mode)
	}
}

func TestUpdateNormal_GNoopWhenEmpty(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	m.mode = ModeNormal
	result, _ := m.updateNormal(keyMsg("g"))
	if result.(model).mode != ModeNormal {
		t.Errorf("mode should stay ModeNormal, got %v", result.(model).mode)
	}
}

func TestUpdateNormal_DeleteSetsPendingAndMode(t *testing.T) {
	m := modelWithSessions(makeSession("sess-1", "A", "/a"))
	result, _ := m.updateNormal(keyMsg("d"))
	rm := result.(model)
	if rm.mode != ModeConfirmDelete {
		t.Errorf("mode: got %v, want ModeConfirmDelete", rm.mode)
	}
	if rm.pendingDeleteID != "sess-1" {
		t.Errorf("pendingDeleteID: got %q, want %q", rm.pendingDeleteID, "sess-1")
	}
}

func TestUpdateNormal_DeleteNoopWhenEmpty(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	m.mode = ModeNormal
	result, _ := m.updateNormal(keyMsg("d"))
	rm := result.(model)
	if rm.mode != ModeNormal {
		t.Errorf("mode should stay ModeNormal, got %v", rm.mode)
	}
	if rm.pendingDeleteID != "" {
		t.Errorf("pendingDeleteID should be empty, got %q", rm.pendingDeleteID)
	}
}

// ---------------------------------------------------------------------------
// updateConfirmDelete
// ---------------------------------------------------------------------------

func TestUpdateConfirmDelete_CancelClearsState(t *testing.T) {
	m := modelWithSessions(makeSession("sess-1", "A", "/a"))
	m.pendingDeleteID = "sess-1"
	m.mode = ModeConfirmDelete

	result, _ := m.updateConfirmDelete(keyMsg("n"))
	rm := result.(model)
	if rm.mode != ModeNormal {
		t.Errorf("mode: got %v, want ModeNormal", rm.mode)
	}
	if rm.pendingDeleteID != "" {
		t.Errorf("pendingDeleteID should be cleared, got %q", rm.pendingDeleteID)
	}
}

func TestUpdateConfirmDelete_ConfirmRemovesSessionOptimistically(t *testing.T) {
	m := modelWithSessions(
		makeSession("sess-1", "A", "/a"),
		makeSession("sess-2", "B", "/b"),
	)
	m.pendingDeleteID = "sess-1"
	m.mode = ModeConfirmDelete

	result, _ := m.updateConfirmDelete(keyMsg("y"))
	rm := result.(model)

	if rm.mode != ModeNormal {
		t.Errorf("mode: got %v, want ModeNormal", rm.mode)
	}
	for _, s := range rm.sessions {
		if s.ID == "sess-1" {
			t.Error("sess-1 should have been optimistically removed from sessions")
		}
	}
	for _, s := range rm.filtered {
		if s.ID == "sess-1" {
			t.Error("sess-1 should have been optimistically removed from filtered")
		}
	}
}

func TestUpdateConfirmDelete_ConfirmEmptyIDNoOp(t *testing.T) {
	m := modelWithSessions(makeSession("sess-1", "A", "/a"))
	m.pendingDeleteID = ""
	m.mode = ModeConfirmDelete

	result, cmd := m.updateConfirmDelete(keyMsg("y"))
	rm := result.(model)
	if rm.mode != ModeNormal {
		t.Errorf("mode: got %v, want ModeNormal", rm.mode)
	}
	if cmd != nil {
		t.Error("expected nil cmd for empty pendingDeleteID")
	}
}

// ---------------------------------------------------------------------------
// updateConfirmDeleteWorkspace
// ---------------------------------------------------------------------------

func TestUpdateConfirmDeleteWorkspace_CancelClearsState(t *testing.T) {
	m := modelWithWorkspaces("/tmp/a", "/tmp/b")
	m.pendingDeleteWorkspace = "/tmp/a"
	m.mode = ModeConfirmDeleteWorkspace

	result, _ := m.updateConfirmDeleteWorkspace(keyMsg("n"))
	rm := result.(model)
	if rm.mode != ModeWorkspaces {
		t.Errorf("mode: got %v, want ModeWorkspaces", rm.mode)
	}
	if rm.pendingDeleteWorkspace != "" {
		t.Errorf("pendingDeleteWorkspace should be cleared, got %q", rm.pendingDeleteWorkspace)
	}
}

func TestUpdateConfirmDeleteWorkspace_ConfirmRemovesSessionsOptimistically(t *testing.T) {
	sessions := []Session{
		makeSession("s1", "A", "/tmp/ws"),
		makeSession("s2", "B", "/tmp/ws"),
		makeSession("s3", "C", "/tmp/other"),
	}
	m := newModel("/tmp/fake.db", true)
	m.sessions = sessions
	m.filtered = sessions
	m.workspaces = buildWorkspaces(sessions)
	m.mode = ModeConfirmDeleteWorkspace
	m.pendingDeleteWorkspace = "/tmp/ws"

	result, _ := m.updateConfirmDeleteWorkspace(keyMsg("y"))
	rm := result.(model)

	if rm.mode != ModeWorkspaces {
		t.Errorf("mode: got %v, want ModeWorkspaces", rm.mode)
	}
	for _, s := range rm.sessions {
		if s.Directory == "/tmp/ws" {
			t.Errorf("session %q in /tmp/ws should have been removed", s.ID)
		}
	}
	// The other workspace's session must survive.
	found := false
	for _, s := range rm.sessions {
		if s.ID == "s3" {
			found = true
		}
	}
	if !found {
		t.Error("session s3 in /tmp/other should NOT have been removed")
	}
}

func TestUpdateConfirmDeleteWorkspace_ConfirmEmptyWSNoOp(t *testing.T) {
	m := modelWithWorkspaces("/tmp/a")
	m.pendingDeleteWorkspace = ""
	m.mode = ModeConfirmDeleteWorkspace

	result, cmd := m.updateConfirmDeleteWorkspace(keyMsg("y"))
	rm := result.(model)
	if rm.mode != ModeWorkspaces {
		t.Errorf("mode: got %v, want ModeWorkspaces", rm.mode)
	}
	if cmd != nil {
		t.Error("expected nil cmd for empty pendingDeleteWorkspace")
	}
}

// ---------------------------------------------------------------------------
// updateSearch
// ---------------------------------------------------------------------------

func TestUpdateSearch_EscReturnsNormalMode(t *testing.T) {
	m := modelWithSessions(makeSession("1", "A", "/a"))
	m.mode = ModeSearch
	m.search.Focus()

	result, _ := m.updateSearch(keyMsg("esc"))
	rm := result.(model)
	if rm.mode != ModeNormal {
		t.Errorf("mode: got %v, want ModeNormal", rm.mode)
	}
}

func TestUpdateSearch_TypingFilters(t *testing.T) {
	sessions := []Session{
		makeSession("1", "Fix login", "/a"),
		makeSession("2", "Refactor DB", "/b"),
	}
	m := newModel("/tmp/fake.db", true)
	m.sessions = sessions
	m.filtered = sessions
	m.mode = ModeSearch
	m.search.Focus()

	// Type "login" one char at a time isn't practical; instead use the public
	// filterSessions function to verify the underlying logic. What we test here
	// is that updateSearch passes the search value through to filtered.
	// Simulate a rune keystroke for "l".
	result, _ := m.updateSearch(keyMsg("l"))
	rm := result.(model)
	// filtered should now only contain sessions matching "l".
	for _, s := range rm.filtered {
		if !strings.Contains(strings.ToLower(s.FilterValue()), "l") {
			t.Errorf("unexpected session in filtered after typing 'l': %q", s.Title)
		}
	}
}

// ---------------------------------------------------------------------------
// updateGoto
// ---------------------------------------------------------------------------

func TestUpdateGoto_EscReturnsNormalMode(t *testing.T) {
	m := modelWithSessions(makeSession("1", "A", "/a"))
	m.mode = ModeGoto

	result, _ := m.updateGoto(keyMsg("esc"))
	if result.(model).mode != ModeNormal {
		t.Errorf("expected ModeNormal, got %v", result.(model).mode)
	}
}

func TestUpdateGoto_WEntersWorkspacesMode(t *testing.T) {
	m := modelWithSessions(makeSession("1", "A", "/tmp/a"))
	m.mode = ModeGoto

	result, _ := m.updateGoto(keyMsg("w"))
	rm := result.(model)
	if rm.mode != ModeWorkspaces {
		t.Errorf("expected ModeWorkspaces, got %v", rm.mode)
	}
}

func TestUpdateGoto_WJumpsToCorrectWorkspaceCursor(t *testing.T) {
	sessions := []Session{
		makeSession("1", "A", "/tmp/aaa"),
		makeSession("2", "B", "/tmp/zzz"),
	}
	m := newModel("/tmp/fake.db", true)
	m.sessions = sessions
	m.filtered = sessions
	m.workspaces = buildWorkspaces(sessions)
	m.cursor = 1 // selected session is in /tmp/zzz
	m.mode = ModeGoto

	result, _ := m.updateGoto(keyMsg("w"))
	rm := result.(model)
	// /tmp/zzz is index 1 in sorted workspaces (/tmp/aaa, /tmp/zzz)
	if rm.workspaceCursor != 1 {
		t.Errorf("workspaceCursor: got %d, want 1", rm.workspaceCursor)
	}
}

func TestUpdateGoto_EmptyFilteredFallsBackToNormal(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	m.mode = ModeGoto
	// filtered is empty

	result, _ := m.updateGoto(keyMsg("w"))
	if result.(model).mode != ModeNormal {
		t.Errorf("expected ModeNormal when filtered is empty, got %v", result.(model).mode)
	}
}

// ---------------------------------------------------------------------------
// updateYank
// ---------------------------------------------------------------------------

func TestUpdateYank_EscReturnsNormalMode(t *testing.T) {
	m := modelWithSessions(makeSession("1", "A", "/a"))
	m.mode = ModeYank

	result, _ := m.updateYank(keyMsg("esc"))
	if result.(model).mode != ModeNormal {
		t.Errorf("expected ModeNormal, got %v", result.(model).mode)
	}
}

func TestUpdateYank_DReturnsNormalMode(t *testing.T) {
	m := modelWithSessions(Session{
		ID: "sess-1", Title: "A", Directory: "/a", DisplayDir: "~/a",
	})
	m.mode = ModeYank

	result, cmd := m.updateYank(keyMsg("d"))
	rm := result.(model)
	if rm.mode != ModeNormal {
		t.Errorf("expected ModeNormal, got %v", rm.mode)
	}
	if cmd == nil {
		t.Error("expected a non-nil yank cmd")
	}
}

func TestUpdateYank_SReturnsNormalMode(t *testing.T) {
	m := modelWithSessions(makeSession("sess-1", "A", "/a"))
	m.mode = ModeYank

	result, cmd := m.updateYank(keyMsg("s"))
	rm := result.(model)
	if rm.mode != ModeNormal {
		t.Errorf("expected ModeNormal, got %v", rm.mode)
	}
	if cmd == nil {
		t.Error("expected a non-nil yank cmd")
	}
}

func TestUpdateYank_EmptyFilteredFallsBackToNormal(t *testing.T) {
	m := newModel("/tmp/fake.db", true)
	m.mode = ModeYank

	result, _ := m.updateYank(keyMsg("d"))
	if result.(model).mode != ModeNormal {
		t.Errorf("expected ModeNormal when filtered is empty, got %v", result.(model).mode)
	}
}
