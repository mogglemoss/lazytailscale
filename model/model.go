package model

import (
	"context"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mogglemoss/lazytailscale/config"
	"github.com/mogglemoss/lazytailscale/ping"
	"github.com/mogglemoss/lazytailscale/rdp"
	"github.com/mogglemoss/lazytailscale/ssh"
	"github.com/mogglemoss/lazytailscale/tailscale"
	"github.com/mogglemoss/lazytailscale/ui"
	"github.com/mogglemoss/lazytailscale/vnc"
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

	sshUsernames      map[string]string // hostname → last used username (persisted)
	sshPorts          map[string]string // hostname → last used port (persisted)
	defaultUser       string            // local system username
	lastConnectedHost string            // hostname of the most recent SSH session

	// Connection modal (nil when closed).
	modal *connectModal

	// Welcome-back message shown briefly after a successful SSH session.
	returnMsg string

	// SSH error panel — shown after a failed session; any key dismisses.
	sshErr *sshErrState

	// Animation state.
	flashPeers   map[string]bool // hostname → briefly highlight this row
	pingFlash    bool            // briefly flash the sparkline after a ping result
	exitFlash    bool            // briefly flash ⬡ in status bar when exit node is toggled
	refreshFlash bool            // briefly flash ◦ in status bar after successful peer fetch
}

// New creates and returns the initial model.
func New(demo bool) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(ui.S.T.Accent)

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

	// SSH error panel intercepts input — any key dismisses it.
	if m.sshErr != nil {
		if ws, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = ws.Width
			m.height = ws.Height
			m = m.recalcLayout()
			m.ready = true
			return m, nil
		}
		if km, ok := msg.(tea.KeyMsg); ok {
			if km.String() == "ctrl+c" {
				return m, tea.Quit
			}
			m.sshErr = nil
			return m, nil
		}
		if _, ok := msg.(sshErrClearMsg); ok {
			m.sshErr = nil
		}
		return m, nil
	}

	// Modal intercepts all messages while active.
	if m.modal != nil {
		if ws, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = ws.Width
			m.height = ws.Height
			m = m.recalcLayout()
			m.ready = true
			return m, nil
		}
		if km, ok := msg.(tea.KeyMsg); ok && km.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m.updateModal(msg)
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
			m.refreshFlash = true
			cmds = append(cmds, refreshFlashClearCmd())
			var notifCmds []tea.Cmd
			m, notifCmds = m.mergePeers(msg.peers, msg.info)
			cmds = append(cmds, notifCmds...)
			if len(notifCmds) == 0 {
				m.errMsg = ""
			}
			// Don't reset items while the user is filtering — bubbles SetItems
			// calls resetFiltering() internally, wiping any active filter query.
			if m.list.FilterState() == list.Unfiltered {
				m.list.SetItems(ui.PeersToItems(m.peers, m.flashPeers))
			}
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
		if m.list.FilterState() == list.Unfiltered {
			m.list.SetItems(ui.PeersToItems(m.peers, m.flashPeers))
		}

	case statusClearMsg:
		m.errMsg = ""

	case returnMsgClearMsg:
		m.returnMsg = ""

	case sshErrClearMsg:
		m.sshErr = nil

	case exitFlashClearMsg:
		m.exitFlash = false

	case refreshFlashClearMsg:
		m.refreshFlash = false

	case exitNodeResultMsg:
		if msg.err != nil {
			m.errMsg = "exit node: " + msg.err.Error()
			cmds = append(cmds, clearStatusCmd())
		} else {
			m.exitFlash = true
			cmds = append(cmds, exitFlashClearCmd())
		}
		cmds = append(cmds, m.fetchPeersCmd())

	case connectionResultMsg:
		if msg.err != nil {
			m.errMsg = "connection: " + msg.err.Error()
			cmds = append(cmds, clearStatusCmd())
		}
		cmds = append(cmds, m.fetchPeersCmd())

	case ssh.SSHErrorMsg:
		m.sshErr = &sshErrState{host: m.lastConnectedHost, err: msg.Err}
		m.lastConnectedHost = ""
		cmds = append(cmds, sshErrClearCmd())

	case ssh.SSHDoneMsg:
		if m.lastConnectedHost != "" {
			m.returnMsg = "connection to " + m.lastConnectedHost + " concluded. welcome back to the substrate."
			m.lastConnectedHost = ""
		} else {
			m.returnMsg = "session concluded. welcome back to the substrate."
		}
		cmds = append(cmds, returnMsgClearCmd())

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
		if m.modal == nil {
			m = m.handleMouse(msg, &cmds)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
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
				m.modal = &connectModal{stage: modalStagePick, target: *p}
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

	case list.FilterMatchesMsg:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// mascotState derives the current mascot animation state from model state.
func (m Model) mascotState() ui.MascotState {
	if m.returnMsg != "" {
		return ui.MascotReturning
	}
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

	statusBar := ui.RenderStatusBar(m.info, m.errMsg, m.returnMsg, m.width, m.mascotFrame, m.mascotState(), m.exitFlash, m.refreshFlash)
	helpBar := m.renderHelpBar()

	// SSH error panel takes over the body when a session fails.
	if m.sshErr != nil {
		hint := ui.RenderModalDismissHint(m.width)
		body := lipgloss.Place(m.width, m.bodyHeight(), lipgloss.Center, lipgloss.Center,
			m.renderSSHErrPanel(),
			lipgloss.WithWhitespaceBackground(ui.S.T.ModalDimColor),
		)
		return lipgloss.JoinVertical(lipgloss.Left, statusBar, body, hint)
	}

	// Modal takes over the body when open.
	if m.modal != nil {
		body := lipgloss.Place(m.width, m.bodyHeight(), lipgloss.Center, lipgloss.Center,
			m.renderModal(),
			lipgloss.WithWhitespaceBackground(ui.S.T.ModalDimColor),
		)
		return lipgloss.JoinVertical(lipgloss.Left, statusBar, body, helpBar)
	}

	listView := m.renderListPane()
	detailView := m.viewport.View()

	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
	return lipgloss.JoinVertical(lipgloss.Left, statusBar, body, helpBar)
}

// renderListPane renders the left peer list with ▲/▼ scroll indicators when needed.
// Runs in View() (value receiver), so SetHeight on the local list copy is safe.
func (m Model) renderListPane() string {
	hasAbove := m.list.Paginator.Page > 0
	hasBelow := m.list.Paginator.Page < m.list.Paginator.TotalPages-1

	// Reduce list height to make room for any indicator lines.
	indH := 0
	if hasAbove {
		indH++
	}
	if hasBelow {
		indH++
	}
	if indH > 0 {
		m.list.SetHeight(m.bodyHeight() - indH)
	}

	var parts []string
	if hasAbove {
		parts = append(parts, ui.S.HelpDesc.Render("  ▲ more"))
	}
	parts = append(parts, m.list.View())
	if hasBelow {
		parts = append(parts, ui.S.HelpDesc.Render("  ▼ more"))
	}

	return ui.S.PanelBorder.
		Width(listPaneWidth).
		Height(m.bodyHeight()).
		Render(strings.Join(parts, "\n"))
}

// renderHelpBar returns the correct help bar for the current state.
func (m Model) renderHelpBar() string {
	if m.modal == nil {
		return ui.RenderHelpBar(m.width, m.showHelp)
	}
	switch m.modal.stage {
	case modalStagePick:
		return ui.RenderModalPickHint(m.width, m.modal.target.Hostname, m.modal.target.OS)
	case modalStageSSH:
		return ui.RenderModalSSHHint(m.width)
	}
	return ui.RenderHelpBar(m.width, m.showHelp)
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

// ── Peer helpers ──────────────────────────────────────────────────────────────

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

	type prev struct {
		online bool
		hist   []time.Duration
		seen   bool
	}
	prevMap := make(map[string]prev, len(m.peers))
	for _, p := range m.peers {
		prevMap[p.TailscaleIP] = prev{online: p.Online, hist: p.PingHistory, seen: true}
	}

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

// ── Mouse ─────────────────────────────────────────────────────────────────────

func (m Model) listPaneBoundary() int {
	return listPaneWidth + 1
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
		row := msg.Y - 3
		if row < 0 {
			break
		}
		pageSize := m.bodyHeight()
		page := m.list.Index() / pageSize
		target := page*pageSize + row
		if target < 0 {
			target = 0
		}
		if target >= len(m.peers) {
			target = len(m.peers) - 1
		}
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

func exitFlashClearCmd() tea.Cmd {
	return tea.Tick(800*time.Millisecond, func(_ time.Time) tea.Msg {
		return exitFlashClearMsg{}
	})
}

func refreshFlashClearCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(_ time.Time) tea.Msg {
		return refreshFlashClearMsg{}
	})
}

func sshErrClearCmd() tea.Cmd {
	return tea.Tick(12*time.Second, func(_ time.Time) tea.Msg {
		return sshErrClearMsg{}
	})
}

func returnMsgClearCmd() tea.Cmd {
	return tea.Tick(4*time.Second, func(_ time.Time) tea.Msg {
		return returnMsgClearMsg{}
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
