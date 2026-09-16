package rp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// AdminService provides access to the SeaRP control-plane API served at
// /admin/v1 by the project gateway.
type AdminService struct {
	client *transport.Client
}

// Request performs a JSON request against the /admin/v1 base and returns the
// decoded JSON object. It is useful for endpoints not yet exposed by a typed
// Admin method.
func (a *AdminService) Request(ctx context.Context, method, path string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, method, path, body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Health returns the gateway health object from GET /admin/v1/health.
func (a *AdminService) Health(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/health", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Whoami returns the current actor role and project ID, if any.
func (a *AdminService) Whoami(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/whoami", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AgentContract returns the published client/agent contract.
func (a *AdminService) AgentContract(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/agent/contract", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListProjects returns the projects visible to the current token.
func (a *AdminService) ListProjects(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/projects", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProject creates a project with an explicit id and token.
func (a *AdminService) CreateProject(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/projects", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProject returns one project by ID.
func (a *AdminService) GetProject(ctx context.Context, id string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/projects/"+projectPathSegment(id), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteProject deletes one project by ID.
func (a *AdminService) DeleteProject(ctx context.Context, id string, opts ...RequestOption) error {
	return adminRequestJSON(ctx, a.client, http.MethodDelete, "/projects/"+projectPathSegment(id), nil, headersFromOptions(opts), nil)
}

// RotateProjectToken rotates the project token and returns the updated project.
func (a *AdminService) RotateProjectToken(ctx context.Context, id string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPost, "/projects/"+projectPathSegment(id)+"/token", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProjectLive returns the currently published live sample for a project.
func (a *AdminService) GetProjectLive(ctx context.Context, id string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodGet, "/projects/"+projectPathSegment(id)+"/live", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateProjectLive updates the published live sample for a project.
func (a *AdminService) UpdateProjectLive(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := adminRequestJSON(ctx, a.client, http.MethodPut, "/projects/"+projectPathSegment(id)+"/live", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func projectPathSegment(id string) string {
	return url.PathEscape(id)
}

func adminRequestJSON(ctx context.Context, c *transport.Client, method, path string, body any, headers http.Header, out any) error {
	status, payload, err := c.Request(ctx, method, path, body, headers)
	if err != nil {
		return publicError(err)
	}
	if status >= 400 {
		return parseAdminError(status, payload)
	}
	if out == nil || len(payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return &Error{Kind: ErrGeneral, Message: "failed to decode response: " + err.Error()}
	}
	return nil
}

func parseAdminError(status int, payload []byte) error {
	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(payload, &envelope); err == nil && (envelope.Error.Code != "" || envelope.Error.Message != "") {
		return newHTTPErrorWithCode(status, envelope.Error.Code, adminErrorMessage(status, envelope.Error.Message))
	}
	return parseRawResponse(status, payload)
}

func adminErrorMessage(status int, message string) string {
	if message == "" {
		return http.StatusText(status)
	}
	return message
}
