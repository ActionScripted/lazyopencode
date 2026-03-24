package main

import "github.com/charmbracelet/lipgloss"

var (
	// Tokyo Night palette
	colorFg       = lipgloss.AdaptiveColor{Light: "#343b58", Dark: "#a9b1d6"} // editor foreground
	colorBright   = lipgloss.AdaptiveColor{Light: "#343b58", Dark: "#c0caf5"} // bright white / variable
	colorDim      = lipgloss.AdaptiveColor{Light: "#68608a", Dark: "#565f89"} // comments — muted text
	colorCyan     = lipgloss.AdaptiveColor{Light: "#0f4b6e", Dark: "#7dcfff"} // cyan — preview pane
	colorBlue     = lipgloss.AdaptiveColor{Light: "#3d59a1", Dark: "#7aa2f7"} // blue — list path + NORMAL badge
	colorBg       = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#1a1b26"} // editor background — dark text on colored badges
	colorBgPanel  = lipgloss.AdaptiveColor{Light: "#d5d6db", Dark: "#1e2030"} // opencode backgroundPanel — list areas, bottom bar
	colorSelected = lipgloss.AdaptiveColor{Light: "#b8cce4", Dark: "#283457"}
	colorBorder   = lipgloss.AdaptiveColor{Light: "#a8aecb", Dark: "#3b4261"}
	colorYellow   = lipgloss.AdaptiveColor{Light: "#8f5e15", Dark: "#e0af68"} // Tokyo Night yellow
	colorGreen    = lipgloss.AdaptiveColor{Light: "#33635c", Dark: "#9ece6a"} // Tokyo Night green
	colorDanger   = lipgloss.AdaptiveColor{Light: "#c53b53", Dark: "#f7768e"} // Tokyo Night red

	// Text
	styleDim    = lipgloss.NewStyle().Foreground(colorDim)
	styleAccent = lipgloss.NewStyle().Foreground(colorCyan)

	styleSearchPrefix = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	styleSeparator = lipgloss.NewStyle().
			Foreground(colorBorder)

	styleListPane = lipgloss.NewStyle().
			Background(colorBgPanel)

	// Preview panes
	stylePreviewPane = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(colorBorder).
				Foreground(colorFg).
				PaddingLeft(1).
				PaddingRight(1)

	stylePreviewPaneStacked = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderTop(true).
				BorderForeground(colorBorder).
				Foreground(colorFg).
				PaddingLeft(1).
				PaddingRight(1)

	stylePreviewTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorFg)

	styleRoleUser = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	styleRoleAssistant = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true)

	// Diff additions/deletions in the preview header
	styleAdd = lipgloss.NewStyle().Foreground(colorGreen)  // green
	styleDel = lipgloss.NewStyle().Foreground(colorDanger) // red

	// Hint bar
	styleHint = lipgloss.NewStyle().
			Foreground(colorDim).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderBottom(true).
			BorderForeground(colorBorder)

	styleHintKey = lipgloss.NewStyle().Foreground(colorBright)

	styleHintBottom = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorBorder)

	// Badge base: shared foundation for all mode badges and modal key badges.
	// Each specific badge derives from this by setting its Background color.
	styleBadgeBase = lipgloss.NewStyle().
			Foreground(colorBg).
			Bold(true).
			Padding(0, 1)

	// Mode badges (hint bar)
	styleModeNormal        = styleBadgeBase.Background(colorBlue)
	styleModeSearch        = styleBadgeBase.Background(colorYellow)
	styleModeWorkspaces    = styleBadgeBase.Background(colorCyan)
	styleModeYank          = styleBadgeBase.Background(colorGreen)
	styleModeConfirmDelete = styleBadgeBase.Background(colorDanger)
	styleModeGoto          = styleBadgeBase.Background(colorYellow)

	// Modal containers (confirm-delete / yank / goto overlays)
	styleModal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDanger).
			Padding(1, 3)

	styleModalYank = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			Padding(1, 3)

	styleModalGoto = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorYellow).
			Padding(1, 3)

	// Modal title text
	styleModalTitle     = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)
	styleModalYankTitle = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	styleModalGotoTitle = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)

	// Modal key badges (action keys rendered inside modals)
	styleModalKey       = styleBadgeBase.Background(colorDanger)
	styleModalKeyYank   = styleBadgeBase.Background(colorCyan)
	styleModalKeyGoto   = styleBadgeBase.Background(colorYellow)
	styleModalKeyCancel = styleBadgeBase.Background(colorDim)
)
