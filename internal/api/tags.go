package api

import (
	"context"
	"fmt"
	"net/url"
)

type tagsListResp struct {
	Tags []Tag `json:"tags"`
}

// ListTags returns all tags with their parent/child relationships.
func (c *Client) ListTags(ctx context.Context, includeEmpty bool) ([]Tag, error) {
	path := "/api/tags"
	if includeEmpty {
		path += "?include_empty=true"
	}
	var resp tagsListResp
	if err := c.doJSON(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Tags, nil
}

// RootTags returns only the top-level tags (those with no parents).
func RootTags(tags []Tag) []Tag {
	var roots []Tag
	for _, t := range tags {
		if len(t.Parents) == 0 {
			roots = append(roots, t)
		}
	}
	return roots
}

// BuildTree reconstructs the full hierarchy from a flat list of tags.
// The API returns a flat list where each tag has its immediate parents and children.
// To support deep hierarchies, we link the objects so each tag contains full nested children.
func BuildTree(tags []Tag) []Tag {
	bySlug := make(map[string]*Tag)
	for i := range tags {
		bySlug[tags[i].Slug] = &tags[i]
	}

	memo := make(map[string]*Tag)
	var build func(slug string) *Tag
	build = func(slug string) *Tag {
		if t, ok := memo[slug]; ok {
			return t
		}
		t, ok := bySlug[slug]
		if !ok {
			return nil
		}
		// Create a copy to build the nested tree.
		newTag := *t
		newTag.Children = make([]Tag, len(t.Children))
		for i, childStub := range t.Children {
			if fullChild := build(childStub.Slug); fullChild != nil {
				newTag.Children[i] = *fullChild
			} else {
				newTag.Children[i] = childStub
			}
		}
		memo[slug] = &newTag
		return &newTag
	}

	var roots []Tag
	for _, t := range tags {
		if len(t.Parents) == 0 {
			if fullRoot := build(t.Slug); fullRoot != nil {
				roots = append(roots, *fullRoot)
			}
		}
	}
	return roots
}

// PostsByTag returns a paginated list of posts for the given tag slug.
func (c *Client) PostsByTag(ctx context.Context, slug string, page, perPage int) (*Paginated[Post], error) {
	v := url.Values{}
	v.Set("page", fmt.Sprintf("%d", page))
	v.Set("per_page", fmt.Sprintf("%d", perPage))
	path := fmt.Sprintf("/api/tags/slug/%s/posts?%s", url.PathEscape(slug), v.Encode())
	var result Paginated[Post]
	if err := c.doJSON(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
