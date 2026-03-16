package ui

import (
	"fmt"
	"github.com/mogglemoss/lazytailscale/tailscale"

	"github.com/charmbracelet/lipgloss"
)

// RenderStatusBar renders the top status bar.
func RenderStatusBar(info tailscale.NetworkInfo, errMsg string, width int, frame int) string {
	// Logo with animated tail — one subtle character that blinks.
	tail := StatusLogoTail(frame)
	logo := S.StatusLogo.Render("◈") + tail + S.StatusLogo.Render("lazytailscale")

	var networkPart string
	if errMsg != "" {
		networkPart = S.StatusOffline.Render(errMsg)
	} else if info.NetworkName != "" {
		dot := onlineDot(info.Online)
		status := "NOMINAL"
		if !info.Online {
			status = "UNREACHABLE"
		}
		networkPart = S.StatusMeta.Render(fmt.Sprintf("%s · %s · %s %s",
			info.NetworkName, info.SelfIP, dot, status))
	} else {
		networkPart = S.StatusMeta.Render("establishing substrate awareness…")
	}

	// Right-align the network part.
	logoWidth := lipgloss.Width(logo)
	metaWidth := lipgloss.Width(networkPart)
	gap := width - logoWidth - metaWidth - 2 // 2 for padding
	if gap < 1 {
		gap = 1
	}

	bar := logo + fmt.Sprintf("%*s", gap, "") + networkPart
	return S.StatusBar.Width(width).Render(bar)
}

func onlineDot(online bool) string {
	if online {
		return S.StatusOnline.Render("●")
	}
	return S.StatusOffline.Render("●")
}
