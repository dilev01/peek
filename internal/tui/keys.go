package tui

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Quit           key.Binding
	VoiceToggle    key.Binding
	TextAnnotation key.Binding
	ShowAnnotation key.Binding
	ScrollDown     key.Binding
	ScrollUp       key.Binding
	HalfPageDown   key.Binding
	HalfPageUp     key.Binding
	GotoTop        key.Binding
	GotoBottom     key.Binding
	Search         key.Binding
	SearchNext     key.Binding
	SearchPrev     key.Binding
	PrevHeading    key.Binding
	NextHeading    key.Binding
	PrevAnnotation key.Binding
	NextAnnotation key.Binding
	Help           key.Binding
}

var DefaultKeyMap = KeyMap{
	Quit:           key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	VoiceToggle:    key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "voice toggle")),
	TextAnnotation: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "text annotation")),
	ShowAnnotation: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "show annotation")),
	ScrollDown:     key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "scroll down")),
	ScrollUp:       key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "scroll up")),
	HalfPageDown:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "half page down")),
	HalfPageUp:     key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "half page up")),
	GotoTop:        key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
	GotoBottom:     key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
	Search:         key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	SearchNext:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "next match")),
	SearchPrev:     key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "prev match")),
	PrevHeading:    key.NewBinding(key.WithKeys("["), key.WithHelp("[", "prev heading")),
	NextHeading:    key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "next heading")),
	PrevAnnotation: key.NewBinding(key.WithKeys("{"), key.WithHelp("{", "prev annotation")),
	NextAnnotation: key.NewBinding(key.WithKeys("}"), key.WithHelp("}", "next annotation")),
	Help:           key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}
