package model

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for lazytailscale.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	SSH      key.Binding
	Ping     key.Binding
	Routes   key.Binding
	Copy     key.Binding
	ExitNode key.Binding
	Filter   key.Binding
	Refresh  key.Binding
	Help     key.Binding
	Quit     key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		SSH: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "ssh"),
		),
		Ping: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "ping"),
		),
		Routes: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "routes"),
		),
		Copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy address"),
		),
		ExitNode: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "toggle exit node"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}
