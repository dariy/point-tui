package api

import (
	"context"
	"fmt"
	"net/url"
)

// GetPublicSettings returns the blog's public configuration.
func (c *Client) GetPublicSettings(ctx context.Context) (*Settings, error) {
	var s Settings
	if err := c.doJSON(ctx, "/api/settings/public", &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetHomePage returns the compound home page response (posts + settings + tags).
func (c *Client) GetHomePage(ctx context.Context, page, perPage int, year int) (*HomePage, error) {
	v := url.Values{}
	v.Set("page", fmt.Sprintf("%d", page))
	v.Set("per_page", fmt.Sprintf("%d", perPage))
	if year != 0 {
		v.Set("year_from", fmt.Sprintf("%d", year))
		v.Set("year_to", fmt.Sprintf("%d", year))
	}
	path := "/api/pages/home?" + v.Encode()
	var home HomePage
	if err := c.doJSON(ctx, path, &home); err != nil {
		return nil, err
	}
	return &home, nil
}

// GetTimeline returns year/count pills, optionally scoped to a tag context.
func (c *Client) GetTimeline(ctx context.Context, tagSlug string) ([]TimelinePill, error) {
	path := "/api/timeline"
	if tagSlug != "" {
		path += "?context=" + url.QueryEscape(tagSlug)
	}
	var resp timelineResponse
	if err := c.doJSON(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Pills, nil
}
