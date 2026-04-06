package setup

import (
	"fmt"

	"opendoc/config"
	"opendoc/ui/styles"

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
	return &SetupModel{cfg: cfg, width: w, height: h, cursor: cursor, onComplete: onComplete}
}

func (m *SetupModel) Init() tea.Cmd { return nil }

func (m *SetupModel) totalItems() int { return len(setupSteps) + 1 } // steps + Start

func (m *SetupModel) isStartIdx(i int) bool { return i == len(setupSteps) }

func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.status = ""
			}

		case "down", "j":
			if m.cursor < m.totalItems()-1 {
				m.cursor++
				m.status = ""
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
		if !m.cfg.IsFullySetUp() {
			m.status = "Complete all steps above before starting"
			return m, nil
		}
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

	setupContainerStyle = lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center)
)

func (m *SetupModel) View() tea.View {
	var rows []string

	rows = append(rows, setupTitleStyle.Render("📄 OpenDoc — Setup"))
	rows = append(rows, setupSubtitleStyle.Render("Complete each step to get started"))
	rows = append(rows, setupHintStyle.Render("↑/↓ navigate  •  enter select  •  q quit"))

	for i, step := range setupSteps {
		done := step.done(m.cfg)

		var checkmark, titleStr, descStr string
		if done {
			checkmark = stepDoneStyle.Render("[✓]")
			titleStr = stepDoneStyle.Render(step.label)
		} else {
			checkmark = stepPendingStyle.Render("[ ]")
			titleStr = styles.ItemTitleNormal.Render(step.label)
		}
		descStr = styles.ItemDescStyle.Render(step.desc)

		label := checkmark + " " + titleStr + "\n    " + descStr

		var row string
		if i == m.cursor {
			row = styles.ItemSelected.Render(label)
		} else {
			row = styles.ItemNormal.Render(label)
		}
		rows = append(rows, row)
	}

	// Start item
	startIdx := len(setupSteps)
	allDone := m.cfg.IsFullySetUp()
	var startTitle string
	if allDone {
		startTitle = startEnabledTitleStyle.Render("▶  Start")
	} else {
		startTitle = startDisabledTitleStyle.Render("▶  Start")
	}

	var startRow string
	if m.cursor == startIdx {
		startRow = styles.ItemSelected.Render(startTitle)
	} else {
		startRow = styles.ItemNormal.Render(startTitle)
	}
	rows = append(rows, startRow)

	if m.status != "" {
		rows = append(rows, setupStatusStyle.Render(m.status))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	v := tea.NewView(setupContainerStyle.
		Width(m.width).
		Height(m.height).
		Render(content))
	v.AltScreen = true
	return v
}
