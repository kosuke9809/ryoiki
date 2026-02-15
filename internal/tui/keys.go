package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Esc      key.Binding
	Describe key.Binding
	Forget   key.Binding
	Add      key.Binding
	Switch   key.Binding
	Refresh  key.Binding
	Search   key.Binding
	Quit     key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "detail"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Describe: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "describe"),
	),
	Forget: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "forget"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add"),
	),
	Switch: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "switch"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Describe, k.Forget, k.Add, k.Search, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Esc},
		{k.Describe, k.Forget, k.Add, k.Switch},
		{k.Search, k.Refresh, k.Quit},
	}
}
