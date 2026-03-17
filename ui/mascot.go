package ui

import "github.com/charmbracelet/lipgloss"

// MascotState controls which animation set the creature uses.
type MascotState int

const (
	MascotNormal  MascotState = iota // gentle blink, tail wags
	MascotPinging                    // eye flickers ◉/◌ (scanning), tail spins
	MascotOffline                    // sad eye, no tail movement
)

// creatureBorder uses AccentSubtle (purple) — visible in both dark and light terminals.
var creatureBorder = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
	Light: "#5B41DF",
	Dark:  "#7B61FF",
})

// creatureBorderOffline uses the unknown/gray color when the network is down.
var creatureBorderOffline = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
	Light: "#9A9A9A",
	Dark:  "#6C6C6C",
})

type creatureFrame struct {
	eye  string
	tail string
}

// normalFrames — gentle blink, tail wags.
var normalFrames = [8]creatureFrame{
	{"◉", " "},
	{"◉", "~"},
	{"◉", "∿"},
	{"─", "~"},
	{"━", " "},
	{"─", "∿"},
	{"◉", "~"},
	{"◉", " "},
}

// pingFrames — eye alternates ◉/◌ (scanning), tail active.
var pingFrames = [8]creatureFrame{
	{"◌", "∿"},
	{"◉", "~"},
	{"◌", "~"},
	{"◉", "∿"},
	{"◌", "∿"},
	{"◉", "~"},
	{"◌", "~"},
	{"◉", "∿"},
}

// offlineFrames — still, defeated.
var offlineFrames = [8]creatureFrame{
	{"•", " "},
	{"•", " "},
	{"•", " "},
	{"•", " "},
	{"•", " "},
	{"•", " "},
	{"•", " "},
	{"•", " "},
}

// CreatureLines returns the 3 lines of the small inline creature.
// Visible width: 4 chars ("╭─╮ " / "│◉│~" / "╰─╯ ").
func CreatureLines(frame int, state MascotState) [3]string {
	var f creatureFrame
	var border lipgloss.Style

	switch state {
	case MascotPinging:
		f = pingFrames[frame%8]
		border = creatureBorder
	case MascotOffline:
		f = offlineFrames[frame%8]
		border = creatureBorderOffline
	default:
		f = normalFrames[frame%8]
		border = creatureBorder
	}

	return [3]string{
		border.Render("╭─╮"),
		border.Render("│") + S.StatusLogo.Render(f.eye) + border.Render("│") + S.StatusMeta.Render(f.tail),
		border.Render("╰─╯"),
	}
}

// StatusLogoTail returns the animated tail for use in the status bar logo.
// Reacts to mascot state: active when pinging, still when offline.
func StatusLogoTail(frame int, state MascotState) string {
	switch state {
	case MascotOffline:
		return S.StatusMeta.Render(" ")
	case MascotPinging:
		tails := [4]string{"~", "∿", "~", "∿"}
		return S.StatusMeta.Render(tails[frame%4])
	default:
		tails := [4]string{" ", "~", " ", "∿"}
		return S.StatusMeta.Render(tails[frame%4])
	}
}
