package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette — works on truecolor; degrades to 256-color approximations on older terminals.
var (
	colorActive   = lipgloss.AdaptiveColor{Light: "#007A3D", Dark: "#00FF87"} // green
	colorInactive = lipgloss.AdaptiveColor{Light: "#767676", Dark: "#626262"} // dim grey
	colorTitle    = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FFFDF5"}
	colorStatus   = lipgloss.AdaptiveColor{Light: "#6B6870", Dark: "#585858"}
	colorError    = lipgloss.AdaptiveColor{Light: "#C0003C", Dark: "#FF5F87"}
	colorTag      = lipgloss.AdaptiveColor{Light: "#2255CC", Dark: "#87AFFF"} // blue
)

// tagDepthColors provides distinct colors for each nesting level in the tags pane.
var tagDepthColors = []lipgloss.AdaptiveColor{
	{Light: "#2255CC", Dark: "#87AFFF"}, // depth 0: blue
	{Light: "#0077AA", Dark: "#5FD7FF"}, // depth 1: cyan
	{Light: "#2E7D5C", Dark: "#87D7AF"}, // depth 2: teal-green
	{Light: "#4B6E2A", Dark: "#AFD787"}, // depth 3+: sage
}

// TagDepthStyle returns a lipgloss style for a tag item at the given nesting depth.
func TagDepthStyle(depth int) lipgloss.Style {
	if depth >= len(tagDepthColors) {
		depth = len(tagDepthColors) - 1
	}
	return lipgloss.NewStyle().Foreground(tagDepthColors[depth])
}

// PaneStyle returns a lipgloss style for a pane border.
func PaneStyle(active bool, width, height int) lipgloss.Style {
	borderColor := colorInactive
	if active {
		borderColor = colorActive
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Height(height)
}

// PaneBodyStyle returns a pane border style without the top border line,
// for use together with PaneTitleBar to render the title inside the border.
func PaneBodyStyle(active bool, width, height int) lipgloss.Style {
	borderColor := colorInactive
	if active {
		borderColor = colorActive
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderTop(false).
		BorderForeground(borderColor).
		Width(width).
		Height(height)
}

// PaneTitleBar renders the top border line of a pane with an embedded title.
// totalWidth must equal the full visible box width (content width + 2 border chars).
func PaneTitleBar(title string, active bool, totalWidth int) string {
	borderColor := colorInactive
	if active {
		borderColor = colorActive
	}
	borderSty := lipgloss.NewStyle().Foreground(borderColor)

	var labelRendered string
	if active {
		labelRendered = lipgloss.NewStyle().Bold(true).Foreground(colorActive).Render(title)
	} else {
		labelRendered = lipgloss.NewStyle().Foreground(colorTitle).Render(title)
	}

	// "╭─ {title} ─...─╮": overhead = "╭─ "(3) + " "(1) + "╮"(1) = 5 visible chars
	dashCount := totalWidth - 5 - len([]rune(title))
	if dashCount < 0 {
		dashCount = 0
	}

	return borderSty.Render("╭─ ") + labelRendered + borderSty.Render(" "+strings.Repeat("─", dashCount)+"╮")
}

// PaneSeparatorBar renders a horizontal separator line with a title and content,
// matching the style of the collapsed log pane.
func PaneSeparatorBar(title string, content string, active bool, width int) string {
	borderColor := colorInactive
	if active {
		borderColor = colorActive
	}
	borderSty := lipgloss.NewStyle().Foreground(borderColor)

	var labelRendered string
	if active {
		labelRendered = lipgloss.NewStyle().Bold(true).Foreground(colorActive).Render(title + ":")
	} else {
		labelRendered = lipgloss.NewStyle().Foreground(colorTitle).Render(title + ":")
	}

	// "── " + label + " " + content + " ─...─"
	fixed := 3 + lipgloss.Width(labelRendered) + 1 // "── " + label + " "
	contentW := lipgloss.Width(content)

	dashCount := width - fixed - contentW
	if dashCount < 0 {
		dashCount = 0
	}

	return borderSty.Render("── ") + labelRendered + " " + content + borderSty.Render(strings.Repeat("─", dashCount))
}

// paneTitle formats a numbered pane title label, e.g. "[1] Tags".
func paneTitle(num int, name string) string {
	if num == 4 && name == "" {
		return "[4]  "
	}
	return fmt.Sprintf("[%d] %s", num, name)
}

// logoChar is the logo mark shown in the top-left of the header bar.
const logoChar = "◉"

// colorHeader is the background for the top header bar.
var colorHeader = lipgloss.AdaptiveColor{Light: "#E8E8E8", Dark: "#1C1C1C"}

// HeaderBarView renders the top header line: breadcrumbs left, author right.
func HeaderBarView(crumbs []string, author string, width int) string {
	sty := lipgloss.NewStyle().
		Background(colorHeader).
		Foreground(colorTitle).
		Width(width).
		PaddingLeft(1).
		PaddingRight(1)

	titleSty := lipgloss.NewStyle().Bold(true).Foreground(colorActive).Background(colorHeader)
	crumbSty := lipgloss.NewStyle().Foreground(colorStatus).Background(colorHeader)
	authorSty := lipgloss.NewStyle().Foreground(colorTitle).Background(colorHeader)

	logo := titleSty.Render(logoChar)

	var leftParts []string
	leftParts = append(leftParts, logo)
	for i, crumb := range crumbs {
		if i == 0 {
			leftParts = append(leftParts, titleSty.Render(crumb))
		} else if i == len(crumbs)-1 {
			leftParts = append(leftParts, authorSty.Render(crumb))
		} else {
			leftParts = append(leftParts, crumbSty.Render(crumb))
		}
	}
	left := strings.Join(leftParts, crumbSty.Render(" "))

	right := authorSty.Render(author)

	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	gap := width - leftLen - rightLen - 2 // -2 for padding
	if gap < 1 {
		gap = 1
	}

	return sty.Render(left + strings.Repeat(" ", gap) + right)
}

// StatusBarStyle is the bottom status-bar style.
var StatusBarStyle = lipgloss.NewStyle().
	Foreground(colorStatus).
	PaddingLeft(1)

// ErrorStyle renders error text in the status bar.
var ErrorStyle = lipgloss.NewStyle().
	Foreground(colorError).
	Bold(true)

// LogLineStyle renders a normal log entry line.
var LogLineStyle = lipgloss.NewStyle().Foreground(colorStatus)

// LogErrorStyle renders an error log entry line.
var LogErrorStyle = lipgloss.NewStyle().Foreground(colorError)

// SelectedItemStyle highlights the focused list item.
var SelectedItemStyle = lipgloss.NewStyle().
	Foreground(colorActive).
	Bold(true)

// NormalItemStyle renders an unselected list item.
var NormalItemStyle = lipgloss.NewStyle().
	Foreground(colorTitle)

// PreviewTagStyle renders a tag reference in the preview pane.
var PreviewTagStyle = lipgloss.NewStyle().Foreground(colorTag)

// PreviewTagFocusedStyle renders the currently focused/navigable tag in the preview pane.
var PreviewTagFocusedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#000000"}).
	Background(colorTag).
	Bold(true)

// Scrollbar styles
var (
	ScrollTrackChar = "│"
	ScrollThumbChar = "┃"
)

// RenderScrollbar returns a slice of strings (of length height) representing
// a vertical scrollbar.
func RenderScrollbar(height, total, visible, offset int) []string {
	if height <= 0 {
		return nil
	}

	res := make([]string, height)
	for i := 0; i < height; i++ {
		res[i] = ScrollTrackChar
	}

	if total <= visible {
		return res
	}

	// Calculate thumb size and position.
	thumbSize := max(1, (visible*height)/total)
	var thumbStart int
	if total > visible {
		thumbStart = (offset * (height - thumbSize)) / (total - visible)
	}

	for i := 0; i < height; i++ {
		if i >= thumbStart && i < thumbStart+thumbSize {
			res[i] = ScrollThumbChar
		}
	}
	return res
}

// RenderPaneBodyWithScrollbar renders a pane body with a scrollbar integrated into the right border.
func RenderPaneBodyWithScrollbar(active bool, width, height int, content string, total, visible, offset int) string {
	borderColor := colorInactive
	if active {
		borderColor = colorActive
	}
	borderSty := lipgloss.NewStyle().Foreground(borderColor)

	// Render the body without the right border.
	bodyStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderTop(false).
		BorderRight(false).
		BorderForeground(borderColor).
		Width(width).
		Height(height)

	body := bodyStyle.Render(content)

	// Construct the right border (scrollbar + corner).
	rightChars := make([]string, height+1) // content lines + bottom border line
	sb := RenderScrollbar(height, total, visible, offset)
	for i := 0; i < height; i++ {
		rightChars[i] = sb[i]
	}
	rightChars[height] = "╯"

	// Style the right border chars.
	for i := 0; i < len(rightChars); i++ {
		rightChars[i] = borderSty.Render(rightChars[i])
	}

	rightBorder := strings.Join(rightChars, "\n")

	return lipgloss.JoinHorizontal(lipgloss.Bottom, body, rightBorder)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
