package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type helpItem struct {
	key  string
	desc string
}

var shortHelp = []helpItem{
	{"↑/k↓/j", "navigate"},
	{"enter", "ssh"},
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
	{"enter", "ssh into node"},
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

// RenderSSHPrompt renders the SSH username prompt, replacing the help bar.
func RenderSSHPrompt(hostname, host, os string, input textinput.Model, width int) string {
	sep := S.HelpSep.Render("  ·  ")

	peer := S.DetailHeader.Render(hostname)
	addr := S.DetailLabel.Render(host)

	var note string
	if strings.ToLower(os) == "windows" {
		note = sep + S.ListDotIdle.Render("rdp may be more appropriate")
	}

	left := S.HelpKey.Render("ssh") + " " + peer + "  " + addr + note
	mid := "  " + S.HelpKey.Render("as:") + " " + input.View()
	right := sep + S.HelpKey.Render("enter") + S.HelpDesc.Render(" connect") +
		sep + S.HelpKey.Render("esc") + S.HelpDesc.Render(" cancel")

	bar := left + mid + right
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
