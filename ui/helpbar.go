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
	{"enter", "ssh"},
	{"p", "ping"},
	{"r", "routes"},
	{"c", "copy ip"},
	{"/", "filter"},
	{"?", "more"},
	{"q", "quit"},
}

var fullHelp = []helpItem{
	{"↑/k", "prev peer"},
	{"↓/j", "next peer"},
	{"enter", "ssh into peer"},
	{"p", "ping now"},
	{"r", "toggle routes"},
	{"c", "copy ip"},
	{"/", "filter list"},
	{"R", "refresh"},
	{"?", "toggle help"},
	{"q / ctrl+c", "quit"},
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
