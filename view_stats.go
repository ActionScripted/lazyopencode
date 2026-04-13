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
	tblTurnW = 6  // "turns" column
	tblInW   = 10 // "tokens in" column
	tblOutW  = 10 // "tokens out" column
	tblTimeW = 8  // "time" column
	tblCostW = 8  // "cost" column

	// Compact-mode column widths (< statsCompactBreakpoint terminal columns).
	// Two token columns are merged into one "↑in/↓out" column; time and cost
	// shrink slightly; sessions and turns also tighten.
	statsCompactBreakpoint = 100 // terminal width that triggers compact layout
	tblSessWC              = 5   // compact sessions column (normal: 8)
	tblTurnWC              = 4   // compact turns column (normal: 6)
	tblTokW                = 12  // merged "in/out" column replacing tblInW+tblOutW; header uses ↑/↓
	tblTimeWC              = 4   // compact time column (normal: 8)
	tblCostWC              = 5   // compact cost column (normal: 8)
	tblGapC                = 2   // compact gap between columns (normal: 4)

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
	turns       int
	inputTokens int
	outTokens   int
	durationMS  int64
	cost        float64
	// subRows is non-nil for project rows; each entry is a model sub-row.
	subRows []statsTableRow
}

// ── Top-level renderer ────────────────────────────────────────────────────────

// renderStatsView is the top-level renderer for ModeStats.
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
		return pad + styleDimPanel.Render("loading stats…")
	}

	gs := m.globalStats
	compact := w < statsCompactBreakpoint

	var sb strings.Builder

	// ── KV summary fieldsets ───────────────────────────────────────────────────

	// avail is the usable content width (inside the 2-space left indent,
	// leaving a 1-char right margin so table content doesn't kiss the edge).
	avail := w - indent - 1
	if avail < 1 {
		avail = 1
	}

	if compact {
		// Narrow: stack ALL TIME on top, LAST 7 DAYS below.
		// Each fieldset takes the full avail width.
		outerW := avail
		if outerW < 20 {
			outerW = 20
		}
		innerW := outerW - 2 - fieldsetPadX*2
		if innerW < 10 {
			innerW = 10
		}

		allTimeKV := buildSummaryKV(
			gs.TotalSessions, gs.TotalMessages,
			gs.TotalInput, gs.TotalOutput,
			gs.TotalCacheRead, gs.TotalCacheWrite,
			gs.TotalDurationMS,
			gs.TotalFiles, gs.TotalAdditions, gs.TotalDeletions,
			innerW,
		)
		recentKV := buildSummaryKV(
			gs.RecentSessions, gs.RecentMessages,
			gs.RecentInput, gs.RecentOutput,
			gs.RecentCacheRead, gs.RecentCacheWrite,
			gs.RecentDurationMS,
			gs.RecentFiles, gs.RecentAdditions, gs.RecentDeletions,
			innerW,
		)

		for _, line := range strings.Split(renderFieldset("ALL TIME", allTimeKV, outerW, innerW), "\n") {
			sb.WriteString(pad + line + "\n")
		}
		sb.WriteString("\n")
		for _, line := range strings.Split(renderFieldset("LAST 7 DAYS", recentKV, outerW, innerW), "\n") {
			sb.WriteString(pad + line + "\n")
		}
	} else {
		// Wide: side-by-side fieldsets.
		const fieldsetGap = 2
		outerW := (avail - fieldsetGap) / 2
		if outerW < 20 {
			outerW = 20
		}
		// Inner text width: outer minus left+right borders (1 each) minus padding (fieldsetPadX each side).
		innerW := outerW - 2 - fieldsetPadX*2
		if innerW < 10 {
			innerW = 10
		}

		allTimeKV := buildSummaryKV(
			gs.TotalSessions, gs.TotalMessages,
			gs.TotalInput, gs.TotalOutput,
			gs.TotalCacheRead, gs.TotalCacheWrite,
			gs.TotalDurationMS,
			gs.TotalFiles, gs.TotalAdditions, gs.TotalDeletions,
			innerW,
		)
		recentKV := buildSummaryKV(
			gs.RecentSessions, gs.RecentMessages,
			gs.RecentInput, gs.RecentOutput,
			gs.RecentCacheRead, gs.RecentCacheWrite,
			gs.RecentDurationMS,
			gs.RecentFiles, gs.RecentAdditions, gs.RecentDeletions,
			innerW,
		)

		left := renderFieldset("ALL TIME", allTimeKV, outerW, innerW)
		right := renderFieldset("LAST 7 DAYS", recentKV, outerW, innerW)
		row := lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", fieldsetGap), right)
		for _, line := range strings.Split(row, "\n") {
			sb.WriteString(pad + line + "\n")
		}
	}

	// ── Models table ──────────────────────────────────────────────────────────
	if len(gs.Models) > 0 {
		sb.WriteString("\n")
		sb.WriteString(renderSectionRule("MODELS", pad, avail))

		// tblNameIndent spaces are prepended to the name column content in header
		// and data rows; +1 for the trailing space after the rightmost column.
		var modelFixedW int
		if compact {
			modelFixedW = tblGapC + tblSessWC + tblGapC + tblTurnWC + tblGapC + tblTokW + tblGapC + tblTimeWC + tblGapC + tblCostWC + tblNameIndent + 1
		} else {
			modelFixedW = tblGap + tblSessW + tblGap + tblTurnW + tblGap + tblInW + tblGap + tblOutW + tblGap + tblTimeW + tblGap + tblCostW + tblNameIndent + 1
		}
		nameW := avail - modelFixedW
		if nameW < nameColMin {
			nameW = nameColMin
		}
		rows := buildModelRows(gs.Models)
		sb.WriteString(renderModelHeader(pad, nameW, compact))
		sb.WriteString(renderTableRule(pad, avail))
		sb.WriteString(renderModelRows(rows, pad, nameW, compact))
	}

	// ── Projects table ────────────────────────────────────────────────────────
	if len(gs.Projects) > 0 {
		sb.WriteString("\n")
		sb.WriteString(renderSectionRule("PROJECTS", pad, avail))

		var projFixedW int
		if compact {
			projFixedW = tblGapC + tblSessWC + tblGapC + tblTurnWC + tblGapC + tblTokW + tblGapC + tblTimeWC + tblGapC + tblCostWC + tblNameIndent + 1
		} else {
			projFixedW = tblGap + tblSessW + tblGap + tblTurnW + tblGap + tblInW + tblGap + tblOutW + tblGap + tblTimeW + tblGap + tblCostW + tblNameIndent + 1
		}
		nameW := avail - projFixedW
		if nameW < nameColMin {
			nameW = nameColMin
		}
		rows := buildProjectRows(gs.Projects)
		sb.WriteString(renderProjectHeader(pad, nameW, compact))
		sb.WriteString(renderTableRule(pad, avail))
		sb.WriteString(renderProjectRows(rows, pad, nameW, compact))
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
	return pad + line + "\n"
}

// renderTableRule renders a dim separator rule: pad + ─×w.
func renderTableRule(pad string, w int) string {
	return pad + styleDimPanel.Render(strings.Repeat("─", w)) + "\n"
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
func renderFieldset(title, content string, outerW, innerW int) string {
	padStr := strings.Repeat(" ", fieldsetPadX)
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
// fieldset. Labels are dim; sessions/turns are bold yellow; other values are
// plain (inherit fg from parent block). Uses *Panel style variants.
func buildSummaryKV(
	sessions, turns int,
	input, output int,
	cacheRead, cacheWrite int,
	durationMS int64,
	files, additions, deletions int,
	innerW int,
) string {
	const labelW = 14

	// kv builds "dim-label<spaces>value" right-aligning value to fill innerW.
	// Uses lipgloss.Width for value measurement so ANSI codes don't break padding.
	kv := func(label, value string) string {
		rawLabel := label + strings.Repeat(" ", max(0, labelW-len(label)))
		styledLabel := styleDimPanel.Render(rawLabel)

		valueW := innerW - labelW
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
	if sessions > 0 && durationMS > 0 {
		perSession = formatDurationMS(durationMS / int64(sessions))
	} else {
		perSession = "—"
	}

	// +N green, -N red — panel-background variants from styles.go to prevent bleed.
	changes := styleAddPanel.Render("+"+fmtCommas(additions)) +
		styleSpPanel.Render(" / ") +
		styleDelPanel.Render("-"+fmtCommas(deletions))

	rows := []string{
		// sessions and turns: bold yellow to match table session count column.
		kv("sessions", styleStatsCountPanel.Render(fmtCommas(sessions))),
		kv("turns", styleStatsCountPanel.Render(fmtCommas(turns))),
		blank,
		kv("tokens in", formatTokens(input)),
		kv("tokens out", formatTokens(output)),
		kv("cache read", formatTokens(cacheRead)),
		kv("cache write", formatTokens(cacheWrite)),
		blank,
		kv("total time", formatDurationMS(durationMS)),
		kv("avg time", perSession),
		blank,
		kv("files", fmtCommas(files)),
		kv("changes", changes),
	}
	return strings.Join(rows, "\n")
}

// ── Models table ──────────────────────────────────────────────────────────────

// buildModelRows converts []ModelStat to []statsTableRow.
func buildModelRows(models []ModelStat) []statsTableRow {
	rows := make([]statsTableRow, len(models))
	for i, ms := range models {
		rows[i] = statsTableRow{
			name:        ms.Name,
			sessions:    ms.Sessions,
			turns:       ms.Turns,
			inputTokens: ms.InputTokens,
			outTokens:   ms.OutputTokens,
			durationMS:  ms.DurationMS,
			cost:        modelCost(ms.Name, ms.InputTokens, ms.OutputTokens),
		}
	}
	return rows
}

// renderModelHeader renders the column header row for the models table.
func renderModelHeader(pad string, nameW int, compact bool) string {
	hdr := styleStatsHeaderPanel
	g := strings.Repeat(" ", tblGap)
	if compact {
		g = strings.Repeat(" ", tblGapC)
	}
	var sb strings.Builder
	sb.WriteString(pad)
	sb.WriteString(styleSpPanel.Render(padRight("", nameW)))
	if compact {
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblSessWC, "sess")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTurnWC, "trns")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTokW, "tokens ↑/↓")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTimeWC, "time")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblCostWC, "cost")))
	} else {
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblSessW, "sessions")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTurnW, "turns")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblInW, "tokens in")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblOutW, "tokens out")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTimeW, "time")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblCostW, "cost")))
	}
	sb.WriteString(styleSpPanel.Render(" "))
	sb.WriteString("\n")
	return sb.String()
}

// renderModelRows renders all model rows.
func renderModelRows(rows []statsTableRow, pad string, nameW int, compact bool) string {
	var sb strings.Builder
	g := strings.Repeat(" ", tblGap)
	if compact {
		g = strings.Repeat(" ", tblGapC)
	}
	for i, r := range rows {
		sp, label, count := styleSpPanel, styleStatsLabelPanel, styleStatsCountPanel
		if i%2 != 0 {
			sp, label, count = styleSpPanelAlt, styleStatsLabelPanelAlt, styleStatsCountPanelAlt
		}
		name := "   " + truncate(r.name, nameW-tblNameIndent)
		sb.WriteString(pad)
		sb.WriteString(label.Render(padRight(name, nameW)))
		if compact {
			tok := formatTokensMerged(r.inputTokens, r.outTokens)
			sb.WriteString(count.Render(g + fmt.Sprintf("%*d", tblSessWC, r.sessions)))
			sb.WriteString(count.Render(g + fmt.Sprintf("%*d", tblTurnWC, r.turns)))
			sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblTokW, tok)))
			sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblTimeWC, formatDurationMS(r.durationMS))))
			sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblCostWC, fmtCost(r.cost))))
		} else {
			sb.WriteString(count.Render(g + fmt.Sprintf("%*d", tblSessW, r.sessions)))
			sb.WriteString(count.Render(g + fmt.Sprintf("%*d", tblTurnW, r.turns)))
			sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblInW, formatTokens(r.inputTokens))))
			sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblOutW, formatTokens(r.outTokens))))
			sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblTimeW, formatDurationMS(r.durationMS))))
			sb.WriteString(sp.Render(g + fmt.Sprintf("%*s", tblCostW, fmtCost(r.cost))))
		}
		sb.WriteString(sp.Render(" "))
		sb.WriteString("\n")
	}
	return sb.String()
}

// ── Projects table ────────────────────────────────────────────────────────────

// buildProjectRows converts []ProjectStat to []statsTableRow, with sub-rows.
func buildProjectRows(projects []ProjectStat) []statsTableRow {
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
				turns:       ms.Turns,
				inputTokens: ms.InputTokens,
				outTokens:   ms.OutputTokens,
				durationMS:  ms.DurationMS,
				cost:        modelCost(ms.Name, ms.InputTokens, ms.OutputTokens),
			}
		}
		// Project-level cost is the sum of sub-row costs. If there are no
		// sub-rows (no model breakdown), fall back to cost with an empty name
		// (which returns 0 — unknown model).
		var totalCost float64
		for _, sr := range subRows {
			totalCost += sr.cost
		}
		rows[i] = statsTableRow{
			name:        dir,
			sessions:    ps.Sessions,
			turns:       ps.Turns,
			inputTokens: ps.InputTokens,
			outTokens:   ps.OutputTokens,
			durationMS:  ps.DurationMS,
			cost:        totalCost,
			subRows:     subRows,
		}
	}
	return rows
}

// renderProjectHeader renders the column header row for the projects table.
func renderProjectHeader(pad string, nameW int, compact bool) string {
	hdr := styleStatsHeaderPanel
	g := strings.Repeat(" ", tblGap)
	if compact {
		g = strings.Repeat(" ", tblGapC)
	}
	var sb strings.Builder
	sb.WriteString(pad)
	sb.WriteString(styleSpPanel.Render(padRight("", nameW)))
	if compact {
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblSessWC, "sess")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTurnWC, "trns")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTokW, "tokens ↑/↓")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTimeWC, "time")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblCostWC, "cost")))
	} else {
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblSessW, "sessions")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTurnW, "turns")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblInW, "tokens in")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblOutW, "tokens out")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblTimeW, "time")))
		sb.WriteString(hdr.Render(g + fmt.Sprintf("%*s", tblCostW, "cost")))
	}
	sb.WriteString(styleSpPanel.Render(" "))
	sb.WriteString("\n")
	return sb.String()
}

// renderProjectRows renders all project rows, each followed by dimmed model sub-rows.
func renderProjectRows(rows []statsTableRow, pad string, nameW int, compact bool) string {
	var sb strings.Builder
	g := strings.Repeat(" ", tblGap)
	if compact {
		g = strings.Repeat(" ", tblGapC)
	}
	for _, r := range rows {
		// Project row — standard background.
		// In compact mode show only the last path component (e.g. "myapp" not "~/code/myapp").
		displayName := r.name
		if compact {
			displayName = filepath.Base(r.name)
		}
		name := "   " + truncate(displayName, nameW-tblNameIndent)
		sb.WriteString(pad)
		sb.WriteString(styleStatsLabelPanel.Render(padRight(name, nameW)))
		if compact {
			tok := formatTokensMerged(r.inputTokens, r.outTokens)
			sb.WriteString(styleStatsCountPanel.Render(g + fmt.Sprintf("%*d", tblSessWC, r.sessions)))
			sb.WriteString(styleStatsCountPanel.Render(g + fmt.Sprintf("%*d", tblTurnWC, r.turns)))
			sb.WriteString(styleSpPanel.Render(g + fmt.Sprintf("%*s", tblTokW, tok)))
			sb.WriteString(styleSpPanel.Render(g + fmt.Sprintf("%*s", tblTimeWC, formatDurationMS(r.durationMS))))
			sb.WriteString(styleSpPanel.Render(g + fmt.Sprintf("%*s", tblCostWC, fmtCost(r.cost))))
		} else {
			sb.WriteString(styleStatsCountPanel.Render(g + fmt.Sprintf("%*d", tblSessW, r.sessions)))
			sb.WriteString(styleStatsCountPanel.Render(g + fmt.Sprintf("%*d", tblTurnW, r.turns)))
			sb.WriteString(styleSpPanel.Render(g + fmt.Sprintf("%*s", tblInW, formatTokens(r.inputTokens))))
			sb.WriteString(styleSpPanel.Render(g + fmt.Sprintf("%*s", tblOutW, formatTokens(r.outTokens))))
			sb.WriteString(styleSpPanel.Render(g + fmt.Sprintf("%*s", tblTimeW, formatDurationMS(r.durationMS))))
			sb.WriteString(styleSpPanel.Render(g + fmt.Sprintf("%*s", tblCostW, fmtCost(r.cost))))
		}
		sb.WriteString(styleSpPanel.Render(" "))
		sb.WriteString("\n")

		// Per-model sub-rows: lighter alt background, tblNameIndent + 2 extra spaces for the sub-indent.
		for _, sr := range r.subRows {
			subName := "     " + truncate(sr.name, nameW-tblNameIndent-2)
			sb.WriteString(pad)
			sb.WriteString(styleDimPanelAlt.Render(padRight(subName, nameW)))
			if compact {
				tok := formatTokensMerged(sr.inputTokens, sr.outTokens)
				sb.WriteString(styleDimPanelAlt.Render(g + fmt.Sprintf("%*d", tblSessWC, sr.sessions)))
				sb.WriteString(styleDimPanelAlt.Render(g + fmt.Sprintf("%*d", tblTurnWC, sr.turns)))
				sb.WriteString(styleDimPanelAlt.Render(g + fmt.Sprintf("%*s", tblTokW, tok)))
				sb.WriteString(styleDimPanelAlt.Render(g + fmt.Sprintf("%*s", tblTimeWC, formatDurationMS(sr.durationMS))))
				sb.WriteString(styleDimPanelAlt.Render(g + fmt.Sprintf("%*s", tblCostWC, fmtCost(sr.cost))))
			} else {
				sb.WriteString(styleDimPanelAlt.Render(g + fmt.Sprintf("%*d", tblSessW, sr.sessions)))
				sb.WriteString(styleDimPanelAlt.Render(g + fmt.Sprintf("%*d", tblTurnW, sr.turns)))
				sb.WriteString(styleDimPanelAlt.Render(g + fmt.Sprintf("%*s", tblInW, formatTokens(sr.inputTokens))))
				sb.WriteString(styleDimPanelAlt.Render(g + fmt.Sprintf("%*s", tblOutW, formatTokens(sr.outTokens))))
				sb.WriteString(styleDimPanelAlt.Render(g + fmt.Sprintf("%*s", tblTimeW, formatDurationMS(sr.durationMS))))
				sb.WriteString(styleDimPanelAlt.Render(g + fmt.Sprintf("%*s", tblCostW, fmtCost(sr.cost))))
			}
			sb.WriteString(styleSpPanelAlt.Render(" "))
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
