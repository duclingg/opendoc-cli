// Package listutil provides shared helpers for rendering scrollable item lists
// in the Bubble Tea TUI screens. All list-based screens use these functions so
// that scroll behaviour, item styling, and viewport clipping stay consistent.
package listutil

import (
	"opendoc/ui/lineutil"
	"opendoc/ui/styles"

	"charm.land/lipgloss/v2"
)

// Item is a standard list entry with a title and a description line.
type Item struct {
	Title string
	Desc  string
}

// RenderItems renders a slice of Items using the standard selected/normal
// styles. It returns the full joined list string, the first terminal line of
// the focused item, and the focused item's height — all in a single pass so
// the list is never rendered twice.
func RenderItems(items []Item, cursor int) (list string, cursorLine int, cursorItemH int) {
	rows := make([]string, 0, len(items))
	lineCount := 0

	for i, item := range items {
		if i == cursor {
			cursorLine = lineCount
		}

		var row string
		if i == cursor {
			title := styles.ItemTitleSelected.Render(item.Title)
			desc := styles.ItemDescStyle.Render(item.Desc)
			row = styles.ItemSelected.Render(title + "\n" + desc)
		} else {
			title := styles.ItemTitleNormal.Render(item.Title)
			desc := styles.ItemDescStyle.Render(item.Desc)
			row = styles.ItemNormal.Render(title + "\n" + desc)
		}
		rows = append(rows, row)
		h := lipgloss.Height(row)
		if i == cursor {
			cursorItemH = h
		}
		lineCount += h
	}

	list = lipgloss.JoinVertical(lipgloss.Left, rows...)
	return list, cursorLine, cursorItemH
}

// UpdateScroll adjusts *scrollOffset so that the focused item occupying
// [cursorLine, cursorLine+cursorItemH) stays within the visible window
// [scrollOffset, scrollOffset+availH).
func UpdateScroll(scrollOffset *int, availH, cursorLine, cursorItemH int) {
	if cursorLine < *scrollOffset {
		*scrollOffset = cursorLine
	}
	if cursorLine+cursorItemH > *scrollOffset+availH {
		*scrollOffset = cursorLine + cursorItemH - availH
	}
	if *scrollOffset < 0 {
		*scrollOffset = 0
	}
}

// RenderList clips the rendered list string to the visible viewport window
// then horizontally centers it within totalWidth. This is the standard pattern
// used in every list-based View().
func RenderList(list string, scrollOffset, availH, totalWidth int) string {
	clipped := lineutil.ClipLines(list, scrollOffset, availH)
	return lipgloss.NewStyle().Width(totalWidth).Align(lipgloss.Center).Render(clipped)
}

// ContentWidth caps the usable content width at 80 columns so that wide
// terminals don't stretch list items uncomfortably.
func ContentWidth(screenWidth int) int {
	const maxWidth = 80
	if screenWidth < maxWidth {
		return screenWidth
	}
	return maxWidth
}
