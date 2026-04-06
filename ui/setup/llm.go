package setup

import (
	"opendoc/config"
	"opendoc/ui/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type llmProvider struct {
	id         string
	name       string
	desc       string
	local      bool
	defaultURL string
}

var llmProviders = []llmProvider{
	{id: "openai", name: "OpenAI", desc: "GPT-4o, GPT-4 Turbo, and more"},
	{id: "anthropic", name: "Anthropic", desc: "Claude 3.5 Sonnet, Claude 3 Opus, and more"},
	{id: "gemini", name: "Google Gemini", desc: "Gemini 1.5 Pro, Gemini Flash, and more"},
	{id: "mistral", name: "Mistral AI", desc: "Mistral Large, Codestral, and more"},
	{id: "cohere", name: "Cohere", desc: "Command R+, Command R, and more"},
	{id: "ollama", name: "Ollama", desc: "Run models locally via Ollama", local: true, defaultURL: "http://localhost:11434"},
	{id: "lmstudio", name: "LM Studio", desc: "Run models locally via LM Studio", local: true, defaultURL: "http://localhost:1234"},
}

type LLMSetupModel struct {
	cfg      *config.Config
	cursor   int
	width    int
	height   int
	returnTo func(*config.Config, int, int) tea.Model
}

func NewLLMSetupModel(cfg *config.Config, w, h int, returnTo func(*config.Config, int, int) tea.Model) *LLMSetupModel {
	cursor := 0
	for i, p := range llmProviders {
		if p.id == cfg.LLMProvider {
			cursor = i
			break
		}
	}

	return &LLMSetupModel{cfg: cfg, cursor: cursor, width: w, height: h, returnTo: returnTo}
}

func (m *LLMSetupModel) Init() tea.Cmd { return nil }

func (m *LLMSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

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
			if m.cursor < len(llmProviders)-1 {
				m.cursor++
			}

		case "enter", " ":
			chosen := llmProviders[m.cursor]
			next := NewLLMCredModel(m.cfg, chosen, m.width, m.height, m.returnTo)
			return next, next.Init()
		}
	}

	return m, nil
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	llmTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1).
			PaddingLeft(2)

	llmHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			PaddingLeft(2).
			MarginBottom(2)

	llmSectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			Bold(true).
			PaddingLeft(2).
			MarginTop(1)

	llmContainerStyle = lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center)
)

func (m *LLMSetupModel) View() tea.View {
	var rows []string

	rows = append(rows, llmTitleStyle.Render("🤖 Setup LLM Provider"))
	rows = append(rows, llmHintStyle.Render("↑/↓ navigate  •  enter select  •  esc back"))

	prevLocal := false
	for i, p := range llmProviders {
		if p.local && !prevLocal {
			rows = append(rows, llmSectionStyle.Render("Local"))
		} else if i == 0 {
			rows = append(rows, llmSectionStyle.Render("Hosted"))
		}
		prevLocal = p.local

		var row string
		if i == m.cursor {
			title := styles.ItemTitleSelected.Render(p.name)
			desc := styles.ItemDescStyle.Render(p.desc)
			row = styles.ItemSelected.Render(title + "\n" + desc)
		} else {
			title := styles.ItemTitleNormal.Render(p.name)
			desc := styles.ItemDescStyle.Render(p.desc)
			row = styles.ItemNormal.Render(title + "\n" + desc)
		}
		rows = append(rows, row)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	v := tea.NewView(llmContainerStyle.
		Width(m.width).
		Height(m.height).
		Render(content))
	v.AltScreen = true
	return v
}
