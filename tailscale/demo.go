package tailscale

import (
	"net/netip"
	"time"
)

// demoStatus returns a fictional but realistic tailnet for screenshots and
// demos. No real network data is accessed.
func (c *Client) demoStatus() ([]Peer, NetworkInfo, error) {
	now := time.Now()

	peers := []Peer{
		// Self node
		{
			Hostname:      "workstation",
			TailscaleIP:   "100.64.0.1",
			DNSName:       "workstation.magpie-lab.ts.net",
			OS:            "darwin",
			Online:        true,
			IsSelf:        true,
			StableNodeID:  "demo-self",
			LastHandshake: now.Add(-30 * time.Second),
		},
		// Online peers
		{
			Hostname:      "homelab-nas",
			TailscaleIP:   "100.64.0.2",
			DNSName:       "homelab-nas.magpie-lab.ts.net",
			OS:            "linux",
			Online:        true,
			CurAddr:       "192.168.1.10:41641",
			StableNodeID:  "demo-nas",
			LastHandshake: now.Add(-12 * time.Second),
			AdvertisedRoutes: []netip.Prefix{
				netip.MustParsePrefix("192.168.1.0/24"),
				netip.MustParsePrefix("192.168.2.0/24"),
				netip.MustParsePrefix("10.0.10.0/24"),
				netip.MustParsePrefix("10.0.20.0/24"),
			},
			PingHistory: []time.Duration{
				3 * time.Millisecond,
				2 * time.Millisecond,
				4 * time.Millisecond,
				3 * time.Millisecond,
				2 * time.Millisecond,
				3 * time.Millisecond,
				4 * time.Millisecond,
				3 * time.Millisecond,
			},
		},
		{
			Hostname:      "cloud-vps",
			TailscaleIP:   "100.64.0.3",
			DNSName:       "cloud-vps.magpie-lab.ts.net",
			OS:            "linux",
			Online:        true,
			CurAddr:       "203.0.113.42:41641",
			IsExitNode:    c.demoExitNodeID == "demo-vps",
			CanBeExitNode: true,
			StableNodeID:  "demo-vps",
			LastHandshake: now.Add(-8 * time.Second),
			PingHistory: []time.Duration{
				42 * time.Millisecond,
				38 * time.Millisecond,
				45 * time.Millisecond,
				PingFailed,
				41 * time.Millisecond,
				39 * time.Millisecond,
				44 * time.Millisecond,
				43 * time.Millisecond,
			},
		},
		{
			Hostname:      "macbook-air",
			TailscaleIP:   "100.64.0.4",
			DNSName:       "macbook-air.magpie-lab.ts.net",
			OS:            "darwin",
			Online:        true,
			CurAddr:       "192.168.1.55:41641",
			StableNodeID:  "demo-mba",
			LastHandshake: now.Add(-2 * time.Minute),
			PingHistory: []time.Duration{
				5 * time.Millisecond,
				6 * time.Millisecond,
				5 * time.Millisecond,
				7 * time.Millisecond,
				5 * time.Millisecond,
				6 * time.Millisecond,
				5 * time.Millisecond,
				6 * time.Millisecond,
			},
		},
		{
			Hostname:      "raspberry-pi",
			TailscaleIP:   "100.64.0.5",
			DNSName:       "raspberry-pi.magpie-lab.ts.net",
			OS:            "linux",
			Online:        true,
			Relay:         "syd",
			StableNodeID:  "demo-pi",
			LastHandshake: now.Add(-45 * time.Second),
			PingHistory: []time.Duration{
				18 * time.Millisecond,
				22 * time.Millisecond,
				19 * time.Millisecond,
				21 * time.Millisecond,
				20 * time.Millisecond,
				18 * time.Millisecond,
				23 * time.Millisecond,
				19 * time.Millisecond,
			},
		},
		{
			Hostname:      "iphone",
			TailscaleIP:   "100.64.0.6",
			DNSName:       "iphone.magpie-lab.ts.net",
			OS:            "ios",
			Online:        true,
			Relay:         "syd",
			StableNodeID:  "demo-iphone",
			LastHandshake: now.Add(-90 * time.Second),
			PingHistory: []time.Duration{
				31 * time.Millisecond,
				PingFailed,
				28 * time.Millisecond,
				35 * time.Millisecond,
				29 * time.Millisecond,
			},
		},
		// Offline peers
		{
			Hostname:     "build-server",
			TailscaleIP:  "100.64.0.7",
			DNSName:      "build-server.magpie-lab.ts.net",
			OS:           "linux",
			Online:       false,
			StableNodeID: "demo-build",
			LastSeen:     now.Add(-3 * time.Hour),
			Tags:         []string{"tag:ci", "tag:server"},
		},
		{
			Hostname:     "old-desktop",
			TailscaleIP:  "100.64.0.8",
			DNSName:      "old-desktop.magpie-lab.ts.net",
			OS:           "windows",
			Online:       false,
			StableNodeID: "demo-desktop",
			LastSeen:     now.Add(-2 * time.Hour),
		},
		{
			Hostname:     "tablet",
			TailscaleIP:  "100.64.0.9",
			DNSName:      "tablet.magpie-lab.ts.net",
			OS:           "android",
			Online:       false,
			StableNodeID: "demo-tablet",
			LastSeen:     now.Add(-5 * 24 * time.Hour),
		},
	}

	info := NetworkInfo{
		NetworkName: "magpie-lab.ts.net",
		SelfIP:      "100.64.0.1",
		SelfName:    "workstation",
		Online:      true,
		TotalPeers:  8,
		OnlinePeers: 5,
	}

	return peers, info, nil
}
