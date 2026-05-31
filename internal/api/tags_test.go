package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListTags_Tree(t *testing.T) {
	// Build a fixture: root has one child which has one grandchild.
	grandchild := Tag{ID: 3, Name: "Grandchild", Slug: "grandchild"}
	child := Tag{ID: 2, Name: "Child", Slug: "child", Children: []Tag{grandchild}}
	root := Tag{ID: 1, Name: "Root", Slug: "root", Children: []Tag{child}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(tagsListResp{Tags: []Tag{root}})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	tags, err := c.ListTags(context.Background(), false)
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 root tag, got %d", len(tags))
	}
	if tags[0].Children[0].Children[0].Slug != "grandchild" {
		t.Errorf("grandchild not found in tree")
	}
}

func TestRootTags_FiltersByParents(t *testing.T) {
	parent := Tag{ID: 1, Name: "Parent", Slug: "parent"}
	child := Tag{ID: 2, Name: "Child", Slug: "child", Parents: []Tag{parent}}
	orphan := Tag{ID: 3, Name: "Orphan", Slug: "orphan"}

	roots := RootTags([]Tag{parent, child, orphan})
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d: %+v", len(roots), roots)
	}
	for _, r := range roots {
		if r.Slug == "child" {
			t.Errorf("child should not be a root tag")
		}
	}
}

func TestBuildTree(t *testing.T) {
	// Simulate flat list where children are stubs (only Slug populated)
	tags := []Tag{
		{ID: 1, Name: "Root", Slug: "root", Children: []Tag{{Slug: "child"}}},
		{ID: 2, Name: "Child", Slug: "child", Parents: []Tag{{Slug: "root"}}, Children: []Tag{{Slug: "grandchild"}}},
		{ID: 3, Name: "Grandchild", Slug: "grandchild", Parents: []Tag{{Slug: "child"}}, Children: []Tag{{Slug: "great-grandchild"}}},
		{ID: 4, Name: "Great Grandchild", Slug: "great-grandchild", Parents: []Tag{{Slug: "grandchild"}}},
	}

	roots := BuildTree(tags)
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}

	root := roots[0]
	if root.Slug != "root" {
		t.Errorf("expected root, got %s", root.Slug)
	}
	if len(root.Children) != 1 || root.Children[0].Slug != "child" {
		t.Fatalf("expected child, got %+v", root.Children)
	}

	child := root.Children[0]
	if len(child.Children) != 1 || child.Children[0].Slug != "grandchild" {
		t.Fatalf("expected grandchild, got %+v", child.Children)
	}

	grandchild := child.Children[0]
	if len(grandchild.Children) != 1 || grandchild.Children[0].Slug != "great-grandchild" {
		t.Fatalf("expected great-grandchild, got %+v", grandchild.Children)
	}
}

func TestRootTags_NoCycles(t *testing.T) {
	tags := []Tag{
		{ID: 1, Slug: "a", Parents: []Tag{}},
		{ID: 2, Slug: "b", Parents: []Tag{}},
		{ID: 3, Slug: "c", Parents: []Tag{{ID: 1}}},
		{ID: 4, Slug: "d", Parents: []Tag{{ID: 2}}},
	}
	roots := RootTags(tags)
	seen := make(map[string]int)
	for _, r := range roots {
		seen[r.Slug]++
		if seen[r.Slug] > 1 {
			t.Errorf("duplicate root: %s", r.Slug)
		}
	}
	if len(roots) != 2 {
		t.Errorf("expected 2 roots, got %d", len(roots))
	}
}
