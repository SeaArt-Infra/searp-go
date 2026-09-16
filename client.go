package rp

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

const (
	defaultBaseURL      = "http://127.0.0.1:8788"
	defaultAdminBaseURL = "http://127.0.0.1:8790/admin/v1"
	defaultTimeout      = 5 * time.Minute
	sdkVersion          = "0.1.0"
)

// Client is the SeaRP engine API client. Create one with New and reuse it.
type Client struct {
	apiKey       string
	baseURL      string
	apiBaseURL   string
	adminBaseURL string
	headers      http.Header
	httpClient   *http.Client

	Sessions   *SessionService
	Operations *OperationService
	Engine     *EngineService
	Cards      *CardsService
	Versions   *VersionsService
	Cinema     *CinemaService
	Admin      *AdminService
}

// ClientConfig configures a Client.
//
// BaseURL defaults to the local engine bind address http://127.0.0.1:8788.
// The SDK derives APIBaseURL as <BaseURL>/v1 unless BaseURL already ends with
// /v1 or APIBaseURL is explicitly set.
type ClientConfig struct {
	APIKey       string
	BaseURL      string
	APIBaseURL   string
	AdminBaseURL string
	Headers      http.Header
	HTTPClient   *http.Client
	Timeout      time.Duration
}

// New creates a Client from an explicit ClientConfig.
func New(cfg *ClientConfig) (*Client, error) {
	if cfg == nil {
		cfg = &ClientConfig{}
	}
	root, err := resolveRootURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	apiBase, err := resolveAPIBaseURL(root, cfg.APIBaseURL)
	if err != nil {
		return nil, err
	}
	adminBase, err := resolveAdminBaseURL(cfg.BaseURL, root, cfg.AdminBaseURL)
	if err != nil {
		return nil, err
	}
	httpClient := buildHTTPClient(cfg)
	headers := cfg.Headers.Clone()
	if headers == nil {
		headers = make(http.Header)
	}

	client := &Client{
		apiKey:       cfg.APIKey,
		baseURL:      root,
		apiBaseURL:   apiBase,
		adminBaseURL: adminBase,
		headers:      headers,
		httpClient:   httpClient,
	}

	newService := func() *transport.Client {
		return &transport.Client{
			APIKey:     client.apiKey,
			BaseURL:    client.apiBaseURL,
			Headers:    client.headers.Clone(),
			UserAgent:  "searp-go/" + sdkVersion,
			HTTPClient: httpClient,
		}
	}
	client.Sessions = &SessionService{client: newService()}
	client.Operations = &OperationService{client: newService()}
	client.Engine = &EngineService{client: newService()}
	client.Cards = &CardsService{client: newService()}
	client.Versions = &VersionsService{client: newService()}
	client.Cinema = &CinemaService{client: newService()}
	client.Admin = &AdminService{client: &transport.Client{
		APIKey:     client.apiKey,
		BaseURL:    client.adminBaseURL,
		Headers:    client.headers.Clone(),
		UserAgent:  "searp-go/" + sdkVersion,
		HTTPClient: httpClient,
	}}
	return client, nil
}

// Raw performs a JSON request against a path relative to APIBaseURL and returns
// the status code and raw body. It is intended for endpoints not yet exposed by
// a typed service method.
func (c *Client) Raw(ctx context.Context, method, path string, body JSONMap, opts ...RequestOption) (int, []byte, error) {
	return c.apiClient().Request(ctx, method, path, body, buildRequestOptions(opts).headers)
}

func (c *Client) apiClient() *transport.Client {
	return &transport.Client{
		APIKey:     c.apiKey,
		BaseURL:    c.apiBaseURL,
		Headers:    c.headers.Clone(),
		UserAgent:  "searp-go/" + sdkVersion,
		HTTPClient: c.httpClient,
	}
}

func resolveRootURL(raw string) (string, error) {
	if raw == "" {
		raw = defaultBaseURL
	}
	return normalizeURL(raw)
}

func resolveAPIBaseURL(root, raw string) (string, error) {
	if raw != "" {
		return normalizeURL(raw)
	}
	if parsed, err := url.Parse(root); err == nil && strings.HasSuffix(parsed.Path, "/v1") {
		return root, nil
	}
	return joinURL(root, "v1")
}

func resolveAdminBaseURL(rawRoot, root, rawAdmin string) (string, error) {
	if rawAdmin != "" {
		return normalizeURL(rawAdmin)
	}
	if strings.TrimSpace(rawRoot) == "" {
		return normalizeURL(defaultAdminBaseURL)
	}
	return joinURL(root, "admin/v1")
}

func normalizeURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", &Error{Kind: ErrGeneral, Message: "invalid URL: " + err.Error()}
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", &Error{Kind: ErrGeneral, Message: "invalid URL: missing scheme or host"}
	}
	parsed.Path = path.Clean("/" + parsed.Path)
	if parsed.Path == "/" {
		parsed.Path = ""
	}
	return parsed.String(), nil
}

func joinURL(baseURL, suffix string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", &Error{Kind: ErrGeneral, Message: "invalid URL: " + err.Error()}
	}
	joined := *parsed
	joined.Path = path.Join(parsed.Path, suffix)
	return normalizeURL(joined.String())
}

func buildHTTPClient(cfg *ClientConfig) *http.Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if cfg.HTTPClient == nil {
		return &http.Client{Timeout: timeout}
	}
	cloned := *cfg.HTTPClient
	cloned.Timeout = timeout
	return &cloned
}

func parseRawResponse(status int, payload []byte) error {
	if status < 400 {
		return nil
	}
	message := strings.TrimSpace(string(payload))
	if message == "" {
		message = http.StatusText(status)
	}
	return newHTTPError(status, message)
}
