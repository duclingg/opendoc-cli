package setup

import (
	"strings"

	"opendoc/config"
	"opendoc/ui/styles"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type llmModelSaveMsg struct{ err error }

// ─── LLMModelModel ───────────────────────────────────────────────────────────
// Shows a list of preset models for cloud providers, or a text-input for local
// providers (Ollama, LM Studio) where the installed model name is unknown.

type LLMModelModel struct {
	cfg      *config.Config
	provider llmProvider
	cursor   int
	input    textinput.Model // used only for local providers
	vp       viewport.Model
	width    int
	height   int
	status   string
	returnTo func(*config.Config, int, int) tea.Model
}

func NewLLMModelModel(cfg *config.Config, provider llmProvider, w, h int, returnTo func(*config.Config, int, int) tea.Model) *LLMModelModel {
	m := &LLMModelModel{
		cfg:      cfg,
		provider: provider,
		width:    w,
		height:   h,
		returnTo: returnTo,
	}

	if provider.local {
		ti := textinput.New()
		ti.CharLimit = 256
		ti.Placeholder = "e.g. llama3, mistral, phi3"
		if cfg.LLMModel != "" && cfg.LLMProvider == provider.id {
			ti.SetValue(cfg.LLMModel)
		}
		ti.Focus()
		m.input = ti
	} else {
		for i, model := range provider.models {
			if model == cfg.LLMModel && cfg.LLMProvider == provider.id {
				m.cursor = i
				break
			}
		}
		m.vp = viewport.New()
		m.syncViewport()
	}

	return m
}

func (m *LLMModelModel) Init() tea.Cmd {
	if m.provider.local {
		return textinput.Blink
	}
	return nil
}

func (m *LLMModelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.provider.local {
			m.syncViewport()
		}

	case llmModelSaveMsg:
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
			next := NewLLMCredModel(m.cfg, m.provider, m.width, m.height, m.returnTo)
			return next, next.Init()

		case "up", "k":
			if !m.provider.local && m.cursor > 0 {
				m.cursor--
				m.syncViewport()
			}

		case "down", "j":
			if !m.provider.local && m.cursor < len(m.provider.models)-1 {
				m.cursor++
				m.syncViewport()
			}

		case "enter", " ":
			return m.handleConfirm()
		}
	}

	if m.provider.local {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *LLMModelModel) handleConfirm() (tea.Model, tea.Cmd) {
	var chosen string
	if m.provider.local {
		chosen = strings.TrimSpace(m.input.Value())
		if chosen == "" {
			m.status = "Model name cannot be empty"
			return m, nil
		}
	} else {
		chosen = m.provider.models[m.cursor]
	}

	cfg := m.cfg
	return m, func() tea.Msg {
		cfg.LLMModel = chosen
		return llmModelSaveMsg{err: cfg.Save()}
	}
}

func (m *LLMModelModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		llmModelTitleStyle.Render("🧠 Select Model — "+m.provider.name),
		llmModelProviderStyle.Render(m.provider.desc),
		llmModelHintStyle.Render("↑/↓ navigate  •  enter select  •  esc back"),
	)
}

func (m *LLMModelModel) buildListContent() (string, int) {
	var rows []string
	lineCount := 0
	cursorLine := 0

	for i, model := range m.provider.models {
		if i == m.cursor {
			cursorLine = lineCount
		}

		var indicator string
		if model == m.cfg.LLMModel && m.cfg.LLMProvider == m.provider.id {
			indicator = llmSelectedIndicatorStyle.Render("●") + " "
		} else {
			indicator = llmUnselectedIndicatorStyle.Render("○") + " "
		}

		var row string
		if i == m.cursor {
			title := indicator + styles.ItemTitleSelected.Render(model)
			row = styles.ItemSelected.Render(title)
		} else {
			title := indicator + styles.ItemTitleNormal.Render(model)
			row = styles.ItemNormal.Render(title)
		}
		rows = append(rows, row)
		lineCount += lipgloss.Height(row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...), cursorLine
}

func (m *LLMModelModel) contentWidth() int {
	const maxWidth = 80
	if m.width < maxWidth {
		return m.width
	}
	return maxWidth
}

func (m *LLMModelModel) syncViewport() {
	if m.width == 0 || m.height == 0 || m.provider.local {
		return
	}
	header := m.renderHeader()
	headerH := lipgloss.Height(header)
	const footerH = 2 // reserve space for status line
	m.vp.SetWidth(m.contentWidth())
	m.vp.SetHeight(m.height - headerH - footerH)
	content, cursorLine := m.buildListContent()
	m.vp.SetContent(content)
	m.vp.EnsureVisible(cursorLine, 0, 0)
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	llmModelTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED")).
				MarginBottom(1)

	llmModelProviderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A78BFA")).
				MarginBottom(1)

	llmModelHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				MarginBottom(2)

	llmModelLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5E7EB")).
				MarginBottom(1)

	llmModelInputBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7C3AED")).
				Padding(0, 1).
				MarginBottom(1)

	llmModelStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				MarginTop(1)

	llmModelContainerStyle = lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center)
)

func (m *LLMModelModel) View() tea.View {
	var content string

	if m.provider.local {
		var rows []string
		rows = append(rows, llmModelTitleStyle.Render("🧠 Select Model — "+m.provider.name))
		rows = append(rows, llmModelProviderStyle.Render(m.provider.desc))
		rows = append(rows, llmModelLabelStyle.Render("Model name"))
		rows = append(rows, llmModelHintStyle.Render("enter confirm  •  esc back"))
		rows = append(rows, llmModelInputBoxStyle.Render(m.input.View()))
		if m.status != "" {
			rows = append(rows, llmModelStatusStyle.Render(m.status))
		}
		content = lipgloss.JoinVertical(lipgloss.Center, rows...)
	} else {
		header := m.renderHeader()
		list, _ := m.buildListContent()
		centeredList := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Render(list)
		parts := []string{header, centeredList}
		if m.status != "" {
			parts = append(parts, llmModelStatusStyle.Render(m.status))
		}
		content = lipgloss.JoinVertical(lipgloss.Center, parts...)
	}

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
