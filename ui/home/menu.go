package home

import (
	"opendoc/config"
	"opendoc/ui/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type menuItem struct {
	title string
	desc  string
}

type MenuModel struct {
	cfg    *config.Config
	items  []menuItem
	cursor int
	width  int
	height int
	status string
}

func NewMenuModel(cfg *config.Config, w, h int) *MenuModel {
	return &MenuModel{
		cfg:    cfg,
		items:  buildMenuItems(cfg),
		width:  w,
		height: h,
	}
}

func buildMenuItems(_ *config.Config) []menuItem {
	return []menuItem{
		{title: "Settings", desc: "Manage app configuration and re-setup options"},
		{title: "Quit", desc: "Exit the application"},
	}
}

func (m *MenuModel) Init() tea.Cmd {
	return nil
}

func (m *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			return m.handleSelect()
		}
	}

	return m, nil
}

func (m *MenuModel) handleSelect() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case 0: // Settings
		next := NewSettingsModel(m.cfg, m.width, m.height)
		return next, next.Init()

	case 1: // Quit
		return m, tea.Quit
	}
	return m, nil
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	menuTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1).
			PaddingLeft(2)

	menuHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			PaddingLeft(2).
			MarginBottom(2)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			PaddingLeft(2).
			MarginTop(1)

	menuContainerStyle = lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center)
)

func (m *MenuModel) View() tea.View {
	var rows []string

	rows = append(rows, menuTitleStyle.Render("📄 OpenDoc"))
	rows = append(rows, menuHintStyle.Render("↑/↓ navigate  •  enter select  •  q quit"))

	for i, item := range m.items {
		var row string
		if i == m.cursor {
			title := styles.ItemTitleSelected.Render(item.title)
			desc := styles.ItemDescStyle.Render(item.desc)
			row = styles.ItemSelected.Render(title + "\n" + desc)
		} else {
			title := styles.ItemTitleNormal.Render(item.title)
			desc := styles.ItemDescStyle.Render(item.desc)
			row = styles.ItemNormal.Render(title + "\n" + desc)
		}
		rows = append(rows, row)
	}

	if m.status != "" {
		rows = append(rows, statusStyle.Render(m.status))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	v := tea.NewView(menuContainerStyle.
		Width(m.width).
		Height(m.height).
		Render(content))
	v.AltScreen = true
	return v
}
