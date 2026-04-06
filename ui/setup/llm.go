package setup

import (
	"opendoc/config"
	"opendoc/ui/listutil"
	"opendoc/ui/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// llmProvider describes an AI backend the user can configure.
type llmProvider struct {
	id         string
	name       string
	desc       string
	local      bool
	defaultURL string
	models     []string // empty means free-text input (local providers)
}

// llmProviders is the ordered list of supported AI providers shown in the
// picker. Hosted providers come first, followed by local ones.
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

// LLMSetupModel is the provider-picker screen. It groups providers under
// "Hosted" and "Local" section headers and uses a scroll-aware viewport.
type LLMSetupModel struct {
	cfg          *config.Config
	cursor       int
	width        int
	height       int
	scrollOffset int
	returnTo     func(*config.Config, int, int) tea.Model
}

// NewLLMSetupModel creates an LLMSetupModel with the cursor pre-positioned on
// the currently configured provider (or the first entry if none is set).
func NewLLMSetupModel(cfg *config.Config, w, h int, returnTo func(*config.Config, int, int) tea.Model) *LLMSetupModel {
	cursor := 0
	for i, p := range llmProviders {
		if p.id == cfg.LLMProvider {
			cursor = i
			break
		}
	}
	m := &LLMSetupModel{cfg: cfg, cursor: cursor, width: w, height: h, returnTo: returnTo}
	m.updateScroll()
	return m
}

// Init satisfies tea.Model; no initial commands are needed.
func (m *LLMSetupModel) Init() tea.Cmd { return nil }

// Update handles window resizing, keyboard navigation, and provider selection.
func (m *LLMSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateScroll()

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
			if m.cursor < len(llmProviders)-1 {
				m.cursor++
				m.updateScroll()
			}

		case "enter", " ":
			chosen := llmProviders[m.cursor]
			next := NewLLMCredModel(m.cfg, chosen, m.width, m.height, m.returnTo)
			return next, next.Init()
		}
	}

	return m, nil
}

// renderHeader builds the title and keyboard-hint block shown above the list.
func (m *LLMSetupModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		styles.TitleStyle.Render("🤖 Setup LLM Provider"),
		styles.HintStyle.Render("↑/↓ navigate  •  enter select  •  esc back"),
	)
}

// buildListContent renders every provider row plus section headers and returns
// the joined list string together with three scroll metrics:
//   - scrollTop: first line of the section header above the cursor item, so
//     scrolling up always reveals the section label.
//   - cursorLine: first line of the cursor item itself.
//   - cursorItemH: height of the cursor item in terminal lines.
func (m *LLMSetupModel) buildListContent() (list string, scrollTop int, cursorLine int, cursorItemH int) {
	var rows []string
	lineCount := 0
	sectionStart := 0 // first line of the current section header
	prevLocal := false

	for i, p := range llmProviders {
		// Insert a section header when transitioning between hosted and local.
		var sectionRow string
		if p.local && !prevLocal {
			sectionRow = llmSectionStyle.Render("Local")
		} else if i == 0 {
			sectionRow = llmSectionStyle.Render("Hosted")
		}
		if sectionRow != "" {
			rows = append(rows, sectionRow)
			sectionStart = lineCount
			lineCount += lipgloss.Height(sectionRow)
		}
		prevLocal = p.local

		if i == m.cursor {
			cursorLine = lineCount
			scrollTop = sectionStart
		}

		var indicator string
		if p.id == m.cfg.LLMProvider {
			indicator = llmSelectedIndicatorStyle.Render("●") + " "
		} else {
			indicator = llmUnselectedIndicatorStyle.Render("○") + " "
		}

		var row string
		if i == m.cursor {
			title := indicator + styles.ItemTitleSelected.Render(p.name)
			desc := styles.ItemDescStyle.Render(p.desc)
			row = styles.ItemSelected.Render(title + "\n" + desc)
		} else {
			title := indicator + styles.ItemTitleNormal.Render(p.name)
			desc := styles.ItemDescStyle.Render(p.desc)
			row = styles.ItemNormal.Render(title + "\n" + desc)
		}
		rows = append(rows, row)
		h := lipgloss.Height(row)
		if i == m.cursor {
			cursorItemH = h
		}
		lineCount += h
	}

	list = lipgloss.JoinVertical(lipgloss.Left, rows...)
	return list, scrollTop, cursorLine, cursorItemH
}

// updateScroll adjusts scrollOffset so that:
//  1. The section header above the cursor item is always revealed when
//     scrolling up (scrollTop constraint).
//  2. The bottom of the cursor item is visible when scrolling down.
func (m *LLMSetupModel) updateScroll() {
	if m.width == 0 || m.height == 0 {
		return
	}
	availH := m.height - lipgloss.Height(m.renderHeader()) - footerH
	_, scrollTop, cursorLine, cursorItemH := m.buildListContent()

	if scrollTop < m.scrollOffset {
		m.scrollOffset = scrollTop
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
	llmSectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			Bold(true).
			MarginTop(1)

	llmSelectedIndicatorStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#10B981"))

	llmUnselectedIndicatorStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#6B7280"))
)

// View renders the LLM provider picker: header then the scrollable list.
func (m *LLMSetupModel) View() tea.View {
	header := m.renderHeader()
	availH := m.height - lipgloss.Height(header) - footerH
	list, _, _, _ := m.buildListContent()
	centeredList := listutil.RenderList(list, m.scrollOffset, availH, m.width)
	content := lipgloss.JoinVertical(lipgloss.Center, header, centeredList)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
