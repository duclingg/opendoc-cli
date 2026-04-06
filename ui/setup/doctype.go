package setup

import (
	"opendoc/config"
	"opendoc/ui/styles"

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
	cfg      *config.Config
	cursor   int
	width    int
	height   int
	status   string
	returnTo func(*config.Config, int, int) tea.Model
}

func NewDocTypeModel(cfg *config.Config, w, h int, returnTo func(*config.Config, int, int) tea.Model) *DocTypeModel {
	return &DocTypeModel{cfg: cfg, width: w, height: h, returnTo: returnTo}
}

func (m *DocTypeModel) Init() tea.Cmd { return nil }

func (m *DocTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

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
			}

		case "down", "j":
			if m.cursor < len(docTypeOptions)-1 {
				m.cursor++
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

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	docTypeTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED")).
				MarginBottom(1).
				PaddingLeft(2)

	docTypeHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				PaddingLeft(2).
				MarginBottom(2)

	docTypeStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				PaddingLeft(2).
				MarginTop(1)

	docTypeContainerStyle = lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center)
)

func (m *DocTypeModel) View() tea.View {
	var rows []string

	rows = append(rows, docTypeTitleStyle.Render("📝 Documentation Output Type"))
	rows = append(rows, docTypeHintStyle.Render("↑/↓ navigate  •  enter select  •  esc back"))

	for i, opt := range docTypeOptions {
		var row string
		if i == m.cursor {
			title := styles.ItemTitleSelected.Render(opt.label)
			desc := styles.ItemDescStyle.Render(opt.desc)
			row = styles.ItemSelected.Render(title + "\n" + desc)
		} else {
			title := styles.ItemTitleNormal.Render(opt.label)
			desc := styles.ItemDescStyle.Render(opt.desc)
			row = styles.ItemNormal.Render(title + "\n" + desc)
		}
		rows = append(rows, row)
	}

	if m.status != "" {
		rows = append(rows, docTypeStatusStyle.Render(m.status))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	v := tea.NewView(docTypeContainerStyle.
		Width(m.width).
		Height(m.height).
		Render(content))
	v.AltScreen = true
	return v
}
