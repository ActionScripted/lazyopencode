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

	if m.mode == ModeWorkspaces {
		return m.renderWorkspacesView(w, h)
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

	// Compute the path column width from the visible session set so the column
	// stays as tight as possible and reflows correctly when the search filter
	// changes.
	const maxPathW = 30
	pathColW := 0
	for _, s := range m.filtered {
		if n := lipgloss.Width(s.ShortDirectory()); n > pathColW {
			pathColW = n
		}
	}
	if pathColW > maxPathW {
		pathColW = maxPathW
	}

	for i := visibleStart; i < len(m.filtered) && i < visibleStart+height; i++ {
		sb.WriteString(formatRow(m.filtered[i], width, pathColW, i == m.cursor) + "\n")
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
	appName := styleDim.Render(" Lazy") + styleDim.Copy().Bold(true).Render("OpenCode")

	var hints string
	switch m.mode {
	case ModeSearch:
		hints = "  enter/esc: back   type to filter"
	case ModeWorkspaces:
		hints = "  j/k: navigate   tab: sessions view   q: quit"
	default:
		hints = "  j/k: navigate   /: search   tab: workspaces   q: quit"
	}
	left := appName + styleDim.Render(hints)

	var badge string
	switch m.mode {
	case ModeSearch:
		badge = styleModeSearch.Render("SEARCH")
	case ModeWorkspaces:
		badge = styleModeWorkspaces.Render("WORKSPACES")
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
//
// Layout (all widths in terminal columns):
//
//	" " date(16) "  " title(titleW) "  " path(pathColW) " " trail
//
// pathColW is computed by the caller from the live session list so the column
// is always as tight as possible and reflows when the search filter changes.
//
// selected controls whether the selection background is applied. It must be
// passed here — not applied by the caller over the assembled string — because
// termenv wraps every styled span with \e[0m (full reset), which would kill a
// background applied by an outer Render call. Instead every span is styled
// independently with the same background so no reset can break the highlight.
//
// All width arithmetic is performed on plain text before any style is applied.
// Adding a new decoration to any cell later only requires touching that cell's
// style call — the numbers never need to change.
func formatRow(s Session, width, pathColW int, selected bool) string {
	const (
		leadSp  = 1  // single leading space
		dateW   = 16 // "2006-01-02 15:04" is always exactly 16 columns
		dateGap = 2  // spaces between date and title
		minGap  = 2  // minimum spaces between title and path
		trailSp = 1  // single trailing space
	)

	titleW := width - leadSp - dateW - dateGap - minGap - pathColW - trailSp
	if titleW < 1 {
		titleW = 1
	}

	// ── plain-text content, truncated to column widths ────────────────────────
	date := s.UpdatedAt.Format("2006-01-02 15:04") // always exactly dateW columns
	rawTitle := truncate(s.Title, titleW)
	rawPath := truncate(s.ShortDirectory(), pathColW)

	// ── pad to exact column widths (plain text, no ANSI yet) ──────────────────
	// Title: left-aligned, space-padded on the right so the path column is pinned.
	paddedTitle := rawTitle + strings.Repeat(" ", titleW-lipgloss.Width(rawTitle))
	// Path: right-aligned, space-padded on the left.
	paddedPath := strings.Repeat(" ", pathColW-lipgloss.Width(rawPath)) + rawPath
	// Trailing fill: keeps the background unbroken to the edge of the list pane.
	assembled := leadSp + dateW + dateGap + titleW + minGap + pathColW + trailSp
	trailFill := strings.Repeat(" ", max(0, width-assembled))

	// ── per-cell styles ───────────────────────────────────────────────────────
	// Every span is styled independently and carries the background color (when
	// selected). This is the key invariant: because termenv emits \e[0m after
	// each span, an outer background Render would be killed by the first inner
	// reset. By giving every span its own background we guarantee the highlight
	// is unbroken across the full row width regardless of what other attributes
	// (bold, color) individual cells carry.
	base := lipgloss.NewStyle().Foreground(colorFg)
	if selected {
		base = base.Background(colorSelected)
	}

	styledLead := base.Render(strings.Repeat(" ", leadSp))
	styledDate := base.Foreground(colorDim).Render(date + strings.Repeat(" ", dateGap))
	styledTitle := base.Foreground(colorTitle).Bold(true).Render(paddedTitle + strings.Repeat(" ", minGap))
	styledPath := base.Foreground(colorBlue).Render(paddedPath)
	styledTrail := base.Render(strings.Repeat(" ", trailSp) + trailFill)

	return styledLead + styledDate + styledTitle + styledPath + styledTrail
}

// truncate clips s to at most maxW terminal columns. If the string is longer
// it is cut and a single "…" character (1 column) is appended.
//
// IMPORTANT: only call this on plain text — never on a string that already
// contains ANSI escape sequences. Width measurements on pre-styled strings
// must use lipgloss.Width(), not this function.
func truncate(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxW {
		return s
	}
	// Reserve 1 column for the ellipsis character.
	var (
		out  []rune
		used = 1 // 1 column reserved for "…"
	)
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if used+rw > maxW {
			break
		}
		out = append(out, r)
		used += rw
	}
	return string(out) + "…"
}

// ── Workspaces view ───────────────────────────────────────────────────────────

// renderWorkspacesView is the top-level render for ModeWorkspaces.
// Left pane: navigable list of unique workspace directories.
// Right pane: read-only list of sessions belonging to the selected workspace.
func (m model) renderWorkspacesView(w, h int) string {
	// hint bar: 2 rows (border + text); pane body gets the rest.
	bodyH := h - 2

	rightW := w * 55 / 100
	if rightW < 30 {
		rightW = 30
	}
	leftW := w - rightW

	left := m.renderWorkspaceList(leftW, bodyH)
	right := m.renderWorkspaceSessions(rightW, bodyH)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	hint := m.renderHint(w)

	return body + "\n" + hint
}

// renderWorkspaceList renders the left pane: a scrollable list of workspace
// directories, with the selected row highlighted.
func (m model) renderWorkspaceList(width, height int) string {
	var sb strings.Builder

	// Header separator (mirrors the search bar separator in the sessions view).
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", width-1)) + "\n")
	height-- // consumed by separator

	if len(m.workspaces) == 0 {
		sb.WriteString(styleDim.Render("  no workspaces") + "\n")
		return sb.String()
	}

	visibleStart := 0
	if m.workspaceCursor >= height {
		visibleStart = m.workspaceCursor - height + 1
	}

	for i := visibleStart; i < len(m.workspaces) && i < visibleStart+height; i++ {
		ws := m.workspaces[i]
		selected := i == m.workspaceCursor
		sb.WriteString(formatWorkspaceRow(ws, width, selected) + "\n")
	}

	return sb.String()
}

// formatWorkspaceRow renders a single workspace directory as a fixed-width row.
// The directory is displayed with "~" substitution and truncated to fit.
func formatWorkspaceRow(dir string, width int, selected bool) string {
	const leadSp = 1
	const trailSp = 1

	// Substitute home directory with "~" for display.
	display := displayDir(dir)

	innerW := width - leadSp - trailSp
	if innerW < 1 {
		innerW = 1
	}
	raw := truncate(display, innerW)
	padded := raw + strings.Repeat(" ", max(0, innerW-lipgloss.Width(raw)))

	base := lipgloss.NewStyle().Foreground(colorAccent)
	if selected {
		base = base.Background(colorSelected)
	}

	lead := base.Copy().Foreground(colorFg).Render(strings.Repeat(" ", leadSp))
	text := base.Bold(selected).Render(padded)
	trail := base.Copy().Foreground(colorFg).Render(strings.Repeat(" ", trailSp))

	return lead + text + trail
}

// renderWorkspaceSessions renders the right pane: a read-only list of sessions
// for the currently selected workspace.
func (m model) renderWorkspaceSessions(width, height int) string {
	inner := width - 4 // account for border + padding

	if len(m.workspaces) == 0 {
		return stylePreviewPane.Width(width - 2).Height(height).Render(
			styleDim.Render("no workspaces"),
		)
	}

	selectedWS := m.workspaces[m.workspaceCursor]

	// Filter sessions belonging to the selected workspace.
	var wsSessions []Session
	for _, s := range m.sessions {
		if s.Directory == selectedWS {
			wsSessions = append(wsSessions, s)
		}
	}

	// ── header ────────────────────────────────────────────────────────────────
	var header strings.Builder
	displayPath := displayDir(selectedWS)
	header.WriteString(stylePreviewTitle.Render(
		lipgloss.NewStyle().MaxWidth(inner).Render(displayPath),
	))
	header.WriteString("\n")
	header.WriteString(styleDim.Render(
		lipgloss.NewStyle().MaxWidth(inner).Render(
			strings.Repeat("─", max(0, inner)),
		),
	))

	// header: path(1) + separator(1) = 2 lines
	const headerLines = 2
	listHeight := height - headerLines
	if listHeight < 1 {
		listHeight = 1
	}

	// ── session rows ──────────────────────────────────────────────────────────
	var sessionRows strings.Builder
	if len(wsSessions) == 0 {
		sessionRows.WriteString("\n" + styleDim.Render("  no sessions"))
	} else {
		for i, s := range wsSessions {
			if i >= listHeight {
				break
			}
			sessionRows.WriteString("\n" + formatWorkspaceSessionRow(s, inner))
		}
	}

	return stylePreviewPane.Width(width - 2).Height(height).Render(
		header.String() + sessionRows.String(),
	)
}

// formatWorkspaceSessionRow renders a compact session row for the workspaces
// right pane. Layout: date(16) "  " title(remaining). No path column needed
// since all sessions share the same workspace.
func formatWorkspaceSessionRow(s Session, width int) string {
	const dateW = 16
	const gap = 2
	titleW := width - dateW - gap
	if titleW < 1 {
		titleW = 1
	}

	date := styleDim.Render(s.UpdatedAt.Format("2006-01-02 15:04"))
	title := lipgloss.NewStyle().Foreground(colorTitle).Bold(true).
		Render(truncate(s.Title, titleW))

	return date + strings.Repeat(" ", gap) + title
}
