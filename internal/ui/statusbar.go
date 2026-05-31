package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// StatusBar manages the bottom line: context info, errors, and search input.
type StatusBar struct {
	searching bool
	input     textinput.Model
	message   string
	isError   bool
	width     int
}

func NewStatusBar(width int) StatusBar {
	ti := textinput.New()
	ti.Placeholder = "search…"
	ti.CharLimit = 100
	return StatusBar{input: ti, width: width}
}

// SetMessage displays a status message (or error) in the bar.
func (sb *StatusBar) SetMessage(msg string, isErr bool) {
	sb.message = msg
	sb.isError = isErr
}

// StartSearch activates the search input.
func (sb *StatusBar) StartSearch() {
	sb.searching = true
	sb.input.SetValue("")
	sb.input.Focus()
}

// searching returns the current search query and clears state.
func (sb *StatusBar) SubmitSearch() string {
	q := sb.input.Value()
	sb.searching = false
	sb.input.Blur()
	return q
}

// CancelSearch aborts the search input.
func (sb *StatusBar) CancelSearch() {
	sb.searching = false
	sb.input.Blur()
	sb.input.SetValue("")
}

// searchSubmittedMsg carries a completed search query.
type searchSubmittedMsg struct{ query string }

func (sb *StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	if !sb.searching {
		return *sb, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			q := sb.SubmitSearch()
			return *sb, func() tea.Msg { return searchSubmittedMsg{query: q} }
		case "esc":
			sb.CancelSearch()
			return *sb, nil
		}
	}
	var cmd tea.Cmd
	sb.input, cmd = sb.input.Update(msg)
	return *sb, cmd
}

func (sb *StatusBar) View(helpView string) string {
	if sb.searching {
		return StatusBarStyle.Render("/ " + sb.input.View())
	}
	if sb.message != "" {
		if sb.isError {
			return ErrorStyle.Render("error: " + sb.message)
		}
		return StatusBarStyle.Render(sb.message)
	}
	return StatusBarStyle.Render(helpView)
}

// IsSearching reports whether the search input is active.
func (sb *StatusBar) IsSearching() bool { return sb.searching }
