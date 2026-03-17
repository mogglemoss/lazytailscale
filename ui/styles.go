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
	ModalDimColor lipgloss.AdaptiveColor // background behind centered modal panels
}

// Default is the Charm Native theme.
var Default = Theme{
	Online:  lipgloss.Color("#04B575"),
	Idle:    lipgloss.Color("#FFBF00"),
	Offline: lipgloss.Color("#FF5F87"),
	Unknown: lipgloss.AdaptiveColor{Light: "#9A9A9A", Dark: "#6C6C6C"},

	Accent:        lipgloss.Color("#FF5F87"),
	AccentSubtle:  lipgloss.AdaptiveColor{Light: "#5B41DF", Dark: "#7B61FF"},
	Selected:      lipgloss.AdaptiveColor{Light: "#DDD9FF", Dark: "#2D2B55"},
	Border:        lipgloss.AdaptiveColor{Light: "#CCCCCC", Dark: "#3D3D3D"},
	TextPrimary:   lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FFFDF5"},
	TextSecondary: lipgloss.AdaptiveColor{Light: "#6C6A62", Dark: "#B4B2A9"},
	ModalDimColor: lipgloss.AdaptiveColor{Light: "#e4e4ee", Dark: "#17171f"},
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

	// Connect popup
	PopupSelected lipgloss.Style
	PopupDim      lipgloss.Style
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

	// Connect popup
	s.PopupSelected = lipgloss.NewStyle().
		Background(t.Selected).
		Foreground(t.Accent).
		Bold(true).
		Padding(0, 1)
	s.PopupDim = lipgloss.NewStyle().Foreground(t.Unknown)

	return s
}

// Presets is the map of named built-in themes selectable via --theme.
var Presets = map[string]Theme{
	"default": Default,
	"catppuccin": {
		Online:        lipgloss.Color("#a6e3a1"),
		Idle:          lipgloss.Color("#f9e2af"),
		Offline:       lipgloss.Color("#f38ba8"),
		Unknown:       lipgloss.AdaptiveColor{Light: "#7c7f93", Dark: "#6c7086"},
		Accent:        lipgloss.Color("#cba6f7"),
		AccentSubtle:  lipgloss.AdaptiveColor{Light: "#1e66f5", Dark: "#89b4fa"},
		Selected:      lipgloss.AdaptiveColor{Light: "#dce0e8", Dark: "#313244"},
		Border:        lipgloss.AdaptiveColor{Light: "#bcc0cc", Dark: "#45475a"},
		TextPrimary:   lipgloss.AdaptiveColor{Light: "#4c4f69", Dark: "#cdd6f4"},
		TextSecondary: lipgloss.AdaptiveColor{Light: "#6c6f85", Dark: "#9399b2"},
		ModalDimColor: lipgloss.AdaptiveColor{Light: "#e0e4ec", Dark: "#14141e"},
	},
	"dracula": {
		Online:        lipgloss.Color("#50fa7b"),
		Idle:          lipgloss.Color("#f1fa8c"),
		Offline:       lipgloss.Color("#ff5555"),
		Unknown:       lipgloss.AdaptiveColor{Light: "#6272a4", Dark: "#6272a4"},
		Accent:        lipgloss.Color("#bd93f9"),
		AccentSubtle:  lipgloss.AdaptiveColor{Light: "#6272a4", Dark: "#8be9fd"},
		Selected:      lipgloss.AdaptiveColor{Light: "#f8f8f2", Dark: "#44475a"},
		Border:        lipgloss.AdaptiveColor{Light: "#6272a4", Dark: "#6272a4"},
		TextPrimary:   lipgloss.AdaptiveColor{Light: "#282a36", Dark: "#f8f8f2"},
		TextSecondary: lipgloss.AdaptiveColor{Light: "#6272a4", Dark: "#6272a4"},
		ModalDimColor: lipgloss.AdaptiveColor{Light: "#e8e8f0", Dark: "#161620"},
	},
	"nord": {
		Online:        lipgloss.Color("#a3be8c"),
		Idle:          lipgloss.Color("#ebcb8b"),
		Offline:       lipgloss.Color("#bf616a"),
		Unknown:       lipgloss.AdaptiveColor{Light: "#7b88a1", Dark: "#4c566a"},
		Accent:        lipgloss.Color("#88c0d0"),
		AccentSubtle:  lipgloss.AdaptiveColor{Light: "#5e81ac", Dark: "#81a1c1"},
		Selected:      lipgloss.AdaptiveColor{Light: "#e5e9f0", Dark: "#3b4252"},
		Border:        lipgloss.AdaptiveColor{Light: "#d8dee9", Dark: "#4c566a"},
		TextPrimary:   lipgloss.AdaptiveColor{Light: "#2e3440", Dark: "#eceff4"},
		TextSecondary: lipgloss.AdaptiveColor{Light: "#4c566a", Dark: "#7b88a1"},
		ModalDimColor: lipgloss.AdaptiveColor{Light: "#e8ecf0", Dark: "#181c22"},
	},
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

// SetTheme applies a named preset theme, overriding Omarchy detection.
// Call this before model.New() so all rendered styles use the new theme.
// Unknown names are silently ignored.
func SetTheme(name string) {
	if t, ok := Presets[name]; ok {
		S = New(t)
	}
}
