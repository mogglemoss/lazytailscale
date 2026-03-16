package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// gustRemarks are the flavor text remarks G.U.S.T. rotates through.
var gustRemarks = []string{
	"G.U.S.T. has been monitoring your tailnet.\nG.U.S.T. finds it adequate.",
	"All nodes accounted for.\nG.U.S.T. has no further questions at this time.",
	"Network topology: noted.\nG.U.S.T. has filed the appropriate documentation.",
	"G.U.S.T. is watching.\nThis is its designated function.\nG.U.S.T. is at peace with this.",
	"Surveillance interval: nominal.\nG.U.S.T. considers this unremarkable.",
	"G.U.S.T. has observed your peers.\nSome of them appear offline.\nG.U.S.T. has noted this without judgment.",
	"Status: OPERATIONAL.\nG.U.S.T. would like you to know it is trying its best.",
	"G.U.S.T. does not sleep.\nG.U.S.T. does not require sleep.\nG.U.S.T. is fine.",
}

// mascotFrames defines the 8-frame animation sequence.
// Each entry: [leftEye, rightEye, tail]
var mascotFrames = [8][3]string{
	{"◉", "◉", "~"},
	{"◉", "◉", "∿"},
	{"◉", "◉", "~"},
	{"─", "─", "~"},
	{"━", "━", "∿"},
	{"─", "─", "~"},
	{"◉", "◉", "∿"},
	{"◉", "◉", "~"},
}

// RenderMascot renders G.U.S.T. centered within width.
func RenderMascot(frame int, width int) string {
	f := mascotFrames[frame%8]
	leftEye := f[0]
	rightEye := f[1]
	tail := f[2]

	border := S.HelpSep
	eyes := S.StatusLogo
	tailStyle := S.StatusMeta

	// Body lines
	line0 := border.Render("  ╔═══════╗  ")
	line1 := border.Render("  ║ ") + eyes.Render(leftEye) + border.Render("   ") + eyes.Render(rightEye) + border.Render(" ║  ")
	line2 := border.Render("  ║   ▾   ║") + tailStyle.Render(tail) + border.Render(" ")
	line3 := border.Render("  ╚═══════╝  ")

	body := strings.Join([]string{line0, line1, line2, line3}, "\n")

	// Name and subtitle
	name := S.DetailHeader.Render("G.U.S.T.")
	subtitle := S.DetailLabel.Render("General Uptime Surveillance Terminal")

	// Flavor remark — rotate by frame/8 mod len
	remarkIdx := (frame / 8) % len(gustRemarks)
	remark := S.DetailLabel.Render(gustRemarks[remarkIdx])

	// Center everything
	center := func(s string) string {
		lines := strings.Split(s, "\n")
		var centered []string
		for _, line := range lines {
			lw := lipgloss.Width(line)
			pad := (width - lw) / 2
			if pad < 0 {
				pad = 0
			}
			centered = append(centered, strings.Repeat(" ", pad)+line)
		}
		return strings.Join(centered, "\n")
	}

	var b strings.Builder
	b.WriteString(center(body))
	b.WriteString("\n\n")
	b.WriteString(center(name))
	b.WriteString("\n")
	b.WriteString(center(subtitle))
	b.WriteString("\n\n")
	b.WriteString(center(remark))

	return b.String()
}
