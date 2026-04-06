package setup

import (
	"os"
	"path/filepath"
	"strings"

	"opendoc/config"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type docPathSaveMsg struct{ err error }

type DocPathModel struct {
	cfg      *config.Config
	input    textinput.Model
	width    int
	height   int
	status   string
	returnTo func(*config.Config, int, int) tea.Model
}

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

func (m *DocPathModel) Init() tea.Cmd {
	return textinput.Blink
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

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
	docPathTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED")).
				MarginBottom(1).
				PaddingLeft(2)

	docPathBodyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				PaddingLeft(2).
				MarginBottom(1)

	docPathHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				PaddingLeft(2).
				MarginBottom(2)

	docPathInputBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7C3AED")).
				Padding(0, 1).
				MarginLeft(2).
				MarginBottom(1)

	docPathStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				PaddingLeft(2).
				MarginTop(1)

	docPathContainerStyle = lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center)
)

func (m *DocPathModel) View() tea.View {
	var rows []string

	rows = append(rows, docPathTitleStyle.Render("📁 Documentation Output Path"))
	rows = append(rows, docPathBodyStyle.Render("Where should generated documentation be saved?"))
	rows = append(rows, docPathHintStyle.Render("enter confirm  •  esc back"))
	rows = append(rows, docPathInputBoxStyle.Render(m.input.View()))

	if m.status != "" {
		rows = append(rows, docPathStatusStyle.Render(m.status))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
