package home

import (
	"fmt"

	"opendoc/config"
	"opendoc/ui/listutil"
	"opendoc/ui/setup"
	"opendoc/ui/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type confirmState int

const (
	confirmNone confirmState = iota
	confirmReset
)

type resetDoneMsg struct{ err error }

type SettingsModel struct {
	cfg          *config.Config
	cursor       int
	width        int
	height       int
	confirm      confirmState
	status       string
	scrollOffset int
}

var settingsItems = []listutil.Item{
	{Title: "Re-connect GitHub", Desc: "Re-authorize with your GitHub account or organization"},
	{Title: "Change LLM Provider", Desc: "Update your AI provider configuration"},
	{Title: "Change Doc Output Type", Desc: "Update the format for generated documentation"},
	{Title: "Change Doc Output Path", Desc: "Update the directory where documentation will be written"},
	{Title: "Reset All Settings", Desc: "Wipe all persisted configuration data"},
	{Title: "← Back", Desc: "Return to the main menu"},
}

func NewSettingsModel(cfg *config.Config, w, h int) *SettingsModel {
	m := &SettingsModel{cfg: cfg, width: w, height: h}
	m.updateScroll()
	return m
}

func (m *SettingsModel) Init() tea.Cmd { return nil }

func (m *SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateScroll()

	case resetDoneMsg:
		m.confirm = confirmNone
		if msg.err != nil {
			m.status = fmt.Sprintf("❌ Error: %v", msg.err)
			return m, nil
		}
		next := setup.NewSetupModel(m.cfg, m.width, m.height, func(c *config.Config, w, h int) tea.Model {
			return NewMenuModel(c, w, h)
		})
		return next, next.Init()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.confirm != confirmNone {
				m.confirm = confirmNone
				m.status = ""
				return m, nil
			}
			next := NewMenuModel(m.cfg, m.width, m.height)
			return next, next.Init()

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.updateScroll()
			}

		case "down", "j":
			if m.cursor < len(settingsItems)-1 {
				m.cursor++
				m.updateScroll()
			}

		case "enter", " ":
			return m.handleSelect()

		case "y", "Y":
			if m.confirm != confirmNone {
				return m, m.doReset()
			}

		case "n", "N":
			m.confirm = confirmNone
			m.status = "Cancelled"
		}
	}

	return m, nil
}

func (m *SettingsModel) settingsReturnTo() func(*config.Config, int, int) tea.Model {
	return func(c *config.Config, w, h int) tea.Model { return NewSettingsModel(c, w, h) }
}

func (m *SettingsModel) handleSelect() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case 0: // re-connect GitHub
		clientID, err := config.ResolveClientID()
		if err != nil {
			m.status = fmt.Sprintf("❌ %v", err)
			return m, nil
		}
		next := setup.NewDeviceAuthModel(m.cfg, clientID, m.width, m.height, m.settingsReturnTo())
		return next, next.Init()

	case 1: // change LLM provider
		next := setup.NewLLMSetupModel(m.cfg, m.width, m.height, m.settingsReturnTo())
		return next, next.Init()

	case 2: // change doc output type
		next := setup.NewDocTypeModel(m.cfg, m.width, m.height, m.settingsReturnTo())
		return next, next.Init()

	case 3: // change doc output path
		next := setup.NewDocPathModel(m.cfg, m.width, m.height, m.settingsReturnTo())
		return next, next.Init()

	case 4: // reset all settings
		m.confirm = confirmReset
		m.status = ""

	case 5: // back
		next := NewMenuModel(m.cfg, m.width, m.height)
		return next, next.Init()
	}
	return m, nil
}

func (m *SettingsModel) doReset() tea.Cmd {
	return func() tea.Msg {
		err := m.cfg.Reset()
		return resetDoneMsg{err: err}
	}
}

func (m *SettingsModel) renderHeader() string {
	var rows []string
	rows = append(rows, settingsTitleStyle.Render("⚙️  Settings"))
	rows = append(rows, settingsHintStyle.Render("↑/↓ navigate  •  enter select  •  esc back"))
	if m.cfg.IsRegistered() {
		info := styles.ItemDescStyle.Render(fmt.Sprintf("GitHub: %s", m.cfg.GitHubLogin))
		rows = append(rows, lipgloss.NewStyle().MarginBottom(1).Render(info))
	}
	return lipgloss.JoinVertical(lipgloss.Center, rows...)
}

func (m *SettingsModel) updateScroll() {
	if m.width == 0 || m.height == 0 {
		return
	}
	const footerH = 2
	availH := m.height - lipgloss.Height(m.renderHeader()) - footerH
	_, cursorLine, cursorItemH := listutil.RenderItems(settingsItems, m.cursor)
	listutil.UpdateScroll(&m.scrollOffset, availH, cursorLine, cursorItemH)
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	settingsTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED")).
				MarginBottom(1)

	settingsHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				MarginBottom(2)

	confirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F59E0B")).
			Padding(1, 2).
			MarginTop(1)

	confirmTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#F59E0B")).
				MarginBottom(1)

	confirmTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5E7EB"))

	settingsStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				MarginTop(1)
)

func (m *SettingsModel) View() tea.View {
	if m.confirm == confirmReset {
		dialog := confirmBoxStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				confirmTitleStyle.Render("⚠️  Confirm Reset"),
				confirmTextStyle.Render("This will wipe all persisted configuration data."),
				confirmTextStyle.Render(""),
				confirmTextStyle.Render("y  confirm  •  n / esc  cancel"),
			),
		)
		v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog))
		v.AltScreen = true
		return v
	}

	header := m.renderHeader()
	const footerH = 2
	availH := m.height - lipgloss.Height(header) - footerH
	list, _, _ := listutil.RenderItems(settingsItems, m.cursor)
	centeredList := listutil.RenderList(list, m.scrollOffset, availH, m.width)

	parts := []string{header, centeredList}
	if m.status != "" {
		parts = append(parts, settingsStatusStyle.Render(m.status))
	}
	content := lipgloss.JoinVertical(lipgloss.Center, parts...)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
