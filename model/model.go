package model

import (
	"context"
	"github.com/mogglemoss/lazytailscale/ping"
	"github.com/mogglemoss/lazytailscale/ssh"
	"github.com/mogglemoss/lazytailscale/tailscale"
	"github.com/mogglemoss/lazytailscale/ui"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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

	// SSH prompt state
	sshPrompting bool
	sshTarget    tailscale.Peer
	sshInput     textinput.Model
	sshUsernames map[string]string // hostname → last used username (session)
	defaultUser  string            // local system username
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
		sshUsernames: make(map[string]string),
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
			m.list.SetItems(ui.PeersToItems(m.peers))
			m = m.refreshDetail()
			if p := m.selectedPeer(); p != nil && p.TailscaleIP != "" && !p.IsSelf && !m.pinging && len(p.PingHistory) == 0 {
				m.pinging = true
				cmds = append(cmds, m.pingCmd(p.TailscaleIP))
			}
		}

	case pingResultMsg:
		m.pinging = false
		m = m.applyPingResult(msg)
		m = m.refreshDetail()

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

	case tea.MouseMsg:
		if !m.sshPrompting {
			m = m.handleMouse(msg, &cmds)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		// SSH prompt intercepts all keys while active.
		if m.sshPrompting {
			return m.handleSSHPromptKey(msg)
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
				m = m.enterSSHPrompt(*p)
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

// View renders the full TUI.
func (m Model) View() string {
	if !m.ready {
		return "\n  " + m.spinner.View() + " Loading…\n"
	}

	statusBar := ui.RenderStatusBar(m.info, m.errMsg, m.width, m.mascotFrame)

	var helpBar string
	if m.sshPrompting {
		host := m.sshTarget.DNSName
		if host == "" {
			host = m.sshTarget.TailscaleIP
		}
		helpBar = ui.RenderSSHPrompt(m.sshTarget.Hostname, host, m.sshTarget.OS, m.sshInput, m.width)
	} else {
		helpBar = ui.RenderHelpBar(m.width, m.showHelp)
	}

	listView := ui.S.PanelBorder.
		Width(listPaneWidth).
		Height(m.bodyHeight()).
		Render(m.list.View())

	detailView := m.viewport.View()

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
	m.viewport.SetContent(ui.RenderDetail(peer, m.info, m.showRoutes, m.viewport.Width, m.mascotFrame))
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

	var cmds []tea.Cmd
	for i := range fresh {
		ip := fresh[i].TailscaleIP
		if p, ok := prevMap[ip]; ok {
			fresh[i].PingHistory = p.hist
			// Notify on status transitions for non-self peers.
			if p.seen && !fresh[i].IsSelf && fresh[i].Online != p.online {
				name := fresh[i].Hostname
				if fresh[i].Online {
					m.errMsg = name + " connected"
				} else {
					m.errMsg = name + " disconnected"
				}
				cmds = append(cmds, clearStatusCmd())
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

// ── SSH prompt ───────────────────────────────────────────────────────────────

func (m Model) enterSSHPrompt(peer tailscale.Peer) Model {
	m.sshTarget = peer
	m.sshPrompting = true

	// Pre-fill with last used username for this host, or local default.
	prefill := m.defaultUser
	if last, ok := m.sshUsernames[peer.Hostname]; ok {
		prefill = last
	}

	ti := textinput.New()
	ti.SetValue(prefill)
	// Select all so the user can immediately type to replace.
	ti.CursorEnd()
	ti.Width = 20
	ti.Prompt = ""
	ti.TextStyle = ui.S.DetailValue
	ti.Cursor.Style = ui.S.StatusLogo
	ti.Focus()

	m.sshInput = ti
	return m
}

func (m Model) handleSSHPromptKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		username := strings.TrimSpace(m.sshInput.Value())
		if username == "" {
			username = m.defaultUser
		}
		m.sshUsernames[m.sshTarget.Hostname] = username
		m.sshPrompting = false
		host := m.sshTarget.DNSName
		if host == "" {
			host = m.sshTarget.TailscaleIP
		}
		return m, ssh.Launch(username, host)

	case "esc", "ctrl+c":
		m.sshPrompting = false
		return m, nil

	default:
		var cmd tea.Cmd
		m.sshInput, cmd = m.sshInput.Update(msg)
		return m, cmd
	}
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
