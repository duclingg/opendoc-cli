package setup

import (
	"strings"

	"opendoc/config"
	"opendoc/ui/listutil"
	"opendoc/ui/styles"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// llmModelSaveMsg carries the result of persisting the chosen model.
type llmModelSaveMsg struct{ err error }

// LLMModelModel shows a list of preset models for cloud providers, or a
// free-text input for local providers (Ollama, LM Studio) where the installed
// model name is not known in advance.
type LLMModelModel struct {
	cfg          *config.Config
	provider     llmProvider
	cursor       int
	input        textinput.Model // used only when provider.local is true
	scrollOffset int
	width        int
	height       int
	status       string
	returnTo     func(*config.Config, int, int) tea.Model
}

// NewLLMModelModel constructs the model-picker for the given provider.
// For cloud providers the cursor is pre-positioned on the currently saved
// model. For local providers a text input is focused instead.
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
		m.updateScroll()
	}

	return m
}

// Init starts the text-input blink cursor for local providers; cloud providers
// need no initial command.
func (m *LLMModelModel) Init() tea.Cmd {
	if m.provider.local {
		return textinput.Blink
	}
	return nil
}

// Update handles window resizing, save results, and keyboard input. For local
// providers the text input also receives every unhandled message.
func (m *LLMModelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.provider.local {
			m.updateScroll()
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
				m.updateScroll()
			}

		case "down", "j":
			if !m.provider.local && m.cursor < len(m.provider.models)-1 {
				m.cursor++
				m.updateScroll()
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

// handleConfirm validates the selection and dispatches an async save command.
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

// renderTitleBlock returns the title and provider description lines that are
// shared between the list and local-provider views.
func (m *LLMModelModel) renderTitleBlock() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		styles.TitleStyle.Render("🧠 Select Model — "+m.provider.name),
		llmModelProviderStyle.Render(m.provider.desc),
	)
}

// renderHeader builds the full header for the list (cloud provider) view:
// title block plus the navigation hint.
func (m *LLMModelModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		m.renderTitleBlock(),
		styles.HintStyle.Render("↑/↓ navigate  •  enter select  •  esc back"),
	)
}

// buildListContent renders each model row and returns the joined list string
// together with the cursor item's top line and height for scroll tracking.
func (m *LLMModelModel) buildListContent() (list string, cursorLine int, cursorItemH int) {
	var rows []string
	lineCount := 0

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
		h := lipgloss.Height(row)
		if i == m.cursor {
			cursorItemH = h
		}
		lineCount += h
	}

	list = lipgloss.JoinVertical(lipgloss.Left, rows...)
	return list, cursorLine, cursorItemH
}

// updateScroll adjusts scrollOffset so the focused model row stays visible.
// Only called for cloud (list) providers; local providers use a text input.
func (m *LLMModelModel) updateScroll() {
	if m.width == 0 || m.height == 0 || m.provider.local {
		return
	}
	availH := m.height - lipgloss.Height(m.renderHeader()) - footerH
	_, cursorLine, cursorItemH := m.buildListContent()
	listutil.UpdateScroll(&m.scrollOffset, availH, cursorLine, cursorItemH)
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	llmModelProviderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A78BFA")).
				MarginBottom(1)

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
)

// View renders either a free-text input (local providers) or a scrollable
// model list (cloud providers).
func (m *LLMModelModel) View() tea.View {
	var content string

	if m.provider.local {
		rows := []string{
			m.renderTitleBlock(),
			llmModelLabelStyle.Render("Model name"),
			styles.HintStyle.Render("enter confirm  •  esc back"),
			llmModelInputBoxStyle.Render(m.input.View()),
		}
		if m.status != "" {
			rows = append(rows, llmModelStatusStyle.Render(m.status))
		}
		content = lipgloss.JoinVertical(lipgloss.Center, rows...)
	} else {
		header := m.renderHeader()
		availH := m.height - lipgloss.Height(header) - footerH
		list, _, _ := m.buildListContent()
		centeredList := listutil.RenderList(list, m.scrollOffset, availH, m.width)
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
