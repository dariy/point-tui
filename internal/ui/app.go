package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dariy/point-tui/internal/api"
	img "github.com/dariy/point-tui/internal/image"
)

type pane int

const (
	tagsPane pane = iota
	postsPane
	previewPane
	messagesPane
	timelinePane
)

const (
	tagsRatio    = 2
	postsRatio   = 3
	previewRatio = 5

	logPaneLines = 4 // visible log lines inside the border

	// narrowModeMinWidth is the terminal width below which the layout collapses
	// to a single pane (the focused one) instead of the three-column view.
	narrowModeMinWidth = 80
)

// App is the root Bubble Tea model: 3-pane miller column layout.
type App struct {
	client *api.Client
	cache  *img.Cache

	focus           pane
	prevFocus       pane
	logExpanded     bool
	fullscreen      bool
	fullscreenImgIdx int
	width           int
	height          int

	timeline TimelinePane
	tags     TagsPane
	posts    PostsPane
	preview  PreviewPane
	log      LogPane
	status   StatusBar

	spinner   spinner.Model
	loadingUI bool
	keys      KeyMap
	help      help.Model

	currentTagSlug string
	currentPostID  int64
	currentYear    int

	animFrames  []string
	animDelays  []time.Duration
	animPlaying bool
	animFrame   int
	animURL     string

	settings *api.Settings
}

func NewApp(baseURL string, login bool) *App {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	client, _ := api.NewClient(baseURL)

	cacheDir, _ := os.UserCacheDir()
	cacheDir = filepath.Join(cacheDir, "point-tui", "images")
	cache := img.NewCache(cacheDir)

	return &App{
		client:   client,
		cache:    cache,
		timeline: NewTimelinePane(),
		tags:     NewTagsPane(),
		posts:    NewPostsPane(10),
		preview:  NewPreviewPane(0, 0),
		log:      NewLogPane(0, logPaneLines+3),
		status:   NewStatusBar(0),
		spinner:  sp,
		loadingUI: true,
		keys:     DefaultKeys,
		help:     NewHelp(),
	}
}

func (a *App) SetClient(c *api.Client) { a.client = c }

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.spinner.Tick,
		a.fetchHomeCmd(1, 0, false),
		a.fetchTagsCmd(),
		a.fetchSettingsCmd(),
		a.fetchTimelineBarCmd(""),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.status.IsSearching() {
		var sbCmd tea.Cmd
		a.status, sbCmd = a.status.Update(msg)
		return a, sbCmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return a.handleResize(msg)

	case tea.KeyMsg:
		return a.handleKey(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd

	case settingsLoadedMsg:
		if msg.err != nil {
			a.log.Log("settings: "+msg.err.Error(), logError)
		} else {
			a.settings = msg.settings
		}
		return a, nil

	case timelineLoadedMsg:
		if msg.err != nil {
			a.log.Log("timeline: "+msg.err.Error(), logError)
		} else {
			a.timeline.SetPills(msg.pills)
			// Timeline bar appeared: recalculate pane heights so layout stays within terminal bounds.
			a.preview.Resize(a.previewPaneWidth(), a.previewPaneHeight())
		}
		return a, nil

	case timelineYearSelectedMsg:
		a.currentYear = msg.year
		a.currentTagSlug = ""
		a.posts = NewPostsPane(a.posts.perPage)
		a.posts.year = msg.year
		a.posts.loading = true
		a.loadingUI = true
		a.setFocus(postsPane)
		if msg.year == 0 {
			return a, tea.Batch(a.fetchHomeCmd(1, 0, false), a.fetchTimelineBarCmd(""))
		}
		return a, tea.Batch(a.fetchHomeCmd(1, msg.year, false), a.fetchTimelineBarCmd(""))

	case tagsLoadedMsg:
		if msg.err != nil {
			a.log.Log(msg.err.Error(), logError)
		} else {
			a.tags.SetTags(msg.tags)
			a.log.Log(fmt.Sprintf("loaded %d tags", len(msg.tags)), logInfo)
		}
		return a, nil

	case postsLoadedMsg:
		if msg.tagSlug != a.posts.tagSlug || msg.year != a.posts.year || msg.search != a.posts.search {
			// Stale message from an old request, ignore it.
			return a, nil
		}
		var cmd tea.Cmd
		if msg.err != nil {
			a.log.Log(msg.err.Error(), logError)
		} else {
			a.posts.SetPosts(msg.posts, msg.total, msg.pages)
			a.log.Log(fmt.Sprintf("loaded %d posts (total %d)", len(msg.posts), msg.total), logInfo)
			if len(msg.posts) > 0 {
				cmd = a.fetchPostCmd(&msg.posts[0])
			}
		}
		a.loadingUI = false
		return a, cmd


	case appendPostsMsg:
		if msg.tagSlug != a.posts.tagSlug || msg.year != a.posts.year || msg.search != a.posts.search {
			// Stale message, ignore.
			return a, nil
		}
		if msg.err != nil {
			a.log.Log(msg.err.Error(), logError)
		} else {
			a.posts.AppendPosts(msg.posts)
		}
		return a, nil

	case loadNextPageMsg:
		return a, a.fetchPostsCmd(msg.tagSlug, msg.search, msg.year, msg.page, msg.perPage, true)

	case postLoadedMsg:
		if msg.err != nil {
			a.log.Log(msg.err.Error(), logError)
		} else {
			ansis := a.cachedANSIs(msg.post)
			a.preview.SetPost(msg.post, ansis)
			if msg.post.ID != a.currentPostID {
				a.animFrames = nil
				a.animDelays = nil
				a.animPlaying = false
				a.animURL = ""
			}
			a.currentPostID = msg.post.ID

			urls := api.MediaURLs(msg.post)
			var cmds []tea.Cmd
			for _, u := range urls {
				if _, ok := ansis[u]; !ok {
					cmds = append(cmds, a.fetchImageCmd(msg.post, u))
				}
			}
			if len(cmds) > 0 {
				return a, tea.Batch(cmds...)
			}
		}
		return a, nil

	case imageRenderedMsg:
		if msg.err != nil {
			a.log.Log("image: "+msg.err.Error(), logError)
			return a, nil
		}
		if msg.postID == a.currentPostID {
			a.preview.SetANSI(msg.url, msg.frames[0])

			// If this is the first image or an animation, set up animation state
			if a.animURL == "" || len(msg.frames) > 1 {
				a.animFrames = msg.frames
				a.animDelays = msg.delays
				a.animFrame = 0
				a.animURL = msg.url
				if len(msg.frames) > 1 {
					if msg.isVideo {
						a.animPlaying = false
					} else {
						a.animPlaying = true
						return a, frameTickCmd(msg.postID, 1, msg.delays[0])
					}
				}
			}
		}
		if len(msg.delays) == 0 {
			a.cache.PutANSI(msg.url, a.previewPaneWidth(), a.previewPaneHeight(), msg.frames[0])
		}
		return a, nil

	case frameTickMsg:
		if msg.postID != a.currentPostID || len(a.animFrames) == 0 {
			return a, nil
		}
		if !a.animPlaying {
			return a, nil
		}
		a.animFrame = msg.frame % len(a.animFrames)
		a.preview.SetANSI(a.animURL, a.animFrames[a.animFrame])
		next := (a.animFrame + 1) % len(a.animFrames)
		return a, frameTickCmd(msg.postID, next, a.animDelays[a.animFrame])

	case tagHoveredMsg:
		return a.handleTagHovered(msg)

	case previewTagSelectedMsg:
		return a.handlePreviewTagSelected(msg)

	case postSelectedMsg:
		a.setFocus(previewPane)
		return a, a.fetchPostCmd(msg.post)

	case postHoveredMsg:
		return a, a.fetchPostCmd(msg.post)

	case searchSubmittedMsg:
		if a.focus == tagsPane {
			a.tags.filter = msg.query
			a.tags.cursor = 0
			// If we filtered to something, select the first match instead of "All / Feed"
			if msg.query != "" {
				if vis := a.tags.visibleEntries(); len(vis) > 1 {
					a.tags.cursor = 1
				}
			}
			if e := a.tags.SelectedEntry(); e != nil {
				if e.kind == "all" {
					return a.selectTag("")
				}
				return a.selectTag(e.slug)
			}
			return a, nil
		}
		a.posts = NewPostsPane(a.posts.perPage)
		a.posts.search = msg.query
		a.posts.loading = true
		a.currentTagSlug = ""
		a.currentYear = 0
		a.loadingUI = true
		a.setFocus(postsPane)
		return a, a.fetchPostsCmd("", msg.query, 0, 1, a.posts.perPage, false)

	case errMsg:
		a.log.Log(msg.Error(), logError)
		return a, nil
	}

	return a.updateFocusedPane(msg)
}

func (a *App) refetchImagesCmd(p *api.Post) tea.Cmd {
	urls := api.MediaURLs(p)
	var cmds []tea.Cmd
	for _, u := range urls {
		a.cache.InvalidateANSI(u)
		cmds = append(cmds, a.fetchImageCmd(p, u))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (a *App) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	a.width = msg.Width
	a.height = msg.Height
	a.help.Width = msg.Width
	a.status = NewStatusBar(msg.Width)

	a.log.Resize(msg.Width, a.logPaneTotalHeight())

	pw := a.previewPaneWidth()
	ph := a.previewPaneHeight()
	a.preview.Resize(pw, ph)

	if a.currentPostID != 0 {
		if p := a.posts.SelectedPost(); p != nil {
			return a, a.refetchImagesCmd(p)
		}
	}
	return a, nil
}

func (a *App) setFocus(p pane) {
	if a.focus == messagesPane && p != messagesPane {
		a.logExpanded = false
	}
	a.focus = p
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Quit):
		return a, tea.Quit
	case a.fullscreen && (key.Matches(msg, a.keys.Down) || key.Matches(msg, a.keys.Right)):
		return a.fullscreenStep(1)
	case a.fullscreen && (key.Matches(msg, a.keys.Up) || key.Matches(msg, a.keys.Left)):
		return a.fullscreenStep(-1)
	case msg.String() == "t" && a.focus != previewPane:
		if a.focus == timelinePane {
			a.setFocus(postsPane)
		} else {
			a.setFocus(timelinePane)
		}
		return a, nil
	case key.Matches(msg, a.keys.Left):
		if a.focus == timelinePane {
			break
		}
		if a.focus > tagsPane {
			a.setFocus(a.focus - 1)
		}
		return a, nil
	case key.Matches(msg, a.keys.Right):
		if a.focus == timelinePane {
			break
		}
		limit := previewPane
		if a.logExpanded {
			limit = messagesPane
		}
		if a.focus < limit {
			if a.focus == previewPane && a.logExpanded {
				a.prevFocus = a.focus
			}
			a.setFocus(a.focus + 1)
		}
		return a, nil
	case key.Matches(msg, a.keys.Tab), key.Matches(msg, a.keys.ShiftTab):
		panes := []pane{tagsPane, postsPane, previewPane}
		if a.timelineBarHeight() > 0 {
			panes = append([]pane{timelinePane}, panes...)
		}
		panes = append(panes, messagesPane)

		idx := -1
		for i, p := range panes {
			if p == a.focus {
				idx = i
				break
			}
		}

		forward := key.Matches(msg, a.keys.Tab)
		var next pane
		if idx == -1 {
			next = panes[0]
		} else {
			if forward {
				next = panes[(idx+1)%len(panes)]
			} else {
				next = panes[(idx-1+len(panes))%len(panes)]
			}
		}

		if next == messagesPane {
			a.prevFocus = a.focus
			if !a.logExpanded {
				a.logExpanded = true
			}
		}
		a.setFocus(next)
		a.preview.Resize(a.previewPaneWidth(), a.previewPaneHeight())
		return a, nil
	case key.Matches(msg, a.keys.Search) && (a.focus == postsPane || a.focus == tagsPane || a.focus == previewPane):
		a.status.StartSearch()
		return a, nil
	case key.Matches(msg, a.keys.Reload):
		a.loadingUI = true
		a.posts.loading = true
		var postsCmd tea.Cmd
		if a.posts.search != "" {
			postsCmd = a.fetchPostsCmd(a.posts.tagSlug, a.posts.search, a.posts.year, 1, a.posts.perPage, false)
		} else if a.posts.tagSlug != "" {
			postsCmd = a.fetchPostsCmd(a.posts.tagSlug, "", a.posts.year, 1, a.posts.perPage, false)
		} else {
			postsCmd = a.fetchHomeCmd(1, a.posts.year, false)
		}
		return a, tea.Batch(postsCmd, a.fetchTagsCmd(), a.fetchTimelineBarCmd(a.posts.tagSlug))
	case key.Matches(msg, a.keys.Pane1):
		a.setFocus(tagsPane)
		return a, nil
	case key.Matches(msg, a.keys.Pane2):
		a.setFocus(postsPane)
		return a, nil
	case key.Matches(msg, a.keys.Pane3):
		a.setFocus(previewPane)
		return a, nil
	case key.Matches(msg, a.keys.Pane0):
		if a.focus == timelinePane {
			// If already focused, let timeline.Update handle "1" to select "All"
			break
		}
		a.setFocus(timelinePane)
		return a, nil
	case key.Matches(msg, a.keys.Pane4):
		if a.logExpanded {
			if a.focus == messagesPane {
				a.setFocus(a.prevFocus)
				// setFocus(prevFocus) will collapse if prevFocus != messagesPane
			} else {
				a.prevFocus = a.focus
				a.setFocus(messagesPane)
			}
		} else {
			a.prevFocus = a.focus
			a.logExpanded = true
			a.setFocus(messagesPane)
		}
		a.preview.Resize(a.previewPaneWidth(), a.previewPaneHeight())
		return a, nil
	case key.Matches(msg, a.keys.PlayPause):
		if a.currentPostID != 0 && len(a.animFrames) > 1 {
			a.animPlaying = !a.animPlaying
			if a.animPlaying {
				next := (a.animFrame + 1) % len(a.animFrames)
				return a, frameTickCmd(a.currentPostID, next, a.animDelays[a.animFrame])
			}
		}
		return a, nil
	case key.Matches(msg, a.keys.OpenBrowser):
		p := a.posts.SelectedPost()
		if p != nil {
			u := a.client.BaseURL() + "/posts/" + p.Slug
			_ = openBrowser(u)
			a.log.Log("opening browser: "+u, logInfo)
		}
		return a, nil
	case key.Matches(msg, a.keys.Escape) && a.fullscreen:
		return a.exitFullscreen()
	case key.Matches(msg, a.keys.Fullscreen):
		if a.currentPostID != 0 {
			if a.fullscreen {
				return a.exitFullscreen()
			}
			a.fullscreen = true
			a.fullscreenImgIdx = 0
			a.setFocus(previewPane)
			a.preview.Resize(a.previewPaneWidth(), a.previewPaneHeight())
			p := a.posts.SelectedPost()
			if p != nil && len(api.MediaURLs(p)) == 0 {
				return a.fullscreenStep(1)
			}
			if p != nil {
				return a, a.refetchImagesCmd(p)
			}
		}
		return a, nil
	}

	switch {
	case a.focus == timelinePane:
		var cmd tea.Cmd
		a.timeline, cmd = a.timeline.Update(msg, true)
		return a, cmd
	}
	return a.updateFocusedPane(msg)
}

func (a *App) exitFullscreen() (tea.Model, tea.Cmd) {
	a.fullscreen = false
	a.preview.Resize(a.previewPaneWidth(), a.previewPaneHeight())
	if p := a.posts.SelectedPost(); p != nil {
		return a, a.refetchImagesCmd(p)
	}
	return a, nil
}

func (a *App) fullscreenStep(delta int) (tea.Model, tea.Cmd) {
	if delta > 0 {
		p := a.posts.SelectedPost()
		if p != nil {
			urls := api.MediaURLs(p)
			if a.fullscreenImgIdx < len(urls)-1 {
				a.fullscreenImgIdx++
				return a, nil
			}
		}

		// Try moving to next posts until we find one with images
		for {
			nextP, cmd := a.posts.StepForward()
			if nextP == nil {
				return a, cmd
			}
			nextURLs := api.MediaURLs(nextP)
			if len(nextURLs) > 0 {
				a.fullscreenImgIdx = 0
				return a, tea.Batch(cmd, a.fetchPostCmd(nextP))
			}
		}
	}

	// delta < 0
	if a.fullscreenImgIdx > 0 {
		a.fullscreenImgIdx--
		return a, nil
	}

	// Try moving to previous posts until we find one with images
	for {
		prevP, cmd := a.posts.StepCursor(-1)
		if prevP == nil {
			return a, nil
		}
		prevURLs := api.MediaURLs(prevP)
		if len(prevURLs) > 0 {
			a.fullscreenImgIdx = len(prevURLs) - 1
			return a, tea.Batch(cmd, a.fetchPostCmd(prevP))
		}
	}
}

func (a *App) selectTag(slug string) (tea.Model, tea.Cmd) {
	a.currentTagSlug = slug
	a.posts = NewPostsPane(a.posts.perPage)
	a.posts.tagSlug = slug
	a.posts.loading = true
	a.loadingUI = true
	a.currentYear = 0

	if slug == "" {
		return a, tea.Batch(a.fetchHomeCmd(1, 0, false), a.fetchTimelineBarCmd(""))
	}

	return a, tea.Batch(
		a.fetchPostsCmd(slug, "", 0, 1, a.posts.perPage, false),
		a.fetchTimelineBarCmd(slug),
	)
}

func (a *App) handlePreviewTagSelected(msg previewTagSelectedMsg) (tea.Model, tea.Cmd) {
	a.tags.SelectBySlug(msg.tag.Slug)
	a.setFocus(postsPane)
	return a.selectTag(msg.tag.Slug)
}

func (a *App) handleTagHovered(msg tagHoveredMsg) (tea.Model, tea.Cmd) {
	e := msg.entry
	if e.kind == "all" {
		return a.selectTag("")
	}
	return a.selectTag(e.slug)
}

func (a *App) updateFocusedPane(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch a.focus {
	case timelinePane:
		a.timeline, cmd = a.timeline.Update(msg, true)
	case tagsPane:
		a.tags, cmd = a.tags.Update(msg, true)
	case postsPane:
		a.posts, cmd = a.posts.Update(msg, true)
	case previewPane:
		a.preview, cmd = a.preview.Update(msg, true)
	case messagesPane:
		if k, ok := msg.(tea.KeyMsg); ok {
			visible := logPaneLines
			switch {
			case key.Matches(k, a.keys.Up):
				a.log.Scroll(1, visible)
			case key.Matches(k, a.keys.Down):
				a.log.Scroll(-1, visible)
			case key.Matches(k, a.keys.PageUp):
				a.log.Scroll(visible, visible)
			case key.Matches(k, a.keys.PageDown):
				a.log.Scroll(-visible, visible)
			case key.Matches(k, a.keys.Top):
				a.log.Scroll(1000000, visible)
			case key.Matches(k, a.keys.Bottom):
				a.log.Scroll(-1000000, visible)
			}
		}
	}
	return a, cmd
}

// ── async commands ────────────────────────────────────────────────────────────

func (a *App) fetchSettingsCmd() tea.Cmd {
	return func() tea.Msg {
		s, err := a.client.GetPublicSettings(context.Background())
		return settingsLoadedMsg{settings: s, err: err}
	}
}

func (a *App) fetchTagsCmd() tea.Cmd {
	return func() tea.Msg {
		tags, err := a.client.ListTags(context.Background(), false)
		return tagsLoadedMsg{tags: tags, err: err}
	}
}

func (a *App) fetchHomeCmd(page, year int, appendMode bool) tea.Cmd {
	return func() tea.Msg {
		home, err := a.client.GetHomePage(context.Background(), page, a.posts.perPage, year)
		if err != nil {
			return postsLoadedMsg{tagSlug: "", year: year, err: err}
		}
		if appendMode {
			return appendPostsMsg{tagSlug: "", year: year, posts: home.Posts, err: nil}
		}
		return postsLoadedMsg{
			tagSlug: "",
			year:    year,
			posts:   home.Posts,
			total:   home.Pagination.Total,
			pages:   home.Pagination.Pages,
		}
	}
}

func (a *App) fetchTimelineBarCmd(tagSlug string) tea.Cmd {
	return func() tea.Msg {
		pills, err := a.client.GetTimeline(context.Background(), tagSlug)
		return timelineLoadedMsg{pills: pills, err: err}
	}
}

func (a *App) fetchPostsCmd(tagSlug, search string, year, page, perPage int, appendMode bool) tea.Cmd {
	return func() tea.Msg {
		var (
			result *api.Paginated[api.Post]
			err    error
		)
		if tagSlug != "" {
			result, err = a.client.PostsByTag(context.Background(), tagSlug, page, perPage)
		} else {
			result, err = a.client.ListPosts(context.Background(), page, perPage, "", search, year)
		}
		if err != nil {
			return postsLoadedMsg{tagSlug: tagSlug, search: search, year: year, err: err}
		}
		if appendMode {
			return appendPostsMsg{tagSlug: tagSlug, search: search, year: year, posts: result.Posts, err: nil}
		}
		return postsLoadedMsg{
			tagSlug: tagSlug,
			search:  search,
			year:    year,
			posts:   result.Posts,
			total:   result.Total,
			pages:   result.Pages,
		}
	}
}

func (a *App) fetchPostCmd(p *api.Post) tea.Cmd {
	return func() tea.Msg {
		full, err := a.client.GetPostByID(context.Background(), p.ID)
		return postLoadedMsg{post: full, err: err}
	}
}

func (a *App) fetchImageCmd(p *api.Post, mediaURL string) tea.Cmd {
	isVideo := api.IsVideoPath(mediaURL)
	if mediaURL == "" {
		return nil
	}
	postID := p.ID
	cols := a.previewPaneWidth()
	rows := a.previewPaneHeight()
	isGIF := strings.HasSuffix(strings.ToLower(mediaURL), ".gif")
	return func() tea.Msg {
		raw, ok := a.cache.GetRaw(mediaURL)
		if !ok {
			var err error
			// Videos are served as-is; skip the thumbnail variant.
			raw, err = a.client.FetchImage(context.Background(), mediaURL, !isVideo)
			if err != nil {
				return imageRenderedMsg{postID: postID, url: mediaURL, err: err}
			}
			_ = a.cache.PutRaw(mediaURL, raw)
		}
		if isVideo {
			frames, delays, err := img.RenderVideoFrames(raw, cols, rows)
			return imageRenderedMsg{postID: postID, url: mediaURL, frames: frames, delays: delays, isVideo: true, err: err}
		}
		if isGIF {
			frames, delays, err := img.RenderGIFFrames(raw, cols, rows)
			return imageRenderedMsg{postID: postID, url: mediaURL, frames: frames, delays: delays, err: err}
		}
		ansi, err := img.Render(raw, cols, rows)
		if err != nil {
			return imageRenderedMsg{postID: postID, url: mediaURL, err: err}
		}
		return imageRenderedMsg{postID: postID, url: mediaURL, frames: []string{ansi}}
	}
}

func frameTickCmd(postID int64, frame int, delay time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(delay)
		return frameTickMsg{postID: postID, frame: frame}
	}
}

func (a *App) cachedANSIs(p *api.Post) map[string]string {
	urls := api.MediaURLs(p)
	ansis := make(map[string]string)
	for _, u := range urls {
		if ansi, ok := a.cache.GetANSI(u, a.previewPaneWidth(), a.previewPaneHeight()); ok {
			ansis[u] = ansi
		}
	}
	return ansis
}

// ── dimensions ────────────────────────────────────────────────────────────────

func (a *App) paneWidths() (tags, posts, preview int) {
	inner := a.width - 6
	if inner < 3 {
		inner = 3
	}
	total := tagsRatio + postsRatio + previewRatio
	tags = inner * tagsRatio / total
	posts = inner * postsRatio / total
	preview = inner - tags - posts
	return
}

func (a *App) narrowMode() bool {
	return a.width < narrowModeMinWidth
}

func (a *App) previewPaneWidth() int {
	if a.fullscreen {
		if a.width < 1 {
			return 1
		}
		return a.width
	}
	if a.narrowMode() {
		return max(a.width-2, 1)
	}
	_, _, w := a.paneWidths()
	return w
}

func (a *App) logPaneTotalHeight() int {
	if a.logExpanded {
		return logPaneLines + 2 // title bar (1) + inner lines + bottom border (1)
	}
	return 1
}

func (a *App) timelineBarHeight() int {
	if len(a.timeline.pills) > 0 {
		return 1
	}
	return 0
}

func (a *App) previewPaneHeight() int {
	if a.fullscreen {
		if a.height < 1 {
			return 1
		}
		return a.height
	}
	// subtract: header (1) + timeline bar + status bar (1) + log pane + top-pane border (2)
	h := a.height - 1 - a.timelineBarHeight() - 1 - a.logPaneTotalHeight() - 2
	if h < 1 {
		return 1
	}
	return h
}

// ── view ──────────────────────────────────────────────────────────────────────

func (a *App) View() string {
	if a.width == 0 {
		return "loading…\n"
	}

	if a.fullscreen {
		ansi := a.preview.ANSIIdx(a.fullscreenImgIdx)
		if ansi == "" {
			return lipgloss.NewStyle().
				Width(a.width).
				Height(a.height).
				Align(lipgloss.Center, lipgloss.Center).
				Render("Loading image…")
		}
		return lipgloss.NewStyle().
			Width(a.width).
			Height(a.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(ansi)
	}

	paneH := a.previewPaneHeight()

	var row string
	if a.narrowMode() {
		w := max(a.width-2, 1)
		inner := max(w-1, 1)
		f := a.focus
		if f == messagesPane {
			f = a.prevFocus
		}
		switch f {
		case tagsPane:
			total := len(a.tags.visibleEntries())
			row = lipgloss.JoinVertical(lipgloss.Left,
				PaneTitleBar(paneTitle(2, "Tags"), a.focus == tagsPane, w+2),
				RenderPaneBodyWithScrollbar(a.focus == tagsPane, w, paneH, a.tags.View(inner, paneH), total, paneH, a.tags.cursor-paneH+1),
			)
		case postsPane:
			row = lipgloss.JoinVertical(lipgloss.Left,
				PaneTitleBar(paneTitle(3, "Posts"), a.focus == postsPane, w+2),
				RenderPaneBodyWithScrollbar(a.focus == postsPane, w, paneH, a.posts.View(inner, paneH), a.posts.total, paneH, a.posts.cursor-paneH+1),
			)
		case previewPane:
			total := a.preview.vp.TotalLineCount()
			visible := a.preview.vp.Height
			offset := a.preview.vp.YOffset
			row = lipgloss.JoinVertical(lipgloss.Left,
				PaneTitleBar(paneTitle(4, ""), a.focus == previewPane, w+2),
				RenderPaneBodyWithScrollbar(a.focus == previewPane, w, paneH, a.preview.View(), total, visible, offset),
			)
		}
	} else {
		tagsW, postsW, previewW := a.paneWidths()

		tagsInner := max(tagsW-1, 1)
		postsInner := max(postsW-1, 1)

		tagsActiveCol := a.focus == tagsPane
		postsActiveCol := a.focus == postsPane
		previewActiveCol := a.focus == previewPane

		tagsTotal := len(a.tags.visibleEntries())
		tagsCol := lipgloss.JoinVertical(lipgloss.Left,
			PaneTitleBar(paneTitle(2, "Tags"), tagsActiveCol, tagsW+2),
			RenderPaneBodyWithScrollbar(tagsActiveCol, tagsW, paneH, a.tags.View(tagsInner, paneH), tagsTotal, paneH, a.tags.cursor-paneH+1),
		)
		postsCol := lipgloss.JoinVertical(lipgloss.Left,
			PaneTitleBar(paneTitle(3, "Posts"), postsActiveCol, postsW+2),
			RenderPaneBodyWithScrollbar(postsActiveCol, postsW, paneH, a.posts.View(postsInner, paneH), a.posts.total, paneH, a.posts.cursor-paneH+1),
		)

		previewTotal := a.preview.vp.TotalLineCount()
		previewVisible := a.preview.vp.Height
		previewOffset := a.preview.vp.YOffset
		previewCol := lipgloss.JoinVertical(lipgloss.Left,
			PaneTitleBar(paneTitle(4, ""), previewActiveCol, previewW+2),
			RenderPaneBodyWithScrollbar(previewActiveCol, previewW, paneH, a.preview.View(), previewTotal, previewVisible, previewOffset),
		)

		row = lipgloss.JoinHorizontal(lipgloss.Top, tagsCol, postsCol, previewCol)
	}

	var logRow string
	if a.logExpanded {
		a.log.Resize(a.width, a.logPaneTotalHeight())
		logRow = a.log.View(a.focus == messagesPane)
	} else {
		logRow = a.log.CollapsedView(a.width, a.focus == messagesPane)
	}

	var spinnerPart string
	if a.loadingUI {
		spinnerPart = a.spinner.View() + " "
	}
	statusLine := a.status.View(spinnerPart + a.help.View(a.keys))

	var crumbs []string
	var author string
	if a.settings != nil {
		crumbs = append(crumbs, a.settings.Title)
		if a.settings.Subtitle != "" {
			crumbs = append(crumbs, a.settings.Subtitle)
		}
		author = a.settings.Author
	}

	p := a.posts.SelectedPost()

	if a.posts.search != "" {
		crumbs = append(crumbs, "search")
		crumbs = append(crumbs, a.posts.search)
	} else {
		if a.currentTagSlug != "" {
			if path := a.tags.PathToTag(a.currentTagSlug); len(path) > 0 {
				crumbs = append(crumbs, path...)
			} else {
				crumbs = append(crumbs, a.currentTagSlug)
			}
		} else if a.currentYear != 0 {
			crumbs = append(crumbs, fmt.Sprintf("%d", a.currentYear))
		}
	}

	if p != nil {
		crumbs = append(crumbs, p.Title)
	}

	header := HeaderBarView(crumbs, author, a.width)

	parts := []string{header}
	if a.timelineBarHeight() > 0 {
		parts = append(parts, a.timeline.View(a.width, a.focus == timelinePane))
	}
	parts = append(parts, row, logRow, statusLine)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

