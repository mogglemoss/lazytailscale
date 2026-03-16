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
}

// PingFailed is stored in PingHistory to indicate a ping that timed out or errored.
const PingFailed = time.Duration(-1)

// NetworkInfo holds information about our own node and the tailnet.
type NetworkInfo struct {
	NetworkName string // MagicDNS suffix, e.g. "magpie-cherimoya.ts.net"
	SelfIP      string
	SelfName    string
	Online      bool
}
