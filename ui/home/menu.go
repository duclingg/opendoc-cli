package home

import (
	"fmt"

	"opendoc/config"
	"opendoc/ui/listutil"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type MenuModel struct {
	cfg          *config.Config
	items        []listutil.Item
	cursor       int
	width        int
	height       int
	status       string
	scrollOffset int
}

func NewMenuModel(cfg *config.Config, w, h int) *MenuModel {
	m := &MenuModel{
		cfg:    cfg,
		items:  buildMenuItems(cfg),
		width:  w,
		height: h,
	}
	m.updateScroll()
	return m
}

func buildMenuItems(_ *config.Config) []listutil.Item {
	return []listutil.Item{
		{Title: "Settings", Desc: "Manage app configuration and re-setup options"},
		{Title: "Quit", Desc: "Exit the application"},
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
		m.updateScroll()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.updateScroll()
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				m.updateScroll()
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

func (m *MenuModel) renderStatusBar() string {
	dot := func(ok bool) string {
		if ok {
			return statusDotOnStyle.Render("●")
		}
		return statusDotOffStyle.Render("●")
	}

	ghOk := m.cfg.IsRegistered()
	ghVal := dot(ghOk) + " " + statusLabelStyle.Render("GitHub:") + " "
	if ghOk {
		ghVal += statusValueStyle.Render("@" + m.cfg.GitHubLogin)
	} else {
		ghVal += statusWarnStyle.Render("not connected")
	}

	llmOk := m.cfg.IsLLMSetUp()
	llmVal := dot(llmOk) + " " + statusLabelStyle.Render("Model:") + " "
	if llmOk {
		llmVal += statusValueStyle.Render(m.cfg.LLMModel)
	} else if m.cfg.LLMProvider != "" {
		llmVal += statusWarnStyle.Render("no model selected")
	} else {
		llmVal += statusWarnStyle.Render("not configured")
	}

	inner := fmt.Sprintf("%s  %s", ghVal, llmVal)
	return lipgloss.NewStyle().MarginBottom(1).Render(
		statusBarStyle.Render(inner),
	)
}

func (m *MenuModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		menuTitleStyle.Render("📄 opendoc cli"),
		menuHintStyle.Render("↑/↓ navigate  •  enter select  •  q quit"),
		m.renderStatusBar(),
	)
}

func (m *MenuModel) updateScroll() {
	if m.width == 0 || m.height == 0 {
		return
	}
	const footerH = 2
	availH := m.height - lipgloss.Height(m.renderHeader()) - footerH
	_, cursorLine, cursorItemH := listutil.RenderItems(m.items, m.cursor)
	listutil.UpdateScroll(&m.scrollOffset, availH, cursorLine, cursorItemH)
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	menuTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1)

	menuHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginBottom(0)

	statusBarStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			PaddingTop(0).
			PaddingBottom(0).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151"))

	statusLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280"))

	statusValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5E7EB")).
				Bold(true)

	statusDotOnStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981"))

	statusDotOffStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B"))

	statusWarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			MarginTop(1)
)

func (m *MenuModel) View() tea.View {
	header := m.renderHeader()
	const footerH = 2
	availH := m.height - lipgloss.Height(header) - footerH
	list, _, _ := listutil.RenderItems(m.items, m.cursor)
	centeredList := listutil.RenderList(list, m.scrollOffset, availH, m.width)

	parts := []string{header, centeredList}
	if m.status != "" {
		parts = append(parts, statusStyle.Render(m.status))
	}
	content := lipgloss.JoinVertical(lipgloss.Center, parts...)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
