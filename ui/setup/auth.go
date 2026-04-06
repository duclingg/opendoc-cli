package setup

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"opendoc/config"
	ghauth "opendoc/github"
	"opendoc/ui/styles"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// registrationDoneMsg is emitted when GitHub auth succeeds or fails.
type registrationDoneMsg struct {
	login string
	err   error
}

// deviceCodeMsg carries the response from the device-authorization endpoint.
type deviceCodeMsg struct {
	code *ghauth.DeviceCodeResponse
	err  error
}

// authTokenMsg carries a single token-poll result. An empty token with a nil
// error means the user hasn't acted yet (authorization_pending).
type authTokenMsg struct {
	token string
	err   error
}

// DeviceAuthModel walks the user through the GitHub OAuth device flow.
// No browser redirect or client secret is required — the user enters a
// one-time code at github.com/login/device.
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

// NewDeviceAuthModel constructs the auth screen and sets up the purple spinner.
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

// Init starts the spinner and fires the device-code request immediately.
func (m *DeviceAuthModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.requestDeviceCode())
}

// requestDeviceCode returns a command that calls the GitHub device-code
// endpoint and delivers a deviceCodeMsg.
func (m *DeviceAuthModel) requestDeviceCode() tea.Cmd {
	clientID := m.clientID
	return func() tea.Msg {
		code, err := ghauth.RequestDeviceCode(clientID)
		return deviceCodeMsg{code: code, err: err}
	}
}

// pollToken returns a command that sleeps for the required interval then polls
// GitHub for the access token, delivering an authTokenMsg.
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

// Update handles window resizing, OAuth flow messages, and keyboard input.
// The spinner receives every message it doesn't recognise so it stays animated.
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
	// authHintStyle has a one-line margin (vs the shared HintStyle's two lines)
	// so the auth screen's tighter layout fits without extra spacing.
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

// View renders one of three states: an error message, an initial spinner while
// requesting the device code, or the verification instructions with the
// one-time code and a polling spinner.
func (m *DeviceAuthModel) View() tea.View {
	var rows []string

	rows = append(rows, styles.TitleStyle.Render("🔑 Connect GitHub"))

	switch {
	case m.err != nil:
		rows = append(rows, authErrorStyle.Render(fmt.Sprintf("❌ %v", m.err)))
		rows = append(rows, authHintStyle.Render("Press esc to go back"))

	case m.userCode == "":
		rows = append(rows, authHintStyle.Render("Requesting authorization code…"))
		rows = append(rows, authHintStyle.Render(m.spinner.View()))

	default:
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

// openURL attempts to open the given URL in the system's default browser.
// Errors are silently ignored — the user can always type the URL manually.
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
