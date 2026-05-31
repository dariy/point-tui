package api

// Post represents a Point blog post.
type Post struct {
	ID              int64    `json:"id"`
	Title           string   `json:"title"`
	Slug            string   `json:"slug"`
	Type            string   `json:"type"`
	Content         string   `json:"content"`
	Excerpt         string   `json:"excerpt"`
	Status          string   `json:"status"`
	PublishedAt     string   `json:"published_at"`
	CreatedAt       string   `json:"created_at"`
	MediaURL        string   `json:"media_url"`
	MetaDescription string   `json:"meta_description"`
	Tags            []TagRef `json:"tags"`
}

// TagRef is the minimal tag representation embedded in Post.Tags.
type TagRef struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Tag is the full tag object returned by /api/tags.
type Tag struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	Slug     string   `json:"slug"`
	Parents  []Tag    `json:"parents"`
	Children []Tag    `json:"children"`
}

// Settings holds the public blog configuration from /api/settings/public.
type Settings struct {
	Title        string `json:"blog_title"`
	Subtitle     string `json:"blog_subtitle"`
	Author       string `json:"author_name"`
	PostsPerPage int    `json:"posts_per_page,string"`
	Theme        string `json:"active_css_theme"`
}

// TimelinePill represents a year-count entry from /api/timeline.
type TimelinePill struct {
	Year  int `json:"year"`
	Count int `json:"post_count"`
}

// timelineResponse is the wrapper returned by /api/timeline.
type timelineResponse struct {
	Pills []TimelinePill `json:"pills"`
}

// Paginated wraps a page of results returned by paginated API endpoints.
type Paginated[T any] struct {
	Posts []T `json:"posts"`
	Total int `json:"total"`
	Pages int `json:"pages"`
}

// Pagination is the nested pagination object returned by /api/pages/home.
type Pagination struct {
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// Navigation holds the previous/next post links for a given post.
type Navigation struct {
	Previous *Post `json:"previous"`
	Next     *Post `json:"next"`
}

// HomePage is the compound response from /api/pages/home.
type HomePage struct {
	Posts      []Post     `json:"posts"`
	Pagination Pagination `json:"pagination"`
}
