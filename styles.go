package main

import "github.com/charmbracelet/lipgloss"

var (
	// Tokyo Night palette
	colorFg        = lipgloss.AdaptiveColor{Light: "#343b58", Dark: "#c0caf5"}
	colorDim       = lipgloss.AdaptiveColor{Light: "#68608a", Dark: "#565f89"}
	colorAccent    = lipgloss.AdaptiveColor{Light: "#0f4b6e", Dark: "#7dcfff"}
	colorSelected  = lipgloss.AdaptiveColor{Light: "#b8cce4", Dark: "#283457"}
	colorBorder    = lipgloss.AdaptiveColor{Light: "#a8aecb", Dark: "#3b4261"}
	colorUser      = lipgloss.AdaptiveColor{Light: "#8f5e15", Dark: "#e0af68"} // Tokyo Night yellow
	colorAssistant = lipgloss.AdaptiveColor{Light: "#33635c", Dark: "#9ece6a"} // Tokyo Night green
	colorSubtle    = lipgloss.AdaptiveColor{Light: "#9699a3", Dark: "#414868"} // Tokyo Night comments

	// Text
	styleDim       = lipgloss.NewStyle().Foreground(colorDim)
	styleAccent    = lipgloss.NewStyle().Foreground(colorAccent)
	styleDateMuted = lipgloss.NewStyle().Foreground(colorSubtle)

	styleSearchPrefix = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	styleSeparator = lipgloss.NewStyle().
			Foreground(colorBorder)

	styleRow = lipgloss.NewStyle().
			Foreground(colorFg)

	styleSelectedRow = lipgloss.NewStyle().
				Background(colorSelected)

	// Preview
	stylePreviewPane = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(colorBorder).
				Foreground(colorFg).
				PaddingLeft(1).
				PaddingRight(1)

	stylePreviewTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorFg)

	styleRoleUser = lipgloss.NewStyle().
			Foreground(colorUser).
			Bold(true)

	styleRoleAssistant = lipgloss.NewStyle().
				Foreground(colorAssistant).
				Bold(true)

	// Hint bar
	styleHint = lipgloss.NewStyle().
			Foreground(colorDim).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(colorBorder)

	styleModeNormal = lipgloss.NewStyle().
			Reverse(true).
			Bold(true).
			Padding(0, 1)

	styleModeSearch = lipgloss.NewStyle().
			Foreground(colorAccent).
			Reverse(true).
			Bold(true).
			Padding(0, 1)
)
