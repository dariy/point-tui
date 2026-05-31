package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dariy/point-tui/internal/api"
)

// TimelinePane is the horizontal year-navigation bar shown below the header.
type TimelinePane struct {
	pills  []api.TimelinePill // sorted chronologically (oldest first)
	cursor int                // -1 = "All" selected, 0..N-1 = specific year
	offset int                // scroll offset (first visible pill index)
}

func NewTimelinePane() TimelinePane {
	return TimelinePane{cursor: -1, offset: 0}
}

func (tp *TimelinePane) SetPills(pills []api.TimelinePill) {
	sorted := make([]api.TimelinePill, len(pills))
	copy(sorted, pills)
	// sort descending (newest first)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Year > sorted[i].Year {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	tp.pills = sorted
	tp.cursor = -1
	tp.offset = 0
}

// SelectedYear returns the selected year, or 0 if "All" is selected.
func (tp *TimelinePane) SelectedYear() int {
	if tp.cursor < 0 || tp.cursor >= len(tp.pills) {
		return 0
	}
	return tp.pills[tp.cursor].Year
}

func (tp *TimelinePane) Update(msg tea.Msg, focused bool) (TimelinePane, tea.Cmd) {
	if !focused {
		return *tp, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			if tp.cursor > -1 {
				tp.cursor--
			}
		case "l", "right":
			if tp.cursor < len(tp.pills)-1 {
				tp.cursor++
			}
		case "g":
			if len(tp.pills) > 0 {
				tp.cursor = 0
			}
		case "G":
			if len(tp.pills) > 0 {
				tp.cursor = len(tp.pills) - 1
			}
		case "0", "1":
			tp.cursor = -1
		case "enter":
			return *tp, tp.selectCmd()
		}
	}
	return *tp, nil
}

func (tp *TimelinePane) selectCmd() tea.Cmd {
	year := tp.SelectedYear()
	return func() tea.Msg {
		return timelineYearSelectedMsg{year: year}
	}
}

// View renders the timeline bar. width is the available terminal width.
// focused controls whether the selected pill shows the cursor indicator.
func (tp *TimelinePane) View(width int, focused bool) string {
	const allLabel = "All"
	const padding = 2 // spaces between pills

	colorTimeline := lipgloss.AdaptiveColor{Light: "#2255CC", Dark: "#5FD7FF"}
	colorTimelineSelected := lipgloss.AdaptiveColor{Light: "#007A3D", Dark: "#00FF87"}
	colorTimelineDim := lipgloss.AdaptiveColor{Light: "#767676", Dark: "#444444"}

	pillSty := lipgloss.NewStyle().Foreground(colorTimeline)
	selSty := lipgloss.NewStyle().Foreground(colorTimelineSelected).Bold(true)
	dimSty := lipgloss.NewStyle().Foreground(colorTimelineDim)

	renderPill := func(label string, selected, isFocused bool) string {
		if selected {
			if isFocused {
				return selSty.Render("▶ " + label + " ◀")
			}
			return selSty.Render(label)
		}
		return pillSty.Render(label)
	}

	type pill struct {
		label    string
		selected bool
	}
	pills := make([]pill, 0, 1+len(tp.pills))
	pills = append(pills, pill{allLabel, tp.cursor == -1})
	for i, p := range tp.pills {
		pills = append(pills, pill{fmt.Sprintf("%d (%d)", p.Year, p.Count), tp.cursor == i})
	}

	selIdx := tp.cursor + 1 // +1 because "All" is index 0

	if tp.offset < 0 {
		tp.offset = 0
	}

	pW := make([]int, len(pills))
	for i, p := range pills {
		pW[i] = lipgloss.Width(renderPill(p.label, p.selected, focused))
	}

	ellLeftW := lipgloss.Width("◁  ")
	ellRightW := lipgloss.Width("  ▷")
	// Overhead from PaneSeparatorBar: "── " (3) + "[1] Timeline:" (13) + " " (1) = 17
	innerW := width - 17
	if innerW < 10 {
		innerW = 10 // safety minimum
	}

	windowWidth := func(start, end int) int {
		w := 0
		for i := start; i <= end; i++ {
			if i > start {
				w += padding
			}
			w += pW[i]
		}
		if start > 0 {
			w += ellLeftW
		}
		if end < len(pills)-1 {
			w += ellRightW
		}
		return w
	}

	if selIdx < tp.offset {
		tp.offset = selIdx
	} else {
		for tp.offset < selIdx && windowWidth(tp.offset, selIdx) > innerW {
			tp.offset++
		}
	}

	end := tp.offset
	for end < len(pills)-1 && windowWidth(tp.offset, end+1) <= innerW {
		end++
	}

	var rendered []string
	for i := tp.offset; i <= end; i++ {
		rendered = append(rendered, renderPill(pills[i].label, pills[i].selected, focused))
	}

	prefix := ""
	if tp.offset > 0 {
		prefix = dimSty.Render("◁  ")
	}
	suffix := ""
	if end < len(pills)-1 {
		suffix = dimSty.Render("  ▷")
	}

	line := prefix + strings.Join(rendered, strings.Repeat(" ", padding)) + suffix

	return PaneSeparatorBar(paneTitle(1, "Timeline"), line, focused, width)
}

// timelineYearSelectedMsg is sent when the user navigates to a different year in the timeline.
type timelineYearSelectedMsg struct {
	year int // 0 = all
}
