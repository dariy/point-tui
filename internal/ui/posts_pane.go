package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"github.com/dariy/point-tui/internal/api"
)

// PostsPane is the middle column showing a paginated post list.
type PostsPane struct {
	posts    []api.Post
	cursor   int
	page     int
	perPage  int
	total    int
	pages    int
	loading  bool
	tagSlug  string // current filter ("" = all/feed)
	search   string
	year     int    // current year filter (0 = none)
}

func NewPostsPane(perPage int) PostsPane {
	if perPage <= 0 {
		perPage = 10
	}
	return PostsPane{perPage: perPage, page: 1}
}

// SetPosts replaces the current post list (first page load or tag switch).
func (pp *PostsPane) SetPosts(posts []api.Post, total, pages int) {
	pp.posts = posts
	pp.total = total
	pp.pages = pages
	pp.cursor = 0
	pp.loading = false
}

// AppendPosts adds the next page of posts (lazy pagination).
func (pp *PostsPane) AppendPosts(posts []api.Post) {
	pp.posts = append(pp.posts, posts...)
	pp.loading = false
}

// SelectedPost returns the currently focused post, or nil.
func (pp *PostsPane) SelectedPost() *api.Post {
	if pp.cursor >= 0 && pp.cursor < len(pp.posts) {
		return &pp.posts[pp.cursor]
	}
	return nil
}

// StepCursor moves the cursor by delta and returns the new post, or nil if at the boundary.
// It also returns a command if it triggers a background page load.
func (pp *PostsPane) StepCursor(delta int) (*api.Post, tea.Cmd) {
	next := pp.cursor + delta
	if next < 0 || next >= len(pp.posts) {
		return nil, nil
	}
	pp.cursor = next
	var cmd tea.Cmd
	if delta > 0 {
		cmd = pp.maybeTriggerLoad()
	}
	return pp.SelectedPost(), cmd
}

// StepForward advances the cursor by one. If already at the last loaded post and more
// pages exist, it triggers a background page load instead and returns (nil, cmd).
func (pp *PostsPane) StepForward() (*api.Post, tea.Cmd) {
	if pp.cursor < len(pp.posts)-1 {
		pp.cursor++
		return pp.SelectedPost(), pp.maybeTriggerLoad()
	}
	if pp.page < pp.pages && !pp.loading {
		pp.loading = true
		pp.page++
		return nil, loadNextPageCmd(pp.tagSlug, pp.search, pp.year, pp.page, pp.perPage)
	}
	return nil, nil
}

func (pp *PostsPane) maybeTriggerLoad() tea.Cmd {
	if pp.page < pp.pages && !pp.loading && pp.cursor >= len(pp.posts)-2 {
		pp.loading = true
		pp.page++
		return loadNextPageCmd(pp.tagSlug, pp.search, pp.year, pp.page, pp.perPage)
	}
	return nil
}

func (pp *PostsPane) Update(msg tea.Msg, focused bool) (PostsPane, tea.Cmd) {
	if !focused {
		return *pp, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if pp.cursor < len(pp.posts)-1 {
				pp.cursor++
			} else if pp.page < pp.pages && !pp.loading {
				pp.loading = true
				pp.page++
				return *pp, loadNextPageCmd(pp.tagSlug, pp.search, pp.year, pp.page, pp.perPage)
			}
			if p := pp.SelectedPost(); p != nil {
				return *pp, tea.Batch(hoverPostCmd(p), pp.maybeTriggerLoad())
			}
		case "k", "up":
			if pp.cursor > 0 {
				pp.cursor--
			}
			if p := pp.SelectedPost(); p != nil {
				return *pp, hoverPostCmd(p)
			}
		case "g":
			pp.cursor = 0
			if p := pp.SelectedPost(); p != nil {
				return *pp, hoverPostCmd(p)
			}
		case "G":
			pp.cursor = len(pp.posts) - 1
			if p := pp.SelectedPost(); p != nil {
				return *pp, tea.Batch(hoverPostCmd(p), pp.maybeTriggerLoad())
			}
		case "enter":
			if p := pp.SelectedPost(); p != nil {
				return *pp, selectPostCmd(p)
			}
		}
	}
	return *pp, nil
}

// postSelectedMsg is emitted when the user opens a post (Enter — also shifts focus).
type postSelectedMsg struct {
	post *api.Post
}

func selectPostCmd(p *api.Post) tea.Cmd {
	return func() tea.Msg { return postSelectedMsg{post: p} }
}

// postHoveredMsg is emitted when the cursor moves to a post (no focus change).
type postHoveredMsg struct {
	post *api.Post
}

func hoverPostCmd(p *api.Post) tea.Cmd {
	return func() tea.Msg { return postHoveredMsg{post: p} }
}

// loadNextPageCmd triggers a paginated posts load.
type loadNextPageMsg struct {
	tagSlug string
	search  string
	year    int
	page    int
	perPage int
}

func loadNextPageCmd(tagSlug, search string, year, page, perPage int) tea.Cmd {
	return func() tea.Msg {
		return loadNextPageMsg{tagSlug: tagSlug, search: search, year: year, page: page, perPage: perPage}
	}
}

func (pp *PostsPane) View(width, height int) string {
	if pp.loading && len(pp.posts) == 0 {
		return NormalItemStyle.Render("Loading…") + "\n"
	}
	if len(pp.posts) == 0 {
		return NormalItemStyle.Render("(no posts)") + "\n"
	}

	var sb strings.Builder

	start := 0
	if pp.cursor >= height {
		start = pp.cursor - height + 1
	}

	linesUsed := 0
	for i := start; i < len(pp.posts) && i < start+height; i++ {
		p := pp.posts[i]
		date := p.PublishedAt
		if len(date) >= 10 {
			date = date[:10]
		}
		line := fmt.Sprintf("%s  %s", date, p.Title)
		line = runewidth.Truncate(line, width, "")
		if i == pp.cursor {
			sb.WriteString(SelectedItemStyle.Render(line))
		} else {
			sb.WriteString(NormalItemStyle.Render(line))
		}
		sb.WriteByte('\n')
		linesUsed++
	}

	if pp.loading && linesUsed < height {
		sb.WriteString(NormalItemStyle.Render("Loading more…"))
		sb.WriteByte('\n')
	}

	return strings.TrimRight(sb.String(), "\n")
}
