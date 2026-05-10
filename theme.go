package main

import "github.com/charmbracelet/lipgloss"

// themePalette holds the semantic color slots for a single theme.
// Every slot is an AdaptiveColor so the theme responds correctly to
// both dark and light terminal backgrounds.
type themePalette struct {
	fg         lipgloss.AdaptiveColor // primary text
	bright     lipgloss.AdaptiveColor // bold/bright text (titles, selected)
	dim        lipgloss.AdaptiveColor // muted/secondary text (comments, hints)
	cyan       lipgloss.AdaptiveColor // accent — preview pane, search prefix
	blue       lipgloss.AdaptiveColor // primary — list path + NORMAL badge
	onBadge    lipgloss.AdaptiveColor // text rendered on top of colored badges
	bgDark     lipgloss.AdaptiveColor // super-dark background (list, stats body)
	bgPanel    lipgloss.AdaptiveColor // panel background (topbar, preview, hint bar)
	bgPanelAlt lipgloss.AdaptiveColor // alternate panel background (zebra rows)
	selected   lipgloss.AdaptiveColor // selected row background
	border     lipgloss.AdaptiveColor // border and separator lines
	yellow     lipgloss.AdaptiveColor // warnings, SEARCH/GOTO badge
	green      lipgloss.AdaptiveColor // additions, assistant label, YANK badge
	danger     lipgloss.AdaptiveColor // errors, deletions, DELETE badge
	purple     lipgloss.AdaptiveColor // STATS badge, accent
	heading    lipgloss.AdaptiveColor // stats section/fieldset title headings
}

// paletteTokyonight is the Tokyo Night theme — the original colour scheme.
var paletteTokyonight = themePalette{
	fg:         lipgloss.AdaptiveColor{Light: "#343b58", Dark: "#a9b1d6"},
	bright:     lipgloss.AdaptiveColor{Light: "#343b58", Dark: "#c0caf5"},
	dim:        lipgloss.AdaptiveColor{Light: "#68608a", Dark: "#565f89"},
	cyan:       lipgloss.AdaptiveColor{Light: "#0f4b6e", Dark: "#7dcfff"},
	blue:       lipgloss.AdaptiveColor{Light: "#3d59a1", Dark: "#7aa2f7"},
	onBadge:    lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#1a1b26"},
	bgDark:     lipgloss.AdaptiveColor{Light: "#cbccd1", Dark: "#16161e"},
	bgPanel:    lipgloss.AdaptiveColor{Light: "#d5d6db", Dark: "#1e2030"},
	bgPanelAlt: lipgloss.AdaptiveColor{Light: "#c8c9cf", Dark: "#222436"},
	selected:   lipgloss.AdaptiveColor{Light: "#b8cce4", Dark: "#283457"},
	border:     lipgloss.AdaptiveColor{Light: "#a8aecb", Dark: "#3b4261"},
	yellow:     lipgloss.AdaptiveColor{Light: "#8f5e15", Dark: "#e0af68"},
	green:      lipgloss.AdaptiveColor{Light: "#33635c", Dark: "#9ece6a"},
	danger:     lipgloss.AdaptiveColor{Light: "#c53b53", Dark: "#f7768e"},
	purple:     lipgloss.AdaptiveColor{Light: "#5a4a78", Dark: "#bb9af7"},
	heading:    lipgloss.AdaptiveColor{Light: "#5a4a78", Dark: "#bb9af7"},
}

// paletteOpencode is the opencode brand theme — warm orange primary on dark,
// blue primary on light. This is the default theme.
var paletteOpencode = themePalette{
	fg:         lipgloss.AdaptiveColor{Light: "#2a2a2a", Dark: "#e0e0e0"},
	bright:     lipgloss.AdaptiveColor{Light: "#2a2a2a", Dark: "#e0e0e0"},
	dim:        lipgloss.AdaptiveColor{Light: "#8a8a8a", Dark: "#6a6a6a"},
	cyan:       lipgloss.AdaptiveColor{Light: "#318795", Dark: "#56b6c2"},
	blue:       lipgloss.AdaptiveColor{Light: "#3b7dd8", Dark: "#fab283"}, // primary — orange dark / blue light
	onBadge:    lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#0a0a0a"},
	bgDark:     lipgloss.AdaptiveColor{Light: "#ebebeb", Dark: "#121212"},
	bgPanel:    lipgloss.AdaptiveColor{Light: "#f8f8f8", Dark: "#212121"},
	bgPanelAlt: lipgloss.AdaptiveColor{Light: "#f0f0f0", Dark: "#252525"},
	selected:   lipgloss.AdaptiveColor{Light: "#e5e5e6", Dark: "#303030"},
	border:     lipgloss.AdaptiveColor{Light: "#d3d3d3", Dark: "#4b4c5c"},
	yellow:     lipgloss.AdaptiveColor{Light: "#d68c27", Dark: "#f5a742"},
	green:      lipgloss.AdaptiveColor{Light: "#3d9a57", Dark: "#7fd88f"},
	danger:     lipgloss.AdaptiveColor{Light: "#d1383d", Dark: "#e06c75"},
	purple:     lipgloss.AdaptiveColor{Light: "#5a4a78", Dark: "#9d7cd8"},
	heading:    lipgloss.AdaptiveColor{Light: "#7b5bb6", Dark: "#5c9cf5"}, // opencode secondary — blue dark / violet light
}

// paletteCatppuccin is the Catppuccin theme — Mocha (dark) / Latte (light).
var paletteCatppuccin = themePalette{
	fg:         lipgloss.AdaptiveColor{Light: "#4c4f69", Dark: "#cdd6f4"}, // Latte Text / Mocha Text
	bright:     lipgloss.AdaptiveColor{Light: "#4c4f69", Dark: "#cdd6f4"},
	dim:        lipgloss.AdaptiveColor{Light: "#8c8fa1", Dark: "#6c7086"}, // Latte Overlay1 / Mocha Overlay0
	cyan:       lipgloss.AdaptiveColor{Light: "#179299", Dark: "#89dceb"}, // Latte Teal / Mocha Sky
	blue:       lipgloss.AdaptiveColor{Light: "#1e66f5", Dark: "#89b4fa"}, // Latte Blue / Mocha Blue
	onBadge:    lipgloss.AdaptiveColor{Light: "#eff1f5", Dark: "#1e1e2e"}, // Latte Base / Mocha Base
	bgDark:     lipgloss.AdaptiveColor{Light: "#dce0e8", Dark: "#181825"}, // Latte Mantle / Mocha Mantle
	bgPanel:    lipgloss.AdaptiveColor{Light: "#e6e9ef", Dark: "#1e1e2e"}, // Latte Base / Mocha Base
	bgPanelAlt: lipgloss.AdaptiveColor{Light: "#eff1f5", Dark: "#313244"}, // Latte Crust / Mocha Surface0
	selected:   lipgloss.AdaptiveColor{Light: "#bcc0cc", Dark: "#313244"}, // Latte Overlay0 / Mocha Surface0
	border:     lipgloss.AdaptiveColor{Light: "#9ca0b0", Dark: "#45475a"}, // Latte Overlay2 / Mocha Surface1
	yellow:     lipgloss.AdaptiveColor{Light: "#df8e1d", Dark: "#f9e2af"}, // Latte Yellow / Mocha Yellow
	green:      lipgloss.AdaptiveColor{Light: "#40a02b", Dark: "#a6e3a1"}, // Latte Green / Mocha Green
	danger:     lipgloss.AdaptiveColor{Light: "#d20f39", Dark: "#f38ba8"}, // Latte Red / Mocha Red
	purple:     lipgloss.AdaptiveColor{Light: "#8839ef", Dark: "#cba6f7"}, // Latte Mauve / Mocha Mauve
	heading:    lipgloss.AdaptiveColor{Light: "#7287fd", Dark: "#b4befe"}, // Latte Lavender / Mocha Lavender
}

// themes maps theme name strings to their palettes.
var themes = map[string]themePalette{
	"opencode":   paletteOpencode,
	"tokyonight": paletteTokyonight,
	"catppuccin": paletteCatppuccin,
}

// themeOrder is the stable cycle order for the theme-toggle key ("t").
var themeOrder = []string{"opencode", "tokyonight", "catppuccin"}

// activeThemeName is the name of the currently active theme.
var activeThemeName = "opencode"

// activePalette holds the color values for the currently active theme.
// initStyles() reads this to populate every color and style var in styles.go.
var activePalette = paletteOpencode
