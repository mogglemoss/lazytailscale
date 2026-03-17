package model

import (
	"context"
	"github.com/mogglemoss/lazytailscale/config"
	"github.com/mogglemoss/lazytailscale/ping"
	"github.com/mogglemoss/lazytailscale/rdp"
	"github.com/mogglemoss/lazytailscale/ssh"
	"github.com/mogglemoss/lazytailscale/tailscale"
	"github.com/mogglemoss/lazytailscale/ui"
	"github.com/mogglemoss/lazytailscale/vnc"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"fmt"
	"regexp"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	peerPollInterval = 5 * time.Second
	pingPollInterval = 10 * time.Second
	statusClearDelay = 3 * time.Second
	maxPingHistory   = 8
	listPaneWidth    = 28
)

// Model is the root Bubbletea model.
type Model struct {
	keys   KeyMap
	client *tailscale.Client
	peers  []tailscale.Peer
	info   tailscale.NetworkInfo
	errMsg string // status bar error, clears after statusClearDelay

	list     list.Model
	viewport viewport.Model
	spinner  spinner.Model

	width  int
	height int

	showHelp    bool
	showRoutes  bool
	pinging     bool
	ready       bool
	mascotFrame int

	// SSH form state (Huh form replacing the old textinput prompt)
	sshForm       *huh.Form
	sshFormValues *sshFormState
	sshTarget     tailscale.Peer
	sshUsernames  map[string]string // hostname → last used username (persisted)
	sshPorts      map[string]string // hostname → last used port (persisted)
	defaultUser   string            // local system username

	// Connect popup state (shown when Enter is pressed on a peer)
	connectPopup  bool
	connectTarget tailscale.Peer

	// Animation state
	flashPeers map[string]bool // hostname → briefly highlight this row
	pingFlash  bool            // briefly flash the sparkline after a ping result
}

// New creates and returns the initial model.
func New(demo bool) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(ui.S.T.Accent)

	// Detect local username as the default SSH user.
	defaultUser := "user"
	if u, err := user.Current(); err == nil {
		defaultUser = u.Username
	}

	return Model{
		keys:         DefaultKeyMap(),
		client:       tailscale.NewClient(demo),
		spinner:      sp,
		defaultUser:  defaultUser,
		sshUsernames: config.LoadUsernames(),
		sshPorts:     config.LoadPorts(),
		flashPeers:   make(map[string]bool),
	}
}

// Init kicks off the first peer fetch and both tick loops.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchPeersCmd(),
		m.tickCmd(),
		m.pingTickCmd(),
		m.mascotTickCmd(),
		m.spinner.Tick,
	)
}

// Update handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// SSH form intercepts all messages while active.
	if m.sshForm != nil {
		// Window resize must be handled even during form use.
		if ws, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = ws.Width
			m.height = ws.Height
			m = m.recalcLayout()
			m.ready = true
			return m, nil
		}
		// ctrl+c always quits.
		if km, ok := msg.(tea.KeyMsg); ok && km.String() == "ctrl+c" {
			return m, tea.Quit
		}
		f, cmd := m.sshForm.Update(msg)
		m.sshForm = f.(*huh.Form)
		switch m.sshForm.State {
		case huh.StateCompleted:
			return m.launchSSHFromForm()
		case huh.StateAborted:
			m.sshForm = nil
			m.sshFormValues = nil
		}
		return m, cmd
	}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.recalcLayout()
		m.ready = true

	case tickMsg:
		cmds = append(cmds, m.tickCmd(), m.fetchPeersCmd())

	case pingTickMsg:
		cmds = append(cmds, m.pingTickCmd())
		if p := m.selectedPeer(); p != nil && p.TailscaleIP != "" && !m.pinging {
			m.pinging = true
			cmds = append(cmds, m.pingCmd(p.TailscaleIP))
		}

	case mascotTickMsg:
		m.mascotFrame++
		cmds = append(cmds, m.mascotTickCmd())
		m = m.refreshDetail()

	case peersLoadedMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			cmds = append(cmds, clearStatusCmd())
		} else {
			var notifCmds []tea.Cmd
			m, notifCmds = m.mergePeers(msg.peers, msg.info)
			cmds = append(cmds, notifCmds...)
			if len(notifCmds) == 0 {
				m.errMsg = ""
			}
			m.list.SetItems(ui.PeersToItems(m.peers, m.flashPeers))
			m = m.refreshDetail()
			if p := m.selectedPeer(); p != nil && p.TailscaleIP != "" && !p.IsSelf && !m.pinging && len(p.PingHistory) == 0 {
				m.pinging = true
				cmds = append(cmds, m.pingCmd(p.TailscaleIP))
			}
		}

	case pingResultMsg:
		m.pinging = false
		m = m.applyPingResult(msg)
		m.pingFlash = true
		cmds = append(cmds, pingFlashClearCmd())
		m = m.refreshDetail()

	case pingFlashClearMsg:
		m.pingFlash = false
		m = m.refreshDetail()

	case peerFlashClearMsg:
		delete(m.flashPeers, msg.hostname)
		m.list.SetItems(ui.PeersToItems(m.peers, m.flashPeers))

	case statusClearMsg:
		m.errMsg = ""

	case exitNodeResultMsg:
		if msg.err != nil {
			m.errMsg = "exit node: " + msg.err.Error()
			cmds = append(cmds, clearStatusCmd())
		}
		cmds = append(cmds, m.fetchPeersCmd())

	case connectionResultMsg:
		if msg.err != nil {
			m.errMsg = "connection: " + msg.err.Error()
			cmds = append(cmds, clearStatusCmd())
		}
		cmds = append(cmds, m.fetchPeersCmd())

	case ssh.SSHErrorMsg:
		m.errMsg = msg.Err.Error()
		cmds = append(cmds, clearStatusCmd())

	case ssh.SSHDoneMsg:
		// TUI already resumed — nothing to do.

	case rdp.ErrMsg:
		m.errMsg = "rdp: " + msg.Err.Error()
		cmds = append(cmds, clearStatusCmd())

	case rdp.DoneMsg:
		// RDP client launched in background — nothing to do.

	case vnc.ErrMsg:
		m.errMsg = "vnc: " + msg.Err.Error()
		cmds = append(cmds, clearStatusCmd())

	case vnc.DoneMsg:
		// VNC viewer launched in background — nothing to do.

	case tea.MouseMsg:
		if m.sshForm == nil && !m.connectPopup {
			m = m.handleMouse(msg, &cmds)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		// Connect popup intercepts all keys while active.
		if m.connectPopup {
			return m.handleConnectPopupKey(msg)
		}

		// While filtering, forward all keys to the list.
		if m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		switch {
		case msg.String() == "ctrl+c":
			return m, tea.Quit

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp

		case key.Matches(msg, m.keys.SSH):
			if p := m.selectedPeer(); p != nil && p.TailscaleIP != "" && !p.IsSelf {
				m = m.enterConnectPopup(*p)
			}

		case key.Matches(msg, m.keys.Ping):
			if p := m.selectedPeer(); p != nil && p.TailscaleIP != "" && !m.pinging {
				m.pinging = true
				cmds = append(cmds, m.pingCmd(p.TailscaleIP))
			}

		case key.Matches(msg, m.keys.Routes):
			m.showRoutes = !m.showRoutes
			m = m.refreshDetail()

		case key.Matches(msg, m.keys.Copy):
			if p := m.selectedPeer(); p != nil {
				addr := p.TailscaleIP
				if p.DNSName != "" {
					addr = p.DNSName
				}
				copyToClipboard(addr)
			}

		case key.Matches(msg, m.keys.ExitNode):
			if p := m.selectedPeer(); p != nil && p.CanBeExitNode && !p.IsSelf {
				cmds = append(cmds, m.toggleExitNodeCmd(p))
			}

		case key.Matches(msg, m.keys.Connection):
			cmds = append(cmds, m.toggleConnectionCmd())

		case key.Matches(msg, m.keys.Refresh):
			cmds = append(cmds, m.fetchPeersCmd())

		case key.Matches(msg, m.keys.Filter):
			// Explicitly activate the list's filter mode via our own keymap.
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)

		default:
			prevIdx := m.list.Index()
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			if m.list.Index() != prevIdx {
				m = m.refreshDetail()
				if p := m.selectedPeer(); p != nil && p.TailscaleIP != "" && !p.IsSelf && !m.pinging {
					m.pinging = true
					cmds = append(cmds, m.pingCmd(p.TailscaleIP))
				}
			}
		}

	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// mascotState derives the current mascot animation state from model state.
func (m Model) mascotState() ui.MascotState {
	if m.info.Stopped || (!m.info.Online && m.info.NetworkName != "") {
		return ui.MascotOffline
	}
	if m.pinging {
		return ui.MascotPinging
	}
	return ui.MascotNormal
}

// View renders the full TUI.
func (m Model) View() string {
	if !m.ready {
		return "\n  " + m.spinner.View() + " Loading…\n"
	}

	statusBar := ui.RenderStatusBar(m.info, m.errMsg, m.width, m.mascotFrame, m.mascotState())

	var helpBar string
	switch {
	case m.connectPopup:
		helpBar = ui.RenderConnectPopup(m.width, m.connectTarget.Hostname, m.connectTarget.OS)
	case m.sshForm != nil:
		helpBar = ui.RenderSSHFormHint(m.width)
	default:
		helpBar = ui.RenderHelpBar(m.width, m.showHelp)
	}

	listView := ui.S.PanelBorder.
		Width(listPaneWidth).
		Height(m.bodyHeight()).
		Render(m.list.View())

	var detailView string
	if m.sshForm != nil {
		header := "\n" + ui.S.DetailHeader.Render("  ssh into "+m.sshTarget.Hostname) + "\n\n"
		detailView = lipgloss.NewStyle().
			Width(m.detailWidth()).
			Height(m.bodyHeight()).
			Render(header + m.sshForm.View())
	} else {
		detailView = m.viewport.View()
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)

	return lipgloss.JoinVertical(lipgloss.Left, statusBar, body, helpBar)
}

// ── Layout ────────────────────────────────────────────────────────────────────

func (m Model) bodyHeight() int {
	h := m.height - 2 // status bar + help bar
	if h < 1 {
		return 1
	}
	return h
}

func (m Model) detailWidth() int {
	w := m.width - listPaneWidth - 1
	if w < 10 {
		return 10
	}
	return w
}

func (m Model) recalcLayout() Model {
	bh := m.bodyHeight()
	dw := m.detailWidth()

	if !m.ready {
		m.list = ui.NewPeerList(m.peers, bh)
	} else {
		m.list.SetHeight(bh)
		m.list.SetWidth(listPaneWidth)
	}

	m.viewport.Width = dw
	m.viewport.Height = bh
	m = m.refreshDetail()
	return m
}

// ── Peer helpers ─────────────────────────────────────────────────────────────

func (m Model) selectedPeer() *tailscale.Peer {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	pi, ok := item.(ui.PeerItem)
	if !ok {
		return nil
	}
	for i := range m.peers {
		if m.peers[i].Hostname == pi.Peer.Hostname {
			return &m.peers[i]
		}
	}
	return nil
}

func (m Model) refreshDetail() Model {
	if len(m.peers) == 0 && m.errMsg != "" {
		m.viewport.SetContent(ui.RenderNoTailscale(m.errMsg, m.viewport.Width))
		return m
	}
	peer := tailscale.Peer{}
	if p := m.selectedPeer(); p != nil {
		peer = *p
	}
	m.viewport.SetContent(ui.RenderDetail(peer, m.info, m.showRoutes, m.viewport.Width, m.mascotFrame, m.mascotState(), m.pingFlash))
	return m
}

func (m Model) mergePeers(fresh []tailscale.Peer, info tailscale.NetworkInfo) (Model, []tea.Cmd) {
	m.info = info

	// Build lookup of previous online state and ping history by IP.
	type prev struct {
		online bool
		hist   []time.Duration
		seen   bool
	}
	prevMap := make(map[string]prev, len(m.peers))
	for _, p := range m.peers {
		prevMap[p.TailscaleIP] = prev{online: p.Online, hist: p.PingHistory, seen: true}
	}

	// Find active exit node.
	m.info.ActiveExitNode = ""
	for i := range fresh {
		if fresh[i].IsExitNode {
			m.info.ActiveExitNode = fresh[i].Hostname
			break
		}
	}

	var cmds []tea.Cmd
	for i := range fresh {
		ip := fresh[i].TailscaleIP
		if p, ok := prevMap[ip]; ok {
			fresh[i].PingHistory = p.hist
			// Notify and flash on status transitions for non-self peers.
			if p.seen && !fresh[i].IsSelf && fresh[i].Online != p.online {
				name := fresh[i].Hostname
				if fresh[i].Online {
					m.errMsg = name + " connected"
				} else {
					m.errMsg = name + " disconnected"
				}
				cmds = append(cmds, clearStatusCmd())
				m.flashPeers[name] = true
				cmds = append(cmds, peerFlashClearCmd(name))
			}
		}
	}
	m.peers = fresh
	return m, cmds
}

func (m Model) applyPingResult(msg pingResultMsg) Model {
	for i := range m.peers {
		if m.peers[i].TailscaleIP == msg.peerIP {
			hist := append(m.peers[i].PingHistory, msg.latency)
			if len(hist) > maxPingHistory {
				hist = hist[len(hist)-maxPingHistory:]
			}
			m.peers[i].PingHistory = hist
			break
		}
	}
	return m
}

// ── Connect popup ────────────────────────────────────────────────────────────

func (m Model) enterConnectPopup(peer tailscale.Peer) Model {
	m.connectPopup = true
	m.connectTarget = peer
	return m
}

func (m Model) handleConnectPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.connectPopup = false
	switch msg.String() {
	case "s", "enter":
		var cmd tea.Cmd
		m, cmd = m.enterSSHForm(m.connectTarget)
		return m, cmd
	case "r":
		return m, rdp.Launch(m.connectTarget.TailscaleIP)
	case "v":
		return m, vnc.Launch(m.connectTarget.TailscaleIP)
	// esc / ctrl+c — popup already closed above, nothing more to do
	}
	return m, nil
}

// ── SSH form (Huh) ───────────────────────────────────────────────────────────

// sshFormState holds the bound values Huh writes into as the user types.
// Stored as a pointer so the values survive model copies (value receiver pattern).
type sshFormState struct {
	username string
	port     string
}

var sshUsernameRe = regexp.MustCompile(`^[a-zA-Z0-9_.\-]+$`)

func (m Model) enterSSHForm(peer tailscale.Peer) (Model, tea.Cmd) {
	username := m.defaultUser
	if last, ok := m.sshUsernames[peer.Hostname]; ok {
		username = last
	}
	port := "22"
	if last, ok := m.sshPorts[peer.Hostname]; ok {
		port = last
	}

	values := &sshFormState{username: username, port: port}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Value(&values.username).
				Validate(func(s string) error {
					if !sshUsernameRe.MatchString(s) {
						return fmt.Errorf("letters, numbers, - _ . only")
					}
					return nil
				}),
			huh.NewInput().
				Title("Port").
				Value(&values.port).
				Validate(func(s string) error {
					n, err := strconv.Atoi(s)
					if err != nil || n < 1 || n > 65535 {
						return fmt.Errorf("must be 1–65535")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm()).WithWidth(m.detailWidth() - 4)

	m.sshForm = form
	m.sshFormValues = values
	m.sshTarget = peer
	return m, form.Init()
}

func (m Model) launchSSHFromForm() (tea.Model, tea.Cmd) {
	username := m.sshFormValues.username
	port := m.sshFormValues.port

	m.sshUsernames[m.sshTarget.Hostname] = username
	m.sshPorts[m.sshTarget.Hostname] = port
	config.SaveUsernames(m.sshUsernames)
	config.SavePorts(m.sshPorts)

	m.sshForm = nil
	m.sshFormValues = nil

	host := m.sshTarget.DNSName
	if host == "" {
		host = m.sshTarget.TailscaleIP
	}
	return m, ssh.Launch(username, host, port)
}

// ── Mouse ─────────────────────────────────────────────────────────────────────

// listPaneBoundary returns the x coordinate (exclusive) where the list pane ends.
// Clicks strictly left of this value hit the list; at or right hit the detail pane.
func (m Model) listPaneBoundary() int {
	return listPaneWidth + 1 // +1 for the border character
}

func (m Model) handleMouse(msg tea.MouseMsg, cmds *[]tea.Cmd) Model {
	inListPane := msg.X < m.listPaneBoundary()

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if inListPane {
			prevIdx := m.list.Index()
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			*cmds = append(*cmds, cmd)
			if m.list.Index() != prevIdx {
				m = m.refreshDetail()
				if p := m.selectedPeer(); p != nil && p.TailscaleIP != "" && !p.IsSelf && !m.pinging {
					m.pinging = true
					*cmds = append(*cmds, m.pingCmd(p.TailscaleIP))
				}
			}
		} else {
			m.viewport.LineUp(3)
		}

	case tea.MouseButtonWheelDown:
		if inListPane {
			prevIdx := m.list.Index()
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			*cmds = append(*cmds, cmd)
			if m.list.Index() != prevIdx {
				m = m.refreshDetail()
				if p := m.selectedPeer(); p != nil && p.TailscaleIP != "" && !p.IsSelf && !m.pinging {
					m.pinging = true
					*cmds = append(*cmds, m.pingCmd(p.TailscaleIP))
				}
			}
		} else {
			m.viewport.LineDown(3)
		}

	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionRelease || !inListPane {
			break
		}
		// Mouse coords are 1-indexed. Row 1 = app status bar.
		// Row 2 = list status bar (SetShowStatusBar(true) renders above items).
		// Row 3 = first list item → row 0.
		row := msg.Y - 3
		if row < 0 {
			break
		}
		// Calculate which absolute peer index the clicked row maps to.
		pageSize := m.bodyHeight()
		page := m.list.Index() / pageSize
		target := page*pageSize + row
		if target < 0 {
			target = 0
		}
		if target >= len(m.peers) {
			target = len(m.peers) - 1
		}
		// Move cursor to target by sending the list enough Up/Down messages.
		current := m.list.Index()
		diff := target - current
		keyType := tea.KeyDown
		if diff < 0 {
			keyType = tea.KeyUp
			diff = -diff
		}
		for i := 0; i < diff; i++ {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(tea.KeyMsg{Type: keyType})
			*cmds = append(*cmds, cmd)
		}
		if m.list.Index() != current {
			m = m.refreshDetail()
			if p := m.selectedPeer(); p != nil && p.TailscaleIP != "" && !p.IsSelf && !m.pinging {
				m.pinging = true
				*cmds = append(*cmds, m.pingCmd(p.TailscaleIP))
			}
		}
	}

	return m
}

// ── Commands ──────────────────────────────────────────────────────────────────

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(peerPollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) pingTickCmd() tea.Cmd {
	return tea.Tick(pingPollInterval, func(t time.Time) tea.Msg {
		return pingTickMsg(t)
	})
}

func (m Model) mascotTickCmd() tea.Cmd {
	return tea.Tick(600*time.Millisecond, func(t time.Time) tea.Msg {
		return mascotTickMsg(t)
	})
}

func (m Model) fetchPeersCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		peers, info, err := m.client.FetchStatus(ctx)
		return peersLoadedMsg{peers: peers, info: info, err: err}
	}
}

func (m Model) pingCmd(ip string) tea.Cmd {
	return func() tea.Msg {
		result := ping.Ping(ip)
		return pingResultMsg{peerIP: result.PeerIP, latency: result.Latency}
	}
}

func (m Model) toggleConnectionCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := m.client.ToggleConnection(ctx, m.info.Stopped)
		return connectionResultMsg{err: err}
	}
}

func (m Model) toggleExitNodeCmd(p *tailscale.Peer) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// If this peer is already the exit node, clear it; otherwise set it.
		id := p.StableNodeID
		if p.IsExitNode {
			id = ""
		}
		err := m.client.SetExitNode(ctx, id)
		return exitNodeResultMsg{err: err}
	}
}

func clearStatusCmd() tea.Cmd {
	return tea.Tick(statusClearDelay, func(_ time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

func pingFlashClearCmd() tea.Cmd {
	return tea.Tick(300*time.Millisecond, func(_ time.Time) tea.Msg {
		return pingFlashClearMsg{}
	})
}

func peerFlashClearCmd(hostname string) tea.Cmd {
	return tea.Tick(800*time.Millisecond, func(_ time.Time) tea.Msg {
		return peerFlashClearMsg{hostname: hostname}
	})
}

// ── Utilities ─────────────────────────────────────────────────────────────────

func copyToClipboard(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		}
	default:
		return
	}
	cmd.Stdin = strings.NewReader(text)
	_ = cmd.Run()
}
