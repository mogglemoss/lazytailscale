package vnc

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

// Launch opens a VNC connection to the given IP address.
// The viewer is launched as a background GUI process — it does not take over
// the terminal, so the TUI remains visible.
//
// Platform behaviour:
//
//	macOS   — open vnc://<ip>  (uses the built-in Screen Sharing app)
//	Linux   — vncviewer <ip>  falling back to xdg-open vnc://<ip>
//	Windows — no built-in VNC client; shows an error
func Launch(ip string) tea.Cmd {
	return func() tea.Msg {
		cmd, err := buildCmd(ip)
		if err != nil {
			return ErrMsg{Err: err}
		}
		if err := cmd.Start(); err != nil {
			return ErrMsg{Err: fmt.Errorf("vnc %s: %w", ip, err)}
		}
		return DoneMsg{}
	}
}

func buildCmd(ip string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", "vnc://"+ip), nil
	case "linux":
		if _, err := exec.LookPath("vncviewer"); err == nil {
			return exec.Command("vncviewer", ip), nil
		}
		return exec.Command("xdg-open", "vnc://"+ip), nil
	default:
		return nil, fmt.Errorf("VNC not supported on %s", runtime.GOOS)
	}
}

// DoneMsg is sent when the VNC viewer launches successfully.
type DoneMsg struct{}

// ErrMsg is sent when the VNC viewer fails to launch.
type ErrMsg struct{ Err error }
