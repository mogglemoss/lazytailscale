package ssh

import (
	"fmt"
	"os/exec"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"
)

// safeUsername allows the character set SSH accepts: alphanumeric, hyphen,
// underscore, dot. Rejects anything that could be interpreted as an SSH flag.
var safeUsername = regexp.MustCompile(`^[a-zA-Z0-9_.\-]+$`)

// Launch returns a tea.ExecCommand that SSHes into the given host as user.
// Bubbletea suspends the TUI, hands off the terminal, and resumes on exit.
// host should be the MagicDNS name when available, falling back to the IP.
// port is the SSH port; pass "" or "22" to use the default.
func Launch(user, host, port string) tea.Cmd {
	if !safeUsername.MatchString(user) {
		return func() tea.Msg {
			return SSHErrorMsg{Err: fmt.Errorf("invalid username: %q", user)}
		}
	}
	// Build args: -l passes the username, -p sets the port if non-default,
	// -- ends SSH's option parsing so host can't be interpreted as a flag.
	args := []string{"-l", user}
	if port != "" && port != "22" {
		args = append(args, "-p", port)
	}
	args = append(args, "--", host)

	cmd := exec.Command("ssh", args...)
	target := user + "@" + host
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
