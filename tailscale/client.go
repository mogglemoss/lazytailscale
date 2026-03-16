package tailscale

import (
	"context"
	"sort"
	"strings"

	ts "tailscale.com/client/tailscale"
	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/tailcfg"
)

// Client wraps the Tailscale LocalClient.
type Client struct {
	lc             ts.LocalClient
	demo           bool
	demoExitNodeID string // tracks the active exit node in demo mode
}

// NewClient returns a new Client using the local tailscaled socket.
// If demo is true, FetchStatus returns static fictional data.
func NewClient(demo bool) *Client {
	return &Client{demo: demo}
}

// FetchStatus returns the current peer list and network info.
func (c *Client) FetchStatus(ctx context.Context) ([]Peer, NetworkInfo, error) {
	if c.demo {
		return c.demoStatus()
	}
	st, err := c.lc.Status(ctx)
	if err != nil {
		return nil, NetworkInfo{}, err
	}

	var peers []Peer

	// Add the local node first (marked as self).
	if st.Self != nil {
		self := peerFromStatus(st.Self)
		self.IsSelf = true
		peers = append(peers, self)
	}

	for _, ps := range st.Peer {
		peers = append(peers, peerFromStatus(ps))
	}

	// Sort: self first, then online, then alphabetical.
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].IsSelf != peers[j].IsSelf {
			return peers[i].IsSelf
		}
		if peers[i].Online != peers[j].Online {
			return peers[i].Online
		}
		return strings.ToLower(peers[i].Hostname) < strings.ToLower(peers[j].Hostname)
	})

	info := NetworkInfo{
		NetworkName: st.MagicDNSSuffix,
		Stopped:     st.BackendState == "Stopped",
	}
	if st.Self != nil {
		info.SelfName = st.Self.HostName
		info.Online = st.Self.Online
		for _, ip := range st.Self.TailscaleIPs {
			if ip.Is4() {
				info.SelfIP = ip.String()
				break
			}
		}
	}

	for _, p := range peers {
		if p.IsSelf {
			continue
		}
		info.TotalPeers++
		if p.Online {
			info.OnlinePeers++
		}
	}

	return peers, info, nil
}

// ToggleConnection connects or disconnects the local node by flipping WantRunning.
func (c *Client) ToggleConnection(ctx context.Context, wantRunning bool) error {
	if c.demo {
		return nil
	}
	_, err := c.lc.EditPrefs(ctx, &ipn.MaskedPrefs{
		WantRunningSet: true,
		Prefs:          ipn.Prefs{WantRunning: wantRunning},
	})
	return err
}

// SetExitNode sets the given peer as the exit node, or clears it if stableID is empty.
func (c *Client) SetExitNode(ctx context.Context, stableID string) error {
	if c.demo {
		c.demoExitNodeID = stableID
		return nil
	}
	_, err := c.lc.EditPrefs(ctx, &ipn.MaskedPrefs{
		ExitNodeIDSet: true,
		Prefs: ipn.Prefs{
			ExitNodeID: tailcfg.StableNodeID(stableID),
		},
	})
	return err
}

func peerFromStatus(ps *ipnstate.PeerStatus) Peer {
	p := Peer{
		Hostname:      ps.HostName,
		DNSName:       strings.TrimSuffix(ps.DNSName, "."),
		OS:            ps.OS,
		Online:        ps.Online,
		LastSeen:      ps.LastSeen,
		CurAddr:       ps.CurAddr,
		Relay:         ps.Relay,
		IsExitNode:    ps.ExitNode,
		CanBeExitNode: ps.ExitNodeOption,
		StableNodeID:  string(ps.ID),
		LastHandshake: ps.LastHandshake,
	}
	if ps.KeyExpiry != nil {
		p.KeyExpiry = *ps.KeyExpiry
	}

	for _, ip := range ps.TailscaleIPs {
		if ip.Is4() {
			p.TailscaleIP = ip.String()
			break
		}
	}

	if ps.AllowedIPs != nil {
		p.AllowedIPs = ps.AllowedIPs.AsSlice()
	}
	if ps.PrimaryRoutes != nil {
		p.AdvertisedRoutes = ps.PrimaryRoutes.AsSlice()
	}
	if ps.Tags != nil {
		p.Tags = ps.Tags.AsSlice()
	}

	return p
}
