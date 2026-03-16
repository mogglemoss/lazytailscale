package ui

import "github.com/charmbracelet/lipgloss"

// Theme holds all color values for a visual theme.
// A --theme flag can swap this struct out without touching rendering logic.
type Theme struct {
	Online  lipgloss.Color
	Idle    lipgloss.Color
	Offline lipgloss.Color
	Unknown lipgloss.AdaptiveColor

	Accent        lipgloss.Color        // logo, section headers
	AccentSubtle  lipgloss.AdaptiveColor // keys in helpbar, tags
	Selected      lipgloss.AdaptiveColor // selected row bg
	Border        lipgloss.AdaptiveColor // panel dividers
	TextPrimary   lipgloss.AdaptiveColor
	TextSecondary lipgloss.AdaptiveColor
}

// Default is the Charm Native theme.
var Default = Theme{
	Online:  lipgloss.Color("#04B575"),
	Idle:    lipgloss.Color("#FFBF00"),
	Offline: lipgloss.Color("#FF5F87"),
	Unknown: lipgloss.AdaptiveColor{Light: "#9A9A9A", Dark: "#6C6C6C"},

	Accent:       lipgloss.Color("#FF5F87"),
	AccentSubtle: lipgloss.AdaptiveColor{Light: "#5B41DF", Dark: "#7B61FF"},
	Selected:     lipgloss.AdaptiveColor{Light: "#DDD9FF", Dark: "#2D2B55"},
	Border:       lipgloss.AdaptiveColor{Light: "#CCCCCC", Dark: "#3D3D3D"},
	TextPrimary:  lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FFFDF5"},
	TextSecondary: lipgloss.AdaptiveColor{Light: "#6C6A62", Dark: "#B4B2A9"},
}

// Styles holds all pre-built lipgloss styles derived from the active theme.
type Styles struct {
	T Theme

	// Status bar
	StatusBar     lipgloss.Style
	StatusLogo    lipgloss.Style
	StatusOnline  lipgloss.Style
	StatusOffline lipgloss.Style
	StatusMeta    lipgloss.Style

	// Peer list
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListDotOnline    lipgloss.Style
	ListDotIdle      lipgloss.Style
	ListDotOffline   lipgloss.Style
	ListDotUnknown   lipgloss.Style
	ListTag          lipgloss.Style
	ListStatusBar    lipgloss.Style

	// Detail panel
	DetailBorder  lipgloss.Style
	DetailHeader  lipgloss.Style
	DetailSection lipgloss.Style
	DetailLabel   lipgloss.Style
	DetailValue   lipgloss.Style

	// Ping sparkline
	SparkGood lipgloss.Style
	SparkMid  lipgloss.Style
	SparkBad  lipgloss.Style

	// Help bar
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style
	HelpSep  lipgloss.Style

	// Panel chrome
	PanelBorder lipgloss.Style
}

// New builds a Styles from the given Theme.
func New(t Theme) Styles {
	s := Styles{T: t}

	// Status bar
	s.StatusBar = lipgloss.NewStyle().
		Background(t.Selected).
		Foreground(t.TextPrimary).
		Padding(0, 1)
	s.StatusLogo = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)
	s.StatusOnline = lipgloss.NewStyle().Foreground(t.Online)
	s.StatusOffline = lipgloss.NewStyle().Foreground(t.Offline)
	s.StatusMeta = lipgloss.NewStyle().Foreground(t.TextSecondary)

	// Peer list items — no padding; the delegate controls exact width.
	s.ListItem = lipgloss.NewStyle().
		Foreground(t.TextPrimary)
	s.ListItemSelected = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)
	s.ListDotOnline = lipgloss.NewStyle().Foreground(t.Online)
	s.ListDotIdle = lipgloss.NewStyle().Foreground(t.Idle)
	s.ListDotOffline = lipgloss.NewStyle().Foreground(t.Offline)
	s.ListDotUnknown = lipgloss.NewStyle().Foreground(t.Unknown)
	s.ListTag = lipgloss.NewStyle().Foreground(t.TextSecondary)
	s.ListStatusBar = lipgloss.NewStyle().Foreground(t.TextSecondary)

	// Detail panel
	s.DetailBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(0, 1)
	s.DetailHeader = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)
	s.DetailSection = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)
	s.DetailLabel = lipgloss.NewStyle().Foreground(t.TextSecondary)
	s.DetailValue = lipgloss.NewStyle().Foreground(t.TextPrimary)

	// Ping sparkline colors
	s.SparkGood = lipgloss.NewStyle().Foreground(t.Online)
	s.SparkMid = lipgloss.NewStyle().Foreground(t.Idle)
	s.SparkBad = lipgloss.NewStyle().Foreground(t.Offline)

	// Help bar
	s.HelpKey = lipgloss.NewStyle().Foreground(t.AccentSubtle).Bold(true)
	s.HelpDesc = lipgloss.NewStyle().Foreground(t.TextSecondary)
	s.HelpSep = lipgloss.NewStyle().Foreground(t.Border)

	// Panel borders (vertical divider between panes)
	s.PanelBorder = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(t.Border)

	return s
}

// S is the active styles singleton. Omarchy theme is used when detected,
// falling back to the built-in Default.
var S = initStyles()

func initStyles() Styles {
	if theme, ok := LoadOmarchyTheme(); ok {
		return New(theme)
	}
	return New(Default)
}
