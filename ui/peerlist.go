package ui

import (
	"fmt"
	"io"
	"github.com/mogglemoss/lazytailscale/tailscale"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PeerItem wraps a tailscale.Peer so it implements list.Item.
type PeerItem struct {
	Peer tailscale.Peer
}

func (p PeerItem) FilterValue() string { return p.Peer.Hostname }
func (p PeerItem) Title() string       { return p.Peer.Hostname }
func (p PeerItem) Description() string { return p.Peer.TailscaleIP }

// PeerDelegate renders each peer row.
type PeerDelegate struct{}

const (
	listWidth   = 28
	// Row layout (28 cols total):
	//   cursor(1) + sp(1) + dot(1) + sp(1) + hostname(hostnameMax) + sp(1) + tag(tagCols)
	//   1 + 1 + 1 + 1 + 16 + 1 + 7 = 28
	hostnameMax = 16
	tagCols     = 7
)

func (d PeerDelegate) Height() int                             { return 1 }
func (d PeerDelegate) Spacing() int                            { return 0 }
func (d PeerDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d PeerDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	pi, ok := item.(PeerItem)
	if !ok {
		return
	}
	peer := pi.Peer
	selected := index == m.Index()

	// Cursor column: 1 char wide.
	cursor := " "
	if selected {
		cursor = S.ListDotOnline.Render("▶")
	}

	dot := statusDot(peer)

	// Hostname: truncated, padded to hostnameMax.
	hostname := truncate(peer.Hostname, hostnameMax)
	if peer.IsSelf {
		hostname = truncate(peer.Hostname, hostnameMax-1) + "~" // trailing ~ marks self
	}
	hostPadded := hostname + strings.Repeat(" ", hostnameMax-utf8.RuneCountInString(hostname))

	// OS tag: fixed tagCols wide, right-padded.
	tag := truncate(osTag(peer.OS), tagCols)
	tagPad := tagCols - utf8.RuneCountInString(tag)
	if tagPad < 0 {
		tagPad = 0
	}
	tagStr := S.ListTag.Render(tag + strings.Repeat(" ", tagPad))

	var hostnameStr string
	if selected {
		hostnameStr = S.ListItemSelected.Render(hostPadded)
	} else {
		hostnameStr = S.ListItem.Render(hostPadded)
	}

	// Final row — exactly listWidth visual chars.
	row := cursor + " " + dot + " " + hostnameStr + " " + tagStr
	fmt.Fprint(w, row)
}

func statusDot(peer tailscale.Peer) string {
	if peer.Online {
		return S.ListDotOnline.Render("●")
	}
	if peer.TailscaleIP == "" {
		return S.ListDotUnknown.Render("○")
	}
	if !peer.LastSeen.IsZero() && time.Since(peer.LastSeen) < 5*time.Minute {
		return S.ListDotIdle.Render("●")
	}
	return S.ListDotOffline.Render("●")
}

func osTag(os string) string {
	switch strings.ToLower(os) {
	case "linux":
		return "Linux"
	case "darwin", "macos":
		return "Mac"
	case "windows":
		return "Win"
	case "android":
		return "Android"
	case "ios":
		return "iOS"
	default:
		if os == "" {
			return ""
		}
		return os
	}
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}

// NewPeerList returns a configured bubbles list.Model for the peer pane.
func NewPeerList(peers []tailscale.Peer, height int) list.Model {
	items := PeersToItems(peers)

	l := list.New(items, PeerDelegate{}, listWidth, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(S.T.TextSecondary).
		Padding(1, 2)

	return l
}

// PeersToItems converts a peer slice to list items.
func PeersToItems(peers []tailscale.Peer) []list.Item {
	items := make([]list.Item, len(peers))
	for i, p := range peers {
		items[i] = PeerItem{Peer: p}
	}
	return items
}
