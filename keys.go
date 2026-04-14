package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
)

// mode represents the current input mode.
type mode int

const (
	modeNormal mode = iota
	modeSearch
	modeWorkspaces
	modeConfirmDelete
	modeConfirmDeleteWorkspace
	modeYank
	modeGoto
	modeStats
	modeError
)

// String returns a human-readable name for the mode, used in logs and errors.
func (m mode) String() string {
	switch m {
	case modeNormal:
		return "Normal"
	case modeSearch:
		return "Search"
	case modeWorkspaces:
		return "Workspaces"
	case modeConfirmDelete:
		return "ConfirmDelete"
	case modeConfirmDeleteWorkspace:
		return "ConfirmDeleteWorkspace"
	case modeYank:
		return "Yank"
	case modeGoto:
		return "Goto"
	case modeStats:
		return "Stats"
	case modeError:
		return "Error"
	default:
		return fmt.Sprintf("mode(%d)", m)
	}
}

// keyMap holds all named key bindings for the application.
type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	Search        key.Binding
	Back          key.Binding
	Quit          key.Binding
	Workspaces    key.Binding
	Delete        key.Binding
	Open          key.Binding
	Yank          key.Binding
	YankDirectory key.Binding
	YankSession   key.Binding
	Confirm       key.Binding
	Cancel        key.Binding
	GotoPrefix    key.Binding
	GotoShell     key.Binding
	GotoWorkspace key.Binding
	Stats         key.Binding
	Refresh       key.Binding
	Theme         key.Binding
}

// defaultKeyMap returns the default key bindings.
func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Workspaces: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "toggle workspace"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Open: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
		Yank: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yank"),
		),
		YankDirectory: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "yank directory"),
		),
		YankSession: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "yank session id"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "esc", "q", "ctrl+c"),
			key.WithHelp("n/esc", "cancel"),
		),
		GotoPrefix: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "go to"),
		),
		GotoShell: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "shell"),
		),
		GotoWorkspace: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "toggle workspace"),
		),
		Stats: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "stats"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Theme: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "theme"),
		),
	}
}
