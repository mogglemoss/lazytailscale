package rdp

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

// Launch opens an RDP connection to the given IP address.
// The client is launched as a background GUI process — it does not take over
// the terminal, so the TUI remains visible.
//
// Platform behaviour:
//
//	macOS   — open rdp://<ip>  (requires Microsoft Remote Desktop from App Store)
//	Linux   — xfreerdp /v:<ip>  falling back to remmina -c rdp://<ip>
//	Windows — mstsc /v:<ip>
func Launch(ip string) tea.Cmd {
	return func() tea.Msg {
		cmd, err := buildCmd(ip)
		if err != nil {
			return ErrMsg{Err: err}
		}
		if err := cmd.Start(); err != nil {
			return ErrMsg{Err: fmt.Errorf("rdp %s: %w", ip, err)}
		}
		return DoneMsg{}
	}
}

func buildCmd(ip string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", "rdp://"+ip), nil
	case "linux":
		if _, err := exec.LookPath("xfreerdp"); err == nil {
			return exec.Command("xfreerdp", "/v:"+ip), nil
		}
		if _, err := exec.LookPath("remmina"); err == nil {
			return exec.Command("remmina", "-c", "rdp://"+ip), nil
		}
		return nil, fmt.Errorf("no RDP client found; install xfreerdp or remmina")
	case "windows":
		return exec.Command("mstsc", "/v:"+ip), nil
	default:
		return nil, fmt.Errorf("RDP not supported on %s", runtime.GOOS)
	}
}

// DoneMsg is sent when the RDP client launches successfully.
type DoneMsg struct{}

// ErrMsg is sent when the RDP client fails to launch.
type ErrMsg struct{ Err error }
