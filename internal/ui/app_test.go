package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dariy/point-tui/internal/api"
)

func newTestApp() *App {
	return NewApp("http://localhost:9999", false)
}

func TestApp_WindowSizeMsg(t *testing.T) {
	a := newTestApp()
	model, _ := a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app := model.(*App)
	if app.width != 120 || app.height != 40 {
		t.Errorf("dimensions not updated: got %dx%d", app.width, app.height)
	}
}

func TestApp_FocusNavigation(t *testing.T) {
	a := newTestApp()
	if a.focus != tagsPane {
		t.Fatalf("initial focus should be tagsPane, got %d", a.focus)
	}

	// l → postsPane
	model, _ := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	app := model.(*App)
	if app.focus != postsPane {
		t.Errorf("after l, focus = %d, want postsPane", app.focus)
	}

	// l → previewPane
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	app = model.(*App)
	if app.focus != previewPane {
		t.Errorf("after second l, focus = %d, want previewPane", app.focus)
	}

	// l should not go past previewPane
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	app = model.(*App)
	if app.focus != previewPane {
		t.Errorf("focus should stay at previewPane, got %d", app.focus)
	}

	// h → postsPane
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	app = model.(*App)
	if app.focus != postsPane {
		t.Errorf("after h, focus = %d, want postsPane", app.focus)
	}
}

func TestApp_TabCyclesFocus(t *testing.T) {
	a := newTestApp()
	// Cycle through Tags -> Posts -> Preview -> Messages -> Tags (since timeline is empty)
	for _, want := range []pane{postsPane, previewPane, messagesPane, tagsPane} {
		model, _ := a.Update(tea.KeyMsg{Type: tea.KeyTab})
		a = model.(*App)
		if a.focus != want {
			t.Errorf("after tab, focus = %d, want %d", a.focus, want)
		}
	}
}

func TestApp_ShiftTabCyclesFocus(t *testing.T) {
	a := newTestApp()
	// Cycle through Tags -> Messages -> Preview -> Posts -> Tags (since timeline is empty)
	for _, want := range []pane{messagesPane, previewPane, postsPane, tagsPane} {
		model, _ := a.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		a = model.(*App)
		if a.focus != want {
			t.Errorf("after shift-tab, focus = %d, want %d", a.focus, want)
		}
	}
}

func TestApp_MessagePaneFocus(t *testing.T) {
	a := newTestApp()
	a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Toggle messages (Pane4 is key "5")
	model, _ := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("5")})
	a = model.(*App)
	if !a.logExpanded {
		t.Error("expected log to be expanded after pressing '5'")
	}
	if a.focus != messagesPane {
		t.Errorf("expected focus to be messagesPane, got %d", a.focus)
	}

	// Tab should cycle through 4 panes (tags, posts, preview, messages)
	for _, want := range []pane{tagsPane, postsPane, previewPane, messagesPane} {
		model, _ = a.Update(tea.KeyMsg{Type: tea.KeyTab})
		a = model.(*App)
		if a.focus != want {
			t.Errorf("after tab, focus = %d, want %d", a.focus, want)
		}
	}

	// Press '5' again to collapse and return focus
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("5")})
	a = model.(*App)
	if a.logExpanded {
		t.Error("expected log to be collapsed after pressing '5' again")
	}
	if a.focus != previewPane {
		t.Errorf("expected focus to return to previewPane, got %d", a.focus)
	}
}

func TestApp_TabCyclesTimeline(t *testing.T) {
	a := newTestApp()
	model, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = model.(*App)

	// Simulate timeline loaded with pills
	model, _ = a.Update(timelineLoadedMsg{pills: []api.TimelinePill{
		{Year: 2024, Count: 10},
		{Year: 2023, Count: 5},
	}})
	a = model.(*App)
	
	// panes should be [timeline, tags, posts, preview, messages]
	// Initial focus is tags (0)
	// Tab 1 -> posts (1)
	// Tab 2 -> preview (2)
	// Tab 3 -> messages (3)
	// Tab 4 -> timeline (4)
	// Tab 5 -> tags (0)
	
	for _, want := range []pane{postsPane, previewPane, messagesPane, timelinePane, tagsPane} {
		model, _ = a.Update(tea.KeyMsg{Type: tea.KeyTab})
		a = model.(*App)
		if a.focus != want {
			t.Errorf("after tab, focus = %d, want %d", a.focus, want)
		}
	}
}

func TestApp_MessagesAutoCollapse(t *testing.T) {
	a := newTestApp()
	a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Tab until messagesPane
	// Tags -> Posts -> Preview -> Messages
	for i := 0; i < 3; i++ {
		model, _ := a.Update(tea.KeyMsg{Type: tea.KeyTab})
		a = model.(*App)
	}
	if a.focus != messagesPane {
		t.Fatalf("expected focus to be messagesPane, got %d", a.focus)
	}
	if !a.logExpanded {
		t.Error("expected log to be expanded when focused")
	}

	// Tab away from messagesPane -> Tags
	model, _ := a.Update(tea.KeyMsg{Type: tea.KeyTab})
	a = model.(*App)
	if a.focus != tagsPane {
		t.Errorf("expected focus to be tagsPane, got %d", a.focus)
	}
	if a.logExpanded {
		t.Error("expected log to be collapsed when focus lost")
	}
}

func TestApp_ErrorInPostsLoaded(t *testing.T) {
	a := newTestApp()
	a.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	model, _ := a.Update(postsLoadedMsg{err: errors.New("network error")})
	app := model.(*App)
	if len(app.log.entries) == 0 {
		t.Error("expected error entry in log pane")
	}
	last := app.log.entries[len(app.log.entries)-1]
	if last.level != logError {
		t.Error("expected logError level for posts load error")
	}
}

func TestApp_ErrorInTagsLoaded(t *testing.T) {
	a := newTestApp()
	model, _ := a.Update(tagsLoadedMsg{err: errors.New("tags failed")})
	app := model.(*App)
	if len(app.log.entries) == 0 {
		t.Error("expected error entry in log pane from tags failure")
	}
	last := app.log.entries[len(app.log.entries)-1]
	if last.level != logError {
		t.Error("expected logError level for tags load error")
	}
}

func TestApp_OpenBrowserKey(t *testing.T) {
	a := newTestApp()
	a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Add a dummy post
	p := api.Post{ID: 1, Title: "Test Post", Slug: "test-post"}
	a.posts.SetPosts([]api.Post{p}, 1, 1)

	// Press 'o'
	model, _ := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	app := model.(*App)

	// Check log for browser opening message
	found := false
	for _, entry := range app.log.entries {
		if strings.Contains(entry.msg, "opening browser: http://localhost:9999/posts/test-post") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected log message for opening browser")
	}
}

func TestApp_TimelineNavigation(t *testing.T) {
	a := newTestApp()
	a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Simulate timeline loaded with pills
	a.Update(timelineLoadedMsg{pills: []api.TimelinePill{
		{Year: 2024, Count: 10},
		{Year: 2023, Count: 5},
	}})

	// Focus timeline (Pane0 is key "1")
	model, _ := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	a = model.(*App)
	if a.focus != timelinePane {
		t.Fatalf("expected focus to be timelinePane, got %d", a.focus)
	}

	// Initial cursor is -1 ("All")
	if a.timeline.cursor != -1 {
		t.Errorf("initial timeline cursor = %d, want -1", a.timeline.cursor)
	}

	// Right arrow -> cursor 0 (2024)
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyRight})
	a = model.(*App)
	if a.timeline.cursor != 0 {
		t.Errorf("after right, timeline cursor = %d, want 0", a.timeline.cursor)
	}
	if a.focus != timelinePane {
		t.Errorf("focus moved from timeline on Right, got %d", a.focus)
	}

	// Left arrow -> cursor -1 ("All")
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyLeft})
	a = model.(*App)
	if a.timeline.cursor != -1 {
		t.Errorf("after left, timeline cursor = %d, want -1", a.timeline.cursor)
	}
	if a.focus != timelinePane {
		t.Errorf("focus moved from timeline on Left, got %d", a.focus)
	}
}

func TestApp_SearchFromPreviewPane(t *testing.T) {
	a := newTestApp()
	a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Set focus to previewPane (usually Tags -> Posts -> Preview)
	model, _ := a.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	a = model.(*App)
	if a.focus != previewPane {
		t.Fatalf("expected focus to be previewPane, got %d", a.focus)
	}

	// Press '/'
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	a = model.(*App)

	if !a.status.IsSearching() {
		t.Error("expected search to start from preview pane")
	}
}
