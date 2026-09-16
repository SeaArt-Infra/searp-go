package rp

import (
	"context"
	"net/url"
	"strconv"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// OperationService provides the /v1/operations APIs.
type OperationService struct {
	client *transport.Client
}

// OperationView is a decoded operation returned by the SeaRP operations API.
// The public engine view excludes recovery internals but keeps execution
// evidence.
type OperationView = JSONMap

// OperationPage is the response from GET /v1/operations.
type OperationPage struct {
	Items      []OperationView `json:"items"`
	NextOffset *uint64         `json:"next_offset,omitempty"`
}

// OperationQuery filters and pages operation lists.
type OperationQuery struct {
	Offset    *uint64
	Limit     *uint64
	SessionID string
	VersionID string
	Status    string
}

// Run executes or replays an idempotent paid operation on a session.
func (o *OperationService) Run(ctx context.Context, sessionID string, body JSONMap, opts ...RequestOption) (OperationView, error) {
	var out OperationView
	if err := requestJSON(ctx, o.client, "POST", "/sessions/"+sessionID+"/operations", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one operation by its stable idempotency key.
func (o *OperationService) Get(ctx context.Context, id string, opts ...RequestOption) (OperationView, error) {
	var out OperationView
	if err := requestJSON(ctx, o.client, "GET", "/operations/"+id, nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// List returns a filtered, chronological page of operations.
func (o *OperationService) List(ctx context.Context, query OperationQuery, opts ...RequestOption) (*OperationPage, error) {
	var out OperationPage
	if err := requestJSON(ctx, o.client, "GET", "/operations"+operationQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Traces returns a local Langfuse-like projection over durable operations.
func (o *OperationService) Traces(ctx context.Context, query OperationQuery, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, o.client, "GET", "/traces"+operationQueryString(query), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Recover commits or abandons an unknown operation after the caller inspected
// its evidence.
func (o *OperationService) Recover(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (OperationView, error) {
	var out OperationView
	if err := requestJSON(ctx, o.client, "POST", "/operations/"+id+"/recover", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func operationQueryString(query OperationQuery) string {
	values := url.Values{}
	if query.Offset != nil {
		values.Set("offset", strconv.FormatUint(*query.Offset, 10))
	}
	if query.Limit != nil {
		values.Set("limit", strconv.FormatUint(*query.Limit, 10))
	}
	if query.SessionID != "" {
		values.Set("session_id", query.SessionID)
	}
	if query.VersionID != "" {
		values.Set("version_id", query.VersionID)
	}
	if query.Status != "" {
		values.Set("status", query.Status)
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}
