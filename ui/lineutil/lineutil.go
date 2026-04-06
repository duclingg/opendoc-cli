// Package lineutil provides low-level line-clipping helpers used by the
// scrollable list viewport logic.
package lineutil

import "strings"

// ClipLines returns the lines [offset, offset+height) from a newline-joined
// string, which is the format produced by lipgloss.JoinVertical. Out-of-range
// indices are clamped gracefully so callers need no boundary checks.
func ClipLines(s string, offset, height int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if offset < 0 {
		offset = 0
	}
	if offset >= len(lines) {
		return ""
	}
	end := offset + height
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[offset:end], "\n")
}
