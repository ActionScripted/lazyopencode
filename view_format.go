package main

import (
	"fmt"
	"strings"
	"time"
)

// ── Duration formatting ───────────────────────────────────────────────────────

// formatDuration formats a time.Duration as a compact human-readable string.
//
//	< 1 minute  → "45s"
//	< 1 hour    → "45m"
//	< 1 day     → "2h 5m"   (omits minutes when zero: "3h")
//	≥ 1 day     → "2d 3h"   (omits hours when zero: "3d")
//
// Negative durations are clamped to zero.
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int64(d.Seconds())
	if total < 60 {
		return fmt.Sprintf("%ds", total)
	}
	minutes := total / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	mins := minutes % 60
	if hours < 24 {
		if mins == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	days := hours / 24
	hrs := hours % 24
	if hrs == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd %dh", days, hrs)
}

// formatDurationMS converts a millisecond duration to time.Duration and
// delegates to formatDuration. Used in the stats dashboard where durations
// are stored as integer milliseconds.
func formatDurationMS(ms int64) string {
	return formatDuration(time.Duration(ms) * time.Millisecond)
}

// ── Number formatting ─────────────────────────────────────────────────────────

// formatCommas formats an integer with thousands-separator commas (e.g. 12400 → "12,400").
func formatCommas(n int) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}
	out := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
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

// formatCost formats a dollar cost as "$0.00" with exactly two decimal places.
// It only formats the value opencode stores in $.cost — no calculation is done.
func formatCost(cost float64) string {
	return fmt.Sprintf("$%.2f", cost)
}

// ── Hint bar ──────────────────────────────────────────────────────────────────

// renderHintSegments parses a hint string of the form
// "  key: desc   key: desc   bare" and returns a styled string where each key
// is rendered in styleHintKey (white) and each description and separator is
// rendered in styleDim.
func renderHintSegments(hints string) string {
	// Split on three-or-more spaces so multi-word keys/descs are kept intact.
	segments := strings.Split(hints, "   ")
	var sb strings.Builder
	for i, seg := range segments {
		if i == 0 {
			// First segment may contain leading padding followed by a key:desc
			// pair (e.g. "  j/k: up/down"). Preserve leading spaces as dim
			// then style the key separately if a colon is present.
			// If there's no colon, render the whole segment dim as before.
			// Find leading space count.
			firstNonSpace := 0
			for firstNonSpace < len(seg) && seg[firstNonSpace] == ' ' {
				firstNonSpace++
			}
			prefix := seg[:firstNonSpace]
			rest := seg[firstNonSpace:]
			sb.WriteString(styleDim.Render(prefix))
			if idx := strings.Index(rest, ": "); idx != -1 {
				sb.WriteString(styleHintKey.Render(rest[:idx]))
				sb.WriteString(styleDim.Render(": " + rest[idx+2:]))
			} else {
				sb.WriteString(styleDim.Render(rest))
			}
			continue
		}
		sb.WriteString(styleDim.Render("   "))
		if idx := strings.Index(seg, ": "); idx != -1 {
			sb.WriteString(styleHintKey.Render(seg[:idx]))
			sb.WriteString(styleDim.Render(": " + seg[idx+2:]))
		} else {
			// No colon — treat as description only (e.g. "type to filter").
			sb.WriteString(styleDim.Render(seg))
		}
	}
	return sb.String()
}
