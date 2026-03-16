package ui

import "github.com/charmbracelet/lipgloss"

// creatureBorder uses AccentSubtle (purple) — visible in both dark and light terminals,
// unlike the dim border color which disappears against dark backgrounds.
var creatureBorder = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
	Light: "#5B41DF",
	Dark:  "#7B61FF",
})

type creatureFrame struct {
	eye  string
	tail string
}

// creatureFrames is the 8-frame blink/tail animation sequence.
var creatureFrames = [8]creatureFrame{
	{"◉", " "},
	{"◉", "~"},
	{"◉", "∿"},
	{"─", "~"},
	{"━", " "},
	{"─", "∿"},
	{"◉", "~"},
	{"◉", " "},
}

// CreatureLines returns the 3 lines of the small inline creature.
// Visible width: 4 chars ("╭─╮ " / "│◉│~" / "╰─╯ ").
func CreatureLines(frame int) [3]string {
	f := creatureFrames[frame%8]
	return [3]string{
		creatureBorder.Render("╭─╮"),
		creatureBorder.Render("│") + S.StatusLogo.Render(f.eye) + creatureBorder.Render("│") + S.StatusMeta.Render(f.tail),
		creatureBorder.Render("╰─╯"),
	}
}

// StatusLogoTail returns the animated tail for use in the status bar logo.
// Cycles through 4 states: space, ~, space, ∿.
func StatusLogoTail(frame int) string {
	tails := [4]string{" ", "~", " ", "∿"}
	return S.StatusMeta.Render(tails[frame%4])
}
