package model

import (
	"context"
	"lazytailscale/ping"
	"lazytailscale/ssh"
	"lazytailscale/tailscale"
	"lazytailscale/ui"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
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

	showHelp   bool
	showRoutes bool
	pinging    bool
	ready      bool
}

// New creates and returns the initial model.
func New() Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(ui.S.T.Accent)

	return Model{
		keys:    DefaultKeyMap(),
		client:  tailscale.NewClient(),
		spinner: sp,
	}
}

// Init kicks off the first peer fetch and both tick loops.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchPeersCmd(),
		m.tickCmd(),
		m.pingTickCmd(),
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

	case peersLoadedMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			cmds = append(cmds, clearStatusCmd())
		} else {
			m.errMsg = ""
			m = m.mergePeers(msg.peers, msg.info)
			m.list.SetItems(ui.PeersToItems(m.peers))
			m = m.refreshDetail()
		}

	case pingResultMsg:
		m.pinging = false
		m = m.applyPingResult(msg)
		m = m.refreshDetail()

	case statusClearMsg:
		m.errMsg = ""

	case ssh.SSHErrorMsg:
		m.errMsg = msg.Err.Error()
		cmds = append(cmds, clearStatusCmd())

	case ssh.SSHDoneMsg:
		// TUI already resumed — nothing to do.

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
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
				return m, ssh.Launch(p.TailscaleIP)
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
				copyToClipboard(p.TailscaleIP)
			}

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

	statusBar := ui.RenderStatusBar(m.info, m.errMsg, m.width)
	helpBar := ui.RenderHelpBar(m.width, m.showHelp)

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
	peer := tailscale.Peer{}
	if p := m.selectedPeer(); p != nil {
		peer = *p
	}
	m.viewport.SetContent(ui.RenderDetail(peer, m.showRoutes, m.viewport.Width))
	return m
}

func (m Model) mergePeers(fresh []tailscale.Peer, info tailscale.NetworkInfo) Model {
	m.info = info
	histMap := make(map[string][]time.Duration, len(m.peers))
	for _, p := range m.peers {
		histMap[p.TailscaleIP] = p.PingHistory
	}
	for i := range fresh {
		if hist, ok := histMap[fresh[i].TailscaleIP]; ok {
			fresh[i].PingHistory = hist
		}
	}
	m.peers = fresh
	return m
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
