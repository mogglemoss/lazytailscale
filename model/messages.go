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

// pingFlashClearMsg clears the sparkline flash indicator after it fires.
type pingFlashClearMsg struct{}

// peerFlashClearMsg clears the flash highlight on a specific peer row.
type peerFlashClearMsg struct{ hostname string }

// returnMsgClearMsg clears the welcome-back message after it has displayed.
type returnMsgClearMsg struct{}

// sshErrClearMsg auto-dismisses the SSH error panel if the user ignores it.
type sshErrClearMsg struct{}

// exitFlashClearMsg clears the ⬡ flash after exit node is toggled.
type exitFlashClearMsg struct{}

// refreshFlashClearMsg clears the ◦ heartbeat after a successful peer fetch.
type refreshFlashClearMsg struct{}

// rdpDoneMsg is sent when an RDP client launches successfully.
type rdpDoneMsg struct{}

// rdpErrorMsg is sent when an RDP client fails to launch.
type rdpErrorMsg struct{ err error }

// vncDoneMsg is sent when a VNC viewer launches successfully.
type vncDoneMsg struct{}

// vncErrorMsg is sent when a VNC viewer fails to launch.
type vncErrorMsg struct{ err error }
