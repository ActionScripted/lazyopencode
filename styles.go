package main

import "github.com/charmbracelet/lipgloss"

var (
	// Tokyo Night palette
	colorFg        = lipgloss.AdaptiveColor{Light: "#343b58", Dark: "#a9b1d6"} // editor foreground
	colorTitle     = lipgloss.AdaptiveColor{Light: "#343b58", Dark: "#c0caf5"} // bright white / variable
	colorDim       = lipgloss.AdaptiveColor{Light: "#68608a", Dark: "#565f89"} // comments — muted text
	colorAccent    = lipgloss.AdaptiveColor{Light: "#0f4b6e", Dark: "#7dcfff"} // cyan — preview pane
	colorBlue      = lipgloss.AdaptiveColor{Light: "#3d59a1", Dark: "#7aa2f7"} // blue — list path + NORMAL badge
	colorBg        = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#1a1b26"} // editor background — dark text on colored badges
	colorBgPanel   = lipgloss.AdaptiveColor{Light: "#d5d6db", Dark: "#1e2030"} // opencode backgroundPanel — list areas, bottom bar
	colorSelected  = lipgloss.AdaptiveColor{Light: "#b8cce4", Dark: "#283457"}
	colorBorder    = lipgloss.AdaptiveColor{Light: "#a8aecb", Dark: "#3b4261"}
	colorUser      = lipgloss.AdaptiveColor{Light: "#8f5e15", Dark: "#e0af68"} // Tokyo Night yellow
	colorAssistant = lipgloss.AdaptiveColor{Light: "#33635c", Dark: "#9ece6a"} // Tokyo Night green

	// Text
	styleDim    = lipgloss.NewStyle().Foreground(colorDim)
	styleAccent = lipgloss.NewStyle().Foreground(colorAccent)

	styleSearchPrefix = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	styleSeparator = lipgloss.NewStyle().
			Foreground(colorBorder)

	styleListPane = lipgloss.NewStyle().
			Background(colorBgPanel)

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

	// Diff additions/deletions in the preview header
	styleAdd = lipgloss.NewStyle().Foreground(colorAssistant) // green
	styleDel = lipgloss.NewStyle().Foreground(colorDanger)    // red

	// Hint bar
	styleHint = lipgloss.NewStyle().
			Foreground(colorDim).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(colorBorder)

	styleModeNormal = lipgloss.NewStyle().
			Foreground(colorBg).
			Background(colorBlue).
			Bold(true).
			Padding(0, 1)

	styleModeSearch = lipgloss.NewStyle().
			Foreground(colorBg).
			Background(colorUser).
			Bold(true).
			Padding(0, 1)

	styleModeWorkspaces = lipgloss.NewStyle().
				Foreground(colorBg).
				Background(colorAccent).
				Bold(true).
				Padding(0, 1)

	styleModeYank = lipgloss.NewStyle().
			Foreground(colorBg).
			Background(colorAssistant).
			Bold(true).
			Padding(0, 1)

	// Modal (confirm-delete overlay)
	colorDanger = lipgloss.AdaptiveColor{Light: "#c53b53", Dark: "#f7768e"} // Tokyo Night red

	styleModal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDanger).
			Padding(1, 3)

	styleModalYank = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 3)

	styleModalGoto = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorUser).
			Padding(1, 3)

	styleModalTitle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	styleModalYankTitle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	styleModalGotoTitle = lipgloss.NewStyle().
				Foreground(colorUser).
				Bold(true)

	styleModalKey = lipgloss.NewStyle().
			Foreground(colorBg).
			Background(colorDanger).
			Bold(true).
			Padding(0, 1)

	styleModalKeyYank = lipgloss.NewStyle().
				Foreground(colorBg).
				Background(colorAccent).
				Bold(true).
				Padding(0, 1)

	styleModalKeyGoto = lipgloss.NewStyle().
				Foreground(colorBg).
				Background(colorUser).
				Bold(true).
				Padding(0, 1)

	styleModalKeyCancel = lipgloss.NewStyle().
				Foreground(colorBg).
				Background(colorDim).
				Bold(true).
				Padding(0, 1)

	styleModeConfirmDelete = lipgloss.NewStyle().
				Foreground(colorBg).
				Background(colorDanger).
				Bold(true).
				Padding(0, 1)

	styleModeGoto = lipgloss.NewStyle().
			Foreground(colorBg).
			Background(colorUser).
			Bold(true).
			Padding(0, 1)
)
