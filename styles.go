package main

import "github.com/charmbracelet/lipgloss"

// Color vars — zero-initialized; populated by initStyles() on startup and on
// every theme switch. All other files reference these vars directly.
var (
	colorFg         lipgloss.AdaptiveColor
	colorBright     lipgloss.AdaptiveColor
	colorDim        lipgloss.AdaptiveColor
	colorCyan       lipgloss.AdaptiveColor
	colorBlue       lipgloss.AdaptiveColor
	colorOnBadge    lipgloss.AdaptiveColor
	colorBgDark     lipgloss.AdaptiveColor
	colorBgPanel    lipgloss.AdaptiveColor
	colorBgPanelAlt lipgloss.AdaptiveColor
	colorSelected   lipgloss.AdaptiveColor
	colorBorder     lipgloss.AdaptiveColor
	colorYellow     lipgloss.AdaptiveColor
	colorGreen      lipgloss.AdaptiveColor
	colorDanger     lipgloss.AdaptiveColor
	colorPurple     lipgloss.AdaptiveColor
	colorHeading    lipgloss.AdaptiveColor
)

// Style vars — zero-initialized; populated by initStyles().
var (
	// Text
	styleDim     lipgloss.Style
	styleDimDark lipgloss.Style
	styleAccent  lipgloss.Style

	// Hint bar app name and decoration dots.
	styleAppNameLazy     lipgloss.Style
	styleAppNameOpenCode lipgloss.Style
	styleDotBlue         lipgloss.Style
	styleDotCyan         lipgloss.Style
	styleDotYellow       lipgloss.Style

	// Workspace session row title (right pane of workspaces view).
	styleWorkspaceSessionTitle lipgloss.Style

	styleSearchPrefix lipgloss.Style
	styleSeparator    lipgloss.Style
	styleListPane     lipgloss.Style

	// Preview panes
	stylePreviewPane        lipgloss.Style
	stylePreviewPaneStacked lipgloss.Style
	stylePreviewTitle       lipgloss.Style

	styleRoleUser      lipgloss.Style
	styleRoleAssistant lipgloss.Style

	// Diff additions/deletions in the preview header
	styleAdd lipgloss.Style
	styleDel lipgloss.Style

	// Panel-background variants of styleAdd/styleDel
	styleAddPanel lipgloss.Style
	styleDelPanel lipgloss.Style

	// Hint bar
	styleHint       lipgloss.Style
	styleHintKey    lipgloss.Style
	styleHintBottom lipgloss.Style
	styleTopBar     lipgloss.Style

	// Badge base
	styleBadgeBase lipgloss.Style

	// Mode badges (hint bar)
	styleModeNormal        lipgloss.Style
	styleModeSearch        lipgloss.Style
	styleModeWorkspaces    lipgloss.Style
	styleModeYank          lipgloss.Style
	styleModeConfirmDelete lipgloss.Style
	styleModeGoto          lipgloss.Style
	styleModeStats         lipgloss.Style
	styleModeError         lipgloss.Style

	// Preview/modal background matches colorBgPanel.
	styleSpBgPanel lipgloss.Style

	// Stats dashboard
	styleSpPanel          lipgloss.Style
	styleDimPanel         lipgloss.Style
	styleStatsTitlePanel  lipgloss.Style
	styleStatsLabelPanel  lipgloss.Style
	styleStatsCountPanel  lipgloss.Style
	styleStatsHeaderPanel lipgloss.Style

	// Zebra-stripe alt-background variants
	styleSpPanelAlt         lipgloss.Style
	styleDimPanelAlt        lipgloss.Style
	styleStatsLabelPanelAlt lipgloss.Style
	styleStatsCountPanelAlt lipgloss.Style

	// Modal containers
	styleModal     lipgloss.Style
	styleModalYank lipgloss.Style
	styleModalGoto lipgloss.Style

	// Modal title text
	styleModalTitle     lipgloss.Style
	styleModalYankTitle lipgloss.Style
	styleModalGotoTitle lipgloss.Style

	// Modal key badges
	styleModalKey       lipgloss.Style
	styleModalKeyYank   lipgloss.Style
	styleModalKeyGoto   lipgloss.Style
	styleModalKeyCancel lipgloss.Style
)

// Row base styles
var (
	styleRowBase         lipgloss.Style
	styleRowSelected     lipgloss.Style
	styleWorkspaceRow    lipgloss.Style
	styleWorkspaceRowSel lipgloss.Style

	styleRowTitleBase     lipgloss.Style
	styleRowTitleSelected lipgloss.Style
	styleRowDateBase      lipgloss.Style
	styleRowDateSelected  lipgloss.Style
	styleRowPathBase      lipgloss.Style
	styleRowPathSelected  lipgloss.Style
)

// initStyles assigns all color vars from activePalette and rebuilds every
// style var. Call once at startup and again on every theme switch.
func initStyles() {
	// ── colors ────────────────────────────────────────────────────────────────
	colorFg = activePalette.fg
	colorBright = activePalette.bright
	colorDim = activePalette.dim
	colorCyan = activePalette.cyan
	colorBlue = activePalette.blue
	colorOnBadge = activePalette.onBadge
	colorBgDark = activePalette.bgDark
	colorBgPanel = activePalette.bgPanel
	colorBgPanelAlt = activePalette.bgPanelAlt
	colorSelected = activePalette.selected
	colorBorder = activePalette.border
	colorYellow = activePalette.yellow
	colorGreen = activePalette.green
	colorDanger = activePalette.danger
	colorPurple = activePalette.purple
	colorHeading = activePalette.heading

	// ── text ──────────────────────────────────────────────────────────────────
	styleDim = lipgloss.NewStyle().Foreground(colorDim).Background(colorBgPanel)
	styleDimDark = lipgloss.NewStyle().Foreground(colorDim).Background(colorBgDark)
	styleAccent = lipgloss.NewStyle().Foreground(colorCyan).Background(colorBgPanel)

	// ── hint bar app name and dots ────────────────────────────────────────────
	styleAppNameLazy = lipgloss.NewStyle().Foreground(colorBlue).Background(colorBgPanel)
	styleAppNameOpenCode = lipgloss.NewStyle().Foreground(colorBright).Bold(true).Background(colorBgPanel)
	styleDotBlue = lipgloss.NewStyle().Foreground(colorBlue).Background(colorBgPanel)
	styleDotCyan = lipgloss.NewStyle().Foreground(colorCyan).Background(colorBgPanel)
	styleDotYellow = lipgloss.NewStyle().Foreground(colorYellow).Background(colorBgPanel)

	// ── workspace ─────────────────────────────────────────────────────────────
	styleWorkspaceSessionTitle = lipgloss.NewStyle().Foreground(colorBright).Bold(true).Background(colorBgPanel)

	// ── search / list ─────────────────────────────────────────────────────────
	styleSearchPrefix = lipgloss.NewStyle().
		Foreground(colorCyan).
		Background(colorBgPanel).
		Bold(true)

	styleSeparator = lipgloss.NewStyle().
		Foreground(colorBorder).
		Background(colorBgPanel)

	styleListPane = lipgloss.NewStyle().
		Background(colorBgDark)

	// ── preview panes ─────────────────────────────────────────────────────────
	stylePreviewPane = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(colorBorder).
		BorderBackground(colorBgPanel).
		Background(colorBgPanel).
		Foreground(colorFg).
		PaddingLeft(1).
		PaddingRight(1)

	stylePreviewPaneStacked = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(colorBorder).
		BorderBackground(colorBgPanel).
		Background(colorBgPanel).
		Foreground(colorFg).
		PaddingLeft(1).
		PaddingRight(1)

	stylePreviewTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorFg).
		Background(colorBgPanel)

	styleRoleUser = lipgloss.NewStyle().
		Foreground(colorYellow).
		Bold(true).
		Background(colorBgPanel)

	styleRoleAssistant = lipgloss.NewStyle().
		Foreground(colorGreen).
		Bold(true).
		Background(colorBgPanel)

	// ── diff ──────────────────────────────────────────────────────────────────
	styleAdd = lipgloss.NewStyle().Foreground(colorGreen).Background(colorBgPanel)
	styleDel = lipgloss.NewStyle().Foreground(colorDanger).Background(colorBgPanel)

	styleAddPanel = lipgloss.NewStyle().Foreground(colorGreen).Background(colorBgDark)
	styleDelPanel = lipgloss.NewStyle().Foreground(colorDanger).Background(colorBgDark)

	// ── hint bar ──────────────────────────────────────────────────────────────
	styleHint = lipgloss.NewStyle().
		Foreground(colorDim).
		Background(colorBgPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderBottom(true).
		BorderForeground(colorBorder).
		BorderBackground(colorBgPanel)

	styleHintKey = lipgloss.NewStyle().Foreground(colorBright).Background(colorBgPanel)

	styleHintBottom = lipgloss.NewStyle().
		Background(colorBgPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(colorBorder).
		BorderBackground(colorBgPanel)

	styleTopBar = lipgloss.NewStyle().
		Background(colorBgPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(colorBorder).
		BorderBackground(colorBgPanel)

	// ── badges ────────────────────────────────────────────────────────────────
	styleBadgeBase = lipgloss.NewStyle().
		Foreground(colorOnBadge).
		Bold(true).
		Padding(0, 1)

	styleModeNormal = styleBadgeBase.Background(colorBlue)
	styleModeSearch = styleBadgeBase.Background(colorYellow)
	styleModeWorkspaces = styleBadgeBase.Background(colorCyan)
	styleModeYank = styleBadgeBase.Background(colorGreen)
	styleModeConfirmDelete = styleBadgeBase.Background(colorDanger)
	styleModeGoto = styleBadgeBase.Background(colorYellow)
	styleModeStats = styleBadgeBase.Background(colorPurple)
	styleModeError = styleBadgeBase.Background(colorDanger)

	styleSpBgPanel = lipgloss.NewStyle().Background(colorBgPanel)

	// ── stats dashboard ───────────────────────────────────────────────────────
	styleSpPanel = lipgloss.NewStyle().Background(colorBgDark)
	styleDimPanel = lipgloss.NewStyle().Foreground(colorDim).Background(colorBgDark)
	styleStatsTitlePanel = lipgloss.NewStyle().Foreground(colorHeading).Bold(true).Background(colorBgDark)
	styleStatsLabelPanel = lipgloss.NewStyle().Foreground(colorFg).Background(colorBgDark)
	styleStatsCountPanel = lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Background(colorBgDark)
	styleStatsHeaderPanel = lipgloss.NewStyle().Foreground(colorDim).Background(colorBgDark)

	styleSpPanelAlt = lipgloss.NewStyle().Background(colorBgPanelAlt)
	styleDimPanelAlt = lipgloss.NewStyle().Foreground(colorDim).Background(colorBgPanelAlt)
	styleStatsLabelPanelAlt = lipgloss.NewStyle().Foreground(colorFg).Background(colorBgPanelAlt)
	styleStatsCountPanelAlt = lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Background(colorBgPanelAlt)

	// ── modals ────────────────────────────────────────────────────────────────
	styleModal = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorDanger).
		BorderBackground(colorBgPanel).
		Background(colorBgPanel).
		Padding(1, 3)

	styleModalYank = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		BorderBackground(colorBgPanel).
		Background(colorBgPanel).
		Padding(1, 3)

	styleModalGoto = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorYellow).
		BorderBackground(colorBgPanel).
		Background(colorBgPanel).
		Padding(1, 3)

	styleModalTitle = lipgloss.NewStyle().Foreground(colorDanger).Bold(true).Background(colorBgPanel)
	styleModalYankTitle = lipgloss.NewStyle().Foreground(colorCyan).Bold(true).Background(colorBgPanel)
	styleModalGotoTitle = lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Background(colorBgPanel)

	styleModalKey = styleBadgeBase.Background(colorDanger)
	styleModalKeyYank = styleBadgeBase.Background(colorCyan)
	styleModalKeyGoto = styleBadgeBase.Background(colorYellow)
	styleModalKeyCancel = styleBadgeBase.Background(colorDim)

	// ── row styles ────────────────────────────────────────────────────────────
	styleRowBase = lipgloss.NewStyle().Foreground(colorFg).Background(colorBgDark)
	styleRowSelected = lipgloss.NewStyle().Foreground(colorFg).Background(colorSelected)
	styleWorkspaceRow = lipgloss.NewStyle().Foreground(colorCyan).Background(colorBgDark)
	styleWorkspaceRowSel = lipgloss.NewStyle().Foreground(colorCyan).Background(colorSelected)

	styleRowTitleBase = styleRowBase.Foreground(colorBright).Bold(true)
	styleRowTitleSelected = styleRowSelected.Foreground(colorBright).Bold(true)
	styleRowDateBase = styleRowBase.Foreground(colorDim)
	styleRowDateSelected = styleRowSelected.Foreground(colorDim)
	styleRowPathBase = styleRowBase.Foreground(colorBlue)
	styleRowPathSelected = styleRowSelected.Foreground(colorBlue)
}
