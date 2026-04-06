package setup

import (
	"os"
	"path/filepath"
	"strings"

	"opendoc/config"
	"opendoc/ui/styles"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// docPathSaveMsg carries the result of persisting the chosen output path.
type docPathSaveMsg struct{ err error }

// DocPathModel collects the filesystem path where generated docs will be saved.
type DocPathModel struct {
	cfg      *config.Config
	input    textinput.Model
	width    int
	height   int
	status   string
	returnTo func(*config.Config, int, int) tea.Model
}

// NewDocPathModel constructs the path input, pre-filled with any previously
// saved path or the default ~/Documents/opendocs.
func NewDocPathModel(cfg *config.Config, w, h int, returnTo func(*config.Config, int, int) tea.Model) *DocPathModel {
	ti := textinput.New()
	ti.Placeholder = "~/Documents/opendocs"
	ti.CharLimit = 256

	defaultPath := "~/Documents/opendocs"
	if cfg.DocOutputPath != "" {
		defaultPath = cfg.DocOutputPath
	}
	ti.SetValue(defaultPath)
	ti.Focus()

	return &DocPathModel{
		cfg:      cfg,
		input:    ti,
		width:    w,
		height:   h,
		returnTo: returnTo,
	}
}

// Init starts the text-input blink cursor.
func (m *DocPathModel) Init() tea.Cmd {
	return textinput.Blink
}

// expandPath expands a leading ~/ to the user's home directory. If the home
// directory cannot be determined the path is returned unchanged.
func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

// Update handles window resizing, the async save result, and keyboard input.
// Unhandled messages are forwarded to the text input component.
func (m *DocPathModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case docPathSaveMsg:
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

		case "enter":
			raw := strings.TrimSpace(m.input.Value())
			if raw == "" {
				m.status = "Path cannot be empty"
				return m, nil
			}
			expanded := expandPath(raw)
			cfg := m.cfg
			return m, func() tea.Msg {
				cfg.DocOutputPath = expanded
				return docPathSaveMsg{err: cfg.Save()}
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	docPathBodyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				MarginBottom(1)

	docPathInputBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7C3AED")).
				Padding(0, 1).
				MarginBottom(1)

	docPathStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				MarginTop(1)
)

// View renders the path input form: title, body text, a labeled text input,
// and an optional status message.
func (m *DocPathModel) View() tea.View {
	rows := []string{
		styles.TitleStyle.Render("📁 Documentation Output Path"),
		docPathBodyStyle.Render("Where should generated documentation be saved?"),
		styles.HintStyle.Render("enter confirm  •  esc back"),
		docPathInputBoxStyle.Render(m.input.View()),
	}
	if m.status != "" {
		rows = append(rows, docPathStatusStyle.Render(m.status))
	}

	content := lipgloss.JoinVertical(lipgloss.Center, rows...)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
