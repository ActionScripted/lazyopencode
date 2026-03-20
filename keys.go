package main

import "github.com/charmbracelet/bubbles/key"

// Mode represents the current input mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeWorkspaces
	ModeConfirmDelete
	ModeConfirmDeleteWorkspace
	ModeYank
)

// KeyMap holds all named key bindings for the application.
type KeyMap struct {
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
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
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
			key.WithHelp("w", "workspaces"),
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
	}
}
