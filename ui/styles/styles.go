package styles

import "charm.land/lipgloss/v2"

var (
	ItemNormal = lipgloss.NewStyle().
			PaddingLeft(4).
			PaddingTop(1).
			PaddingBottom(1)

	ItemSelected = lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingTop(1).
			PaddingBottom(1).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("#7C3AED"))

	ItemTitleNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	ItemTitleSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED"))

	ItemDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))
)
