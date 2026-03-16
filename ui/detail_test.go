package ui

import (
	"testing"
	"time"

	"github.com/mogglemoss/lazytailscale/tailscale"
)

func TestFmtDur(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Microsecond, "500µs"},
		{999 * time.Microsecond, "999µs"},
		{time.Millisecond, "1ms"},
		{10 * time.Millisecond, "10ms"},
		{250 * time.Millisecond, "250ms"},
		{0, "0µs"},
	}

	for _, tc := range tests {
		got := fmtDur(tc.d)
		if got != tc.want {
			t.Errorf("fmtDur(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestLastSeenStr_Online(t *testing.T) {
	peer := tailscale.Peer{Online: true}
	got := lastSeenStr(peer)
	if got != "now" {
		t.Errorf("online peer: want %q, got %q", "now", got)
	}
}

func TestLastSeenStr_UnknownTime(t *testing.T) {
	peer := tailscale.Peer{Online: false} // zero LastSeen
	got := lastSeenStr(peer)
	if got != "unknown" {
		t.Errorf("zero LastSeen: want %q, got %q", "unknown", got)
	}
}

func TestLastSeenStr_Offline(t *testing.T) {
	tests := []struct {
		ago  time.Duration
		want string
	}{
		{30 * time.Second, "30s ago"},
		{2 * time.Minute, "2m ago"},
		{3 * time.Hour, "3h ago"},
		{48 * time.Hour, "2d ago"},
	}

	for _, tc := range tests {
		peer := tailscale.Peer{
			Online:   false,
			LastSeen: time.Now().Add(-tc.ago).Round(time.Second),
		}
		got := lastSeenStr(peer)
		if got != tc.want {
			t.Errorf("lastSeenStr offset %v: want %q, got %q", tc.ago, tc.want, got)
		}
	}
}
