package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	w := m.width
	if w == 0 {
		w = 80
	}
	h := m.height
	if h == 0 {
		h = 24
	}

	previewW := w * 45 / 100
	if previewW < 30 {
		previewW = 30
	}
	listW := w - previewW

	// hint bar occupies 2 rows (border + text)
	// search + separator occupy 2 rows
	list := m.renderList(listW, h-4)
	preview := m.renderPreview(previewW, h-2)

	body := lipgloss.JoinHorizontal(lipgloss.Top, list, preview)
	hint := m.renderHint(w)

	return body + "\n" + hint
}

func (m model) renderList(width, height int) string {
	var sb strings.Builder

	// search prefix — accent in search mode, dim in normal mode
	prefix := styleSeparator.Render("> ")
	if m.mode == ModeSearch {
		prefix = styleSearchPrefix.Render("> ")
	}
	sb.WriteString(prefix + m.search.View() + "\n")
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", width-1)) + "\n")

	if m.err != nil {
		sb.WriteString(styleDim.Render("  error: "+m.err.Error()) + "\n")
		return sb.String()
	}

	if len(m.filtered) == 0 && len(m.sessions) == 0 {
		sb.WriteString(styleDim.Render("  no sessions found") + "\n")
		return sb.String()
	}

	if len(m.filtered) == 0 {
		sb.WriteString(styleDim.Render("  no matches") + "\n")
		return sb.String()
	}

	visibleStart := 0
	if m.cursor >= height {
		visibleStart = m.cursor - height + 1
	}

	for i := visibleStart; i < len(m.filtered) && i < visibleStart+height; i++ {
		row := formatRow(m.filtered[i], width)
		pad := width - lipgloss.Width(row)
		if pad < 0 {
			pad = 0
		}
		row += strings.Repeat(" ", pad)
		if i == m.cursor {
			sb.WriteString(styleSelectedRow.Render(row) + "\n")
		} else {
			sb.WriteString(styleRow.Render(row) + "\n")
		}
	}

	return sb.String()
}

func (m model) renderPreview(width, height int) string {
	inner := width - 4 // account for border + padding

	if len(m.filtered) == 0 {
		return stylePreviewPane.Width(width - 2).Height(height).Render(
			styleDim.Render("no selection"),
		)
	}

	s := m.filtered[m.cursor]

	// ── header ────────────────────────────────────────────────────────────────
	var header strings.Builder
	header.WriteString(stylePreviewTitle.Render(
		lipgloss.NewStyle().MaxWidth(inner).Render(s.Title),
	))
	header.WriteString("\n\n")
	header.WriteString(styleDim.Render("path     ") + styleAccent.Render(
		lipgloss.NewStyle().MaxWidth(inner).Render(s.DisplayDirectory()),
	))
	header.WriteString("\n")
	header.WriteString(styleDim.Render("updated  ") + s.UpdatedAt.Format("2006-01-02 15:04"))
	header.WriteString("\n\n")
	separator := styleDim.Render("─── messages " + strings.Repeat("─", max(0, inner-13)))
	header.WriteString(separator)

	// header occupies: title(1) + blank(1) + dir(1) + updated(1) + blank(1) + separator(1) = 6 lines
	const headerLines = 6
	msgHeight := height - headerLines
	if msgHeight < 1 {
		msgHeight = 1
	}

	// ── messages ──────────────────────────────────────────────────────────────
	var msgSection string
	switch {
	case m.messages == nil:
		msgSection = "\n" + styleDim.Render("  loading…")

	case len(m.messages) == 0:
		msgSection = "\n" + styleDim.Render("  no messages")

	default:
		// Render each message into a block, then fill from the bottom up.
		blocks := make([]string, len(m.messages))
		for i, msg := range m.messages {
			var label string
			if msg.Role == "user" {
				label = styleRoleUser.Render("[user]")
			} else {
				label = styleRoleAssistant.Render("[asst]")
			}
			wrapped := lipgloss.NewStyle().Width(inner).Render(msg.Text)
			blocks[i] = label + "\n" + wrapped
		}

		// Walk backwards, accumulate blocks that fit.
		// Each block costs: lines(block) + 1 blank separator.
		used := 0
		first := len(blocks)
		for i := len(blocks) - 1; i >= 0; i-- {
			cost := strings.Count(blocks[i], "\n") + 1 + 1 // block lines + blank
			if used+cost > msgHeight {
				break
			}
			used += cost
			first = i
		}

		var sb strings.Builder
		for i := first; i < len(blocks); i++ {
			sb.WriteString("\n")
			sb.WriteString(blocks[i])
			sb.WriteString("\n")
		}
		msgSection = sb.String()
	}

	return stylePreviewPane.Width(width - 2).Height(height).Render(header.String() + msgSection)
}

func (m model) renderHint(width int) string {
	appName := styleDim.Render("Lazy") + styleDim.Copy().Bold(true).Render("OpenCode")

	var hints string
	switch m.mode {
	case ModeSearch:
		hints = "  enter/esc: back   type to filter"
	default:
		hints = "  j/k: navigate   /: search   q: quit"
	}
	left := appName + styleDim.Render(hints)

	var badge string
	switch m.mode {
	case ModeSearch:
		badge = styleModeSearch.Render("SEARCH")
	default:
		badge = styleModeNormal.Render("NORMAL")
	}

	space := width - 1 - lipgloss.Width(left) - lipgloss.Width(badge)
	if space < 1 {
		space = 1
	}

	return styleHint.Width(width - 1).Render(left + strings.Repeat(" ", space) + badge)
}

// formatRow renders a single session as a fixed-layout row.
// Layout: " " date " " title " " dir " "
// Columns: date=16, dir=20, title=whatever remains. No Bold, no Width() calls.
func formatRow(s Session, width int) string {
	const dateW = 16 // "2006-01-02 15:04"
	const dirW = 20
	const spacing = 4 // 1 leading + 3 separators between columns

	titleW := width - dateW - dirW - spacing
	if titleW < 1 {
		titleW = 1
	}

	date := s.UpdatedAt.Format("2006-01-02 15:04")
	title := lipgloss.NewStyle().MaxWidth(titleW).Render(s.Title)
	dir := lipgloss.NewStyle().MaxWidth(dirW).Render(s.ShortDirectory())

	return " " + date + " " + title + " " + dir + " "
}
