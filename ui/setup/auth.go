package setup

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"opendoc/config"
	ghauth "opendoc/github"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// registrationDoneMsg is emitted by DeviceAuthModel when auth succeeds or fails.
type registrationDoneMsg struct {
	login string
	err   error
}

type deviceCodeMsg struct {
	code *ghauth.DeviceCodeResponse
	err  error
}

type authTokenMsg struct {
	token string
	err   error
}

// DeviceAuthModel is a full-screen TUI that walks the user through the GitHub
// OAuth device flow. No browser redirect or client secret is required.
type DeviceAuthModel struct {
	cfg        *config.Config
	clientID   string
	width      int
	height     int
	deviceCode string
	interval   int
	userCode   string
	verifyURI  string
	spinner    spinner.Model
	err        error
	returnTo   func(cfg *config.Config, w, h int) tea.Model
}

func NewDeviceAuthModel(cfg *config.Config, clientID string, w, h int, returnTo func(*config.Config, int, int) tea.Model) *DeviceAuthModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))
	return &DeviceAuthModel{
		cfg:      cfg,
		clientID: clientID,
		width:    w,
		height:   h,
		spinner:  sp,
		returnTo: returnTo,
	}
}

func (m *DeviceAuthModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.requestDeviceCode())
}

func (m *DeviceAuthModel) requestDeviceCode() tea.Cmd {
	clientID := m.clientID
	return func() tea.Msg {
		code, err := ghauth.RequestDeviceCode(clientID)
		return deviceCodeMsg{code: code, err: err}
	}
}

func (m *DeviceAuthModel) pollToken() tea.Cmd {
	clientID := m.clientID
	deviceCode := m.deviceCode
	interval := m.interval
	return func() tea.Msg {
		time.Sleep(time.Duration(interval) * time.Second)
		token, err := ghauth.PollForToken(clientID, deviceCode)
		return authTokenMsg{token: token, err: err}
	}
}

func (m *DeviceAuthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		}

	case deviceCodeMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.deviceCode = msg.code.DeviceCode
		m.userCode = msg.code.UserCode
		m.verifyURI = msg.code.VerificationURI
		m.interval = msg.code.Interval
		return m, tea.Batch(openURL(m.verifyURI), m.pollToken())

	case authTokenMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if msg.token == "" {
			// Still waiting — poll again after the interval.
			return m, m.pollToken()
		}
		token := msg.token
		cfg := m.cfg
		return m, func() tea.Msg {
			login, err := ghauth.FetchAuthenticatedUser(token)
			if err != nil {
				return registrationDoneMsg{err: err}
			}
			cfg.GitHubToken = token
			cfg.GitHubLogin = login
			if err := cfg.Save(); err != nil {
				return registrationDoneMsg{err: err}
			}
			return registrationDoneMsg{login: login}
		}

	case registrationDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		next := m.returnTo(m.cfg, m.width, m.height)
		return next, next.Init()

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// ─── styles ──────────────────────────────────────────────────────────────────

var (
	authTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1)

	authHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginBottom(1)

	codeBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 4).
			MarginTop(1).
			MarginBottom(1).
			Align(lipgloss.Center)

	codeLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			MarginBottom(1)

	codeValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F9FAFB"))

	authURLStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED"))

	authStatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			MarginTop(1)

	authErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			MarginTop(1)
)

func (m *DeviceAuthModel) View() tea.View {
	var rows []string

	rows = append(rows, authTitleStyle.Render("🔑 Connect GitHub"))

	if m.err != nil {
		rows = append(rows, authErrorStyle.Render(fmt.Sprintf("❌ %v", m.err)))
		rows = append(rows, authHintStyle.Render("Press esc to go back"))
	} else if m.userCode == "" {
		rows = append(rows, authHintStyle.Render("Requesting authorization code…"))
		rows = append(rows, authHintStyle.Render(m.spinner.View()))
	} else {
		rows = append(rows, authHintStyle.Render("1. Your browser should open automatically."))
		rows = append(rows, authHintStyle.Render("   Or visit: "+authURLStyle.Render(m.verifyURI)))
		rows = append(rows, authHintStyle.Render("2. Enter the one-time code shown below:"))

		box := codeBoxStyle.Render(lipgloss.JoinVertical(
			lipgloss.Center,
			codeLabelStyle.Render("Authorization code"),
			codeValueStyle.Render(m.userCode),
		))
		rows = append(rows, box)

		rows = append(rows, authStatusStyle.Render(m.spinner.View()+" Waiting for authorization…"))
		rows = append(rows, authHintStyle.Render("Press esc to cancel"))
	}

	content := lipgloss.JoinVertical(lipgloss.Center, rows...)

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}

func openURL(u string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", u)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
		default:
			cmd = exec.Command("xdg-open", u)
		}
		_ = cmd.Start()
		return nil
	}
}
