package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"github.com/dariy/point-tui/internal/api"
)

// tagEntry is one row in the tags pane.
type tagEntry struct {
	label    string
	slug     string   // empty for synthetic entries (All, Timeline)
	kind     string   // "all", "timeline", "tag"
	indent   int
	expanded bool
	children []*tagEntry
}

// TagsPane holds the state for the left (tags) column.
type TagsPane struct {
	entries []*tagEntry
	cursor  int
	loaded  bool
	filter  string
}

func NewTagsPane() TagsPane {
	return TagsPane{
		entries: []*tagEntry{
			{label: "All / Feed", kind: "all"},
		},
	}
}

// SetTags rebuilds the tag list from the API response.
func (tp *TagsPane) SetTags(tags []api.Tag) {
	roots := api.BuildTree(tags)
	sortTags(roots)
	extra := tp.entries[:1] // keep All/Feed at top
	tp.entries = append(extra, buildEntries(roots, 0)...)
	tp.loaded = true
}

func sortTags(tags []api.Tag) {
	sort.Slice(tags, func(i, j int) bool {
		iYear, iErr := strconv.Atoi(tags[i].Name)
		jYear, jErr := strconv.Atoi(tags[j].Name)
		if iErr == nil && jErr == nil {
			return iYear > jYear // Descending years
		}
		if iErr == nil {
			return true // Years first
		}
		if jErr == nil {
			return false // Years first
		}
		return strings.ToLower(tags[i].Name) < strings.ToLower(tags[j].Name)
	})
	for i := range tags {
		if len(tags[i].Children) > 0 {
			sortTags(tags[i].Children)
		}
	}
}

func buildEntries(tags []api.Tag, depth int) []*tagEntry {
	var entries []*tagEntry
	for _, t := range tags {
		e := &tagEntry{
			label:  t.Name,
			slug:   t.Slug,
			kind:   "tag",
			indent: depth,
		}
		if len(t.Children) > 0 {
			e.children = buildEntries(t.Children, depth+1)
		}
		entries = append(entries, e)
	}
	return entries
}

// visibleEntries returns the currently visible entries (respecting expanded state and filter).
func (tp *TagsPane) visibleEntries() []*tagEntry {
	var result []*tagEntry
	f := strings.ToLower(tp.filter)
	for _, e := range tp.entries {
		result = tp.appendVisible(result, e, f)
	}
	return result
}

func (tp *TagsPane) appendVisible(result []*tagEntry, e *tagEntry, filter string) []*tagEntry {
	if filter != "" {
		if e.kind == "tag" {
			matches := strings.Contains(strings.ToLower(e.label), filter)
			hasMatchingChild := false
			for _, c := range e.children {
				if tp.anyChildMatches(c, filter) {
					hasMatchingChild = true
					break
				}
			}
			if !matches && !hasMatchingChild {
				return result
			}
		}
	}

	result = append(result, e)
	if e.expanded || filter != "" {
		for _, c := range e.children {
			result = tp.appendVisible(result, c, filter)
		}
	}
	return result
}

func (tp *TagsPane) anyChildMatches(e *tagEntry, filter string) bool {
	if strings.Contains(strings.ToLower(e.label), filter) {
		return true
	}
	for _, c := range e.children {
		if tp.anyChildMatches(c, filter) {
			return true
		}
	}
	return false
}

// SelectBySlug finds the entry with the given slug, expanding parent entries as needed,
// and moves the cursor to it. Returns true if the slug was found.
func (tp *TagsPane) SelectBySlug(slug string) bool {
	if !tp.expandTo(tp.entries, slug) {
		return false
	}
	vis := tp.visibleEntries()
	for i, e := range vis {
		if e.slug == slug {
			tp.cursor = i
			return true
		}
	}
	return false
}

// expandTo recursively expands entries so the given slug becomes visible.
func (tp *TagsPane) expandTo(entries []*tagEntry, slug string) bool {
	for _, e := range entries {
		if e.slug == slug {
			return true
		}
		if len(e.children) > 0 && tp.expandTo(e.children, slug) {
			e.expanded = true
			return true
		}
	}
	return false
}

// SelectedEntry returns the currently focused tag entry, or nil.
func (tp *TagsPane) SelectedEntry() *tagEntry {
	vis := tp.visibleEntries()
	if tp.cursor >= 0 && tp.cursor < len(vis) {
		return vis[tp.cursor]
	}
	return nil
}

// PathToTag returns the labels of the tag and its parents for the given slug.
func (tp *TagsPane) PathToTag(slug string) []string {
	return findPath(tp.entries, slug)
}

func findPath(entries []*tagEntry, slug string) []string {
	for _, e := range entries {
		if e.slug == slug {
			return []string{e.label}
		}
		if len(e.children) > 0 {
			if path := findPath(e.children, slug); path != nil {
				return append([]string{e.label}, path...)
			}
		}
	}
	return nil
}

func (tp *TagsPane) Update(msg tea.Msg, focused bool) (TagsPane, tea.Cmd) {
	if !focused {
		return *tp, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		vis := tp.visibleEntries()
		switch msg.String() {
		case "j", "down":
			if tp.cursor < len(vis)-1 {
				tp.cursor++
			}
			if e := tp.SelectedEntry(); e != nil {
				return *tp, hoverTagCmd(e)
			}
		case "k", "up":
			if tp.cursor > 0 {
				tp.cursor--
			}
			if e := tp.SelectedEntry(); e != nil {
				return *tp, hoverTagCmd(e)
			}
		case "g":
			tp.cursor = 0
			if e := tp.SelectedEntry(); e != nil {
				return *tp, hoverTagCmd(e)
			}
		case "G":
			tp.cursor = len(vis) - 1
			if e := tp.SelectedEntry(); e != nil {
				return *tp, hoverTagCmd(e)
			}
		case "enter":
			if tp.cursor < len(vis) {
				e := vis[tp.cursor]
				if len(e.children) > 0 {
					e.expanded = !e.expanded
				}
				// Stay in tags pane; only right arrow moves focus to posts.
				return *tp, hoverTagCmd(e)
			}
		}
	}
	return *tp, nil
}

// tagHoveredMsg is emitted when the cursor moves to a tag (no focus change).
type tagHoveredMsg struct {
	entry *tagEntry
}

func hoverTagCmd(e *tagEntry) tea.Cmd {
	return func() tea.Msg {
		return tagHoveredMsg{entry: e}
	}
}

func (tp *TagsPane) View(width, height int) string {
	vis := tp.visibleEntries()
	var sb strings.Builder

	start := 0
	if tp.cursor >= height {
		start = tp.cursor - height + 1
	}

	for i := start; i < len(vis) && i < start+height; i++ {
		e := vis[i]
		prefix := strings.Repeat("  ", e.indent)
		indicator := " "
		if len(e.children) > 0 {
			if e.expanded {
				indicator = "▾"
			} else {
				indicator = "▸"
			}
		}
		line := fmt.Sprintf("%s%s %s", prefix, indicator, e.label)
		line = runewidth.Truncate(line, width, "")
		if i == tp.cursor {
			sb.WriteString(SelectedItemStyle.Render(line))
		} else if e.kind == "tag" {
			sb.WriteString(TagDepthStyle(e.indent).Render(line))
		} else {
			sb.WriteString(NormalItemStyle.Render(line))
		}
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n")
}
