package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Layout constants ───────────────────────────────────────────────────────────

const (
	// Fixed column widths for the stats tables (right-aligned values).
	tblGap   = 3  // spaces before each value column
	tblSessW = 8  // "sessions" column
	tblPrmtW = 8  // "prompts" column
	tblInW   = 10 // "tokens in" column
	tblOutW  = 10 // "tokens out" column
	tblTimeW = 8  // "time" column

	// Compact-mode column widths (< statsCompactBreakpoint terminal columns).
	// Two token columns are merged into one "↑in/↓out" column; time shrinks
	// slightly; sessions and prompts also tighten.
	statsCompactBreakpoint = 100 // terminal width that triggers compact layout
	tblSessWC              = 5   // compact sessions column (normal: 8)
	tblPrmtWC              = 5   // compact prompts column (normal: 8)
	tblTokW                = 7   // compact merged tokens column (in+out total, e.g. "123.4M")
	tblTimeWC              = 6   // compact time column (normal: 8) — fits "1h 38m"
	tblGapC                = 2   // compact gap between columns (normal: 3)

	// Cost column — same width in both normal and compact layouts.
	tblCostW = 9 // fits "$9,999.99"

	// Fieldset padding (inside the border, each side).
	fieldsetPadX = 2

	// Minimum name column width for tables.
	nameColMin = 10

	// Extra left indent applied inside the name column for table header and data rows.
	// The section rule and separator stay at the outer pad; only column content shifts.
	tblNameIndent = 3
)

// statsTableRow holds a single row for the generic stats table renderer.
type statsTableRow struct {
	name        string
	sessions    int
	prompts     int
	inputTokens int
	outTokens   int
	durationMS  int64
	cost        float64
	// subRows is non-nil for project rows; each entry is a model sub-row.
	subRows []statsTableRow
}

// ── Top-level renderer ────────────────────────────────────────────────────────

// renderStatsView is the top-level renderer for modeStats.
func (m model) renderStatsView(w, h int) string {
	topBar := renderTopBar(w)
	hint := m.renderHint(w)
	hintH := strings.Count(hint, "\n") + 1
	bodyH := h - topBarH - hintH

	body := styleListPane.Width(w).Height(bodyH).Render(m.renderStats(w, bodyH))
	return topBar + "\n" + body + "\n" + hint
}

// renderStats renders the full stats dashboard body, windowed to bodyH lines
// starting at m.statsScrollOffset. Content that fits entirely is returned as-is.
func (m model) renderStats(w, h int) string {
	const indent = 2
	pad := strings.Repeat(" ", indent)

	if m.globalStats == nil {
		return styleSpPanel.Render(pad) + styleDimPanel.Render("loading stats…")
	}

	gs := m.globalStats
	compact := w < statsCompactBreakpoint

	var sb strings.Builder

	// ── KV summary fieldsets ───────────────────────────────────────────────────

	// avail is the usable content width (inside the 2-space left+right indent).
	avail := w - indent*2
	if avail < 1 {
		avail = 1
	}
	// Wide mode halves avail between two fieldsets. Round avail down so
	// (avail - fieldsetGap) is even — otherwise a stray column leaks and
	// the gap jumps on resize. Compact mode uses avail directly, so skip
	// the rounding there (it would just steal a col from the right margin).
	if !compact && (avail-2)%2 != 0 {
		avail--
	}

	// Inner horizontal padding inside fieldset borders and around the
	// outermost table columns. Compact mode tightens this so narrow
	// terminals don't waste columns on empty margins.
	padX := fieldsetPadX
	if compact {
		padX = 1
	}

	if compact {
		// Narrow: stack ALL TIME on top, LAST 7 DAYS below.
		// Each fieldset takes the full avail width.
		outerW := avail
		if outerW < 20 {
			outerW = 20
		}
		innerW := outerW - 2 - padX*2
		if innerW < 10 {
			innerW = 10
		}

		allTimeKV := buildSummaryKV(summaryData{
			sessions: gs.TotalSessions, prompts: gs.TotalPrompts,
			input: gs.TotalInput, output: gs.TotalOutput,
			cacheRead: gs.TotalCacheRead, cacheWrite: gs.TotalCacheWrite,
			durationMS: gs.TotalDurationMS,
			files:      gs.TotalFiles, additions: gs.TotalAdditions, deletions: gs.TotalDeletions,
			cost:   gs.TotalCost,
			innerW: innerW,
		}, compact)
		recentKV := buildSummaryKV(summaryData{
			sessions: gs.RecentSessions, prompts: gs.RecentPrompts,
			input: gs.RecentInput, output: gs.RecentOutput,
			cacheRead: gs.RecentCacheRead, cacheWrite: gs.RecentCacheWrite,
			durationMS: gs.RecentDurationMS,
			files:      gs.RecentFiles, additions: gs.RecentAdditions, deletions: gs.RecentDeletions,
			cost:   gs.RecentCost,
			innerW: innerW,
		}, compact)

		for _, line := range strings.Split(renderFieldset("ALL TIME", allTimeKV, outerW, innerW, padX), "\n") {
			sb.WriteString(styleSpPanel.Render(pad) + line + "\n")
		}
		sb.WriteString(styleSpPanel.Render("\n"))
		for _, line := range strings.Split(renderFieldset("LAST 7 DAYS", recentKV, outerW, innerW, padX), "\n") {
			sb.WriteString(styleSpPanel.Render(pad) + line + "\n")
		}
	} else {
		// Wide: side-by-side fieldsets.
		const fieldsetGap = 2
		outerW := (avail - fieldsetGap) / 2
		if outerW < 20 {
			outerW = 20
		}
		// Inner text width: outer minus left+right borders (1 each) minus padding (fieldsetPadX each side).
		innerW := outerW - 2 - padX*2
		if innerW < 10 {
			innerW = 10
		}

		allTimeKV := buildSummaryKV(summaryData{
			sessions: gs.TotalSessions, prompts: gs.TotalPrompts,
			input: gs.TotalInput, output: gs.TotalOutput,
			cacheRead: gs.TotalCacheRead, cacheWrite: gs.TotalCacheWrite,
			durationMS: gs.TotalDurationMS,
			files:      gs.TotalFiles, additions: gs.TotalAdditions, deletions: gs.TotalDeletions,
			cost:   gs.TotalCost,
			innerW: innerW,
		}, compact)
		recentKV := buildSummaryKV(summaryData{
			sessions: gs.RecentSessions, prompts: gs.RecentPrompts,
			input: gs.RecentInput, output: gs.RecentOutput,
			cacheRead: gs.RecentCacheRead, cacheWrite: gs.RecentCacheWrite,
			durationMS: gs.RecentDurationMS,
			files:      gs.RecentFiles, additions: gs.RecentAdditions, deletions: gs.RecentDeletions,
			cost:   gs.RecentCost,
			innerW: innerW,
		}, compact)

		left := renderFieldset("ALL TIME", allTimeKV, outerW, innerW, padX)
		right := renderFieldset("LAST 7 DAYS", recentKV, outerW, innerW, padX)
		leftLines := strings.Split(left, "\n")
		rightLines := strings.Split(right, "\n")
		gapStr := styleSpPanel.Render(strings.Repeat(" ", fieldsetGap))
		for len(leftLines) < len(rightLines) {
			leftLines = append(leftLines, "")
		}
		for len(rightLines) < len(leftLines) {
			rightLines = append(rightLines, "")
		}
		for i := range leftLines {
			sb.WriteString(styleSpPanel.Render(pad) + leftLines[i] + gapStr + rightLines[i] + "\n")
		}
	}

	// tblNameIndent spaces are prepended to the name column content in header
	// and data rows; +1 for the trailing space after the rightmost column.
	var fixedW int
	if compact {
		fixedW = tblGapC + tblSessWC + tblGapC + tblPrmtWC + tblGapC + tblTokW + tblGapC + tblTimeWC + tblGapC + tblCostW + 1
	} else {
		fixedW = tblGap + tblSessW + tblGap + tblPrmtW + tblGap + tblInW + tblGap + tblOutW + tblGap + tblTimeW + tblGap + tblCostW + 3
	}

	// Tables get an extra 2-space left indent (tblPad) so they visually align
	// with the fieldset content above. tblAvail shrinks by the same amount so
	// the right margin stays symmetric.
	tblPad := pad
	tblAvail := avail

	// ── Models table ──────────────────────────────────────────────────────────
	if len(gs.Models) > 0 {
		sb.WriteString("\n")
		sb.WriteString(renderSectionRule("MODELS", tblPad, tblAvail))

		nameW := tblAvail - fixedW
		if nameW < nameColMin {
			nameW = nameColMin
		}
		rows := buildModelRows(gs.Models)
		sb.WriteString(renderTableHeader(tblPad, nameW, compact))
		sb.WriteString(renderTableRule(tblPad, tblAvail))
		sb.WriteString(renderModelRows(rows, tblPad, nameW, compact))
	}

	// ── Projects table ────────────────────────────────────────────────────────
	if len(gs.Projects) > 0 {
		sb.WriteString("\n")
		sb.WriteString(renderSectionRule("PROJECTS", tblPad, tblAvail))

		nameW := tblAvail - fixedW
		if nameW < nameColMin {
			nameW = nameColMin
		}
		rows := buildProjectRows(gs.Projects)
		sb.WriteString(renderTableHeader(tblPad, nameW, compact))
		sb.WriteString(renderTableRule(tblPad, tblAvail))
		sb.WriteString(renderProjectRows(rows, tblPad, nameW, compact))
	}

	// Window the full content to bodyH lines at the current scroll offset.
	return scrollContent(sb.String(), m.statsScrollOffset, h)
}

// scrollContent windows a multi-line string to bodyH visible lines starting
// at offset. The offset is clamped so it can never push the window past the
// last line. An empty string is returned as-is.
func scrollContent(content string, offset, bodyH int) string {
	lines := strings.Split(content, "\n")
	// Strip a single trailing empty line produced by the final "\n" so the
	// total count reflects real content rows.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	total := len(lines)
	if total == 0 || bodyH <= 0 {
		return content
	}
	maxOffset := max(0, total-bodyH)
	if offset > maxOffset {
		offset = maxOffset
	}
	end := offset + bodyH
	if end > total {
		end = total
	}
	return strings.Join(lines[offset:end], "\n")
}

// ── Section rule ──────────────────────────────────────────────────────────────

// renderSectionRule renders "── TITLE ─────...──\n".
// pad is prepended; the rule fills exactly avail columns after the pad.
// The title word is purple; the surrounding dashes are dim.
// Uses *Panel style variants so the panel background carries through.
func renderSectionRule(title, pad string, avail int) string {
	prefix := "── "
	suffix := " "
	// Use lipgloss.Width (not len) — "─" is U+2500, 3 bytes in UTF-8 but 1 visible column.
	labelVisW := lipgloss.Width(prefix) + lipgloss.Width(title) + lipgloss.Width(suffix)
	remaining := avail - labelVisW
	if remaining < 0 {
		remaining = 0
	}
	line := styleDimPanel.Render(prefix) +
		styleStatsTitlePanel.Render(title) +
		styleDimPanel.Render(suffix+strings.Repeat("─", remaining))
	return styleSpPanel.Render(pad) + line + "\n"
}

// renderTableRule renders a dim separator rule: pad + ─×w.
func renderTableRule(pad string, w int) string {
	return styleSpPanel.Render(pad) + styleDimPanel.Render(strings.Repeat("─", w)) + "\n"
}

// ── Fieldset ──────────────────────────────────────────────────────────────────

// renderFieldset builds a box manually using plain box-drawing characters so
// the title can be embedded in the top border without ANSI-parsing tricks.
//
//	┌─ TITLE ─────────────────────────────┐
//	│ content line                        │
//	└─────────────────────────────────────┘
//
// outerW is the total column width of the box; innerW is the text content width.
// Uses *Panel style variants throughout so the panel background carries through.
func renderFieldset(title, content string, outerW, innerW, padX int) string {
	padStr := strings.Repeat(" ", padX)
	borderW := outerW - 2 // dashes between the two corner characters

	// ── top border with title ──────────────────────────────────────────────
	// Spaces flanking the title are folded into adjacent styled spans so no
	// bare unstyled space can bleed the terminal's native background.
	titleVisW := lipgloss.Width(title) + 2 // +2 for the flanking spaces
	dashesLeft := 1
	dashesRight := borderW - dashesLeft - titleVisW
	if dashesRight < 0 {
		dashesRight = 0
	}
	top := styleDimPanel.Render("┌"+strings.Repeat("─", dashesLeft)+" ") +
		styleStatsTitlePanel.Render(title) +
		styleDimPanel.Render(" "+strings.Repeat("─", dashesRight)+"┐")

	// ── content lines ──────────────────────────────────────────────────────
	contentLines := strings.Split(content, "\n")
	var body strings.Builder
	for _, cl := range contentLines {
		vis := lipgloss.Width(cl)
		fill := ""
		if vis < innerW {
			fill = styleSpPanel.Render(strings.Repeat(" ", innerW-vis))
		}
		body.WriteString(
			styleDimPanel.Render("│") +
				styleSpPanel.Render(padStr) +
				cl + fill +
				styleSpPanel.Render(padStr) +
				styleDimPanel.Render("│") + "\n",
		)
	}

	// ── bottom border ──────────────────────────────────────────────────────
	bot := styleDimPanel.Render("└" + strings.Repeat("─", borderW) + "┘")

	return top + "\n" + body.String() + bot
}

// buildSummaryKV returns a pre-formatted multi-line string of KV rows for a
// fieldset. Labels are dim; sessions/prompts are bold yellow; other values are
// plain (inherit fg from parent block). Uses *Panel style variants.
type summaryData struct {
	sessions, prompts           int
	input, output               int
	cacheRead, cacheWrite       int
	durationMS                  int64
	files, additions, deletions int
	cost                        float64
	innerW                      int
}

func buildSummaryKV(d summaryData, compact bool) string {
	const labelW = 14

	// kv builds "dim-label<spaces>value" right-aligning value to fill innerW.
	// Uses lipgloss.Width for value measurement so ANSI codes don't break padding.
	kv := func(label, value string) string {
		rawLabel := label + strings.Repeat(" ", max(0, labelW-len(label)))
		styledLabel := styleDimPanel.Render(rawLabel)

		valueW := d.innerW - labelW
		if valueW < 1 {
			valueW = 1
		}
		visW := lipgloss.Width(value)
		spaces := ""
		if visW < valueW {
			spaces = styleSpPanel.Render(strings.Repeat(" ", valueW-visW))
		}
		return styledLabel + spaces + value
	}

	blank := ""

	var perSession string
	if d.sessions > 0 && d.durationMS > 0 {
		perSession = formatDurationMS(d.durationMS / int64(d.sessions))
	} else {
		perSession = "—"
	}

	// +N green, -N red — panel-background variants from styles.go to prevent bleed.
	changes := styleAddPanel.Render("+"+formatCommas(d.additions)) +
		styleSpPanel.Render(" / ") +
		styleDelPanel.Render("-"+formatCommas(d.deletions))

	rows := []string{
		// sessions and prompts: bold yellow to match table session count column.
		kv("sessions", styleStatsCountPanel.Render(formatCommas(d.sessions))),
		kv("prompts", styleStatsCountPanel.Render(formatCommas(d.prompts))),
		blank,
	}

	if compact {
		rows = append(rows,
			kv("tokens in/out", styleSpPanel.Render(formatTokens(d.input)+" / "+formatTokens(d.output))),
			kv("cache r/w", styleSpPanel.Render(formatTokens(d.cacheRead)+" / "+formatTokens(d.cacheWrite))),
		)
	} else {
		rows = append(rows,
			kv("tokens in", styleSpPanel.Render(formatTokens(d.input))),
			kv("tokens out", styleSpPanel.Render(formatTokens(d.output))),
			kv("cache read", styleSpPanel.Render(formatTokens(d.cacheRead))),
			kv("cache write", styleSpPanel.Render(formatTokens(d.cacheWrite))),
		)
	}

	rows = append(rows,
		blank,
		kv("total time", styleSpPanel.Render(formatDurationMS(d.durationMS))),
		kv("avg time", styleSpPanel.Render(perSession)),
		kv("cost", styleSpPanel.Render(formatCost(d.cost))),
		blank,
		kv("files", styleSpPanel.Render(formatCommas(d.files))),
		kv("changes", changes),
	)
	return strings.Join(rows, "\n")
}

// ── Models table ──────────────────────────────────────────────────────────────

// buildModelRows converts []modelStat to []statsTableRow.
func buildModelRows(models []modelStat) []statsTableRow {
	rows := make([]statsTableRow, len(models))
	for i, ms := range models {
		rows[i] = statsTableRow{
			name:        ms.Name,
			sessions:    ms.Sessions,
			prompts:     ms.Prompts,
			inputTokens: ms.InputTokens,
			outTokens:   ms.OutputTokens,
			durationMS:  ms.DurationMS,
		}
	}
	return rows
}

// ── Projects table ────────────────────────────────────────────────────────────

// buildProjectRows converts []projectStat to []statsTableRow, with sub-rows.
func buildProjectRows(projects []projectStat) []statsTableRow {
	rows := make([]statsTableRow, len(projects))
	for i, ps := range projects {
		dir := ps.DisplayDir
		if dir == "" {
			dir = ps.Dir
		}
		subRows := make([]statsTableRow, len(ps.Models))
		for j, ms := range ps.Models {
			subRows[j] = statsTableRow{
				name:        ms.Name,
				sessions:    ms.Sessions,
				prompts:     ms.Prompts,
				inputTokens: ms.InputTokens,
				outTokens:   ms.OutputTokens,
				durationMS:  ms.DurationMS,
			}
		}
		rows[i] = statsTableRow{
			name:        dir,
			sessions:    ps.Sessions,
			prompts:     ps.Prompts,
			inputTokens: ps.InputTokens,
			outTokens:   ps.OutputTokens,
			durationMS:  ps.DurationMS,
			subRows:     subRows,
		}
	}
	return rows
}

// ── Shared table rendering ────────────────────────────────────────────────────

// renderTableHeader renders the shared column header row for both stats tables.
func renderTableHeader(pad string, nameW int, compact bool) string {
	hdr := styleStatsHeaderPanel
	g := strings.Repeat(" ", tblGap)
	if compact {
		g = strings.Repeat(" ", tblGapC)
	}
	var sb strings.Builder
	sb.WriteString(styleSpPanel.Render(pad))
	sb.WriteString(styleSpPanel.Render(padRight("", nameW)))
	if compact {
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblSessWC, "sess")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblPrmtWC, "prmt")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTokW, "tokens")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTimeWC, "time")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblCostW, "cost")))
	} else {
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblSessW, "sessions")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblPrmtW, "prompts")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblInW, "tokens in")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblOutW, "tokens out")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTimeW, "time")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblCostW, "cost")))
	}
	sb.WriteString(styleSpPanel.Render(" "))
	sb.WriteString("\n")
	return sb.String()
}

// renderTableCols writes the value columns for a single row into sb, using the
// provided style for each cell. It handles both compact and normal layouts.
func renderTableCols(sb *strings.Builder, r statsTableRow, g string, compact bool, sp, count lipgloss.Style) {
	if compact {
		tok := formatTokens(r.inputTokens + r.outTokens)
		sb.WriteString(count.Render(g + fmt.Sprintf("%*d", tblSessWC, r.sessions)))
		sb.WriteString(count.Render(g + fmt.Sprintf("%*d", tblPrmtWC, r.prompts)))
		sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblTokW, tok)))
		sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblTimeWC, formatDurationMS(r.durationMS))))
		sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblCostW, formatCost(r.cost))))
	} else {
		sb.WriteString(count.Render(g + fmt.Sprintf("%*d", tblSessW, r.sessions)))
		sb.WriteString(count.Render(g + fmt.Sprintf("%*d", tblPrmtW, r.prompts)))
		sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblInW, formatTokens(r.inputTokens))))
		sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblOutW, formatTokens(r.outTokens))))
		sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblTimeW, formatDurationMS(r.durationMS))))
		sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblCostW, formatCost(r.cost))))
	}
}

// renderModelRows renders all model rows with alternating row backgrounds.
func renderModelRows(rows []statsTableRow, pad string, nameW int, compact bool) string {
	var sb strings.Builder
	g := strings.Repeat(" ", tblGap)
	indent := tblNameIndent
	trailing := 3
	if compact {
		g = strings.Repeat(" ", tblGapC)
		indent = 1
		trailing = 1
	}
	indentStr := strings.Repeat(" ", indent)
	for i, r := range rows {
		sp, label, count := styleSpPanel, styleStatsLabelPanel, styleStatsCountPanel
		if i%2 != 0 {
			sp, label, count = styleSpPanelAlt, styleStatsLabelPanelAlt, styleStatsCountPanelAlt
		}
		name := indentStr + truncate(r.name, nameW-indent)
		sb.WriteString(styleSpPanel.Render(pad))
		sb.WriteString(label.Render(padRight(name, nameW)))
		renderTableCols(&sb, r, g, compact, sp, count)
		sb.WriteString(sp.Render(strings.Repeat(" ", trailing)))
		sb.WriteString("\n")
	}
	return sb.String()
}

// renderProjectRows renders all project rows, each followed by dimmed model sub-rows.
// In compact mode the project name is shortened to its last path component.
func renderProjectRows(rows []statsTableRow, pad string, nameW int, compact bool) string {
	var sb strings.Builder
	g := strings.Repeat(" ", tblGap)
	indent := tblNameIndent
	trailing := 3
	if compact {
		g = strings.Repeat(" ", tblGapC)
		indent = 1
		trailing = 1
	}
	indentStr := strings.Repeat(" ", indent)
	subIndentStr := strings.Repeat(" ", indent+2)
	for _, r := range rows {
		// In compact mode show only the last path component (e.g. "myapp" not "~/code/myapp").
		displayName := r.name
		if compact {
			displayName = filepath.Base(r.name)
		}
		name := indentStr + truncate(displayName, nameW-indent)
		sb.WriteString(styleSpPanel.Render(pad))
		sb.WriteString(styleStatsLabelPanel.Render(padRight(name, nameW)))
		renderTableCols(&sb, r, g, compact, styleSpPanel, styleStatsCountPanel)
		sb.WriteString(styleSpPanel.Render(strings.Repeat(" ", trailing)))
		sb.WriteString("\n")

		// Per-model sub-rows: alt background, extra indent.
		for _, sr := range r.subRows {
			subName := subIndentStr + truncate(sr.name, nameW-indent-2)
			sb.WriteString(styleSpPanel.Render(pad))
			sb.WriteString(styleDimPanelAlt.Render(padRight(subName, nameW)))
			renderTableCols(&sb, sr, g, compact, styleDimPanelAlt, styleDimPanelAlt)
			sb.WriteString(styleSpPanelAlt.Render(strings.Repeat(" ", trailing)))
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// padRight pads s with trailing spaces to exactly w visible columns.
func padRight(s string, w int) string {
	n := lipgloss.Width(s)
	if n >= w {
		return s
	}
	return s + strings.Repeat(" ", w-n)
}
