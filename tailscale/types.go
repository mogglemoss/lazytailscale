package tailscale

import (
	"net/netip"
	"time"
)

// Peer is our internal representation of a Tailscale peer node.
type Peer struct {
	Hostname         string
	TailscaleIP      string // first IPv4 from TailscaleIPs
	DNSName          string // trimmed trailing dot
	OS               string
	Online           bool
	IsSelf           bool // true for the local machine
	LastSeen         time.Time
	AdvertisedRoutes []netip.Prefix
	AllowedIPs       []netip.Prefix
	PingHistory      []time.Duration // last 8 results, most recent last; -1 = failed
	Tags             []string

	CurAddr        string    // direct address if peer is connected directly
	Relay          string    // DERP relay region name if relayed
	IsExitNode     bool      // true if this is the currently active exit node
	CanBeExitNode  bool      // true if this peer advertises exit node capability
	StableNodeID   string    // stable identifier used to set exit node via API
	LastHandshake  time.Time // time of last WireGuard handshake
	KeyExpiry      time.Time // when this peer's Tailscale key expires (zero = no expiry)
}

// PingFailed is stored in PingHistory to indicate a ping that timed out or errored.
const PingFailed = time.Duration(-1)

// NetworkInfo holds information about our own node and the tailnet.
type NetworkInfo struct {
	NetworkName  string // MagicDNS suffix, e.g. "magpie-cherimoya.ts.net"
	SelfIP       string
	SelfName     string
	Online       bool
	Stopped      bool // true when the user has explicitly disconnected (WantRunning=false)
	TotalPeers   int  // excludes self
	OnlinePeers  int  // excludes self
}
