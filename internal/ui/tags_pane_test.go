package ui

import (
	"testing"

	"github.com/dariy/point-tui/internal/api"
)

func TestTagsPane_HierarchyDepth(t *testing.T) {
	// Simulate flat list from API where children are just stubs
	tags := []api.Tag{
		{ID: 1, Name: "Location", Slug: "location", Children: []api.Tag{{Slug: "country"}}},
		{ID: 2, Name: "Country", Slug: "country", Parents: []api.Tag{{Slug: "location"}}, Children: []api.Tag{{Slug: "portugal"}}},
		{ID: 3, Name: "Portugal", Slug: "portugal", Parents: []api.Tag{{Slug: "country"}}, Children: []api.Tag{{Slug: "lisbon"}}},
		{ID: 4, Name: "Lisbon", Slug: "lisbon", Parents: []api.Tag{{Slug: "portugal"}}},
	}

	tp := NewTagsPane()
	tp.SetTags(tags)

	// Initial state: only All/Feed and root Location are visible
	vis := tp.visibleEntries()
	if len(vis) != 2 {
		t.Fatalf("expected 2 visible entries (All/Feed + Location), got %d", len(vis))
	}
	if vis[1].label != "Location" {
		t.Errorf("expected Location, got %s", vis[1].label)
	}

	// Expand Location
	vis[1].expanded = true
	vis = tp.visibleEntries()
	if len(vis) != 3 {
		t.Fatalf("expected 3 visible entries, got %d", len(vis))
	}
	if vis[2].label != "Country" || vis[2].indent != 1 {
		t.Errorf("expected Country at indent 1, got %s at indent %d", vis[2].label, vis[2].indent)
	}

	// Expand Country
	vis[2].expanded = true
	vis = tp.visibleEntries()
	if len(vis) != 4 {
		t.Fatalf("expected 4 visible entries, got %d", len(vis))
	}
	if vis[3].label != "Portugal" || vis[3].indent != 2 {
		t.Errorf("expected Portugal at indent 2, got %s at indent %d", vis[3].label, vis[3].indent)
	}

	// Expand Portugal
	vis[3].expanded = true
	vis = tp.visibleEntries()
	if len(vis) != 5 {
		t.Fatalf("expected 5 visible entries, got %d", len(vis))
	}
	if vis[4].label != "Lisbon" || vis[4].indent != 3 {
		t.Errorf("expected Lisbon at indent 3, got %s at indent %d", vis[4].label, vis[4].indent)
	}
}

func TestTagsPane_Filtering(t *testing.T) {
	tags := []api.Tag{
		{ID: 1, Name: "Apple", Slug: "apple"},
		{ID: 2, Name: "Banana", Slug: "banana", Children: []api.Tag{{Slug: "split"}}},
		{ID: 3, Name: "Banana Split", Slug: "split", Parents: []api.Tag{{Slug: "banana"}}},
		{ID: 4, Name: "Cherry", Slug: "cherry"},
	}

	tp := NewTagsPane()
	tp.SetTags(tags)

	// Filter by "ana"
	tp.filter = "ana"
	vis := tp.visibleEntries()

	// Should show "All/Feed", "Banana", and "Banana Split" (because it's a child and it matches)
	if len(vis) != 3 {
		t.Fatalf("expected 3 visible entries for 'ana', got %d", len(vis))
	}
	if vis[1].label != "Banana" || vis[2].label != "Banana Split" {
		t.Errorf("incorrect filtered labels: %s, %s", vis[1].label, vis[2].label)
	}

	// Filter by "split"
	tp.filter = "split"
	vis = tp.visibleEntries()
	// Should show "All/Feed", "Banana" (as parent of match), and "Banana Split"
	if len(vis) != 3 {
		t.Fatalf("expected 3 visible entries for 'split', got %d", len(vis))
	}

	// Filter by "xyz"
	tp.filter = "xyz"
	vis = tp.visibleEntries()
	// Should show only "All/Feed"
	if len(vis) != 1 {
		t.Fatalf("expected only 1 visible entry for 'xyz', got %d", len(vis))
	}
}
