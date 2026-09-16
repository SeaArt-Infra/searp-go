package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// Client owns the shared HTTP transport state used by SeaRP service packages.
type Client struct {
	APIKey     string
	BaseURL    string
	Headers    http.Header
	UserAgent  string
	HTTPClient *http.Client
}

func (c *Client) buildRequest(ctx context.Context, method, path string, body any, headers http.Header) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, &Error{Kind: "general", Message: "failed to marshal request: " + err.Error()}
		}
		bodyReader = bytes.NewReader(b)
	}
	return c.newRequest(ctx, method, path, bodyReader, headers)
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader, headers http.Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return nil, &Error{Kind: "general", Message: "failed to build request: " + err.Error()}
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", c.UserAgent)
	for key, values := range c.Headers {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	for key, values := range headers {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return req, nil
}

// Request executes one JSON HTTP request and returns the raw status code and body.
func (c *Client) Request(ctx context.Context, method, path string, body any, headers http.Header) (int, []byte, error) {
	req, err := c.buildRequest(ctx, method, path, body, headers)
	if err != nil {
		return 0, nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, nil, &Error{Kind: "network", Message: "request failed: " + err.Error()}
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, &Error{Kind: "general", Message: "failed to read response: " + err.Error()}
	}
	return resp.StatusCode, payload, nil
}

// RequestStream executes one JSON HTTP request and returns the open response body.
func (c *Client) RequestStream(ctx context.Context, method, path string, body any, headers http.Header) (*http.Response, error) {
	req, err := c.buildRequest(ctx, method, path, body, headers)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &Error{Kind: "network", Message: "request failed: " + err.Error()}
	}
	return resp, nil
}

// Error mirrors the public error contract without importing the root package.
type Error struct {
	Kind    string
	Message string
	Status  int
}

func (e *Error) Error() string {
	if e.Message == "" {
		return "HTTP " + statusText(e.Status)
	}
	return e.Message
}

func statusText(status int) string {
	switch status {
	case 400:
		return "bad request"
	case 401:
		return "unauthorized"
	case 403:
		return "forbidden"
	case 404:
		return "not found"
	case 409:
		return "conflict"
	case 429:
		return "too many requests"
	case 500:
		return "internal server error"
	default:
		return strings.TrimPrefix(http.StatusText(status), "StatusCode: ")
	}
}

// NewError maps an HTTP status to a transport error.
func NewError(status int, message string) *Error {
	kind := "general"
	switch status {
	case 400:
		kind = "invalid"
	case 401, 403:
		kind = "auth"
	case 404:
		kind = "not_found"
	case 409:
		kind = "conflict"
	case 429:
		kind = "quota"
	case 408, 504:
		kind = "timeout"
	}
	return &Error{Kind: kind, Status: status, Message: message}
}
