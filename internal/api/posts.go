package api

import (
	"context"
	"fmt"
	"net/url"
)

// ListPosts returns a paginated list of posts, optionally filtered by tag, search query, or year.
func (c *Client) ListPosts(ctx context.Context, page, perPage int, tag, search string, year int) (*Paginated[Post], error) {
	v := url.Values{}
	v.Set("page", fmt.Sprintf("%d", page))
	v.Set("per_page", fmt.Sprintf("%d", perPage))
	if tag != "" {
		v.Set("tag", tag)
	}
	if search != "" {
		v.Set("q", search)
	}
	if year != 0 {
		v.Set("year", fmt.Sprintf("%d", year))
	}
	path := "/api/posts?" + v.Encode()
	var result Paginated[Post]
	if err := c.doJSON(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPostBySlug returns a single post by its slug.
func (c *Client) GetPostBySlug(ctx context.Context, slug string) (*Post, error) {
	var post Post
	if err := c.doJSON(ctx, "/api/posts/slug/"+url.PathEscape(slug), &post); err != nil {
		return nil, err
	}
	return &post, nil
}

// GetPostByID returns a single post by its numeric ID.
func (c *Client) GetPostByID(ctx context.Context, id int64) (*Post, error) {
	var post Post
	if err := c.doJSON(ctx, fmt.Sprintf("/api/posts/%d", id), &post); err != nil {
		return nil, err
	}
	return &post, nil
}

// ListPostsByYear returns a paginated list of posts published in the given year.
func (c *Client) ListPostsByYear(ctx context.Context, year, page, perPage int) (*Paginated[Post], error) {
	v := url.Values{}
	v.Set("page", fmt.Sprintf("%d", page))
	v.Set("per_page", fmt.Sprintf("%d", perPage))
	v.Set("year", fmt.Sprintf("%d", year))
	path := "/api/posts?" + v.Encode()
	var result Paginated[Post]
	if err := c.doJSON(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetNavigation returns the previous/next posts relative to postID, optionally within a tag.
func (c *Client) GetNavigation(ctx context.Context, postID int64, tag string) (*Navigation, error) {
	path := fmt.Sprintf("/api/posts/%d/navigation", postID)
	if tag != "" {
		path += "?tag=" + url.QueryEscape(tag)
	}
	var nav Navigation
	if err := c.doJSON(ctx, path, &nav); err != nil {
		return nil, err
	}
	return &nav, nil
}
