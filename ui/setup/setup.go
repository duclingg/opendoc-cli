package setup

import (
	"fmt"

	"opendoc/config"
	"opendoc/ui/styles"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type setupStep struct {
	label string
	desc  string
	done  func(*config.Config) bool
}

var setupSteps = []setupStep{
	{
		label: "Authorize GitHub",
		desc:  "Connect your GitHub account via OAuth device flow",
		done:  func(c *config.Config) bool { return c.IsGitHubSetUp() },
	},
	{
		label: "Setup LLM provider",
		desc:  "Configure an AI provider for documentation generation",
		done:  func(c *config.Config) bool { return c.IsLLMSetUp() },
	},
	{
		label: "Documentation output type",
		desc:  "Choose the format for generated documentation",
		done:  func(c *config.Config) bool { return c.IsDocTypeSetUp() },
	},
	{
		label: "Documentation output path",
		desc:  "Set the directory where documentation will be written",
		done:  func(c *config.Config) bool { return c.IsDocPathSetUp() },
	},
}

type SetupModel struct {
	cfg        *config.Config
	cursor     int
	width      int
	height     int
	status     string
	vp         viewport.Model
	onComplete func(*config.Config, int, int) tea.Model
}

func NewSetupModel(cfg *config.Config, w, h int, onComplete func(*config.Config, int, int) tea.Model) *SetupModel {
	cursor := len(setupSteps) // default to Start
	for i, step := range setupSteps {
		if !step.done(cfg) {
			cursor = i
			break
		}
	}
	if onComplete == nil {
		// Callers must supply onComplete; fall back to quitting rather than panicking.
		onComplete = func(_ *config.Config, _, _ int) tea.Model { return nil }
	}
	m := &SetupModel{cfg: cfg, width: w, height: h, cursor: cursor, onComplete: onComplete}
	m.vp = viewport.New()
	m.syncViewport()
	return m
}

func (m *SetupModel) Init() tea.Cmd { return nil }

func (m *SetupModel) totalItems() int { return len(setupSteps) + 1 } // steps + Start

func (m *SetupModel) isStartIdx(i int) bool { return i == len(setupSteps) }

func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncViewport()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.status = ""
				m.syncViewport()
			}

		case "down", "j":
			if m.cursor < m.totalItems()-1 {
				m.cursor++
				m.status = ""
				m.syncViewport()
			}

		case "enter", " ":
			return m.handleSelect()
		}
	}

	return m, nil
}

func (m *SetupModel) selfReturnTo() func(*config.Config, int, int) tea.Model {
	onComplete := m.onComplete
	return func(c *config.Config, w, h int) tea.Model {
		return NewSetupModel(c, w, h, onComplete)
	}
}

func (m *SetupModel) handleSelect() (tea.Model, tea.Cmd) {
	if m.isStartIdx(m.cursor) {
		next := m.onComplete(m.cfg, m.width, m.height)
		if next == nil {
			return m, tea.Quit
		}
		return next, next.Init()
	}

	switch m.cursor {
	case 0: // GitHub auth
		clientID, err := config.ResolveClientID()
		if err != nil {
			m.status = fmt.Sprintf("❌ %v", err)
			return m, nil
		}
		next := NewDeviceAuthModel(m.cfg, clientID, m.width, m.height, m.selfReturnTo())
		return next, next.Init()

	case 1: // LLM provider
		next := NewLLMSetupModel(m.cfg, m.width, m.height, m.selfReturnTo())
		return next, next.Init()

	case 2: // doc output type
		next := NewDocTypeModel(m.cfg, m.width, m.height, m.selfReturnTo())
		return next, next.Init()

	case 3: // doc output path
		next := NewDocPathModel(m.cfg, m.width, m.height, m.selfReturnTo())
		return next, next.Init()
	}

	return m, nil
}

func (m *SetupModel) contentWidth() int {
	const maxWidth = 80
	if m.width < maxWidth {
		return m.width
	}
	return maxWidth
}

func (m *SetupModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		setupTitleStyle.Render("📄 OpenDoc — Setup"),
		setupSubtitleStyle.Render("Complete each step to get started"),
		setupHintStyle.Render("↑/↓ navigate  •  enter select  •  q quit"),
	)
}

func (m *SetupModel) buildListContent() (string, int) {
	var rows []string
	lineCount := 0
	cursorLine := 0

	for i, step := range setupSteps {
		if i == m.cursor {
			cursorLine = lineCount
		}

		done := step.done(m.cfg)
		var checkmark, titleStr string
		if done {
			checkmark = stepDoneStyle.Render("[✓]")
			titleStr = stepDoneStyle.Render(step.label)
		} else {
			checkmark = stepPendingStyle.Render("[ ]")
			titleStr = styles.ItemTitleNormal.Render(step.label)
		}
		descStr := styles.ItemDescStyle.Render(step.desc)
		label := checkmark + " " + titleStr + "\n    " + descStr

		var row string
		if i == m.cursor {
			row = styles.ItemSelected.Render(label)
		} else {
			row = styles.ItemNormal.Render(label)
		}
		rows = append(rows, row)
		lineCount += lipgloss.Height(row)
	}

	// Start item
	startIdx := len(setupSteps)
	if m.cursor == startIdx {
		cursorLine = lineCount
	}
	startTitle := startEnabledTitleStyle.Render("▶  Start")
	var startRow string
	if m.cursor == startIdx {
		startRow = styles.ItemSelected.Render(startTitle)
	} else {
		startRow = styles.ItemNormal.Render(startTitle)
	}
	rows = append(rows, startRow)

	return lipgloss.JoinVertical(lipgloss.Left, rows...), cursorLine
}

func (m *SetupModel) syncViewport() {
	if m.width == 0 || m.height == 0 {
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
	setupTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1).
			PaddingLeft(2)

	setupSubtitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A78BFA")).
				PaddingLeft(2).
				MarginBottom(2)

	setupHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			PaddingLeft(2).
			MarginBottom(2)

	stepDoneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	stepPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280"))

	startEnabledTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#10B981"))

	startDisabledTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4B5563"))

	setupStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B")).
				PaddingLeft(2).
				MarginTop(1)
)

func (m *SetupModel) View() tea.View {
	header := m.renderHeader()
	body := m.vp.View()
	parts := []string{header, body}
	if m.status != "" {
		parts = append(parts, setupStatusStyle.Render(m.status))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	if off := (m.width - lipgloss.Width(content)) / 2; off > 0 {
		content = lipgloss.NewStyle().PaddingLeft(off).Render(content)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
