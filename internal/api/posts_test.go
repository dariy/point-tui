package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPosts(t *testing.T) {
	want := Paginated[Post]{
		Posts: []Post{{ID: 1, Title: "Hello", Slug: "hello"}},
		Total: 1,
		Pages: 1,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/posts" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.ListPosts(context.Background(), 1, 10, "", "", 0)
	if err != nil {
		t.Fatalf("ListPosts: %v", err)
	}
	if got.Total != 1 || len(got.Posts) != 1 || got.Posts[0].Slug != "hello" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestGetPostBySlug(t *testing.T) {
	want := Post{ID: 42, Title: "Test Post", Slug: "test-post"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/posts/slug/test-post" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetPostBySlug(context.Background(), "test-post")
	if err != nil {
		t.Fatalf("GetPostBySlug: %v", err)
	}
	if got.ID != 42 {
		t.Errorf("got ID %d, want 42", got.ID)
	}
}

func TestGetPostByID(t *testing.T) {
	want := Post{ID: 7, Title: "Seven", Slug: "seven"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/posts/7" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetPostByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetPostByID: %v", err)
	}
	if got.Slug != "seven" {
		t.Errorf("got slug %q, want seven", got.Slug)
	}
}

func TestGetNavigation(t *testing.T) {
	prev := Post{ID: 1, Slug: "prev"}
	next := Post{ID: 3, Slug: "next"}
	want := Navigation{Previous: &prev, Next: &next}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/posts/2/navigation" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetNavigation(context.Background(), 2, "")
	if err != nil {
		t.Fatalf("GetNavigation: %v", err)
	}
	if got.Previous == nil || got.Previous.Slug != "prev" {
		t.Errorf("unexpected previous: %+v", got.Previous)
	}
	if got.Next == nil || got.Next.Slug != "next" {
		t.Errorf("unexpected next: %+v", got.Next)
	}
}

func TestListPosts_Pagination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		perPage := r.URL.Query().Get("per_page")
		result := Paginated[Post]{
			Posts: []Post{{ID: 1, Title: fmt.Sprintf("page=%s per_page=%s", page, perPage)}},
			Total: 100,
			Pages: 10,
		}
		json.NewEncoder(w).Encode(result)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.ListPosts(context.Background(), 3, 10, "", "", 0)
	if err != nil {
		t.Fatalf("ListPosts: %v", err)
	}
	if got.Pages != 10 || got.Total != 100 {
		t.Errorf("pagination fields wrong: pages=%d total=%d", got.Pages, got.Total)
	}
}

func TestListPosts_SearchEncoding(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("q")
		if search != "hello world" {
			t.Errorf("expected search 'hello world', got %q", search)
		}
		json.NewEncoder(w).Encode(Paginated[Post]{})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.ListPosts(context.Background(), 1, 10, "", "hello world", 0)
	if err != nil {
		t.Fatalf("ListPosts: %v", err)
	}
}
