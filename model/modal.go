package model

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/mogglemoss/lazytailscale/config"
	"github.com/mogglemoss/lazytailscale/rdp"
	"github.com/mogglemoss/lazytailscale/ssh"
	"github.com/mogglemoss/lazytailscale/tailscale"
	"github.com/mogglemoss/lazytailscale/ui"
	"github.com/mogglemoss/lazytailscale/vnc"
)

// sshErrState holds context for the SSH error panel.
type sshErrState struct {
	host string
	err  error
}

// renderSSHErrPanel renders a centered error panel after a failed SSH session.
func (m Model) renderSSHErrPanel() string {
	host := m.sshErr.host
	if host == "" {
		host = "remote host"
	}

	title := lipgloss.NewStyle().Foreground(ui.S.T.Offline).Bold(true).Render("connection failed")

	body := ui.S.DetailLabel.Render("ssh to "+host+" returned an error.") + "\n\n" +
		ui.S.DetailLabel.Render("the remote host may have declined the inquiry,") + "\n" +
		ui.S.DetailLabel.Render("or sshd may not be running on the expected port.") + "\n\n" +
		ui.S.HelpKey.Render("any key") + ui.S.HelpDesc.Render("  continue")

	inner := title + "\n\n" + body

	w := modalInnerWidth(m.width)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.S.T.Offline).
		Padding(1, 2).
		Width(w).
		Render(inner)
}

// modalStage controls which panel the connect modal shows.
type modalStage int

const (
	modalStagePick modalStage = iota // type picker: SSH / RDP / VNC
	modalStageSSH                    // SSH credentials form
)

// connectModal is the single source of state for the connection modal.
// A nil pointer means no modal is open.
type connectModal struct {
	stage  modalStage
	target tailscale.Peer

	// SSH form state — pointer so mutations survive value-receiver copies.
	sshValues *sshFormValues
	sshForm   *huh.Form
}

type sshFormValues struct {
	username string
	port     string
}

var sshUsernameRe = regexp.MustCompile(`^[a-zA-Z0-9_.\-]+$`)

// updateModal dispatches messages to the correct stage handler.
func (m Model) updateModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.modal.stage {
	case modalStagePick:
		return m.updatePickStage(msg)
	case modalStageSSH:
		return m.updateSSHStage(msg)
	}
	return m, nil
}

func (m Model) updatePickStage(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch km.String() {
	case "esc", "q":
		m.modal = nil
	case "enter":
		// Fast-path: if credentials have been used before, launch directly.
		if tm, cmd, ok := m.tryFastSSH(); ok {
			return tm, cmd
		}
		// No saved credentials — fall through to form.
		var cmd tea.Cmd
		m, cmd = m.enterSSHStage()
		return m, cmd
	case "s":
		// Always show the form so the user can review / edit credentials.
		var cmd tea.Cmd
		m, cmd = m.enterSSHStage()
		return m, cmd
	case "r":
		if strings.EqualFold(m.modal.target.OS, "windows") {
			target := m.modal.target
			m.modal = nil
			return m, rdp.Launch(target.TailscaleIP)
		}
		// RDP is dimmed for non-Windows — ignore silently.
	case "v":
		target := m.modal.target
		m.modal = nil
		return m, vnc.Launch(target.TailscaleIP)
	}
	return m, nil
}

func (m Model) updateSSHStage(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Esc goes back to the type picker, not away from the modal entirely.
	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "esc" {
		m.modal.stage = modalStagePick
		m.modal.sshForm = nil
		m.modal.sshValues = nil
		return m, nil
	}

	f, cmd := m.modal.sshForm.Update(msg)
	m.modal.sshForm = f.(*huh.Form)

	switch m.modal.sshForm.State {
	case huh.StateCompleted:
		return m.launchSSHFromModal()
	case huh.StateAborted:
		// Huh's own abort handling — go back to pick stage.
		m.modal.stage = modalStagePick
		m.modal.sshForm = nil
		m.modal.sshValues = nil
	}
	return m, cmd
}

// tryFastSSH launches SSH immediately using saved credentials without showing the form.
// Returns ok=true if a previously saved username was found.
func (m Model) tryFastSSH() (tea.Model, tea.Cmd, bool) {
	peer := m.modal.target
	username, hasSaved := m.sshUsernames[peer.Hostname]
	if !hasSaved {
		return m, nil, false
	}
	port := "22"
	if saved, ok := m.sshPorts[peer.Hostname]; ok {
		port = saved
	}
	m.modal.sshValues = &sshFormValues{username: username, port: port}
	tm, cmd := m.launchSSHFromModal()
	return tm, cmd, true
}

// enterSSHStage transitions the modal to the SSH credentials form.
func (m Model) enterSSHStage() (Model, tea.Cmd) {
	peer := m.modal.target

	username := m.defaultUser
	if last, ok := m.sshUsernames[peer.Hostname]; ok {
		username = last
	}
	port := "22"
	if last, ok := m.sshPorts[peer.Hostname]; ok {
		port = last
	}

	values := &sshFormValues{username: username, port: port}
	formW := modalInnerWidth(m.width)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Value(&values.username).
				Validate(func(s string) error {
					if !sshUsernameRe.MatchString(s) {
						return fmt.Errorf("letters, numbers, - _ . only")
					}
					return nil
				}),
			huh.NewInput().
				Title("Port").
				Value(&values.port).
				Validate(func(s string) error {
					n, err := strconv.Atoi(s)
					if err != nil || n < 1 || n > 65535 {
						return fmt.Errorf("must be 1–65535")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm()).WithWidth(formW)

	m.modal.stage = modalStageSSH
	m.modal.sshValues = values
	m.modal.sshForm = form
	return m, form.Init()
}

// launchSSHFromModal launches SSH with the values stored in the modal.
func (m Model) launchSSHFromModal() (tea.Model, tea.Cmd) {
	username := m.modal.sshValues.username
	port := m.modal.sshValues.port
	target := m.modal.target

	m.sshUsernames[target.Hostname] = username
	m.sshPorts[target.Hostname] = port
	config.SaveUsernames(m.sshUsernames)
	config.SavePorts(m.sshPorts)

	m.modal = nil
	m.lastConnectedHost = target.Hostname

	host := target.DNSName
	if host == "" {
		host = target.TailscaleIP
	}
	return m, ssh.Launch(username, host, port)
}

// modalInnerWidth returns the inner content width for modal panels.
// Gives comfortable margins without being overly wide on large terminals.
func modalInnerWidth(termWidth int) int {
	w := termWidth - 20
	if w < 36 {
		return 36
	}
	if w > 56 {
		return 56
	}
	return w
}

// renderModal renders the appropriate panel for the current modal stage.
func (m Model) renderModal() string {
	switch m.modal.stage {
	case modalStagePick:
		return m.renderPickPanel()
	case modalStageSSH:
		return m.renderSSHPanel()
	}
	return ""
}

// renderPickPanel renders the connection type picker.
func (m Model) renderPickPanel() string {
	t := m.modal.target
	sep := ui.S.HelpSep.Render("  ·  ")

	title := ui.S.DetailHeader.Render("connect to " + t.Hostname)

	sshLine := ui.S.HelpKey.Render("s") + ui.S.HelpDesc.Render("  ssh")

	var rdpLine string
	if strings.EqualFold(t.OS, "windows") {
		rdpLine = ui.S.HelpKey.Render("r") + ui.S.HelpDesc.Render("  rdp")
	} else {
		rdpLine = ui.S.PopupDim.Render("r  rdp")
	}

	vncLine := ui.S.HelpKey.Render("v") + ui.S.HelpDesc.Render("  vnc")

	options := sshLine + sep + rdpLine + sep + vncLine

	// Hint for returning users — fast-path is available.
	var hint string
	if _, hasSaved := m.sshUsernames[t.Hostname]; hasSaved {
		hint = "\n" + ui.S.DetailLabel.Render("enter connects as "+m.sshUsernames[t.Hostname])
	}

	inner := title + "\n\n" + options + hint

	w := modalInnerWidth(m.width)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.S.T.Accent).
		Padding(1, 2).
		Width(w).
		Render(inner)
}

// shimmerColors cycles through the accent palette for the SSH form border.
var shimmerColors = []lipgloss.Color{
	"#FF5F87", // hot pink
	"#D44FC8", // pink-purple
	"#9B5DE5", // purple
	"#7B61FF", // violet
	"#9B5DE5", // purple
	"#D44FC8", // pink-purple
}

// renderSSHPanel renders the SSH credentials form with an animated border shimmer.
func (m Model) renderSSHPanel() string {
	t := m.modal.target
	title := ui.S.DetailHeader.Render("ssh into " + t.Hostname)

	inner := title + "\n\n" + m.modal.sshForm.View()

	borderColor := shimmerColors[m.mascotFrame%len(shimmerColors)]
	w := modalInnerWidth(m.width)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(w).
		Render(inner)
}
