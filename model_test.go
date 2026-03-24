package main

import (
	"errors"
	"testing"
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
	if len(gotArgs) != len(wantArgs) {
		t.Fatalf("args length: got %d, want %d", len(gotArgs), len(wantArgs))
	}
	for i, a := range wantArgs {
		if gotArgs[i] != a {
			t.Errorf("arg[%d]: got %q, want %q", i, gotArgs[i], a)
		}
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
