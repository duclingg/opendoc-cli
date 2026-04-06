package setup

import (
	"opendoc/config"
	"opendoc/ui/lineutil"
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

func (m *DocTypeModel) contentWidth() int {
	const maxWidth = 80
	if m.width < maxWidth {
		return m.width
	}
	return maxWidth
}

func (m *DocTypeModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		docTypeTitleStyle.Render("📝 Documentation Output Type"),
		docTypeHintStyle.Render("↑/↓ navigate  •  enter select  •  esc back"),
	)
}

func (m *DocTypeModel) buildListContent() (string, int) {
	var rows []string
	lineCount := 0
	cursorLine := 0

	for i, opt := range docTypeOptions {
		if i == m.cursor {
			cursorLine = lineCount
		}

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
		lineCount += lipgloss.Height(row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...), cursorLine
}

func (m *DocTypeModel) cursorItemHeight() int {
	opt := docTypeOptions[m.cursor]
	title := styles.ItemTitleSelected.Render(opt.label)
	desc := styles.ItemDescStyle.Render(opt.desc)
	return lipgloss.Height(styles.ItemSelected.Render(title + "\n" + desc))
}

func (m *DocTypeModel) updateScroll() {
	if m.width == 0 || m.height == 0 {
		return
	}
	header := m.renderHeader()
	const footerH = 2
	availH := m.height - lipgloss.Height(header) - footerH
	_, cursorLine := m.buildListContent()
	cursorItemH := m.cursorItemHeight()
	if cursorLine < m.scrollOffset {
		m.scrollOffset = cursorLine
	}
	if cursorLine+cursorItemH > m.scrollOffset+availH {
		m.scrollOffset = cursorLine + cursorItemH - availH
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
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
	list, _ := m.buildListContent()
	clipped := lineutil.ClipLines(list, m.scrollOffset, availH)
	centeredList := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(clipped)
	parts := []string{header, centeredList}
	if m.status != "" {
		parts = append(parts, docTypeStatusStyle.Render(m.status))
	}
	content := lipgloss.JoinVertical(lipgloss.Center, parts...)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
