package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchImage_CorrectURL(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.RequestURI()
		w.Write([]byte("fakeimgdata"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	data, err := c.FetchImage(context.Background(), "/2024/05/photo.jpg", false)
	if err != nil {
		t.Fatalf("FetchImage: %v", err)
	}
	if string(data) != "fakeimgdata" {
		t.Errorf("unexpected body: %q", data)
	}
	if gotPath != "/2024/05/photo.jpg" {
		t.Errorf("path = %q, want /2024/05/photo.jpg", gotPath)
	}
}

func TestFetchImage_ThumbParam(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.RequestURI()
		w.Write([]byte("thumbdata"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.FetchImage(context.Background(), "/2024/05/photo.jpg", true)
	if err != nil {
		t.Fatalf("FetchImage(thumb): %v", err)
	}
	if gotPath != "/2024/05/photo.jpg?thumb" {
		t.Errorf("path = %q, want /2024/05/photo.jpg?thumb", gotPath)
	}
}
