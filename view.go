package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// ── Layout constants ──────────────────────────────────────────────────────────
//
// All sizing magic lives here. Touch these if you need to adjust proportions,
// breakpoints, or overhead accounting — nothing else should need to change.

const (
	// Terminal size fallbacks (classic VT100 dimensions).
	defaultTermW = 80
	defaultTermH = 24

	// Layout breakpoints (terminal width in columns).
	wideLayoutBreakpoint     = 120 // wide (side-by-side) vs. narrow (stacked)
	singleLineHintBreakpoint = 182 // hint bar: one row vs. two rows

	// Split percentages (integer arithmetic out of 100).
	previewWidthPct = 45 // preview pane share of terminal width in wide mode
	listHeightPct   = 40 // list pane share of body height in stacked mode

	// Minimum list pane height in stacked mode (prevents the list from
	// collapsing to nothing when the terminal is very short).
	listMinH = 6

	// Minimum body height required to show both list + preview in stacked
	// mode; below this threshold the preview is dropped entirely.
	minStackedBodyH = 24

	// Top border bar — a single row rendered at the very top of the viewport so
	// content is not clipped by window chrome (titlebars, tab bars, etc.).
	topBarH = 1

	// Hint bar.
	hintBarWidthAdj = 1 // intentional 1-col right margin for badge placement; borders still extend to full width

	// Rows consumed by the search input + separator drawn at the top of
	// renderList. Wide mode passes (paneH - listHeaderRows) as the content
	// height so those rows are not double-counted.
	listHeaderRows = 2

	// Pane border/padding overhead — derived from the lipgloss box model.
	//
	// stylePreviewPane (BorderLeft + PaddingLeft(1) + PaddingRight(1)):
	//   outer width  = Width() + 1   → paneHorizBorder = 1
	//   text area    = Width() - 2   → paneHorizPadding = 2
	//   outer height = Height()      → no vertical border cost
	//
	// stylePreviewPaneStacked (BorderTop + PaddingLeft(1) + PaddingRight(1)):
	//   outer width  = Width()       → no horizontal border cost
	//   text area    = Width() - 2   → same padding overhead
	//   outer height = Height() + 1  → paneVertBorderStacked = 1
	//
	// Callers always pass the desired outer height; renderPreview subtracts
	// paneVertBorderStacked internally for the stacked case.
	paneHorizBorder       = 1 // left border col cost (side-by-side mode)
	paneHorizPadding      = 2 // PaddingLeft(1)+PaddingRight(1) inner cost
	paneVertBorderStacked = 1 // top border row cost (stacked mode)

	// Path column caps in renderList.
	maxPathWidthWide   = 30
	maxPathWidthNarrow = 20

	// Modal inner content width (max chars for a title inside a delete modal).
	modalInnerWidth = 46

	// Minimum right-pane width in the workspaces view.
	workspacesRightMinWidth = 30

	// Date/time column widths (shared by formatSessionRow and formatWorkspaceSessionRow).
	dateFullWidth  = 16 // "2006-01-02 15:04" — always exactly 16 columns
	dateShortWidth = 6  // "Jan 02"            — always exactly 6 columns
	dateGap        = 2  // spaces between the date column and the title column
)

func (m model) View() string {
	w := m.width
	if w == 0 {
		w = defaultTermW
	}
	h := m.height
	if h == 0 {
		h = defaultTermH
	}

	if m.mode == modeWorkspaces {
		return m.renderWorkspacesView(w, h)
	}

	if m.mode == modeStats {
		return m.renderStatsView(w, h)
	}

	if m.mode == modeConfirmDeleteWorkspace {
		base := m.renderWorkspacesView(w, h)
		return overlayModal(base, m.renderWorkspaceModal(), w, h)
	}

	topBar := renderTopBar(w)
	h -= topBarH

	previewW := w * previewWidthPct / 100
	listW := w - previewW

	// Render the hint first to measure its height. Row counts are additive:
	// body rows + hint rows = h. The "\n" joining them is just a line terminator.
	hint := m.renderHint(w)
	hintH := strings.Count(hint, "\n") + 1

	var body string
	if w < wideLayoutBreakpoint {
		// Narrow: stack list on top, preview below.
		bodyH := h - hintH

		var listH, previewH int
		if bodyH < minStackedBodyH {
			// Too short for both panes — give the list the full height, drop preview.
			listH = bodyH
			previewH = 0
		} else {
			listH = bodyH * listHeightPct / 100
			if listH < listMinH {
				listH = listMinH
			}
			previewH = bodyH - listH
		}

		list := styleListPane.Width(w).Height(listH).Render(m.renderList(w, listH-listHeaderRows))
		if previewH == 0 {
			body = list
		} else {
			preview := m.renderPreview(w, previewH, true)
			body = list + "\n" + preview
		}
	} else {
		// Wide: side-by-side list (left) and preview (right).
		paneH := h - hintH
		// renderList draws listHeaderRows lines internally, so pass paneH minus
		// that overhead as the scrollable content height.
		list := styleListPane.Width(listW).Height(paneH).Render(m.renderList(listW, paneH-listHeaderRows))
		preview := m.renderPreview(previewW, paneH, false)
		body = lipgloss.JoinHorizontal(lipgloss.Top, list, preview)
	}

	base := topBar + "\n" + body + "\n" + hint

	if m.mode == modeError {
		return overlayModal(base, m.renderErrorModal(), w, h)
	}

	if m.mode == modeConfirmDelete {
		return overlayModal(base, m.renderSessionModal(), w, h)
	}

	if m.mode == modeYank {
		return overlayModal(base, m.renderYankModal(), w, h)
	}

	if m.mode == modeGoto {
		return overlayModal(base, m.renderGotoModal(), w, h)
	}

	return base
}

func (m model) renderList(width, height int) string {
	var sb strings.Builder

	// search prefix — accent in search mode, dim in normal mode
	prefix := styleSeparator.Render("> ")
	if m.mode == modeSearch {
		prefix = styleSearchPrefix.Render("> ")
	}
	sb.WriteString(prefix + m.search.View() + "\n")
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", width)) + "\n")

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

	// Compute the path column width from the cached value (pre-computed in
	// Update whenever m.filtered changes — see model.recomputePathColW).
	// Cap is tighter on narrower list panes (< wideLayoutBreakpoint cols).
	maxPathW := maxPathWidthWide
	if width < wideLayoutBreakpoint {
		maxPathW = maxPathWidthNarrow
	}
	pathColW := m.pathColW
	if pathColW > maxPathW {
		pathColW = maxPathW
	}

	for i := visibleStart; i < len(m.filtered) && i < visibleStart+height; i++ {
		sb.WriteString(formatSessionRow(m.filtered[i], width, pathColW, i == m.cursor) + "\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

func (m model) renderPreview(width, height int, stacked bool) string {
	// Callers pass the outer height they want (total rows including borders).
	// stacked mode has a top border (1 row) that eats into the content height;
	// side-by-side mode has only a left border (no vertical cost).
	paneStyle := stylePreviewPane
	paneW := width - paneHorizBorder // outer == width (left border adds 1 col)
	if stacked {
		paneStyle = stylePreviewPaneStacked
		paneW = width                   // top border adds no horizontal cost
		height -= paneVertBorderStacked // top border costs 1 outer row
	}
	inner := paneW - paneHorizPadding // text area for both styles

	if len(m.filtered) == 0 {
		return paneStyle.Width(paneW).Height(height).Render(
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

	// stats (async — show loading until statsLoadedMsg arrives)
	if m.stats == nil {
		header.WriteString(styleDim.Render("  loading stats…"))
		header.WriteString("\n")
	} else {
		st := m.stats

		// models
		if len(st.Models) > 0 {
			header.WriteString(metaLabel("model"))
			header.WriteString(st.Models[0])
			header.WriteString("\n")
			for _, mdl := range st.Models[1:] {
				header.WriteString(strings.Repeat(" ", labelW))
				header.WriteString(mdl)
				header.WriteString("\n")
			}
			header.WriteString("\n")
		}

		// messages + prompts — both derived from the display slice so the
		// counts match exactly what's rendered in the preview pane.
		var msgCount, promptCount int
		for _, msg := range m.messages {
			msgCount++
			if msg.Role == "user" {
				promptCount++
			}
		}
		header.WriteString(metaLabel("messages"))
		fmt.Fprintf(&header, "%d", msgCount)
		header.WriteString("\n")

		header.WriteString(metaLabel("prompts"))
		fmt.Fprintf(&header, "%d", promptCount)
		header.WriteString("\n")

		// context + tokens (only when AI turns exist)
		if st.ContextTokens > 0 {
			header.WriteString(metaLabel("context"))
			header.WriteString(formatTokens(st.ContextTokens))
			header.WriteString("\n")

			header.WriteString(metaLabel("tokens"))
			fmt.Fprintf(&header, "%s in / %s out", formatTokens(st.InputTokens), formatTokens(st.OutputTokens))
			header.WriteString("\n")

			header.WriteString(metaLabel("cost"))
			header.WriteString(formatCost(st.Cost))
			header.WriteString("\n")
		}
	}

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

	header.WriteString("\n")

	// changes
	header.WriteString(metaLabel("changes"))
	if s.SummaryFiles > 0 {
		fmt.Fprintf(&header, "%d files (", s.SummaryFiles)
		header.WriteString(styleAdd.Render(fmt.Sprintf("+%d", s.SummaryAdditions)))
		header.WriteString(" ")
		header.WriteString(styleDel.Render(fmt.Sprintf("-%d", s.SummaryDeletions)))
		header.WriteString(")")
	} else {
		header.WriteString(styleDim.Render("none"))
	}
	header.WriteString("\n")

	header.WriteString("\n")
	header.WriteString(styleDim.Render(strings.Repeat("─", max(0, inner))))
	header.WriteString("\n")

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
		// Reserve 1 line for the "showing last N messages" indicator —
		// it is always shown when we truncate, so deduct it upfront.
		walkHeight := msgHeight - 1

		// Determine which messages are visible without rendering all of them.
		// Walk backwards from the last message, accumulating line cost, until
		// we exceed the available height. Only the visible subset is rendered.
		costs := make([]int, len(m.messages))
		for i, msg := range m.messages {
			// Line count: 1 for the label + wrapped text lines + 1 blank separator.
			// Approximate text lines as the number of "\n" in the raw text plus 1,
			// then add any wrapping overflow using integer ceiling division.
			rawLines := strings.Count(msg.Text, "\n") + 1
			wrapExtra := 0
			if inner > 0 {
				for _, line := range strings.Split(msg.Text, "\n") {
					visLen := lipgloss.Width(line)
					if visLen > inner {
						wrapExtra += (visLen - 1) / inner // extra wrap rows beyond the first
					}
				}
			}
			costs[i] = 1 + rawLines + wrapExtra + 1 // label line + text lines + blank sep
		}

		used := 0
		first := len(m.messages)
		for i := len(m.messages) - 1; i >= 0; i-- {
			if used+costs[i] > walkHeight {
				break
			}
			used += costs[i]
			first = i
		}
		// Fallback: if even the last message alone exceeds the available
		// height, force it visible so the pane never renders blank.
		clipped := false
		if first == len(m.messages) {
			first = len(m.messages) - 1
			clipped = true
		}

		var sb strings.Builder
		if first > 0 {
			n := len(m.messages) - first
			indicator := lipgloss.NewStyle().
				Width(inner).
				Align(lipgloss.Center).
				Italic(true).
				Foreground(colorDim).
				Render(fmt.Sprintf("showing last %d messages", n))
			sb.WriteString(indicator)
			sb.WriteString("\n")
		}
		for i := first; i < len(m.messages); i++ {
			msg := m.messages[i]
			var label string
			if msg.Role == "user" {
				label = styleRoleUser.Render("[user]")
			} else {
				label = styleRoleAssistant.Render("[asst]")
			}

			// In the fallback case the last (and only) message may be taller
			// than the pane. Clip its text to the lines that actually fit:
			// available = walkHeight minus label line and trailing blank.
			text := msg.Text
			if clipped && i == first {
				avail := walkHeight - 2 // label + blank sep
				if avail < 1 {
					avail = 1
				}
				// Expand the text into wrapped display lines and keep only the first avail.
				var displayLines []string
				for _, raw := range strings.Split(text, "\n") {
					if inner > 0 && lipgloss.Width(raw) > inner {
						for len(raw) > 0 {
							cut := raw
							if lipgloss.Width(cut) > inner {
								cut = raw[:inner]
							}
							displayLines = append(displayLines, cut)
							raw = raw[len(cut):]
						}
					} else {
						displayLines = append(displayLines, raw)
					}
				}
				if len(displayLines) > avail {
					displayLines = displayLines[:avail-1]
					displayLines = append(displayLines, styleDim.Render("…"))
				}
				text = strings.Join(displayLines, "\n")
			}

			wrapped := lipgloss.NewStyle().Width(inner).Render(text)
			sb.WriteString("\n")
			sb.WriteString(label + "\n" + wrapped)
			sb.WriteString("\n")
		}
		msgSection = sb.String()
	}

	return paneStyle.Width(paneW).Height(height).Render(strings.TrimRight(header.String()+msgSection, "\n"))
}

// renderTopBar renders the single-row top border bar that pads the viewport
// against window chrome (titlebars, tab bars, etc.) that may clip the top row.
func renderTopBar(width int) string {
	return styleTopBar.Width(width).Render("")
}

func (m model) renderHint(width int) string {
	appName := styleDim.Render(" ") +
		styleAppNameLazy.Render("Lazy") +
		styleAppNameOpenCode.Render("OpenCode")

	var hints string
	if m.notice != "" {
		hints = "  " + m.notice
	} else {
		switch m.mode {
		case modeSearch:
			hints = "  enter/esc: back   type to filter"
		case modeWorkspaces:
			hints = "  j/k: up/down   d: del   w: workspace   q: quit"
		case modeConfirmDelete, modeConfirmDeleteWorkspace:
			hints = "  y/d: confirm   n/esc: cancel"
		case modeYank:
			hints = "  d: directory   s: session id   esc: cancel"
		case modeGoto:
			hints = "  s: shell   w: workspace   esc: cancel"
		case modeStats:
			hints = "  j/k: scroll   r: refresh   esc: back"
		case modeError:
			hints = "  q: quit"
		default:
			hints = "  j/k: up/down   enter: open   /: search   s: stats   y: yank   g: goto   d: del   w: workspace   q: quit"
		}
	}
	dots := "  " +
		styleDotBlue.Render("•") +
		" " +
		styleDotCyan.Render("•") +
		" " +
		styleDotYellow.Render("•")

	var badge string
	switch m.mode {
	case modeSearch:
		badge = styleModeSearch.Render("SEARCH")
	case modeWorkspaces:
		badge = styleModeWorkspaces.Render("WORKSPACES")
	case modeConfirmDelete, modeConfirmDeleteWorkspace:
		badge = styleModeConfirmDelete.Render("DELETE")
	case modeYank:
		badge = styleModeYank.Render("YANK")
	case modeGoto:
		badge = styleModeGoto.Render("GOTO")
	case modeStats:
		badge = styleModeStats.Render("STATS")
	case modeError:
		badge = styleModeError.Render("ERROR")
	default:
		badge = styleModeNormal.Render("NORMAL")
	}

	// innerW: content width for badge placement — 1-col right margin so the
	// mode badge is not flush to the terminal edge.
	// The border-bearing styleHint uses the full width so borders reach the edge.
	innerW := width - hintBarWidthAdj

	if width < singleLineHintBreakpoint {
		// Narrow layout: hints on top (with border), logo + mode badge below.
		hintLine := styleHint.Width(width).Render(renderHintSegments(hints))
		logo := appName + dots
		space := innerW - lipgloss.Width(logo) - lipgloss.Width(badge)
		if space < 1 {
			space = 1
		}
		logoLine := styleHintBottom.Foreground(colorDim).Width(width).Render(logo + strings.Repeat(" ", space) + badge)
		return hintLine + "\n" + logoLine
	}

	left := appName + dots + renderHintSegments(hints)
	space := innerW - lipgloss.Width(left) - lipgloss.Width(badge)
	if space < 1 {
		space = 1
	}

	return styleHint.Width(width).Render(left + strings.Repeat(" ", space) + badge)
}

// formatSessionRow renders a single session as a fixed-layout row.
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
func formatSessionRow(s session, width, pathColW int, selected bool) string {
	const (
		leadSp  = 1 // single leading space
		minGap  = 2 // minimum spaces between title and path
		trailSp = 1 // single trailing space
	)

	// Responsive column hiding/shortening:
	//   >= wideLayoutBreakpoint: full date "2006-01-02 15:04"
	//   >= 80:                   short date "Jan 02"
	//   <  80:                   no date
	//   >= 80:                   path column shown
	showDateFull := width >= wideLayoutBreakpoint
	showDateShort := width >= defaultTermW && !showDateFull
	showDate := showDateFull || showDateShort
	showPath := width >= defaultTermW

	effectiveDateW := 0
	effectiveDateGap := 0
	switch {
	case showDateFull:
		effectiveDateW = dateFullWidth
		effectiveDateGap = dateGap
	case showDateShort:
		effectiveDateW = dateShortWidth
		effectiveDateGap = dateGap
	}

	effectiveMinGap := minGap
	effectivePathW := pathColW
	if !showPath {
		effectiveMinGap = 0
		effectivePathW = 0
	}

	titleW := width - leadSp - effectiveDateW - effectiveDateGap - effectiveMinGap - effectivePathW - trailSp
	if titleW < 1 {
		titleW = 1
	}

	// ── plain-text content, truncated to column widths ────────────────────────
	var date string
	switch {
	case showDateFull:
		date = s.UpdatedAt.Format("2006-01-02 15:04")
	case showDateShort:
		date = s.UpdatedAt.Format("Jan 02")
	}
	rawTitle := truncate(s.Title, titleW)
	rawPath := truncate(s.ShortDirectory(), effectivePathW)

	// ── pad to exact column widths (plain text, no ANSI yet) ──────────────────
	// Title: left-aligned, space-padded on the right so the path column is pinned.
	paddedTitle := rawTitle + strings.Repeat(" ", titleW-lipgloss.Width(rawTitle))
	// Path: right-aligned, space-padded on the left.
	paddedPath := strings.Repeat(" ", effectivePathW-lipgloss.Width(rawPath)) + rawPath
	// Trailing fill: keeps the background unbroken to the edge of the list pane.
	fixedCols := leadSp + effectiveDateW + effectiveDateGap + titleW + effectiveMinGap + effectivePathW + trailSp
	trailFill := strings.Repeat(" ", max(0, width-fixedCols))

	// ── per-cell styles ───────────────────────────────────────────────────────
	// Every span is styled independently and carries the background color (when
	// selected). This is the key invariant: because termenv emits \e[0m after
	// each span, an outer background Render would be killed by the first inner
	// reset. By giving every span its own background we guarantee the highlight
	// is unbroken across the full row width regardless of what other attributes
	// (bold, color) individual cells carry.
	//
	// Use pre-declared package-level style variants to avoid allocating mutated
	// Style copies on every rendered row every frame.
	base := styleRowBase
	styledTitle := styleRowTitleBase
	styledDate := styleRowDateBase
	styledPath := styleRowPathBase
	if selected {
		base = styleRowSelected
		styledTitle = styleRowTitleSelected
		styledDate = styleRowDateSelected
		styledPath = styleRowPathSelected
	}

	styledLead := base.Render(strings.Repeat(" ", leadSp))
	styledTitleStr := styledTitle.Render(paddedTitle)
	styledTrail := base.Render(strings.Repeat(" ", trailSp) + trailFill)

	row := styledLead
	if showDate {
		row += styledDate.Render(date + strings.Repeat(" ", effectiveDateGap))
	}
	row += styledTitleStr
	if showPath {
		row += styledPath.Render(strings.Repeat(" ", effectiveMinGap) + paddedPath)
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

// ── Modals ────────────────────────────────────────────────────────────────────

// renderGotoModal returns the styled modal for the g-prefix goto menu.
func (m model) renderGotoModal() string {
	sKey := styleModalKeyGoto.Render("s")
	wKey := styleModalKeyGoto.Render("w")

	content := styleModalGotoTitle.Render("Go to…") + "\n\n" +
		sKey + "  open shell" + "\n" +
		styleDim.Render("     type 'exit' to return to lazyopencode") + "\n" +
		wKey + "  workspace view"

	return styleModalGoto.Render(content)
}

// renderYankModal returns the styled modal for the yank-to-clipboard prompt.
func (m model) renderYankModal() string {
	dKey := styleModalKeyYank.Render("d")
	sKey := styleModalKeyYank.Render("s")

	content := styleModalYankTitle.Render("Yank to clipboard") + "\n\n" +
		dKey + "  directory" + "\n" +
		sKey + "  session id"

	return styleModalYank.Render(content)
}

// renderSessionModal returns the styled modal for a single-session delete.
func (m model) renderSessionModal() string {
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
		stylePreviewTitle.Render(truncate(sessionTitle, modalInnerWidth)) + "\n\n" +
		confirm

	return styleModal.Render(content)
}

// renderErrorModal returns the styled modal for a hard error state.
// The app is non-interactive until the user quits.
func (m model) renderErrorModal() string {
	qKey := styleModalKey.Render("q")

	msg := m.err.Error()
	content := styleModalTitle.Render("Fatal Error") + "\n\n" +
		styleDim.Render(truncate(msg, modalInnerWidth)) + "\n\n" +
		qKey + "  quit"

	return styleModal.Render(content)
}

// renderWorkspaceModal returns the styled modal for a workspace delete.
func (m model) renderWorkspaceModal() string {
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

	// Use pre-computed DisplayDir from the workspace to avoid calling
	// os.UserHomeDir in the render path (AGENTS.md convention).
	displayPath := m.pendingDeleteWorkspace
	for _, ws := range m.workspaces {
		if ws.Dir == m.pendingDeleteWorkspace {
			displayPath = ws.DisplayDir
			break
		}
	}

	content := styleModalTitle.Render("Delete workspace?") + "\n\n" +
		stylePreviewTitle.Render(truncate(displayPath, modalInnerWidth)) + "\n" +
		countLine + "\n\n" +
		confirm

	return styleModal.Render(content)
}
