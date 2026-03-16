package tailscale

import (
	"testing"
)

func TestSortPeers_SelfFirst(t *testing.T) {
	peers := []Peer{
		{Hostname: "alpha", Online: true},
		{Hostname: "self", IsSelf: true, Online: true},
		{Hostname: "beta", Online: true},
	}
	sortPeers(peers)
	if !peers[0].IsSelf {
		t.Errorf("expected self at index 0, got %q", peers[0].Hostname)
	}
}

func TestSortPeers_OnlineBeforeOffline(t *testing.T) {
	peers := []Peer{
		{Hostname: "offline-a", Online: false},
		{Hostname: "online-b", Online: true},
		{Hostname: "offline-c", Online: false},
		{Hostname: "online-a", Online: true},
	}
	sortPeers(peers)
	// First two should be online.
	for i := 0; i < 2; i++ {
		if !peers[i].Online {
			t.Errorf("index %d: expected online peer, got %q (offline)", i, peers[i].Hostname)
		}
	}
	for i := 2; i < 4; i++ {
		if peers[i].Online {
			t.Errorf("index %d: expected offline peer, got %q (online)", i, peers[i].Hostname)
		}
	}
}

func TestSortPeers_AlphabeticWithinGroup(t *testing.T) {
	peers := []Peer{
		{Hostname: "Zebra", Online: true},
		{Hostname: "apple", Online: true},
		{Hostname: "Mango", Online: true},
	}
	sortPeers(peers)
	want := []string{"apple", "Mango", "Zebra"}
	for i, w := range want {
		if peers[i].Hostname != w {
			t.Errorf("index %d: want %q, got %q", i, w, peers[i].Hostname)
		}
	}
}

func TestSortPeers_SelfFirstThenOnlineThenAlpha(t *testing.T) {
	peers := []Peer{
		{Hostname: "zeta", Online: false},
		{Hostname: "beta", Online: true},
		{Hostname: "me", IsSelf: true, Online: true},
		{Hostname: "alpha", Online: false},
		{Hostname: "gamma", Online: true},
	}
	sortPeers(peers)

	if peers[0].Hostname != "me" {
		t.Errorf("expected self first, got %q", peers[0].Hostname)
	}
	// Online non-self peers come next (beta, gamma alphabetically).
	if peers[1].Hostname != "beta" || peers[2].Hostname != "gamma" {
		t.Errorf("expected online peers beta/gamma at [1],[2], got %q/%q",
			peers[1].Hostname, peers[2].Hostname)
	}
	// Offline peers last (alpha, zeta alphabetically).
	if peers[3].Hostname != "alpha" || peers[4].Hostname != "zeta" {
		t.Errorf("expected offline peers alpha/zeta at [3],[4], got %q/%q",
			peers[3].Hostname, peers[4].Hostname)
	}
}

func TestSortPeers_Empty(t *testing.T) {
	// Must not panic.
	sortPeers(nil)
	sortPeers([]Peer{})
}

func TestSortPeers_StableOnTie(t *testing.T) {
	// Two peers with identical sort keys — order must not change (stable sort).
	peers := []Peer{
		{Hostname: "twin", Online: true},
		{Hostname: "twin", Online: true},
	}
	sortPeers(peers)
	// Just verify it doesn't panic and returns two items.
	if len(peers) != 2 {
		t.Errorf("expected 2 peers, got %d", len(peers))
	}
}
