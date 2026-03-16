package ui

import (
	"fmt"
	"github.com/mogglemoss/lazytailscale/tailscale"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// creatureVisibleWidth is the rendered width of CreatureLines entries.
const creatureVisibleWidth = 4

const sparkChars = "▁▂▃▄▅▆▇█"

// RenderDetail returns the full content string for the detail viewport.
func RenderDetail(peer tailscale.Peer, info tailscale.NetworkInfo, showRoutes bool, width int, mascotFrame int) string {
	if peer.Hostname == "" {
		return S.DetailLabel.Render("\n  No peer selected")
	}

	var b strings.Builder

	// ── Header (with inline creature for self node) ───────────────────────
	heading := peer.Hostname
	if peer.IsSelf {
		heading += "  " + S.DetailLabel.Render("(this device)")
	}

	if peer.IsSelf {
		// Render heading + IP line alongside the creature (right-aligned).
		creature := CreatureLines(mascotFrame)
		leftWidth := width - creatureVisibleWidth - 1
		if leftWidth < 1 {
			leftWidth = 1
		}

		headerLines := [3]string{
			S.DetailHeader.Render(heading),
			S.DetailLabel.Render(peer.TailscaleIP + "  ·  " + peer.DNSName),
			"",
		}
		for i := 0; i < 3; i++ {
			left := lipgloss.NewStyle().Width(leftWidth).Render(headerLines[i])
			b.WriteString(left + creature[i] + "\n")
		}
		b.WriteString("\n")

		// ── Network substrate stats ──────────────────────────────────────────
		b.WriteString(S.DetailSection.Render("NODE RECORD"))
		b.WriteString("\n")
		b.WriteString(metaRow("PLATFORM", peer.OS))
		b.WriteString(metaRow("ADDRESS", peer.TailscaleIP))
		b.WriteString(metaRow("TAILNET", info.NetworkName))
		b.WriteString("\n")

		b.WriteString(S.DetailSection.Render("NETWORK SUBSTRATE"))
		b.WriteString("\n")
		offline := info.TotalPeers - info.OnlinePeers
		b.WriteString(fmt.Sprintf("  %s  %s · %s · %s\n",
			S.DetailLabel.Render(fmt.Sprintf("%-12s", "NODES")),
			S.DetailValue.Render(fmt.Sprintf("%d enrolled", info.TotalPeers)),
			S.ListDotOnline.Render(fmt.Sprintf("%d nominal", info.OnlinePeers)),
			S.ListDotOffline.Render(fmt.Sprintf("%d unreachable", offline)),
		))
		b.WriteString("\n")
		b.WriteString(S.DetailLabel.Render("  network substrate nominal. this assessment is considered final."))
		return b.String()
	}

	// ── Non-self header ───────────────────────────────────────────────────
	b.WriteString(S.DetailHeader.Render(heading))
	b.WriteString("\n")
	b.WriteString(S.DetailLabel.Render(peer.TailscaleIP))
	if peer.DNSName != "" {
		b.WriteString(S.DetailLabel.Render("  ·  " + peer.DNSName))
	}
	b.WriteString("\n\n")

	// ── Connection (online peers only) ────────────────────────────────────
	if peer.Online {
		b.WriteString(S.DetailSection.Render("CONNECTION"))
		b.WriteString("\n")
		switch {
		case peer.CurAddr != "":
			b.WriteString("  " + S.ListDotOnline.Render("◈") + S.DetailLabel.Render(" direct") + "   " + S.DetailValue.Render(peer.CurAddr) + "\n")
		case peer.Relay != "":
			b.WriteString("  " + S.ListDotIdle.Render("◌") + S.DetailLabel.Render(" relayed") + "  " + S.DetailValue.Render("via "+peer.Relay) + "\n")
		default:
			b.WriteString("  " + S.ListDotUnknown.Render("○") + S.DetailLabel.Render(" unknown") + "\n")
		}
		b.WriteString("\n")
	}

	// ── Node Record ─────────────────────────────────────────────────────────
	b.WriteString(S.DetailSection.Render("NODE RECORD"))
	b.WriteString("\n")
	b.WriteString(metaRow("PLATFORM", peer.OS))
	b.WriteString(metaRow("LAST CONTACT", lastSeenStr(peer)))
	b.WriteString(metaRow("HANDSHAKE", lastHandshakeStr(peer)))
	if peer.IsExitNode {
		exitLabel := S.DetailLabel.Render(fmt.Sprintf("  %-12s", "EXIT NODE"))
		exitVal := lipgloss.NewStyle().Foreground(S.T.Online).Render("active")
		b.WriteString(exitLabel + "  " + exitVal + "\n")
	}
	b.WriteString("\n")

	// ── Latency Assessment ───────────────────────────────────────────────────
	b.WriteString(S.DetailSection.Render("LATENCY ASSESSMENT"))
	b.WriteString("\n")
	b.WriteString(renderSparkline(peer.PingHistory))
	b.WriteString("\n")
	b.WriteString(renderPingStats(peer.PingHistory))
	b.WriteString("\n\n")

	// ── Claimed Prefixes ─────────────────────────────────────────────────────
	// Only show subnet routes (AdvertisedRoutes = PrimaryRoutes from Tailscale).
	// We deliberately skip AllowedIPs — those are just the peer's own /32 Tailscale
	// IPs and aren't useful to display.
	if len(peer.AdvertisedRoutes) > 0 {
		b.WriteString(S.DetailSection.Render("CLAIMED PREFIXES"))
		b.WriteString("\n")
		routes := peer.AdvertisedRoutes
		display := routes
		if !showRoutes && len(routes) > 3 {
			display = routes[:3]
		}
		for _, r := range display {
			b.WriteString("  " + S.DetailValue.Render(r.String()) + "\n")
		}
		if !showRoutes && len(routes) > 3 {
			b.WriteString(S.DetailLabel.Render(fmt.Sprintf("  … %d more  (r to expand)", len(routes)-3)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// ── Classifications ──────────────────────────────────────────────────────
	if len(peer.Tags) > 0 {
		b.WriteString(S.DetailSection.Render("CLASSIFICATIONS"))
		b.WriteString("\n")
		for _, tag := range peer.Tags {
			b.WriteString("  " + S.HelpKey.Render(tag) + "\n")
		}
	}

	_ = width // reserved for future wrapping
	return b.String()
}

func metaRow(label, value string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf("  %s  %s\n",
		S.DetailLabel.Render(fmt.Sprintf("%-12s", label)),
		S.DetailValue.Render(value),
	)
}

func lastSeenStr(peer tailscale.Peer) string {
	if peer.Online {
		return "now"
	}
	if peer.LastSeen.IsZero() {
		return "unknown"
	}
	d := time.Since(peer.LastSeen).Round(time.Second)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func lastHandshakeStr(peer tailscale.Peer) string {
	if peer.LastHandshake.IsZero() {
		return ""
	}
	d := time.Since(peer.LastHandshake).Round(time.Second)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func renderSparkline(history []time.Duration) string {
	if len(history) == 0 {
		return S.DetailLabel.Render("  awaiting telemetry")
	}

	var maxD time.Duration
	for _, d := range history {
		if d != tailscale.PingFailed && d > maxD {
			maxD = d
		}
	}

	runes := []rune(sparkChars)
	var bars strings.Builder
	bars.WriteString("  ")

	for _, d := range history {
		if d == tailscale.PingFailed {
			bars.WriteString(S.SparkBad.Render("✕"))
			continue
		}
		idx := 0
		if maxD > 0 {
			idx = int(float64(d) / float64(maxD) * float64(len(runes)-1))
		}
		if idx >= len(runes) {
			idx = len(runes) - 1
		}
		bars.WriteString(sparkColorFor(d).Render(string(runes[idx])))
	}

	return bars.String()
}

func renderPingStats(history []time.Duration) string {
	if len(history) == 0 {
		return ""
	}

	var total time.Duration
	minD := time.Duration(1<<63 - 1)
	var maxD time.Duration
	count := 0

	for _, d := range history {
		if d == tailscale.PingFailed || d < 0 {
			continue
		}
		total += d
		count++
		if d < minD {
			minD = d
		}
		if d > maxD {
			maxD = d
		}
	}
	if count == 0 {
		return S.DetailLabel.Render("  node unresponsive to inquiry")
	}
	avg := total / time.Duration(count)
	return fmt.Sprintf("  %s  avg %s  min %s  max %s",
		S.DetailLabel.Render("latency"),
		sparkColorFor(avg).Render(fmtDur(avg)),
		S.DetailValue.Render(fmtDur(minD)),
		S.DetailValue.Render(fmtDur(maxD)),
	)
}

func fmtDur(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func sparkColorFor(d time.Duration) lipgloss.Style {
	switch {
	case d < 10*time.Millisecond:
		return S.SparkGood
	case d < 50*time.Millisecond:
		return S.SparkMid
	default:
		return S.SparkBad
	}
}
