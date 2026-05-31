package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin_SendsSHA256Hash(t *testing.T) {
	var gotBody loginRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	const password = "s3cr3t"
	if err := c.Login(context.Background(), "", password, true); err != nil {
		t.Fatalf("Login: %v", err)
	}

	want := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	if gotBody.Name != want {
		t.Errorf("hash sent = %q, want %q", gotBody.Name, want)
	}
	if len(gotBody.Name) != 64 {
		t.Errorf("hash length = %d, want 64", len(gotBody.Name))
	}
	if gotBody.Username != "" {
		t.Errorf("username = %q, want empty", gotBody.Username)
	}
	if !gotBody.RememberMe {
		t.Error("remember_me should be true")
	}
}

func TestLogin_CookiePersisted(t *testing.T) {
	var cookieSeen bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "tok123", Path: "/"})
			w.WriteHeader(http.StatusOK)
		case "/api/posts":
			if c, err := r.Cookie("session"); err == nil && c.Value == "tok123" {
				cookieSeen = true
			}
			json.NewEncoder(w).Encode(Paginated[Post]{})
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.Login(context.Background(), "", "pass", false); err != nil {
		t.Fatalf("Login: %v", err)
	}
	if _, err := c.ListPosts(context.Background(), 1, 10, "", "", 0); err != nil {
		t.Fatalf("ListPosts after login: %v", err)
	}
	if !cookieSeen {
		t.Error("session cookie was not replayed on follow-up request")
	}
}
