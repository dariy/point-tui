package ui

import (
	"time"

	"github.com/dariy/point-tui/internal/api"
)

// tagsLoadedMsg carries the tag list fetched from the API.
type tagsLoadedMsg struct {
	tags []api.Tag
	err  error
}

// postsLoadedMsg carries a paginated post list.
type postsLoadedMsg struct {
	tagSlug string
	search  string
	year    int
	posts   []api.Post
	total   int
	pages   int
	err     error
}

// appendPostsMsg carries a next page of posts for lazy pagination.
type appendPostsMsg struct {
	tagSlug string
	search  string
	year    int
	posts   []api.Post
	err     error
}

// postLoadedMsg carries a single fully-loaded post.
type postLoadedMsg struct {
	post *api.Post
	err  error
}

// imageRenderedMsg carries rendered ANSI frames for a post image.
// Static images have one frame and nil delays; animated GIFs have multiple.
type imageRenderedMsg struct {
	postID  int64
	url     string
	frames  []string
	delays  []time.Duration
	isVideo bool
	err     error
}

// frameTickMsg advances the animation to the next frame.
type frameTickMsg struct {
	postID int64
	frame  int
}

// errMsg wraps a non-fatal error to display in the status bar.
type errMsg struct {
	err error
}

func (e errMsg) Error() string { return e.err.Error() }

// previewTagSelectedMsg is emitted when the user opens a tag from the preview pane.
type previewTagSelectedMsg struct {
	tag api.TagRef
}

// settingsLoadedMsg carries the public blog settings from the API.
type settingsLoadedMsg struct {
	settings *api.Settings
	err      error
}

// timelineLoadedMsg carries the year/count pills for the timeline bar.
type timelineLoadedMsg struct {
	pills []api.TimelinePill
	err   error
}
