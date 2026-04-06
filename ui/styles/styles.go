package styles

import "charm.land/lipgloss/v2"

var (
	// TitleStyle is the shared heading style used across all screens: bold,
	// purple, with a one-line bottom margin.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1)

	// HintStyle is the shared keyboard-hint style used across most screens:
	// muted gray with a two-line bottom margin.
	HintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginBottom(2)

	// ItemNormal is the container style for a list row that is not focused.
	// The extra left padding mirrors the one-char border used by ItemSelected
	// so that unselected rows stay visually aligned.
	ItemNormal = lipgloss.NewStyle().
			PaddingLeft(3).
			PaddingRight(3).
			PaddingTop(1).
			PaddingBottom(1)

	// ItemSelected is the container style for the focused list row. It adds a
	// left purple border to draw the eye without using a background fill.
	ItemSelected = lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(3).
			PaddingTop(1).
			PaddingBottom(1).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("#7C3AED"))

	// ItemTitleNormal is the text style for an unselected item's title.
	ItemTitleNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	// ItemTitleSelected is the text style for the focused item's title.
	ItemTitleSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED"))

	// ItemDescStyle is the secondary description line rendered below each
	// item title, used for both selected and unselected rows.
	ItemDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))
)
