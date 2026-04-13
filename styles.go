package main

import "github.com/charmbracelet/lipgloss"

var (
	// Tokyo Night palette
	colorFg         = lipgloss.AdaptiveColor{Light: "#343b58", Dark: "#a9b1d6"} // editor foreground
	colorBright     = lipgloss.AdaptiveColor{Light: "#343b58", Dark: "#c0caf5"} // bright white / variable
	colorDim        = lipgloss.AdaptiveColor{Light: "#68608a", Dark: "#565f89"} // comments — muted text
	colorCyan       = lipgloss.AdaptiveColor{Light: "#0f4b6e", Dark: "#7dcfff"} // cyan — preview pane
	colorBlue       = lipgloss.AdaptiveColor{Light: "#3d59a1", Dark: "#7aa2f7"} // blue — list path + NORMAL badge
	colorOnBadge    = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#1a1b26"} // editor background — dark text on colored badges
	colorBgPanel    = lipgloss.AdaptiveColor{Light: "#d5d6db", Dark: "#1e2030"} // opencode backgroundPanel — list areas, bottom bar
	colorBgPanelAlt = lipgloss.AdaptiveColor{Light: "#c8c9cf", Dark: "#222436"} // slightly lighter stripe for stats table zebra rows
	colorSelected   = lipgloss.AdaptiveColor{Light: "#b8cce4", Dark: "#283457"}
	colorBorder     = lipgloss.AdaptiveColor{Light: "#a8aecb", Dark: "#3b4261"}
	colorYellow     = lipgloss.AdaptiveColor{Light: "#8f5e15", Dark: "#e0af68"} // Tokyo Night yellow
	colorGreen      = lipgloss.AdaptiveColor{Light: "#33635c", Dark: "#9ece6a"} // Tokyo Night green
	colorDanger     = lipgloss.AdaptiveColor{Light: "#c53b53", Dark: "#f7768e"} // Tokyo Night red
	colorPurple     = lipgloss.AdaptiveColor{Light: "#5a4a78", Dark: "#bb9af7"} // Tokyo Night purple

	// Text
	styleDim    = lipgloss.NewStyle().Foreground(colorDim)
	styleAccent = lipgloss.NewStyle().Foreground(colorCyan)

	// Hint bar app name ("LazyOpenCode" logo) and decoration dots.
	// These are declared here rather than inline in renderHint so that a theme
	// change is a single-file edit.
	styleAppNameLazy     = lipgloss.NewStyle().Foreground(colorBlue)
	styleAppNameOpenCode = lipgloss.NewStyle().Foreground(colorBright).Bold(true)
	styleDotBlue         = lipgloss.NewStyle().Foreground(colorBlue)
	styleDotCyan         = lipgloss.NewStyle().Foreground(colorCyan)
	styleDotYellow       = lipgloss.NewStyle().Foreground(colorYellow)

	// Workspace session row title (right pane of workspaces view).
	styleWorkspaceSessionTitle = lipgloss.NewStyle().Foreground(colorBright).Bold(true)

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

	// Panel-background variants of styleAdd/styleDel for use inside the stats
	// dashboard where every span must carry colorBgPanel explicitly to prevent
	// terminal-background bleed-through on ANSI reset sequences.
	styleAddPanel = lipgloss.NewStyle().Foreground(colorGreen).Background(colorBgPanel)
	styleDelPanel = lipgloss.NewStyle().Foreground(colorDanger).Background(colorBgPanel)

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

	styleTopBar = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorBorder)

	// Badge base: shared foundation for all mode badges and modal key badges.
	// Each specific badge derives from this by setting its Background color.
	styleBadgeBase = lipgloss.NewStyle().
			Foreground(colorOnBadge).
			Bold(true).
			Padding(0, 1)

	// Mode badges (hint bar)
	styleModeNormal        = styleBadgeBase.Background(colorBlue)
	styleModeSearch        = styleBadgeBase.Background(colorYellow)
	styleModeWorkspaces    = styleBadgeBase.Background(colorCyan)
	styleModeYank          = styleBadgeBase.Background(colorGreen)
	styleModeConfirmDelete = styleBadgeBase.Background(colorDanger)
	styleModeGoto          = styleBadgeBase.Background(colorYellow)
	styleModeStats         = styleBadgeBase.Background(colorPurple)
	styleModeError         = styleBadgeBase.Background(colorDanger)

	// Stats dashboard
	// Uses *Panel variants throughout so every span carries colorBgPanel —
	// in lipgloss v1.x the outer styleListPane.Background() only fills padding
	// cells, so inner styled spans must carry the bg explicitly to prevent
	// terminal-background bleed-through on ANSI reset sequences.
	styleSpPanel          = lipgloss.NewStyle().Background(colorBgPanel)
	styleDimPanel         = lipgloss.NewStyle().Foreground(colorDim).Background(colorBgPanel)
	styleStatsTitlePanel  = lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Background(colorBgPanel)
	styleStatsLabelPanel  = lipgloss.NewStyle().Foreground(colorFg).Background(colorBgPanel)
	styleStatsCountPanel  = lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Background(colorBgPanel)
	styleStatsHeaderPanel = lipgloss.NewStyle().Foreground(colorDim).Background(colorBgPanel)

	// Zebra-stripe alt-background variants for stats table data rows (odd rows).
	styleSpPanelAlt         = lipgloss.NewStyle().Background(colorBgPanelAlt)
	styleDimPanelAlt        = lipgloss.NewStyle().Foreground(colorDim).Background(colorBgPanelAlt)
	styleStatsLabelPanelAlt = lipgloss.NewStyle().Foreground(colorFg).Background(colorBgPanelAlt)
	styleStatsCountPanelAlt = lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Background(colorBgPanelAlt)

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

// Row base styles — non-selected and selected variants.
// formatSessionRow and formatWorkspaceRow pick between these instead of
// constructing lipgloss.NewStyle() inline.
var (
	styleRowBase         = lipgloss.NewStyle().Foreground(colorFg).Background(colorBgPanel)
	styleRowSelected     = lipgloss.NewStyle().Foreground(colorFg).Background(colorSelected)
	styleWorkspaceRow    = lipgloss.NewStyle().Foreground(colorCyan).Background(colorBgPanel)
	styleWorkspaceRowSel = lipgloss.NewStyle().Foreground(colorCyan).Background(colorSelected)
)
