package api

import "testing"

func TestPhotoURL_MediaURL(t *testing.T) {
	p := &Post{MediaURL: "/2024/05/photo.jpg", Content: "![img](/2024/05/other.jpg)"}
	got := PhotoURL(p)
	if got != "/2024/05/photo.jpg" {
		t.Errorf("expected media_url, got %q", got)
	}
}

func TestPhotoURL_InlineImage(t *testing.T) {
	p := &Post{Content: "Some text ![alt](/2024/01/inline.jpg) more text"}
	got := PhotoURL(p)
	if got != "/2024/01/inline.jpg" {
		t.Errorf("expected inline image, got %q", got)
	}
}

func TestPhotoURL_MultilineContent(t *testing.T) {
	p := &Post{Content: "/2026/03/hideaway2.jpeg\n/2026/03/hideaway.jpeg"}
	got := PhotoURL(p)
	if got != "/2026/03/hideaway2.jpeg" {
		t.Errorf("expected first image path, got %q", got)
	}
}

func TestPhotoURL_NoImage(t *testing.T) {
	p := &Post{Content: "No images here, just text."}
	got := PhotoURL(p)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestPhotoURL_ThumbnailSet(t *testing.T) {
	p := &Post{MediaURL: "/2023/08/thumb.png"}
	got := PhotoURL(p)
	if got != "/2023/08/thumb.png" {
		t.Errorf("got %q, want /2023/08/thumb.png", got)
	}
}
