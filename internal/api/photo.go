package api

import (
	"regexp"
	"strings"
)

// reMarkdownImage matches the first markdown image in a post body.
var reMarkdownImage = regexp.MustCompile(`!\[[^\]]*\]\((/[^\)]+)\)`)

// PhotoURL returns the best single image URL for a post.
func PhotoURL(p *Post) string {
	urls := MediaURLs(p)
	for _, u := range urls {
		if IsImagePath(u) {
			return u
		}
	}
	return ""
}

// MediaURLs returns all image and video URLs for a post in order of appearance.
func MediaURLs(p *Post) []string {
	var urls []string
	seen := make(map[string]bool)
	add := func(u string) {
		if u != "" && !seen[u] {
			urls = append(urls, u)
			seen[u] = true
		}
	}

	if p.MediaURL != "" && (IsImagePath(p.MediaURL) || IsVideoPath(p.MediaURL)) {
		add(p.MediaURL)
	}
	for _, line := range strings.Split(p.Content, "\n") {
		line = strings.TrimSpace(line)
		if IsImagePath(line) || IsVideoPath(line) {
			add(line)
		}
	}
	matches := reMarkdownImage.FindAllStringSubmatch(p.Content, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			add(m[1])
		}
	}
	return urls
}

// VideoURL returns the video URL for a post, or "" if none.
func VideoURL(p *Post) string {
	urls := MediaURLs(p)
	for _, u := range urls {
		if IsVideoPath(u) {
			return u
		}
	}
	return ""
}

// IsImagePath reports whether s is a bare server-relative image path.
func IsImagePath(s string) bool {
	if !strings.HasPrefix(s, "/") {
		return false
	}
	lower := strings.ToLower(s)
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// IsVideoPath reports whether s is a bare server-relative video path.
func IsVideoPath(s string) bool {
	if !strings.HasPrefix(s, "/") {
		return false
	}
	lower := strings.ToLower(s)
	for _, ext := range []string{".mp4", ".webm", ".mov", ".mkv"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}
