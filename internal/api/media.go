package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// FetchImage downloads the image at the given path (e.g. "/2024/05/photo.jpg").
// If thumb is true, the thumbnail variant (?thumb) is requested instead.
func (c *Client) FetchImage(ctx context.Context, path string, thumb bool) ([]byte, error) {
	u, err := c.joinURL(path)
	if err != nil {
		return nil, err
	}
	if thumb {
		u += "?thumb"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating image request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{Code: resp.StatusCode}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading image body: %w", err)
	}
	return data, nil
}
