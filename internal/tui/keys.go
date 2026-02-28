package tui

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Quit key.Binding
}

var DefaultKeyMap = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
