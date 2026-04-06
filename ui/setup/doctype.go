package setup

import (
	"opendoc/config"
	"opendoc/ui/listutil"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type docTypeOption struct {
	label string
	value string
	desc  string
}

var docTypeOptions = []docTypeOption{
	{label: "Markdown (.md)", value: "md", desc: "Standard Markdown format, widely supported"},
}

type docTypeSaveMsg struct{ err error }

type DocTypeModel struct {
	cfg          *config.Config
	cursor       int
	width        int
	height       int
	status       string
	scrollOffset int
	returnTo     func(*config.Config, int, int) tea.Model
}

func NewDocTypeModel(cfg *config.Config, w, h int, returnTo func(*config.Config, int, int) tea.Model) *DocTypeModel {
	m := &DocTypeModel{cfg: cfg, width: w, height: h, returnTo: returnTo}
	m.updateScroll()
	return m
}

func (m *DocTypeModel) Init() tea.Cmd { return nil }

func (m *DocTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateScroll()

	case docTypeSaveMsg:
		if msg.err != nil {
			m.status = "❌ Failed to save setting"
			return m, nil
		}
		next := m.returnTo(m.cfg, m.width, m.height)
		return next, next.Init()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			next := m.returnTo(m.cfg, m.width, m.height)
			return next, next.Init()

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.updateScroll()
			}

		case "down", "j":
			if m.cursor < len(docTypeOptions)-1 {
				m.cursor++
				m.updateScroll()
			}

		case "enter", " ":
			chosen := docTypeOptions[m.cursor]
			cfg := m.cfg
			value := chosen.value
			return m, func() tea.Msg {
				cfg.DocOutputType = value
				return docTypeSaveMsg{err: cfg.Save()}
			}
		}
	}

	return m, nil
}

func docTypeItems() []listutil.Item {
	items := make([]listutil.Item, len(docTypeOptions))
	for i, opt := range docTypeOptions {
		items[i] = listutil.Item{Title: opt.label, Desc: opt.desc}
	}
	return items
}

func (m *DocTypeModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		docTypeTitleStyle.Render("📝 Documentation Output Type"),
		docTypeHintStyle.Render("↑/↓ navigate  •  enter select  •  esc back"),
	)
}

func (m *DocTypeModel) updateScroll() {
	if m.width == 0 || m.height == 0 {
		return
	}
	const footerH = 2
	availH := m.height - lipgloss.Height(m.renderHeader()) - footerH
	_, cursorLine, cursorItemH := listutil.RenderItems(docTypeItems(), m.cursor)
	listutil.UpdateScroll(&m.scrollOffset, availH, cursorLine, cursorItemH)
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	docTypeTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED")).
				MarginBottom(1)

	docTypeHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				MarginBottom(2)

	docTypeStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				MarginTop(1)
)

func (m *DocTypeModel) View() tea.View {
	header := m.renderHeader()
	const footerH = 2
	availH := m.height - lipgloss.Height(header) - footerH
	list, _, _ := listutil.RenderItems(docTypeItems(), m.cursor)
	centeredList := listutil.RenderList(list, m.scrollOffset, availH, m.width)

	parts := []string{header, centeredList}
	if m.status != "" {
		parts = append(parts, docTypeStatusStyle.Render(m.status))
	}
	content := lipgloss.JoinVertical(lipgloss.Center, parts...)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
