package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := NewClient(srv.URL)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestDoJSON_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"hello": "world"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var got map[string]string
	if err := c.doJSON(context.Background(), "/test", &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["hello"] != "world" {
		t.Errorf("got %v, want {hello:world}", got)
	}
}

func TestDoJSON_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var got any
	err := c.doJSON(context.Background(), "/test", &got)
	se, ok := err.(*StatusError)
	if !ok {
		t.Fatalf("expected *StatusError, got %T: %v", err, err)
	}
	if se.Code != 401 {
		t.Errorf("expected code 401, got %d", se.Code)
	}
}

func TestDoJSON_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var got any
	err := c.doJSON(context.Background(), "/missing", &got)
	se, ok := err.(*StatusError)
	if !ok {
		t.Fatalf("expected *StatusError, got %T: %v", err, err)
	}
	if se.Code != 404 {
		t.Errorf("expected code 404, got %d", se.Code)
	}
}

func TestDoJSON_500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var got any
	err := c.doJSON(context.Background(), "/broken", &got)
	se, ok := err.(*StatusError)
	if !ok {
		t.Fatalf("expected *StatusError, got %T: %v", err, err)
	}
	if se.Code != 500 {
		t.Errorf("expected code 500, got %d", se.Code)
	}
}

func TestDoJSON_ContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var got any
	err := c.doJSON(ctx, "/slow", &got)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
