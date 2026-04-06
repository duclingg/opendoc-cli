package home

import (
	"fmt"

	"opendoc/config"
	"opendoc/ui/listutil"
	"opendoc/ui/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// footerH is the number of terminal lines reserved at the bottom of every
// home screen for breathing room.
const footerH = 2

// menuItems is the fixed list of main-menu actions.
var menuItems = []listutil.Item{
	{Title: "Settings", Desc: "Manage app configuration and re-setup options"},
	{Title: "Quit", Desc: "Exit the application"},
}

// MenuModel is the main menu shown after setup completes. It presents a status
// bar with the current GitHub and LLM configuration plus a short action list.
type MenuModel struct {
	cfg          *config.Config
	cursor       int
	width        int
	height       int
	scrollOffset int
}

// NewMenuModel constructs the main menu.
func NewMenuModel(cfg *config.Config, w, h int) *MenuModel {
	m := &MenuModel{
		cfg:    cfg,
		width:  w,
		height: h,
	}
	m.updateScroll()
	return m
}

// Init satisfies tea.Model; no initial commands are needed.
func (m *MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles window resizing and keyboard navigation.
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
			if m.cursor < len(menuItems)-1 {
				m.cursor++
				m.updateScroll()
			}

		case "enter", " ":
			return m.handleSelect()
		}
	}

	return m, nil
}

// handleSelect executes the action for the currently focused menu item.
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

// renderStatusBar builds the compact status pill showing the GitHub login and
// active LLM model. Missing values are highlighted in amber.
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

// renderHeader builds the title, keyboard-hint, and status-bar block shown
// above the menu list.
func (m *MenuModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		styles.TitleStyle.Render("📄 opendoc cli"),
		menuHintStyle.Render("↑/↓ navigate  •  enter select  •  q quit"),
		m.renderStatusBar(),
	)
}

// updateScroll adjusts scrollOffset so the focused row stays visible.
func (m *MenuModel) updateScroll() {
	if m.width == 0 || m.height == 0 {
		return
	}
	availH := m.height - lipgloss.Height(m.renderHeader()) - footerH
	_, cursorLine, cursorItemH := listutil.RenderItems(menuItems, m.cursor)
	listutil.UpdateScroll(&m.scrollOffset, availH, cursorLine, cursorItemH)
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	// menuHintStyle uses zero bottom margin so the status bar sits immediately
	// below the hint without extra spacing.
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
)

// View renders the main menu: header with status bar then the scrollable
// action list.
func (m *MenuModel) View() tea.View {
	header := m.renderHeader()
	availH := m.height - lipgloss.Height(header) - footerH
	list, _, _ := listutil.RenderItems(menuItems, m.cursor)
	centeredList := listutil.RenderList(list, m.scrollOffset, availH, m.width)

	content := lipgloss.JoinVertical(lipgloss.Center, header, centeredList)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
