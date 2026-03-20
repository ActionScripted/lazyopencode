package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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

	if m.mode == ModeConfirmDeleteWorkspace {
		base := m.renderWorkspacesView(w, h)
		return overlayModal(base, m.renderWorkspaceModal(), w, h)
	}

	previewW := w * 45 / 100
	if previewW < 30 {
		previewW = 30
	}
	listW := w - previewW

	// hint bar occupies 2 rows (border + text)
	// search + separator occupy 2 rows
	list := styleListPane.Width(listW).Height(h - 2).Render(m.renderList(listW, h-4))
	preview := m.renderPreview(previewW, h-2)

	body := lipgloss.JoinHorizontal(lipgloss.Top, list, preview)
	hint := m.renderHint(w)

	base := body + "\n" + hint

	if m.mode == ModeConfirmDelete {
		return overlayModal(base, m.renderSessionModal(), w, h)
	}

	if m.mode == ModeYank {
		return overlayModal(base, m.renderYankModal(), w, h)
	}

	return base
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
	const labelW = 9 // "messages " — widest label + 1 space

	metaLabel := func(l string) string {
		return styleDim.Render(l + strings.Repeat(" ", labelW-len(l)))
	}

	var header strings.Builder

	// title
	header.WriteString(stylePreviewTitle.Render(
		lipgloss.NewStyle().MaxWidth(inner).Render(s.Title),
	))
	header.WriteString("\n\n")

	// path
	header.WriteString(metaLabel("path"))
	header.WriteString(styleAccent.Render(
		lipgloss.NewStyle().MaxWidth(inner - labelW).Render(s.DisplayDirectory()),
	))
	header.WriteString("\n")

	// created
	header.WriteString(metaLabel("created"))
	header.WriteString(s.CreatedAt.Format("2006-01-02 15:04"))
	header.WriteString("\n")

	// updated
	header.WriteString(metaLabel("updated"))
	header.WriteString(s.UpdatedAt.Format("2006-01-02 15:04"))
	header.WriteString("\n")

	// duration
	header.WriteString(metaLabel("duration"))
	header.WriteString(formatDuration(s.UpdatedAt.Sub(s.CreatedAt)))
	header.WriteString("\n")

	// stats (async — show loading until statsLoadedMsg arrives)
	if m.stats == nil {
		header.WriteString(styleDim.Render("  loading stats…"))
		header.WriteString("\n")
	} else {
		st := m.stats

		// messages
		header.WriteString(metaLabel("messages"))
		fmt.Fprintf(&header, "%d", st.MsgCount)
		header.WriteString("\n")

		// context + output tokens (only when AI turns exist)
		if st.ContextTokens > 0 {
			header.WriteString(metaLabel("context"))
			header.WriteString(formatTokens(st.ContextTokens))
			header.WriteString("\n")

			header.WriteString(metaLabel("output"))
			header.WriteString(formatTokens(st.OutputTokens))
			header.WriteString("\n")
		}

		// changes (only when files were modified)
		if s.SummaryFiles > 0 {
			header.WriteString(metaLabel("changes"))
			fmt.Fprintf(&header, "%d files (", s.SummaryFiles)
			header.WriteString(styleAdd.Render(fmt.Sprintf("+%d", s.SummaryAdditions)))
			header.WriteString(" ")
			header.WriteString(styleDel.Render(fmt.Sprintf("-%d", s.SummaryDeletions)))
			header.WriteString(")")
			header.WriteString("\n")
		}
	}

	// separator
	header.WriteString("\n")
	header.WriteString(styleDim.Render("─── messages " + strings.Repeat("─", max(0, inner-13))))

	// ── messages ──────────────────────────────────────────────────────────────
	// Compute header height dynamically so it never drifts out of sync.
	headerLines := strings.Count(header.String(), "\n") + 1
	msgHeight := height - headerLines
	if msgHeight < 1 {
		msgHeight = 1
	}

	var msgSection string
	switch {
	case m.messages == nil:
		msgSection = "\n" + styleDim.Render("  loading…")

	case len(m.messages) == 0:
		msgSection = "\n" + styleDim.Render("  no messages")

	default:
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

		used := 0
		first := len(blocks)
		for i := len(blocks) - 1; i >= 0; i-- {
			cost := strings.Count(blocks[i], "\n") + 1 + 1
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

// formatDuration formats a duration for display in the preview header.
// Sub-minute durations show as "< 1m"; otherwise "Xh Ym" or "Ym".
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Minute {
		return "< 1m"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// formatTokens formats a token count with K/M suffix.
func formatTokens(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func (m model) renderHint(width int) string {
	appName := styleDim.Render(" Lazy") + styleDim.Bold(true).Render("OpenCode")

	var hints string
	switch m.mode {
	case ModeSearch:
		hints = "  enter/esc: back   type to filter"
	case ModeWorkspaces:
		hints = "  j/k: navigate   w: sessions view   q: quit"
	case ModeConfirmDelete, ModeConfirmDeleteWorkspace:
		hints = "  Y/d: confirm delete   n/esc: cancel"
	case ModeYank:
		hints = "  d: yank directory   s: yank session id   esc: cancel"
	default:
		hints = "  j/k: navigate   enter: open   /: search   y: yank   g: workspace   dd: delete   w: workspaces   q: quit"
	}
	left := appName + styleDim.Render(hints)

	var badge string
	switch m.mode {
	case ModeSearch:
		badge = styleModeSearch.Render("SEARCH")
	case ModeWorkspaces:
		badge = styleModeWorkspaces.Render("WORKSPACES")
	case ModeConfirmDelete, ModeConfirmDeleteWorkspace:
		badge = styleModeConfirmDelete.Render("DELETE?")
	case ModeYank:
		badge = styleModeYank.Render("YANK")
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

	// Responsive column hiding: hide date below 100 cols, hide path below 77.
	showDate := width >= 100
	showPath := width >= 77

	effectiveDateW := dateW
	effectiveDateGap := dateGap
	effectiveMinGap := minGap
	effectivePathW := pathColW
	if !showDate {
		effectiveDateW = 0
		effectiveDateGap = 0
	}
	if !showPath {
		effectiveMinGap = 0
		effectivePathW = 0
	}

	titleW := width - leadSp - effectiveDateW - effectiveDateGap - effectiveMinGap - effectivePathW - trailSp
	if titleW < 1 {
		titleW = 1
	}

	// ── plain-text content, truncated to column widths ────────────────────────
	date := s.UpdatedAt.Format("2006-01-02 15:04") // always exactly dateW columns
	rawTitle := truncate(s.Title, titleW)
	rawPath := truncate(s.ShortDirectory(), effectivePathW)

	// ── pad to exact column widths (plain text, no ANSI yet) ──────────────────
	// Title: left-aligned, space-padded on the right so the path column is pinned.
	paddedTitle := rawTitle + strings.Repeat(" ", titleW-lipgloss.Width(rawTitle))
	// Path: right-aligned, space-padded on the left.
	paddedPath := strings.Repeat(" ", effectivePathW-lipgloss.Width(rawPath)) + rawPath
	// Trailing fill: keeps the background unbroken to the edge of the list pane.
	assembled := leadSp + effectiveDateW + effectiveDateGap + titleW + effectiveMinGap + effectivePathW + trailSp
	trailFill := strings.Repeat(" ", max(0, width-assembled))

	// ── per-cell styles ───────────────────────────────────────────────────────
	// Every span is styled independently and carries the background color (when
	// selected). This is the key invariant: because termenv emits \e[0m after
	// each span, an outer background Render would be killed by the first inner
	// reset. By giving every span its own background we guarantee the highlight
	// is unbroken across the full row width regardless of what other attributes
	// (bold, color) individual cells carry.
	base := lipgloss.NewStyle().Foreground(colorFg).Background(colorBgPanel)
	if selected {
		base = base.Background(colorSelected)
	}

	styledLead := base.Render(strings.Repeat(" ", leadSp))
	styledTitle := base.Foreground(colorTitle).Bold(true).Render(paddedTitle)
	styledTrail := base.Render(strings.Repeat(" ", trailSp) + trailFill)

	row := styledLead
	if showDate {
		row += base.Foreground(colorDim).Render(date + strings.Repeat(" ", dateGap))
	}
	row += styledTitle
	if showPath {
		row += base.Foreground(colorBlue).Render(strings.Repeat(" ", effectiveMinGap) + paddedPath)
	}
	row += styledTrail

	return row
}

// truncate clips s to at most maxW terminal columns, appending "…" if the
// string is longer. Delegates to ansi.Truncate which is ANSI-safe and
// grapheme-cluster-aware.
func truncate(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	return ansi.Truncate(s, maxW, "…")
}

// overlayModal centers a pre-rendered modal string over the base view.
// The base is split by newlines; the modal rows are spliced over the middle
// lines so the background content remains partially visible around it.
func overlayModal(base, rendered string, w, h int) string {
	modalW := lipgloss.Width(rendered)
	modalH := lipgloss.Height(rendered)

	baseLines := strings.Split(base, "\n")
	for len(baseLines) < h {
		baseLines = append(baseLines, "")
	}

	startRow := (h - modalH) / 2
	if startRow < 0 {
		startRow = 0
	}
	startCol := (w - modalW) / 2
	if startCol < 0 {
		startCol = 0
	}

	modalLines := strings.Split(rendered, "\n")

	for i, ml := range modalLines {
		row := startRow + i
		if row >= len(baseLines) {
			break
		}

		// Pad the base line to full terminal width (plain spaces; ANSI already
		// in there so we can't use lipgloss.Width reliably — we just ensure
		// enough trailing spaces).
		bl := baseLines[row]
		blW := ansi.StringWidth(bl)
		if blW < w {
			bl += strings.Repeat(" ", w-blW)
			blW = w
		}

		// Splice: keep startCol columns from the left, then the modal line,
		// then the remainder — all ANSI-safe via ansi.Cut.
		left := ansi.Cut(bl, 0, startCol)
		right := ansi.Cut(bl, startCol+modalW, blW)
		baseLines[row] = left + ml + right
	}

	return strings.Join(baseLines, "\n")
}

// renderYankModal returns the styled modal for the yank-to-clipboard prompt.
func (m model) renderYankModal() string {
	dKey := styleModalKeyYank.Render("d")
	sKey := styleModalKeyYank.Render("s")

	content := styleModalYankTitle.Render("Yank to clipboard") + "\n\n" +
		dKey + styleDim.Render("  directory") + "\n" +
		sKey + styleDim.Render("  session id")

	return styleModalYank.Render(content)
}

// renderSessionModal returns the styled modal for a single-session delete.
func (m model) renderSessionModal() string {
	const modalInnerW = 46

	sessionTitle := m.pendingDeleteID
	for _, s := range m.sessions {
		if s.ID == m.pendingDeleteID {
			sessionTitle = s.Title
			break
		}
	}

	confirm := styleModalKey.Render("Yes [y/d]") + "   " +
		styleModalKeyCancel.Render("No [n]")

	content := styleModalTitle.Render("Delete session?") + "\n\n" +
		stylePreviewTitle.Render(truncate(sessionTitle, modalInnerW)) + "\n\n" +
		confirm

	return styleModal.Render(content)
}

// renderWorkspaceModal returns the styled modal for a workspace delete.
func (m model) renderWorkspaceModal() string {
	const modalInnerW = 46

	count := 0
	for _, s := range m.sessions {
		if s.Directory == m.pendingDeleteWorkspace {
			count++
		}
	}

	noun := "sessions"
	if count == 1 {
		noun = "session"
	}
	countLine := stylePreviewTitle.Render(fmt.Sprintf("%d", count)) +
		styleDim.Render(" "+noun+" will be deleted")

	confirm := styleModalKey.Render("Yes [y/d]") + "   " +
		styleModalKeyCancel.Render("No [n]")

	content := styleModalTitle.Render("Delete workspace?") + "\n\n" +
		stylePreviewTitle.Render(truncate(displayDir(m.pendingDeleteWorkspace), modalInnerW)) + "\n" +
		countLine + "\n\n" +
		confirm

	return styleModal.Render(content)
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

	left := styleListPane.Width(leftW).Height(bodyH).Render(m.renderWorkspaceList(leftW, bodyH))
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
		sb.WriteString(formatWorkspaceRow(ws.Dir, ws.DisplayDir, width, selected) + "\n")
	}

	return sb.String()
}

// formatWorkspaceRow renders a single workspace directory as a fixed-width row.
// displayDir is the pre-computed display string (with "~" substitution).
func formatWorkspaceRow(dir, displayDir string, width int, selected bool) string {
	const leadSp = 1
	const trailSp = 1

	_ = dir // raw dir available if needed in future; display uses displayDir

	innerW := width - leadSp - trailSp
	if innerW < 1 {
		innerW = 1
	}
	raw := truncate(displayDir, innerW)
	padded := raw + strings.Repeat(" ", max(0, innerW-lipgloss.Width(raw)))

	base := lipgloss.NewStyle().Foreground(colorAccent).Background(colorBgPanel)
	if selected {
		base = base.Background(colorSelected)
	}

	lead := base.Foreground(colorFg).Render(strings.Repeat(" ", leadSp))
	text := base.Bold(selected).Render(padded)
	trail := base.Foreground(colorFg).Render(strings.Repeat(" ", trailSp))

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
		if s.Directory == selectedWS.Dir {
			wsSessions = append(wsSessions, s)
		}
	}

	// ── header ────────────────────────────────────────────────────────────────
	var header strings.Builder
	header.WriteString(stylePreviewTitle.Render(
		lipgloss.NewStyle().MaxWidth(inner).Render(selectedWS.DisplayDir),
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
