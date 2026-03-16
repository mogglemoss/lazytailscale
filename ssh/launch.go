package ssh

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// Launch returns a tea.ExecCommand that SSHes into the given host as user.
// Bubbletea suspends the TUI, hands off the terminal, and resumes on exit.
// host should be the MagicDNS name when available, falling back to the IP.
func Launch(user, host string) tea.Cmd {
	target := user + "@" + host
	cmd := exec.Command("ssh", target)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return SSHErrorMsg{Err: fmt.Errorf("ssh %s: %w", target, err)}
		}
		return SSHDoneMsg{}
	})
}

// SSHDoneMsg is sent when an SSH session exits cleanly.
type SSHDoneMsg struct{}

// SSHErrorMsg is sent when SSH fails to launch or exits with an error.
type SSHErrorMsg struct{ Err error }
