package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderWorkspacesView is the top-level renderer for modeWorkspaces. It
// composes a top bar, a side-by-side left/right pane body, and the hint bar.
func (m model) renderWorkspacesView(w, h int) string {
	hint := m.renderHint(w)
	hintH := strings.Count(hint, "\n") + 1
	topBar := renderTopBar(w)
	bodyH := h - hintH - topBarH

	// Secondary (right) pane takes the same share as the preview pane in the
	// main view; primary (left) pane takes the remainder. Mirrors View()'s
	// previewW / listW split but for the workspaces layout.
	rightW := w * previewWidthPct / 100
	if rightW < workspacesRightMinWidth {
		rightW = workspacesRightMinWidth
	}
	leftW := w - rightW

	left := styleListPane.Width(leftW).Height(bodyH).Render(m.renderWorkspaceList(leftW, bodyH))
	right := m.renderWorkspaceSessions(rightW, bodyH)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return topBar + "\n" + body + "\n" + hint
}

// renderWorkspaceList renders the left pane: a scrollable list of workspace
// directories, with the selected row highlighted.
func (m model) renderWorkspaceList(width, height int) string {
	var sb strings.Builder

	// Header separator (mirrors the search bar separator in the sessions view).
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", width)) + "\n")
	height-- // consumed by separator

	if len(m.workspaces) == 0 {
		sb.WriteString(styleDimDark.Render("  no workspaces") + "\n")
		return sb.String()
	}

	visibleStart := 0
	if m.workspaceCursor >= height {
		visibleStart = m.workspaceCursor - height + 1
	}

	for i := visibleStart; i < len(m.workspaces) && i < visibleStart+height; i++ {
		ws := m.workspaces[i]
		selected := i == m.workspaceCursor
		sb.WriteString(formatWorkspaceRow(ws.DisplayDir, width, selected) + "\n")
	}

	return sb.String()
}

// formatWorkspaceRow renders a single workspace directory as a fixed-width row.
// displayDir is the pre-computed display string (with "~" substitution).
func formatWorkspaceRow(displayDir string, width int, selected bool) string {
	const leadSp = 1
	const trailSp = 1

	innerW := width - leadSp - trailSp
	if innerW < 1 {
		innerW = 1
	}
	raw := truncate(displayDir, innerW)
	padded := raw + strings.Repeat(" ", max(0, innerW-lipgloss.Width(raw)))

	base := styleWorkspaceRow
	if selected {
		base = styleWorkspaceRowSel
	}

	lead := base.Foreground(colorFg).Render(strings.Repeat(" ", leadSp))
	text := base.Bold(selected).Render(padded)
	trail := base.Foreground(colorFg).Render(strings.Repeat(" ", trailSp))

	return lead + text + trail
}

// renderWorkspaceSessions renders the right pane: a read-only list of sessions
// for the currently selected workspace.
func (m model) renderWorkspaceSessions(width, height int) string {
	paneW := width - paneHorizBorder  // same overhead as side-by-side preview
	inner := paneW - paneHorizPadding // text area

	if len(m.workspaces) == 0 {
		return stylePreviewPane.Width(paneW).Height(height).Render(
			styleDim.Render("no workspaces"),
		)
	}

	selectedWS := m.workspaces[m.workspaceCursor]

	// Filter sessions belonging to the selected workspace.
	var wsSessions []session
	for _, s := range m.sessions {
		if s.Directory == selectedWS.Dir {
			wsSessions = append(wsSessions, s)
		}
	}

	// ── header ────────────────────────────────────────────────────────────────
	var header strings.Builder
	header.WriteString(stylePreviewTitle.Render(
		lipgloss.NewStyle().MaxWidth(inner).Background(colorBgPanel).Render(selectedWS.DisplayDir),
	))
	header.WriteString("\n")
	header.WriteString(styleDim.Render(
		lipgloss.NewStyle().MaxWidth(inner).Background(colorBgPanel).Render(
			strings.Repeat("─", max(0, inner)),
		),
	))

	// Compute header height dynamically so it never drifts out of sync with
	// the header content (mirrors the approach used in renderPreview).
	headerLines := strings.Count(header.String(), "\n") + 1
	listHeight := height - headerLines
	if listHeight < 1 {
		listHeight = 1
	}

	// ── session rows ──────────────────────────────────────────────────────────
	var sessionRows strings.Builder
	if len(wsSessions) == 0 {
		sessionRows.WriteString(styleSpBgPanel.Render("\n") + styleDim.Render("  no sessions"))
	} else {
		for i, s := range wsSessions {
			if i >= listHeight {
				break
			}
			sessionRows.WriteString(styleSpBgPanel.Render("\n") + formatWorkspaceSessionRow(s, inner))
		}
	}

	return stylePreviewPane.Width(paneW).Height(height).Render(
		header.String() + sessionRows.String(),
	)
}

// formatWorkspaceSessionRow renders a compact session row for the workspaces
// right pane. Layout: date(dateFullWidth) "  " title(remaining). No path
// column needed since all sessions share the same workspace directory.
func formatWorkspaceSessionRow(s session, width int) string {
	titleW := width - dateFullWidth - dateGap
	if titleW < 1 {
		titleW = 1
	}

	date := styleDim.Render(s.UpdatedAt.Format("2006-01-02 15:04"))
	title := styleWorkspaceSessionTitle.Render(truncate(s.Title, titleW))

	return date + styleDim.Render(strings.Repeat(" ", dateGap)) + title
}
