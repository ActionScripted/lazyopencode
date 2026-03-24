package main

import (
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

func TestUpdateWorkspaces_DownMovesCursor(t *testing.T) {
	m := modelWithWorkspaces("/tmp/a", "/tmp/b", "/tmp/c")
	m.workspaceCursor = 0

	result, _ := m.updateWorkspaces(keyMsg("j"))
	got := result.(model).workspaceCursor
	if got != 1 {
		t.Errorf("cursor after j: got %d, want 1", got)
	}
}

func TestUpdateWorkspaces_UpMovesCursor(t *testing.T) {
	m := modelWithWorkspaces("/tmp/a", "/tmp/b", "/tmp/c")
	m.workspaceCursor = 2

	result, _ := m.updateWorkspaces(keyMsg("k"))
	got := result.(model).workspaceCursor
	if got != 1 {
		t.Errorf("cursor after k: got %d, want 1", got)
	}
}

func TestUpdateWorkspaces_UpClampsAtTop(t *testing.T) {
	m := modelWithWorkspaces("/tmp/a", "/tmp/b")
	m.workspaceCursor = 0

	result, _ := m.updateWorkspaces(keyMsg("k"))
	got := result.(model).workspaceCursor
	if got != 0 {
		t.Errorf("cursor should stay at 0, got %d", got)
	}
}

func TestUpdateWorkspaces_DownClampsAtBottom(t *testing.T) {
	m := modelWithWorkspaces("/tmp/a", "/tmp/b")
	m.workspaceCursor = 1

	result, _ := m.updateWorkspaces(keyMsg("j"))
	got := result.(model).workspaceCursor
	if got != 1 {
		t.Errorf("cursor should stay at 1, got %d", got)
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
