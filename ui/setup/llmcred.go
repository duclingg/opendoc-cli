package setup

import (
	"strings"

	"opendoc/config"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type llmCredSaveMsg struct{ err error }

type LLMCredModel struct {
	cfg      *config.Config
	provider llmProvider
	input    textinput.Model
	width    int
	height   int
	status   string
	returnTo func(*config.Config, int, int) tea.Model
}

func NewLLMCredModel(cfg *config.Config, provider llmProvider, w, h int, returnTo func(*config.Config, int, int) tea.Model) *LLMCredModel {
	ti := textinput.New()
	ti.CharLimit = 512
	ti.Focus()

	if provider.local {
		existing := cfg.LLMBaseURL
		if existing == "" {
			existing = provider.defaultURL
		}
		ti.Placeholder = provider.defaultURL
		ti.SetValue(existing)
	} else {
		ti.Placeholder = "sk-..."
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
		if cfg.LLMAPIKey != "" && cfg.LLMProvider == provider.id {
			ti.SetValue(cfg.LLMAPIKey)
		}
	}

	return &LLMCredModel{
		cfg:      cfg,
		provider: provider,
		input:    ti,
		width:    w,
		height:   h,
		returnTo: returnTo,
	}
}

func (m *LLMCredModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *LLMCredModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case llmCredSaveMsg:
		if msg.err != nil {
			m.status = "❌ Failed to save setting"
			return m, nil
		}
		next := NewLLMModelModel(m.cfg, m.provider, m.width, m.height, m.returnTo)
		return next, next.Init()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			next := NewLLMSetupModel(m.cfg, m.width, m.height, m.returnTo)
			return next, next.Init()

		case "enter":
			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				if m.provider.local {
					m.status = "Base URL cannot be empty"
				} else {
					m.status = "API key cannot be empty"
				}
				return m, nil
			}
			cfg := m.cfg
			provider := m.provider
			return m, func() tea.Msg {
				if cfg.LLMProvider != provider.id {
					cfg.LLMModel = "" // clear model only when switching providers
				}
				cfg.LLMProvider = provider.id
				if provider.local {
					cfg.LLMBaseURL = val
					cfg.LLMAPIKey = ""
				} else {
					cfg.LLMAPIKey = val
					cfg.LLMBaseURL = ""
				}
				return llmCredSaveMsg{err: cfg.Save()}
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	llmCredTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED")).
				MarginBottom(1)

	llmCredProviderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A78BFA")).
				MarginBottom(1)

	llmCredLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5E7EB")).
				MarginBottom(1)

	llmCredHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				MarginBottom(2)

	llmCredInputBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7C3AED")).
				Padding(0, 1).
				MarginBottom(1)

	llmCredStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				MarginTop(1)

	llmCredContainerStyle = lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center)
)

func (m *LLMCredModel) View() tea.View {
	var rows []string

	rows = append(rows, llmCredTitleStyle.Render("🔑 Configure "+m.provider.name))
	rows = append(rows, llmCredProviderStyle.Render(m.provider.desc))

	if m.provider.local {
		rows = append(rows, llmCredLabelStyle.Render("Base URL"))
	} else {
		rows = append(rows, llmCredLabelStyle.Render("API Key"))
	}

	rows = append(rows, llmCredHintStyle.Render("enter confirm  •  esc back"))
	rows = append(rows, llmCredInputBoxStyle.Render(m.input.View()))

	if m.status != "" {
		rows = append(rows, llmCredStatusStyle.Render(m.status))
	}

	content := lipgloss.JoinVertical(lipgloss.Center, rows...)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
