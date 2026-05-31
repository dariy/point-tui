package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dariy/point-tui/internal/api"
)

func TestPostsPane_InfiniteScrolling(t *testing.T) {
	pp := NewPostsPane(10)
	posts := make([]api.Post, 10)
	for i := 0; i < 10; i++ {
		posts[i] = api.Post{ID: int64(i + 1), Title: "Post"}
	}
	pp.SetPosts(posts, 20, 2) // 2 pages total

	// Move cursor to 7
	for i := 0; i < 7; i++ {
		pp, _ = pp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, true)
	}
	
	if pp.loading {
		t.Errorf("should not be loading yet at cursor %d", pp.cursor)
	}

	// Move to 8 (item before last)
	var cmd tea.Cmd
	pp, cmd = pp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, true)
	
	if pp.cursor != 8 {
		t.Fatalf("expected cursor 8, got %d", pp.cursor)
	}

	if cmd == nil {
		t.Fatal("expected command at cursor 8")
	}
	if !pp.loading {
		t.Error("expected pp.loading to be true")
	}

	// In bubbletea, we can't easily inspect tea.Batch results without more machinery.
	// But we know pp.loading is true only if loadNextPageCmd was triggered.
}

func TestPostsPane_InfiniteScrolling_G(t *testing.T) {
	pp := NewPostsPane(10)
	posts := make([]api.Post, 10)
	for i := 0; i < 10; i++ {
		posts[i] = api.Post{ID: int64(i + 1), Title: "Post"}
	}
	pp.SetPosts(posts, 20, 2)

	// Press G to go to end
	pp, cmd := pp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")}, true)

	if pp.cursor != 9 {
		t.Fatalf("expected cursor 9, got %d", pp.cursor)
	}

	if cmd == nil {
		t.Fatal("expected command when jumping to end with G")
	}
	if !pp.loading {
		t.Error("expected pp.loading to be true when jumping to end with G")
	}
}
