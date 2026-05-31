package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

const defaultTimeout = 15 * time.Second

// StatusError is returned when the server responds with a non-2xx status.
type StatusError struct {
	Code int
	Body string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Body)
}

// Client is a cookie-aware HTTP client for a Point instance.
type Client struct {
	http    *http.Client
	baseURL string
}

// NewClient creates a Client for the given base URL.
func NewClient(baseURL string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}
	return &Client{
		http: &http.Client{
			Jar:     jar,
			Timeout: defaultTimeout,
		},
		baseURL: baseURL,
	}, nil
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

// joinURL joins a path (with optional query string) to the base URL.
func (c *Client) joinURL(path string) (string, error) {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("parsing base URL: %w", err)
	}
	ref, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("parsing path %q: %w", path, err)
	}
	return base.ResolveReference(ref).String(), nil
}

// doJSON performs a GET request and JSON-decodes the response body into dst.
func (c *Client) doJSON(ctx context.Context, path string, dst any) error {
	u, err := c.joinURL(path)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &StatusError{Code: resp.StatusCode, Body: string(body)}
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decoding JSON: %w", err)
	}
	return nil
}

// doPost performs a POST request with a JSON body and decodes the response.
func (c *Client) doPost(ctx context.Context, path string, payload, dst any) error {
	u, err := c.joinURL(path)
	if err != nil {
		return err
	}

	pr, pw := io.Pipe()
	go func() {
		enc := json.NewEncoder(pw)
		pw.CloseWithError(enc.Encode(payload))
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, pr)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &StatusError{Code: resp.StatusCode, Body: string(body)}
	}

	if dst != nil {
		if err := json.Unmarshal(body, dst); err != nil {
			return fmt.Errorf("decoding JSON: %w", err)
		}
	}
	return nil
}
