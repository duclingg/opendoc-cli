package ui

import (
	"time"

	"opendoc/config"
	"opendoc/ui/home"
	"opendoc/ui/setup"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// splashDoneMsg is sent after the splash delay to trigger model routing.
type splashDoneMsg struct{}

// SplashModel is the entry-point screen. It shows the logo and a spinner for
// one second while the config is available, then routes to either the setup
// wizard or the main menu.
type SplashModel struct {
	spinner spinner.Model
	cfg     *config.Config
	width   int
	height  int
	done    bool
}

// NewSplashModel constructs the initial splash screen from the loaded config.
func NewSplashModel(cfg *config.Config) *SplashModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	return &SplashModel{
		spinner: sp,
		cfg:     cfg,
	}
}

// Init starts the spinner tick and kicks off the one-second loading delay.
func (m *SplashModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		simulateLoading(),
	)
}

// simulateLoading fires splashDoneMsg after a short delay so the logo is
// always visible for at least one second before transitioning.
func simulateLoading() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return splashDoneMsg{}
	})
}

// Update handles window sizing, the loading timeout, and quit keys.
func (m *SplashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case splashDoneMsg:
		// Skip the wizard when already fully configured or when the user has
		// previously clicked Start (SetupDismissed). The wizard only reappears
		// after an explicit Reset All Settings.
		if m.cfg.IsFullySetUp() || m.cfg.SetupDismissed {
			next := home.NewMenuModel(m.cfg, m.width, m.height)
			return next, next.Init()
		}
		next := setup.NewSetupModel(m.cfg, m.width, m.height, func(c *config.Config, w, h int) tea.Model {
			return home.NewMenuModel(c, w, h)
		})
		return next, next.Init()

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			MarginBottom(2)

	splashContainerStyle = lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center)

	loadingTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6D6D6D"))
)

const logo = `
  ██████╗ ██████╗ ███████╗███╗   ██╗██████╗  ██████╗  ██████╗
 ██╔═══██╗██╔══██╗██╔════╝████╗  ██║██╔══██╗██╔═══██╗██╔════╝
 ██║   ██║██████╔╝█████╗  ██╔██╗ ██║██║  ██║██║   ██║██║     
 ██║   ██║██╔═══╝ ██╔══╝  ██║╚██╗██║██║  ██║██║   ██║██║     
 ╚██████╔╝██║     ███████╗██║ ╚████║██████╔╝╚██████╔╝╚██████╗
  ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚═══╝╚═════╝  ╚═════╝  ╚═════╝`

// View renders the splash screen: the ASCII logo, subtitle, and a spinner.
func (m *SplashModel) View() tea.View {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		logoStyle.Render(logo),
		subtitleStyle.Render("opendoc cli - cli-based open sourced documentation manager"),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.spinner.View(),
			loadingTextStyle.Render(" Initializing..."),
		),
	)

	v := tea.NewView(splashContainerStyle.
		Width(m.width).
		Height(m.height).
		Render(content))
	v.AltScreen = true
	return v
}
