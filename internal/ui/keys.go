package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/help"
)

// KeyMap defines all keybindings for point-tui.
type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Tab        key.Binding
	ShiftTab   key.Binding
	Enter      key.Binding
	Search     key.Binding
	Top        key.Binding
	Bottom     key.Binding
	Reload     key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Quit       key.Binding
	Pane1      key.Binding
	Pane2      key.Binding
	Pane3      key.Binding
	Pane4      key.Binding
	Pane0      key.Binding
	Fullscreen key.Binding
	PlayPause  key.Binding
	OpenBrowser key.Binding
	Escape     key.Binding
}

// DefaultKeys returns the default vim-inspired keymap.
var DefaultKeys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("h/←", "prev pane"),
	),
	Right: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("l/→", "next pane"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "cycle pane"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "cycle pane back"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Top: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "bottom"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reload"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "ctrl+u"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "ctrl+d"),
		key.WithHelp("pgdn", "page down"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Pane1: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "tags pane"),
	),
	Pane2: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "posts pane"),
	),
	Pane3: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "preview pane"),
	),
	Pane4: key.NewBinding(
		key.WithKeys("5"),
		key.WithHelp("5", "messages"),
	),
	Pane0: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "timeline"),
	),
	Fullscreen: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "fullscreen"),
	),
	PlayPause: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "play/pause video"),
	),
	OpenBrowser: key.NewBinding(
		key.WithKeys("o", "u"),
		key.WithHelp("o/u", "open browser"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close fullscreen"),
	),
}

// ShortHelp returns the abbreviated keybinding list for the help bar.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Left, k.Right, k.Enter, k.Search, k.Reload, k.Quit}
}

// FullHelp returns the full keybinding list.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown, k.Top, k.Bottom},
		{k.Left, k.Right, k.Tab, k.ShiftTab, k.Enter},
		{k.Search, k.Reload, k.Fullscreen, k.PlayPause, k.Quit},
	}
}

// NewHelp returns a configured help model.
func NewHelp() help.Model {
	h := help.New()
	h.ShowAll = false
	return h
}
