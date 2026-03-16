package model

import (
	"testing"
	"time"

	"github.com/mogglemoss/lazytailscale/tailscale"
)

// minModel returns a Model with just enough state for unit tests
// that don't need the TUI components.
func minModel(peers []tailscale.Peer) Model {
	return Model{peers: peers}
}

// ── applyPingResult ───────────────────────────────────────────────────────────

func TestApplyPingResult_AppendToHistory(t *testing.T) {
	m := minModel([]tailscale.Peer{
		{TailscaleIP: "100.1.1.1", PingHistory: nil},
	})
	m = m.applyPingResult(pingResultMsg{peerIP: "100.1.1.1", latency: 5 * time.Millisecond})

	if len(m.peers[0].PingHistory) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(m.peers[0].PingHistory))
	}
	if m.peers[0].PingHistory[0] != 5*time.Millisecond {
		t.Errorf("expected 5ms, got %v", m.peers[0].PingHistory[0])
	}
}

func TestApplyPingResult_CapsAtMaxHistory(t *testing.T) {
	hist := make([]time.Duration, maxPingHistory)
	for i := range hist {
		hist[i] = time.Duration(i) * time.Millisecond
	}
	m := minModel([]tailscale.Peer{
		{TailscaleIP: "100.1.1.1", PingHistory: hist},
	})

	m = m.applyPingResult(pingResultMsg{peerIP: "100.1.1.1", latency: 99 * time.Millisecond})

	if len(m.peers[0].PingHistory) != maxPingHistory {
		t.Errorf("expected history capped at %d, got %d", maxPingHistory, len(m.peers[0].PingHistory))
	}
	last := m.peers[0].PingHistory[len(m.peers[0].PingHistory)-1]
	if last != 99*time.Millisecond {
		t.Errorf("expected newest entry 99ms at tail, got %v", last)
	}
}

func TestApplyPingResult_WrongIP(t *testing.T) {
	m := minModel([]tailscale.Peer{
		{TailscaleIP: "100.1.1.1"},
	})
	m = m.applyPingResult(pingResultMsg{peerIP: "100.9.9.9", latency: 10 * time.Millisecond})

	// No peer matched — history must remain empty.
	if len(m.peers[0].PingHistory) != 0 {
		t.Errorf("unexpected history entry for non-matching IP")
	}
}

func TestApplyPingResult_RecordsFailed(t *testing.T) {
	m := minModel([]tailscale.Peer{
		{TailscaleIP: "100.1.1.1"},
	})
	m = m.applyPingResult(pingResultMsg{peerIP: "100.1.1.1", latency: tailscale.PingFailed})

	if len(m.peers[0].PingHistory) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(m.peers[0].PingHistory))
	}
	if m.peers[0].PingHistory[0] != tailscale.PingFailed {
		t.Errorf("expected PingFailed sentinel, got %v", m.peers[0].PingHistory[0])
	}
}

// ── mergePeers ────────────────────────────────────────────────────────────────

func TestMergePeers_PreservesPingHistory(t *testing.T) {
	hist := []time.Duration{1 * time.Millisecond, 2 * time.Millisecond}
	m := minModel([]tailscale.Peer{
		{TailscaleIP: "100.1.1.1", PingHistory: hist},
	})

	fresh := []tailscale.Peer{
		{TailscaleIP: "100.1.1.1", Hostname: "node-a", Online: true},
	}
	m, _ = m.mergePeers(fresh, tailscale.NetworkInfo{})

	if len(m.peers[0].PingHistory) != len(hist) {
		t.Errorf("ping history not preserved: got %d entries, want %d",
			len(m.peers[0].PingHistory), len(hist))
	}
}

func TestMergePeers_NotifiesOnConnect(t *testing.T) {
	m := minModel([]tailscale.Peer{
		{TailscaleIP: "100.1.1.1", Hostname: "node-a", Online: false},
	})

	fresh := []tailscale.Peer{
		{TailscaleIP: "100.1.1.1", Hostname: "node-a", Online: true},
	}
	m, cmds := m.mergePeers(fresh, tailscale.NetworkInfo{})

	if m.errMsg != "node-a connected" {
		t.Errorf("expected connected notification, got %q", m.errMsg)
	}
	if len(cmds) == 0 {
		t.Error("expected a clearStatusCmd to be returned")
	}
}

func TestMergePeers_NotifiesOnDisconnect(t *testing.T) {
	m := minModel([]tailscale.Peer{
		{TailscaleIP: "100.1.1.1", Hostname: "node-a", Online: true},
	})

	fresh := []tailscale.Peer{
		{TailscaleIP: "100.1.1.1", Hostname: "node-a", Online: false},
	}
	m, cmds := m.mergePeers(fresh, tailscale.NetworkInfo{})

	if m.errMsg != "node-a disconnected" {
		t.Errorf("expected disconnected notification, got %q", m.errMsg)
	}
	if len(cmds) == 0 {
		t.Error("expected a clearStatusCmd to be returned")
	}
}

func TestMergePeers_NoNotifyForSelf(t *testing.T) {
	m := minModel([]tailscale.Peer{
		{TailscaleIP: "100.1.1.1", Hostname: "me", Online: true, IsSelf: true},
	})

	fresh := []tailscale.Peer{
		{TailscaleIP: "100.1.1.1", Hostname: "me", Online: false, IsSelf: true},
	}
	m, cmds := m.mergePeers(fresh, tailscale.NetworkInfo{})

	if m.errMsg != "" {
		t.Errorf("self transition should not trigger notification, got %q", m.errMsg)
	}
	if len(cmds) != 0 {
		t.Error("expected no cmds for self transition")
	}
}

func TestMergePeers_NoNotifyWhenUnchanged(t *testing.T) {
	m := minModel([]tailscale.Peer{
		{TailscaleIP: "100.1.1.1", Hostname: "node-a", Online: true},
	})

	fresh := []tailscale.Peer{
		{TailscaleIP: "100.1.1.1", Hostname: "node-a", Online: true},
	}
	_, cmds := m.mergePeers(fresh, tailscale.NetworkInfo{})

	if len(cmds) != 0 {
		t.Error("no notification expected when online state is unchanged")
	}
}
