package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dariy/point-tui/internal/api"
)

func TestApp_FullscreenNavigation(t *testing.T) {
	a := newTestApp()
	a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	posts := []api.Post{
		{ID: 1, Title: "Post 1", Content: "/1.jpg"},
		{ID: 2, Title: "Post 2", Content: "/2-1.jpg\n/2-2.jpg"},
		{ID: 3, Title: "Post 3", Content: "No images here"},
		{ID: 4, Title: "Post 4", Content: "/4.jpg"},
	}

	// Load posts
	model, _ := a.Update(postsLoadedMsg{posts: posts, total: 4, pages: 1})
	a = model.(*App)

	// Initial state: Post 1 selected (cursor 0)
	if a.posts.cursor != 0 {
		t.Fatalf("expected initial cursor 0, got %d", a.posts.cursor)
	}

	// Enter fullscreen
	a.currentPostID = 1
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	a = model.(*App)

	if !a.fullscreen {
		t.Fatal("expected app to be in fullscreen mode")
	}
	if a.fullscreenImgIdx != 0 {
		t.Errorf("expected fullscreenImgIdx 0, got %d", a.fullscreenImgIdx)
	}

	// Right -> Post 2, Image 1 (Post 1 only had 1 image)
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyRight})
	a = model.(*App)
	if a.posts.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", a.posts.cursor)
	}
	if a.fullscreenImgIdx != 0 {
		t.Errorf("expected fullscreenImgIdx 0 for post 2, got %d", a.fullscreenImgIdx)
	}

	// Right -> Post 2, Image 2
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyRight})
	a = model.(*App)
	if a.posts.cursor != 1 {
		t.Errorf("expected cursor to stay 1, got %d", a.posts.cursor)
	}
	if a.fullscreenImgIdx != 1 {
		t.Errorf("expected fullscreenImgIdx 1 for post 2, got %d", a.fullscreenImgIdx)
	}

	// Right -> Post 4, Image 1 (Post 3 skipped because no images)
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyRight})
	a = model.(*App)
	if a.posts.cursor != 3 {
		t.Errorf("expected cursor 3 (skipped post 3), got %d", a.posts.cursor)
	}
	if a.fullscreenImgIdx != 0 {
		t.Errorf("expected fullscreenImgIdx 0 for post 4, got %d", a.fullscreenImgIdx)
	}

	// Left -> Post 2, Image 2
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyLeft})
	a = model.(*App)
	if a.posts.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", a.posts.cursor)
	}
	if a.fullscreenImgIdx != 1 {
		t.Errorf("expected fullscreenImgIdx 1 for post 2, got %d", a.fullscreenImgIdx)
	}
}
