package ssh

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// Launch returns a tea.ExecCommand that SSHes into the given Tailscale IP.
// Bubbletea suspends the TUI, hands off the terminal, and resumes on exit.
func Launch(ip string) tea.Cmd {
	cmd := exec.Command("ssh", ip)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return SSHErrorMsg{Err: fmt.Errorf("ssh %s: %w", ip, err)}
		}
		return SSHDoneMsg{}
	})
}

// SSHDoneMsg is sent when an SSH session exits cleanly.
type SSHDoneMsg struct{}

// SSHErrorMsg is sent when SSH fails to launch or exits with an error.
type SSHErrorMsg struct{ Err error }
