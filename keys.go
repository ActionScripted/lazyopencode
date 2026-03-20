package main

import "github.com/charmbracelet/bubbles/key"

// Mode represents the current input mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeWorkspaces
)

// KeyMap holds all named key bindings for the application.
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Search key.Binding
	Back   key.Binding
	Quit   key.Binding
	Tab    key.Binding
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
			key.WithKeys("esc", "enter"),
			key.WithHelp("esc/enter", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle workspaces view"),
		),
	}
}
