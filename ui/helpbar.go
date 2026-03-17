package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpItem struct {
	key  string
	desc string
}

var shortHelp = []helpItem{
	{"↑/k↓/j", "navigate"},
	{"enter", "connect"},
	{"e", "exit node"},
	{"u", "connect/disconnect"},
	{"c", "copy"},
	{"/", "filter"},
	{"?", "more"},
	{"q", "quit"},
}

var fullHelp = []helpItem{
	{"↑/k", "prev node"},
	{"↓/j", "next node"},
	{"enter", "connect to node"},
	{"p", "ping selected node"},
	{"e", "toggle exit node"},
	{"u", "connect / disconnect tailscale"},
	{"r", "toggle subnet routes"},
	{"c", "copy address (dns preferred)"},
	{"/", "filter node list"},
	{"R", "refresh node list"},
	{"?", "toggle help"},
	{"q / ctrl+c", "quit"},
}

// RenderConnectPopup renders the connection type selector in place of the help bar.
// Each option has a direct letter key — press it and the action fires immediately.
// RDP is dimmed when the peer is not running Windows.
func RenderConnectPopup(width int, peerHostname, peerOS string) string {
	sep := S.HelpSep.Render("  ·  ")

	label := S.HelpKey.Render("Connect to") + " " + S.DetailHeader.Render(peerHostname) + ":"

	sshOpt := S.HelpKey.Render("s") + S.HelpDesc.Render(" ssh")

	var rdpOpt string
	if strings.EqualFold(peerOS, "windows") {
		rdpOpt = S.HelpKey.Render("r") + S.HelpDesc.Render(" rdp")
	} else {
		rdpOpt = S.PopupDim.Render("r rdp")
	}

	vncOpt := S.HelpKey.Render("v") + S.HelpDesc.Render(" vnc")
	cancel := S.HelpKey.Render("esc") + S.HelpDesc.Render(" cancel")

	bar := label + "  " + sshOpt + sep + rdpOpt + sep + vncOpt + sep + cancel
	return lipgloss.NewStyle().
		Width(width).
		Foreground(S.T.TextSecondary).
		Padding(0, 1).
		Render(bar)
}

// RenderSSHFormHint renders the help bar while the SSH Huh form is active.
func RenderSSHFormHint(width int) string {
	sep := S.HelpSep.Render("  ·  ")
	bar := S.HelpKey.Render("tab") + S.HelpDesc.Render(" next field") +
		sep + S.HelpKey.Render("enter") + S.HelpDesc.Render(" connect") +
		sep + S.HelpKey.Render("esc") + S.HelpDesc.Render(" cancel")
	return lipgloss.NewStyle().
		Width(width).
		Foreground(S.T.TextSecondary).
		Padding(0, 1).
		Render(bar)
}

// RenderHelpBar renders the bottom help bar.
func RenderHelpBar(width int, showFull bool) string {
	items := shortHelp
	if showFull {
		items = fullHelp
	}

	sep := S.HelpSep.Render("  ·  ")
	var parts []string
	for _, item := range items {
		k := S.HelpKey.Render(item.key)
		d := S.HelpDesc.Render(" " + item.desc)
		parts = append(parts, k+d)
	}

	bar := strings.Join(parts, sep)
	barWidth := lipgloss.Width(bar)
	if barWidth > width {
		// Fall back to minimal help if too narrow.
		bar = S.HelpKey.Render("?") + S.HelpDesc.Render(" help") +
			sep + S.HelpKey.Render("q") + S.HelpDesc.Render(" quit")
	}

	return lipgloss.NewStyle().
		Width(width).
		Foreground(S.T.TextSecondary).
		Padding(0, 1).
		Render(bar)
}
