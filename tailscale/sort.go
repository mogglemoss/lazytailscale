package tailscale

import (
	"sort"
	"strings"
)

// sortPeers sorts peers: self first, then online before offline, then
// alphabetically by hostname (case-insensitive).
func sortPeers(peers []Peer) {
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].IsSelf != peers[j].IsSelf {
			return peers[i].IsSelf
		}
		if peers[i].Online != peers[j].Online {
			return peers[i].Online
		}
		return strings.ToLower(peers[i].Hostname) < strings.ToLower(peers[j].Hostname)
	})
}
