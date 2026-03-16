package ping

import (
	"context"
	"net/netip"
	"time"

	ts "tailscale.com/client/tailscale"
	"tailscale.com/tailcfg"
)

const Timeout = 5 * time.Second

// Result holds the outcome of a single ping.
type Result struct {
	PeerIP  string
	Latency time.Duration // -1 on failure
}

// Ping sends a single TSMP ping to the given IP and returns the result.
// It is safe to call from a tea.Cmd goroutine.
func Ping(ip string) Result {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return Result{PeerIP: ip, Latency: -1}
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	lc := &ts.LocalClient{}
	pr, err := lc.Ping(ctx, addr, tailcfg.PingTSMP)
	if err != nil || pr == nil {
		return Result{PeerIP: ip, Latency: -1}
	}

	latency := time.Duration(pr.LatencySeconds * float64(time.Second))
	return Result{PeerIP: ip, Latency: latency}
}
