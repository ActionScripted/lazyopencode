package main

import (
	"math"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// formatDuration
// ---------------------------------------------------------------------------

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0s"},
		{30 * time.Second, "30s"},
		{59*time.Second + 999*time.Millisecond, "59s"},
		{time.Minute, "1m"},
		{90 * time.Second, "1m"},
		{45 * time.Minute, "45m"},
		{time.Hour, "1h"},
		{time.Hour + 30*time.Minute, "1h 30m"},
		{2*time.Hour + 5*time.Minute, "2h 5m"},
		{-time.Minute, "0s"}, // negative clamped to 0
	}
	for _, tc := range tests {
		got := formatDuration(tc.d)
		if got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// formatTokens
// ---------------------------------------------------------------------------

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1_000, "1.0K"},
		{1_500, "1.5K"},
		{999_999, "1000.0K"},
		{1_000_000, "1.0M"},
		{1_234_567, "1.2M"},
	}
	for _, tc := range tests {
		got := formatTokens(tc.n)
		if got != tc.want {
			t.Errorf("formatTokens(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// truncate
// ---------------------------------------------------------------------------

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		maxW int
		want string
	}{
		{"hello", 10, "hello"},         // shorter than limit — unchanged
		{"hello", 5, "hello"},          // exactly at limit — unchanged
		{"hello world", 8, "hello w…"}, // truncated with ellipsis
		{"hello", 0, ""},               // zero width — empty
		{"hello", -1, ""},              // negative — empty
		{"", 10, ""},                   // empty input
	}
	for _, tc := range tests {
		got := truncate(tc.s, tc.maxW)
		if got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.s, tc.maxW, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// renderHintSegments
// ---------------------------------------------------------------------------

func TestRenderHintSegments_ContainsKeyAndDesc(t *testing.T) {
	// The function applies lipgloss styles, so we strip ANSI and check the
	// plain-text content is present.
	raw := "  j/k: up/down   enter: open   q: quit"
	got := renderHintSegments(raw)
	// Strip ANSI escape codes with a basic check: rendered output should
	// contain the key and description text somewhere.
	plain := stripANSI(got)
	for _, want := range []string{"j/k", "up/down", "enter", "open", "q", "quit"} {
		if !strings.Contains(plain, want) {
			t.Errorf("renderHintSegments output missing %q; plain text: %q", want, plain)
		}
	}
}

func TestRenderHintSegments_BareSegmentNoColon(t *testing.T) {
	raw := "  type to filter"
	got := renderHintSegments(raw)
	plain := stripANSI(got)
	if !strings.Contains(plain, "type to filter") {
		t.Errorf("bare segment not present in output; plain text: %q", plain)
	}
}

// stripANSI removes ANSI/VT100 escape sequences from s for plain-text
// assertions. It handles both CSI sequences (ESC [ ... <final>) and simple
// two-byte ESC sequences (ESC <char>).
func stripANSI(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] != '\x1b' {
			b.WriteRune(runes[i])
			continue
		}
		// ESC: peek at the next character.
		i++
		if i >= len(runes) {
			break
		}
		if runes[i] == '[' {
			// CSI sequence: consume until a byte in 0x40–0x7E (the final byte).
			i++
			for i < len(runes) && (runes[i] < 0x40 || runes[i] > 0x7E) {
				i++
			}
			// i now points at the final byte; the outer loop will i++ past it.
		}
		// Any other ESC <char> pair is consumed by the outer i++ above.
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// formatWorkspaceRow
// ---------------------------------------------------------------------------

func TestFormatWorkspaceRow_ContainsDir(t *testing.T) {
	got := formatWorkspaceRow("~/projects/myapp", 40, false)
	plain := stripANSI(got)
	if !strings.Contains(plain, "myapp") {
		t.Errorf("formatWorkspaceRow output missing dir; plain: %q", plain)
	}
}

func TestFormatWorkspaceRow_TruncatesLongDir(t *testing.T) {
	long := "~/projects/" + strings.Repeat("a", 50)
	got := formatWorkspaceRow(long, 20, false)
	plain := stripANSI(got)
	// The plain text width should not exceed the column width.
	if len([]rune(plain)) > 22 { // 20 + some slack for lead/trail spaces
		t.Errorf("formatWorkspaceRow did not truncate; plain: %q", plain)
	}
	if !strings.Contains(got, "…") {
		t.Errorf("expected ellipsis in truncated row; got: %q", plain)
	}
}

func TestFormatWorkspaceRow_SelectedContainsDir(t *testing.T) {
	// Lipgloss strips ANSI in non-TTY environments so we can't assert on color
	// differences, but we can verify the selected row still contains the dir.
	got := formatWorkspaceRow("~/projects/myapp", 40, true)
	plain := stripANSI(got)
	if !strings.Contains(plain, "myapp") {
		t.Errorf("selected formatWorkspaceRow missing dir; plain: %q", plain)
	}
}

// ---------------------------------------------------------------------------
// formatWorkspaceSessionRow
// ---------------------------------------------------------------------------

func TestFormatWorkspaceSessionRow_ContainsDateAndTitle(t *testing.T) {
	s := session{
		Title:     "My session",
		UpdatedAt: time.Date(2024, 3, 15, 9, 30, 0, 0, time.UTC),
	}
	got := formatWorkspaceSessionRow(s, 60)
	plain := stripANSI(got)
	if !strings.Contains(plain, "2024-03-15") {
		t.Errorf("expected date in row; plain: %q", plain)
	}
	if !strings.Contains(plain, "My session") {
		t.Errorf("expected title in row; plain: %q", plain)
	}
}

func TestFormatWorkspaceSessionRow_TruncatesLongTitle(t *testing.T) {
	s := session{
		Title:     strings.Repeat("x", 100),
		UpdatedAt: time.Now(),
	}
	got := formatWorkspaceSessionRow(s, 40)
	if !strings.Contains(got, "…") {
		t.Errorf("expected ellipsis for long title; plain: %q", stripANSI(got))
	}
}

// ---------------------------------------------------------------------------
// formatSessionRow
// ---------------------------------------------------------------------------

func TestFormatSessionRow(t *testing.T) {
	s := session{
		Title:     "Fix the login bug",
		ShortDir:  "myapp",
		UpdatedAt: time.Date(2024, 6, 1, 14, 30, 0, 0, time.UTC),
	}

	tests := []struct {
		name      string
		width     int
		pathColW  int
		wantDate  string // expected date fragment in plain text; "" = not shown
		wantPath  bool   // whether path column should appear
		wantTitle string // fragment that must appear in title area
	}{
		{
			name:      "wide layout - full date and path",
			width:     wideLayoutBreakpoint + 20, // >= 120
			pathColW:  10,
			wantDate:  "2024-06-01",
			wantPath:  true,
			wantTitle: "Fix the login bug",
		},
		{
			name:      "medium layout - short date and path",
			width:     defaultTermW + 10, // >= 80, < 120
			pathColW:  10,
			wantDate:  "Jun 01",
			wantPath:  true,
			wantTitle: "Fix the login bug",
		},
		{
			name:      "narrow layout - no date no path",
			width:     defaultTermW - 10, // < 80
			pathColW:  10,
			wantDate:  "",
			wantPath:  false,
			wantTitle: "Fix",
		},
		{
			name:      "long title truncated",
			width:     defaultTermW,
			pathColW:  10,
			wantDate:  "Jun 01",
			wantPath:  true,
			wantTitle: "Fix",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatSessionRow(s, tc.width, tc.pathColW, false)
			plain := stripANSI(got)

			if tc.wantDate != "" && !strings.Contains(plain, tc.wantDate) {
				t.Errorf("expected date %q in row; plain: %q", tc.wantDate, plain)
			}
			if tc.wantDate == "" && (strings.Contains(plain, "2024") || strings.Contains(plain, "Jun")) {
				t.Errorf("expected no date in narrow row; plain: %q", plain)
			}
			if tc.wantPath && !strings.Contains(plain, "myapp") {
				t.Errorf("expected path %q in row; plain: %q", "myapp", plain)
			}
			if !tc.wantPath && strings.Contains(plain, "myapp") {
				t.Errorf("expected no path in narrow row; plain: %q", plain)
			}
			if !strings.Contains(plain, tc.wantTitle) {
				t.Errorf("expected title fragment %q in row; plain: %q", tc.wantTitle, plain)
			}
		})
	}
}

func TestFormatSessionRow_TitleMinWidth(t *testing.T) {
	// Even at extreme narrowness titleW should floor at 1, not go negative.
	s := session{Title: "X", ShortDir: "y", UpdatedAt: time.Now()}
	got := formatSessionRow(s, 1, 0, false)
	if got == "" {
		t.Error("expected non-empty row even at width=1")
	}
}

// ---------------------------------------------------------------------------
// formatCommas
// ---------------------------------------------------------------------------

func TestFormatCommas(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{12400, "12,400"},
		{1000000, "1,000,000"},
		{1234567, "1,234,567"},
	}
	for _, tc := range tests {
		got := formatCommas(tc.n)
		if got != tc.want {
			t.Errorf("formatCommas(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// formatDurationMS
// ---------------------------------------------------------------------------

func TestFormatDurationMS(t *testing.T) {
	tests := []struct {
		ms   int64
		want string
	}{
		{0, "0s"},
		{30000, "30s"},      // 30 seconds
		{90000, "1m"},       // 90 seconds
		{3600000, "1h"},     // 1 hour
		{5400000, "1h 30m"}, // 1.5 hours
	}
	for _, tc := range tests {
		got := formatDurationMS(tc.ms)
		if got != tc.want {
			t.Errorf("formatDurationMS(%d) = %q, want %q", tc.ms, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// modelCost
// ---------------------------------------------------------------------------

func TestModelCost(t *testing.T) {
	tests := []struct {
		name         string
		model        string
		inputTokens  int
		outputTokens int
		want         float64
	}{
		{"sonnet-4 input price", "claude-sonnet-4-6", 1_000_000, 0, 3.0},
		{"opus-4 output price", "claude-opus-4-6", 0, 1_000_000, 75.0},
		{"unknown model returns zero", "gpt-4o", 1_000_000, 1_000_000, 0},
		{"zero tokens returns zero", "claude-sonnet-4-6", 0, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := modelCost(tc.model, tc.inputTokens, tc.outputTokens)
			if math.Abs(got-tc.want) >= 0.001 {
				t.Errorf("modelCost(%q, %d, %d) = %f, want %f",
					tc.model, tc.inputTokens, tc.outputTokens, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// formatCost
// ---------------------------------------------------------------------------

func TestFormatCost(t *testing.T) {
	tests := []struct {
		f    float64
		want string
	}{
		{0, "—"},
		{-1.0, "—"},
		{0.001, "<$0.01"},
		{0.42, "$0.42"},
		{3.5, "$3.50"},
		{42.1, "$42"},
		{1500.0, "$1.5K"},
	}
	for _, tc := range tests {
		got := formatCost(tc.f)
		if got != tc.want {
			t.Errorf("formatCost(%v) = %q, want %q", tc.f, got, tc.want)
		}
	}
}
