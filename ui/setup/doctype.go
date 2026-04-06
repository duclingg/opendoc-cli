package setup

import (
	"opendoc/config"
	"opendoc/ui/listutil"
	"opendoc/ui/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// docTypeOption describes a supported documentation output format.
type docTypeOption struct {
	label string
	value string
	desc  string
}

// docTypeOptions is the list of available output formats. Additional formats
// can be added here without changing any other code.
var docTypeOptions = []docTypeOption{
	{label: "Markdown (.md)", value: "md", desc: "Standard Markdown format, widely supported"},
}

// docTypeSaveMsg carries the result of persisting the chosen output type.
type docTypeSaveMsg struct{ err error }

// DocTypeModel lets the user choose the file format for generated docs.
type DocTypeModel struct {
	cfg          *config.Config
	cursor       int
	width        int
	height       int
	status       string
	scrollOffset int
	returnTo     func(*config.Config, int, int) tea.Model
}

// NewDocTypeModel constructs the doc-type picker.
func NewDocTypeModel(cfg *config.Config, w, h int, returnTo func(*config.Config, int, int) tea.Model) *DocTypeModel {
	m := &DocTypeModel{cfg: cfg, width: w, height: h, returnTo: returnTo}
	m.updateScroll()
	return m
}

// Init satisfies tea.Model; no initial commands are needed.
func (m *DocTypeModel) Init() tea.Cmd { return nil }

// Update handles window resizing, the async save result, and keyboard input.
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

// docTypeItems converts docTypeOptions to the listutil.Item format expected by
// the shared rendering helpers.
func docTypeItems() []listutil.Item {
	items := make([]listutil.Item, len(docTypeOptions))
	for i, opt := range docTypeOptions {
		items[i] = listutil.Item{Title: opt.label, Desc: opt.desc}
	}
	return items
}

// renderHeader builds the title and keyboard-hint block shown above the list.
func (m *DocTypeModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		styles.TitleStyle.Render("📝 Documentation Output Type"),
		styles.HintStyle.Render("↑/↓ navigate  •  enter select  •  esc back"),
	)
}

// updateScroll adjusts scrollOffset so the focused row stays visible.
func (m *DocTypeModel) updateScroll() {
	if m.width == 0 || m.height == 0 {
		return
	}
	availH := m.height - lipgloss.Height(m.renderHeader()) - footerH
	_, cursorLine, cursorItemH := listutil.RenderItems(docTypeItems(), m.cursor)
	listutil.UpdateScroll(&m.scrollOffset, availH, cursorLine, cursorItemH)
}

// ─── styles ──────────────────────────────────────────────────────────────────

var docTypeStatusStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#EF4444")).
	MarginTop(1)

// View renders the doc-type picker: header, scrollable format list, and an
// optional status message.
func (m *DocTypeModel) View() tea.View {
	header := m.renderHeader()
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
