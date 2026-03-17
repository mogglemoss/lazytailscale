package ui

import (
	"fmt"
	"github.com/mogglemoss/lazytailscale/tailscale"

	"github.com/charmbracelet/lipgloss"
)

// RenderStatusBar renders the top status bar.
func RenderStatusBar(info tailscale.NetworkInfo, errMsg, returnMsg string, width int, frame int, state MascotState, exitFlash, refreshFlash bool) string {
	// Logo with animated tail — reacts to mascot state.
	tail := StatusLogoTail(frame, state)
	logo := S.StatusLogo.Render("◈") + tail + S.StatusLogo.Render("lazytailscale")

	var networkPart string
	if returnMsg != "" {
		networkPart = S.ListDotOnline.Render("✦ ") + creatureBorderReturning.Render(returnMsg)
	} else if errMsg != "" {
		networkPart = S.StatusOffline.Render(errMsg)
	} else if info.Stopped {
		networkPart = S.StatusOffline.Render("● DISCONNECTED  ·  u to reconnect")
	} else if info.NetworkName != "" {
		dot := onlineDot(info.Online)
		status := "NODE NOMINAL"
		if !info.Online {
			status = "NODE UNREACHABLE"
		}
		networkPart = S.StatusMeta.Render(fmt.Sprintf("%s · %s · %s %s",
			info.NetworkName, info.SelfIP, dot, status))

		// Exit node indicator — shown when routing traffic through a peer.
		if info.ActiveExitNode != "" {
			sep := S.StatusMeta.Render("  ·  ")
			exitHex := lipgloss.NewStyle().Foreground(S.T.Online).Render("⬡")
			if exitFlash {
				exitHex = lipgloss.NewStyle().Foreground(S.T.Online).Bold(true).Render("⬡")
			}
			networkPart += sep + exitHex + S.StatusMeta.Render(" via "+info.ActiveExitNode)
		} else if exitFlash {
			// Just deactivated — show a brief ⬡ fade.
			sep := S.StatusMeta.Render("  ·  ")
			networkPart += sep + S.StatusMeta.Render("⬡ off")
		}

		// Refresh heartbeat — brief ◦ after a successful data fetch.
		if refreshFlash {
			networkPart += lipgloss.NewStyle().Foreground(S.T.Online).Render(" ◦")
		}
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
