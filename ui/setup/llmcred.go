package setup

import (
	"strings"

	"opendoc/config"
	"opendoc/ui/styles"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// llmCredSaveMsg carries the result of persisting the credential.
type llmCredSaveMsg struct{ err error }

// LLMCredModel collects the credential needed for the chosen provider:
// an API key for hosted providers or a base URL for local ones.
type LLMCredModel struct {
	cfg      *config.Config
	provider llmProvider
	input    textinput.Model
	width    int
	height   int
	status   string
	returnTo func(*config.Config, int, int) tea.Model
}

// NewLLMCredModel constructs the credential input screen. The text input is
// pre-filled with any previously saved value and uses password masking for
// API keys.
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

// Init starts the text-input blink cursor.
func (m *LLMCredModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles window resizing, the async save result, and keyboard input.
// Unhandled messages are forwarded to the text input component.
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
				// Clear the model only when the user switches to a different provider.
				if cfg.LLMProvider != provider.id {
					cfg.LLMModel = ""
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
	llmCredProviderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A78BFA")).
				MarginBottom(1)

	llmCredLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5E7EB")).
				MarginBottom(1)

	llmCredInputBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7C3AED")).
				Padding(0, 1).
				MarginBottom(1)

	llmCredStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				MarginTop(1)
)

// View renders the credential input form: title, provider description, a
// labeled text input, and an optional status message.
func (m *LLMCredModel) View() tea.View {
	var fieldLabel string
	if m.provider.local {
		fieldLabel = "Base URL"
	} else {
		fieldLabel = "API Key"
	}

	rows := []string{
		styles.TitleStyle.Render("🔑 Configure " + m.provider.name),
		llmCredProviderStyle.Render(m.provider.desc),
		llmCredLabelStyle.Render(fieldLabel),
		styles.HintStyle.Render("enter confirm  •  esc back"),
		llmCredInputBoxStyle.Render(m.input.View()),
	}
	if m.status != "" {
		rows = append(rows, llmCredStatusStyle.Render(m.status))
	}

	content := lipgloss.JoinVertical(lipgloss.Center, rows...)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
