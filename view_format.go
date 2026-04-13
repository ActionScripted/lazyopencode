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

// fmtCommas formats an integer with thousands-separator commas (e.g. 12400 → "12,400").
func fmtCommas(n int) string {
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

// ── Cost ──────────────────────────────────────────────────────────────────────

// modelPricing holds per-token USD prices for a model family.
// Prices are per 1M tokens (input / output). Cache read/write prices are not
// included here since the DB aggregates only input+output for the breakdown
// tables; the fieldset already shows cache tokens separately.
type modelPricing struct {
	inputPer1M  float64
	outputPer1M float64
}

// knownModelPricing maps model ID substrings (matched in order) to pricing.
// Prices are sourced from the Anthropic public pricing page.
// Unknown models return zero cost rather than an error.
var knownModelPricing = []struct {
	substr  string
	pricing modelPricing
}{
	// Claude 4 series
	{"claude-opus-4", modelPricing{inputPer1M: 15.0, outputPer1M: 75.0}},
	{"claude-sonnet-4", modelPricing{inputPer1M: 3.0, outputPer1M: 15.0}},
	// Claude 3.7 series
	{"claude-sonnet-3-7", modelPricing{inputPer1M: 3.0, outputPer1M: 15.0}},
	// Claude 3.5 series
	{"claude-opus-3-5", modelPricing{inputPer1M: 15.0, outputPer1M: 75.0}},
	{"claude-sonnet-3-5", modelPricing{inputPer1M: 3.0, outputPer1M: 15.0}},
	{"claude-haiku-3-5", modelPricing{inputPer1M: 0.8, outputPer1M: 4.0}},
	// Claude 3 series
	{"claude-opus-3", modelPricing{inputPer1M: 15.0, outputPer1M: 75.0}},
	{"claude-sonnet-3", modelPricing{inputPer1M: 3.0, outputPer1M: 15.0}},
	{"claude-haiku-3", modelPricing{inputPer1M: 0.25, outputPer1M: 1.25}},
}

// modelCost returns the estimated USD cost for the given model ID and token
// counts. Returns 0 if the model is not in the pricing table.
func modelCost(name string, inputTokens, outputTokens int) float64 {
	for _, entry := range knownModelPricing {
		if strings.Contains(name, entry.substr) {
			input := float64(inputTokens) / 1_000_000 * entry.pricing.inputPer1M
			output := float64(outputTokens) / 1_000_000 * entry.pricing.outputPer1M
			return input + output
		}
	}
	return 0
}

// fmtCost formats a USD cost as a compact string.
//
//	< $0.01  → "<$0.01"
//	< $1     → "$0.42"
//	< $10    → "$3.50"
//	< $1000  → "$42.10"
//	≥ $1000  → "$1.2K"
func fmtCost(f float64) string {
	if f <= 0 {
		return "—"
	}
	if f < 0.01 {
		return "<$0.01"
	}
	if f < 10 {
		return fmt.Sprintf("$%.2f", f)
	}
	if f < 1000 {
		return fmt.Sprintf("$%.0f", f)
	}
	return fmt.Sprintf("$%.1fK", f/1000)
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
			// Leading padding / empty prefix — always dim.
			sb.WriteString(styleDim.Render(seg))
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
