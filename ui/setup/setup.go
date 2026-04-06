package setup

import (
	"fmt"

	"opendoc/config"
	"opendoc/ui/listutil"
	"opendoc/ui/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// footerH is the number of terminal lines reserved at the bottom of every
// setup screen for status messages and breathing room.
const footerH = 2

// setupStep describes a single required configuration step in the wizard.
type setupStep struct {
	label string
	desc  string
	done  func(*config.Config) bool
}

// setupSteps is the ordered list of steps the user must complete before the
// wizard considers itself done.
var setupSteps = []setupStep{
	{
		label: "Authorize GitHub",
		desc:  "Connect your GitHub account via OAuth device flow",
		done:  func(c *config.Config) bool { return c.IsRegistered() },
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

// SetupModel is the setup wizard checklist screen. It shows every required
// step with a done/pending indicator and a "Start" item at the bottom.
type SetupModel struct {
	cfg          *config.Config
	cursor       int
	width        int
	height       int
	status       string
	scrollOffset int
	onComplete   func(*config.Config, int, int) tea.Model
}

// NewSetupModel creates a SetupModel with the cursor pre-positioned on the
// first incomplete step. onComplete is called when the user selects "Start".
func NewSetupModel(cfg *config.Config, w, h int, onComplete func(*config.Config, int, int) tea.Model) *SetupModel {
	cursor := len(setupSteps) // default to the Start item
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
	m.updateScroll()
	return m
}

// Init satisfies tea.Model; the setup list requires no initial commands.
func (m *SetupModel) Init() tea.Cmd { return nil }

// totalItems returns the count of navigable rows (steps + the Start item).
func (m *SetupModel) totalItems() int { return len(setupSteps) + 1 }

// isStartIdx reports whether index i points to the Start item.
func (m *SetupModel) isStartIdx(i int) bool { return i == len(setupSteps) }

// Update handles window resizing, keyboard navigation, and step selection.
func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateScroll()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.status = ""
				m.updateScroll()
			}

		case "down", "j":
			if m.cursor < m.totalItems()-1 {
				m.cursor++
				m.status = ""
				m.updateScroll()
			}

		case "enter", " ":
			return m.handleSelect()
		}
	}

	return m, nil
}

// selfReturnTo returns a returnTo callback that recreates this SetupModel,
// preserving the original onComplete handler. Sub-screens use it to navigate
// back to the wizard after finishing a step.
func (m *SetupModel) selfReturnTo() func(*config.Config, int, int) tea.Model {
	onComplete := m.onComplete
	return func(c *config.Config, w, h int) tea.Model {
		return NewSetupModel(c, w, h, onComplete)
	}
}

// handleSelect launches the sub-screen for the selected step, or calls
// onComplete when the user activates the Start item. Pressing Start also
// persists the SetupDismissed flag so future launches skip the wizard.
func (m *SetupModel) handleSelect() (tea.Model, tea.Cmd) {
	if m.isStartIdx(m.cursor) {
		m.cfg.SetupDismissed = true
		_ = m.cfg.Save() // best-effort; failure should not block the transition
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

// renderHeader builds the title, subtitle, and keyboard-hint block shown
// above the step list.
func (m *SetupModel) renderHeader() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		styles.TitleStyle.Render("📄 opendoc — Setup"),
		setupSubtitleStyle.Render("Complete each step to get started"),
		styles.HintStyle.Render("↑/↓ navigate  •  enter select  •  q quit"),
	)
}

// buildListContent renders every step row plus the Start item and returns the
// joined list string together with the cursor item's top line and height.
// These are used by updateScroll to keep the focused row visible.
func (m *SetupModel) buildListContent() (list string, cursorLine int, cursorItemH int) {
	var rows []string
	lineCount := 0

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
		label := checkmark + " " + titleStr + "\n    " + styles.ItemDescStyle.Render(step.desc)

		var row string
		if i == m.cursor {
			row = styles.ItemSelected.Render(label)
		} else {
			row = styles.ItemNormal.Render(label)
		}
		rows = append(rows, row)
		h := lipgloss.Height(row)
		if i == m.cursor {
			cursorItemH = h
		}
		lineCount += h
	}

	// Start item
	startTitle := startEnabledTitleStyle.Render("▶  Start")
	if m.isStartIdx(m.cursor) {
		cursorLine = lineCount
		row := styles.ItemSelected.Render(startTitle)
		cursorItemH = lipgloss.Height(row)
		rows = append(rows, row)
	} else {
		rows = append(rows, styles.ItemNormal.Render(startTitle))
	}

	list = lipgloss.JoinVertical(lipgloss.Left, rows...)
	return list, cursorLine, cursorItemH
}

// updateScroll adjusts scrollOffset so the focused row stays within the
// visible viewport.
func (m *SetupModel) updateScroll() {
	if m.width == 0 || m.height == 0 {
		return
	}
	availH := m.height - lipgloss.Height(m.renderHeader()) - footerH
	_, cursorLine, cursorItemH := m.buildListContent()
	listutil.UpdateScroll(&m.scrollOffset, availH, cursorLine, cursorItemH)
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	setupSubtitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A78BFA")).
				MarginBottom(2)

	stepDoneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	stepPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280"))

	startEnabledTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#10B981"))

	setupStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B")).
				MarginTop(1)
)

// View renders the setup wizard: header, scrollable step list, and an optional
// status message.
func (m *SetupModel) View() tea.View {
	header := m.renderHeader()
	availH := m.height - lipgloss.Height(header) - footerH
	list, _, _ := m.buildListContent()
	centeredList := listutil.RenderList(list, m.scrollOffset, availH, m.width)

	parts := []string{header, centeredList}
	if m.status != "" {
		parts = append(parts, setupStatusStyle.Render(m.status))
	}
	content := lipgloss.JoinVertical(lipgloss.Center, parts...)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
