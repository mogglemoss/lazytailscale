package tailscale

import (
	"context"
	"sort"
	"strings"

	ts "tailscale.com/client/tailscale"
	"tailscale.com/ipn/ipnstate"
)

// Client wraps the Tailscale LocalClient.
type Client struct {
	lc ts.LocalClient
}

// NewClient returns a new Client using the local tailscaled socket.
func NewClient() *Client {
	return &Client{}
}

// FetchStatus returns the current peer list and network info.
func (c *Client) FetchStatus(ctx context.Context) ([]Peer, NetworkInfo, error) {
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
		LastHandshake: ps.LastHandshake,
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
