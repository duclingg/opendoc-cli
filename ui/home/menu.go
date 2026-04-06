package home

import (
	"fmt"

	"opendoc/config"
	"opendoc/ui/styles"

	"charm.land/bubbles/v2/viewport"
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
	vp     viewport.Model
}

func NewMenuModel(cfg *config.Config, w, h int) *MenuModel {
	m := &MenuModel{
		cfg:    cfg,
		items:  buildMenuItems(cfg),
		width:  w,
		height: h,
	}
	m.vp = viewport.New()
	m.syncViewport()
	return m
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
		m.syncViewport()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.syncViewport()
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				m.syncViewport()
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
	return lipgloss.NewStyle().PaddingLeft(2).MarginBottom(1).Render(
		statusBarStyle.Render(inner),
	)
}

func (m *MenuModel) contentWidth() int {
	const maxWidth = 80
	if m.width < maxWidth {
		return m.width
	}
	return maxWidth
}

func (m *MenuModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		menuTitleStyle.Render("📄 opendoc cli"),
		menuHintStyle.Render("↑/↓ navigate  •  enter select  •  q quit"),
		m.renderStatusBar(),
	)
}

func (m *MenuModel) buildListContent() (string, int) {
	var rows []string
	lineCount := 0
	cursorLine := 0

	for i, item := range m.items {
		if i == m.cursor {
			cursorLine = lineCount
		}

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
		lineCount += lipgloss.Height(row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...), cursorLine
}

func (m *MenuModel) syncViewport() {
	if m.width == 0 || m.height == 0 {
		return
	}
	header := m.renderHeader()
	headerH := lipgloss.Height(header)
	const footerH = 2 // reserve space for status line
	m.vp.SetWidth(m.contentWidth())
	m.vp.SetHeight(m.height - headerH - footerH)
	content, cursorLine := m.buildListContent()
	m.vp.SetContent(content)
	m.vp.EnsureVisible(cursorLine, 0, 0)
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
			PaddingLeft(2).
			MarginTop(1)
)

func (m *MenuModel) View() tea.View {
	header := m.renderHeader()
	body := m.vp.View()
	parts := []string{header, body}
	if m.status != "" {
		parts = append(parts, statusStyle.Render(m.status))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
