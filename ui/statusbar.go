package ui

import (
	"fmt"
	"lazytailscale/tailscale"

	"github.com/charmbracelet/lipgloss"
)

// RenderStatusBar renders the top status bar.
func RenderStatusBar(info tailscale.NetworkInfo, errMsg string, width int) string {
	logo := S.StatusLogo.Render("◈ lazytailscale")

	var networkPart string
	if errMsg != "" {
		networkPart = S.StatusOffline.Render(errMsg)
	} else if info.NetworkName != "" {
		dot := onlineDot(info.Online)
		status := "online"
		if !info.Online {
			status = "offline"
		}
		meta := S.StatusMeta.Render(fmt.Sprintf("%s · %s · %s %s",
			info.NetworkName, info.SelfIP, dot, status))
		networkPart = meta
	} else {
		networkPart = S.StatusMeta.Render("connecting…")
	}

	// Right-align the network part.
	logoWidth := lipgloss.Width(logo)
	metaWidth := lipgloss.Width(networkPart)
	gap := width - logoWidth - metaWidth - 2 // 2 for padding
	if gap < 1 {
		gap = 1
	}
	spaces := fmt.Sprintf("%*s", gap, "")

	bar := logo + spaces + networkPart
	return S.StatusBar.Width(width).Render(bar)
}

func onlineDot(online bool) string {
	if online {
		return S.StatusOnline.Render("●")
	}
	return S.StatusOffline.Render("●")
}
