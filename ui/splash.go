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

type splashDoneMsg struct{}

type SplashModel struct {
	spinner   spinner.Model
	cfg       *config.Config
	width     int
	height    int
	done      bool
	nextModel tea.Model
}

func NewSplashModel(cfg *config.Config) *SplashModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	return &SplashModel{
		spinner: sp,
		cfg:     cfg,
	}
}

func (m *SplashModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		simulateLoading(),
	)
}

func simulateLoading() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return splashDoneMsg{}
	})
}

func (m *SplashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case splashDoneMsg:
		if m.cfg.IsFullySetUp() {
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

func (m *SplashModel) View() tea.View {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		logoStyle.Render(logo),
		subtitleStyle.Render("OpenDoc - Open sourced documentation manager"),
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
