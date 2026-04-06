package setup

import (
	"opendoc/config"
	"opendoc/ui/styles"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type llmProvider struct {
	id         string
	name       string
	desc       string
	local      bool
	defaultURL string
	models     []string // empty means free-text input (local providers)
}

var llmProviders = []llmProvider{
	{
		id: "openai", name: "OpenAI", desc: "GPT-4o, GPT-4 Turbo, and more",
		models: []string{
			"gpt-4o",
			"gpt-4o-mini",
			"gpt-4-turbo",
			"gpt-4",
			"gpt-3.5-turbo",
			"o1",
			"o1-mini",
			"o3-mini",
		},
	},
	{
		id: "anthropic", name: "Anthropic", desc: "Claude Sonnet, Claude Opus, and more",
		models: []string{
			"claude-opus-4-5",
			"claude-sonnet-4-5",
			"claude-haiku-3-5",
			"claude-3-5-sonnet-20241022",
			"claude-3-opus-20240229",
			"claude-3-haiku-20240307",
		},
	},
	{
		id: "gemini", name: "Google Gemini", desc: "Gemini 2.0 Flash, Gemini 1.5 Pro, and more",
		models: []string{
			"gemini-2.0-flash",
			"gemini-2.0-flash-lite",
			"gemini-1.5-pro",
			"gemini-1.5-flash",
			"gemini-1.5-flash-8b",
		},
	},
	{
		id: "mistral", name: "Mistral AI", desc: "Mistral Large, Codestral, and more",
		models: []string{
			"mistral-large-latest",
			"mistral-medium-latest",
			"mistral-small-latest",
			"codestral-latest",
			"open-mixtral-8x22b",
			"open-mistral-7b",
		},
	},
	{
		id: "cohere", name: "Cohere", desc: "Command R+, Command R, and more",
		models: []string{
			"command-r-plus",
			"command-r",
			"command",
			"command-light",
		},
	},
	{id: "ollama", name: "Ollama", desc: "Run models locally via Ollama", local: true, defaultURL: "http://localhost:11434"},
	{id: "lmstudio", name: "LM Studio", desc: "Run models locally via LM Studio", local: true, defaultURL: "http://localhost:1234"},
}

type LLMSetupModel struct {
	cfg      *config.Config
	cursor   int
	width    int
	height   int
	vp       viewport.Model
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

	m := &LLMSetupModel{cfg: cfg, cursor: cursor, width: w, height: h, returnTo: returnTo}
	m.vp = viewport.New()
	m.syncViewport()
	return m
}

func (m *LLMSetupModel) Init() tea.Cmd { return nil }

func (m *LLMSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncViewport()

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
				m.syncViewport()
			}

		case "down", "j":
			if m.cursor < len(llmProviders)-1 {
				m.cursor++
				m.syncViewport()
			}

		case "enter", " ":
			chosen := llmProviders[m.cursor]
			next := NewLLMCredModel(m.cfg, chosen, m.width, m.height, m.returnTo)
			return next, next.Init()
		}
	}

	return m, nil
}

func (m *LLMSetupModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		llmTitleStyle.Render("🤖 Setup LLM Provider"),
		llmHintStyle.Render("↑/↓ navigate  •  enter select  •  esc back"),
	)
}

func (m *LLMSetupModel) buildListContent() (string, int) {
	var rows []string
	lineCount := 0
	cursorLine := 0
	prevLocal := false

	for i, p := range llmProviders {
		var sectionRow string
		if p.local && !prevLocal {
			sectionRow = llmSectionStyle.Render("Local")
		} else if i == 0 {
			sectionRow = llmSectionStyle.Render("Hosted")
		}
		if sectionRow != "" {
			rows = append(rows, sectionRow)
			lineCount += lipgloss.Height(sectionRow)
		}
		prevLocal = p.local

		if i == m.cursor {
			cursorLine = lineCount
		}

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
		lineCount += lipgloss.Height(row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...), cursorLine
}

func (m *LLMSetupModel) contentWidth() int {
	const maxWidth = 80
	if m.width < maxWidth {
		return m.width
	}
	return maxWidth
}

func (m *LLMSetupModel) syncViewport() {
	if m.width == 0 || m.height == 0 {
		return
	}
	header := m.renderHeader()
	headerH := lipgloss.Height(header)
	m.vp.SetWidth(m.contentWidth())
	m.vp.SetHeight(m.height - headerH)
	content, cursorLine := m.buildListContent()
	m.vp.SetContent(content)
	m.vp.EnsureVisible(cursorLine, 0, 0)
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
)

func (m *LLMSetupModel) View() tea.View {
	header := m.renderHeader()
	body := m.vp.View()
	content := lipgloss.JoinVertical(lipgloss.Left, header, body)
	if off := (m.width - lipgloss.Width(content)) / 2; off > 0 {
		content = lipgloss.NewStyle().PaddingLeft(off).Render(content)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
