package model

import (
	"github.com/mogglemoss/lazytailscale/tailscale"
	"time"
)

// tickMsg fires on the peer-poll interval (5s).
type tickMsg time.Time

// pingTickMsg fires on the ping interval (10s).
type pingTickMsg time.Time

// peersLoadedMsg carries a freshly-fetched peer list and network info.
type peersLoadedMsg struct {
	peers []tailscale.Peer
	info  tailscale.NetworkInfo
	err   error
}

// pingResultMsg carries the result of a single ping.
type pingResultMsg struct {
	peerIP  string
	latency time.Duration // PingFailed (-1) on error
}

// statusClearMsg clears the status bar error message.
type statusClearMsg struct{}

// exitNodeResultMsg carries the result of a SetExitNode call.
type exitNodeResultMsg struct{ err error }

// connectionResultMsg carries the result of a ToggleConnection call.
type connectionResultMsg struct{ err error }

// mascotTickMsg fires on the mascot animation interval (600ms).
type mascotTickMsg time.Time
