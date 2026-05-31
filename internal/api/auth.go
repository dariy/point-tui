package api

import (
	"context"
	"crypto/sha256"
	"fmt"
)

type loginRequest struct {
	Username   string `json:"username"`
	Name       string `json:"name"`        // sha256 hex of password
	RememberMe bool   `json:"remember_me"`
}

// Login authenticates with the Point instance. The password is SHA-256-hashed
// client-side (as a 64-char hex string) before transmission, matching the
// server's expectation. On success, the session cookie is stored in the client's
// jar and replayed on subsequent requests.
func (c *Client) Login(ctx context.Context, username, password string, rememberMe bool) error {
	hash := sha256.Sum256([]byte(password))
	hexHash := fmt.Sprintf("%x", hash)

	payload := loginRequest{
		Username:   username,
		Name:       hexHash,
		RememberMe: rememberMe,
	}
	return c.doPost(ctx, "/api/auth/login", payload, nil)
}
