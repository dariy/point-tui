package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type logLevel int

const (
	logInfo logLevel = iota
	logError
)

type logEntry struct {
	t     time.Time
	level logLevel
	msg   string
}

// LogPane shows a scrollable history of messages, errors, and log entries.
type LogPane struct {
	entries      []logEntry
	width        int
	height       int
	scrollOffset int // lines from bottom (0 = show tail)
}

func NewLogPane(width, height int) LogPane {
	return LogPane{width: width, height: height}
}

func (lp *LogPane) Log(msg string, level logLevel) {
	atBottom := lp.scrollOffset == 0
	lp.entries = append(lp.entries, logEntry{t: time.Now(), level: level, msg: msg})
	if atBottom {
		lp.scrollOffset = 0
	}
}

func (lp *LogPane) Resize(width, height int) {
	lp.width = width
	lp.height = height
}

func (lp *LogPane) Scroll(delta int, visible int) {
	lp.scrollOffset += delta
	if lp.scrollOffset < 0 {
		lp.scrollOffset = 0
	}
	// We'll cap the max scroll in View once we know the total line count.
}

func (lp LogPane) View(focused bool) string {
	inner := max(lp.width-2, 1)
	visible := lp.height - 2 // subtract top border+title row and bottom border row
	if visible < 1 {
		visible = 1
	}

	allLines := lp.renderAllLines(inner)
	total := len(allLines)

	// Cap scroll offset
	if lp.scrollOffset > total-visible {
		lp.scrollOffset = total - visible
	}
	if lp.scrollOffset < 0 {
		lp.scrollOffset = 0
	}

	var lines []lineWithInfo
	start := 0
	if total <= visible {
		lines = allLines
		// pad to fill height so the border stays stable
		for len(lines) < visible {
			lines = append(lines, lineWithInfo{text: ""})
		}
	} else {
		end := total - lp.scrollOffset
		start = end - visible
		if start < 0 {
			start = 0
			end = visible
		}
		lines = allLines[start:end]
	}

	rendered := make([]string, len(lines))
	for i, l := range lines {
		if l.text == "" {
			rendered[i] = ""
			continue
		}
		if l.level == logError {
			rendered[i] = LogErrorStyle.Render(l.text)
		} else {
			rendered[i] = LogLineStyle.Render(l.text)
		}
	}

	body := strings.Join(rendered, "\n")
	titleBar := PaneTitleBar(paneTitle(5, "Messages"), focused, lp.width)
	bodyBox := RenderPaneBodyWithScrollbar(focused, lp.width-2, visible, body, total, visible, start)
	return lipgloss.JoinVertical(lipgloss.Left, titleBar, bodyBox)
}

type lineWithInfo struct {
	text  string
	level logLevel
}

func (lp LogPane) renderAllLines(inner int) []lineWithInfo {
	var lines []lineWithInfo
	for _, e := range lp.entries {
		prefix := fmt.Sprintf("[%s] ", e.t.Format("15:04:05"))
		for i, part := range strings.Split(e.msg, "\n") {
			pfx := prefix
			if i > 0 {
				pfx = strings.Repeat(" ", len(prefix))
			}
			text := pfx + part
			for len(text) > inner {
				lines = append(lines, lineWithInfo{text: text[:inner], level: e.level})
				text = strings.Repeat(" ", len(prefix)) + text[inner:]
			}
			lines = append(lines, lineWithInfo{text: text, level: e.level})
		}
	}
	return lines
}

// CollapsedView renders the log pane as a single status line:
// "─ [5] Messages: [ts] last message ─────────────────────"
func (lp LogPane) CollapsedView(width int, focused bool) string {
	title := paneTitle(5, "Messages")

	var contentStr string
	var contentSty lipgloss.Style
	if len(lp.entries) > 0 {
		last := lp.entries[len(lp.entries)-1]
		msg := last.msg
		if i := strings.IndexByte(msg, '\n'); i >= 0 {
			msg = msg[:i]
		}
		contentStr = fmt.Sprintf("[%s] %s", last.t.Format("15:04:05"), msg)
		if last.level == logError {
			contentSty = LogErrorStyle
		} else {
			contentSty = LogLineStyle
		}
	}

	return PaneSeparatorBar(title, contentSty.Render(contentStr), focused, width)
}

