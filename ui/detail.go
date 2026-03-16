package ui

import (
	"fmt"
	"github.com/mogglemoss/lazytailscale/tailscale"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const sparkChars = "▁▂▃▄▅▆▇█"

// RenderDetail returns the full content string for the detail viewport.
func RenderDetail(peer tailscale.Peer, showRoutes bool, width int) string {
	if peer.Hostname == "" {
		return S.DetailLabel.Render("\n  No peer selected")
	}

	var b strings.Builder

	// ── Header ──────────────────────────────────────────────────────────────
	heading := peer.Hostname
	if peer.IsSelf {
		heading += "  " + S.DetailLabel.Render("(this device)")
	}
	b.WriteString(S.DetailHeader.Render(heading))
	b.WriteString("\n")
	b.WriteString(S.DetailLabel.Render(peer.TailscaleIP))
	if peer.DNSName != "" {
		b.WriteString(S.DetailLabel.Render("  ·  " + peer.DNSName))
	}
	b.WriteString("\n\n")

	// ── Meta ─────────────────────────────────────────────────────────────────
	b.WriteString(S.DetailSection.Render("META"))
	b.WriteString("\n")
	b.WriteString(metaRow("OS", peer.OS))
	b.WriteString(metaRow("Last seen", lastSeenStr(peer)))
	b.WriteString("\n")

	// ── Ping ─────────────────────────────────────────────────────────────────
	b.WriteString(S.DetailSection.Render("PING"))
	b.WriteString("\n")
	b.WriteString(renderSparkline(peer.PingHistory))
	b.WriteString("\n")
	b.WriteString(renderPingStats(peer.PingHistory))
	b.WriteString("\n\n")

	// ── Routes ───────────────────────────────────────────────────────────────
	// Only show subnet routes (AdvertisedRoutes = PrimaryRoutes from Tailscale).
	// We deliberately skip AllowedIPs — those are just the peer's own /32 Tailscale
	// IPs and aren't useful to display.
	if len(peer.AdvertisedRoutes) > 0 {
		b.WriteString(S.DetailSection.Render("ROUTES"))
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

	// ── Tags ─────────────────────────────────────────────────────────────────
	if len(peer.Tags) > 0 {
		b.WriteString(S.DetailSection.Render("TAGS"))
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
		S.DetailLabel.Render(fmt.Sprintf("%-10s", label)),
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

func renderSparkline(history []time.Duration) string {
	if len(history) == 0 {
		return S.DetailLabel.Render("  no data yet")
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
		return S.DetailLabel.Render("  all pings failed")
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
