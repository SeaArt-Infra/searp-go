package rp

import (
	"context"
	"net/url"
	"strconv"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// SessionService provides the /v1/sessions and /v1/experience/sessions APIs.
type SessionService struct {
	client *transport.Client
}

// Create creates a playground, card, or version session with POST /v1/sessions.
func (s *SessionService) Create(ctx context.Context, body JSONMap, opts ...RequestOption) (*SessionSnapshot, error) {
	var out SessionSnapshot
	if err := requestJSON(ctx, s.client, "POST", "/sessions", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateExperience creates a session for the currently published version with
// POST /v1/experience/sessions.
func (s *SessionService) CreateExperience(ctx context.Context, body JSONMap, opts ...RequestOption) (*SessionSnapshot, error) {
	var out SessionSnapshot
	if err := requestJSON(ctx, s.client, "POST", "/experience/sessions", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a session snapshot with GET /v1/sessions/{id}.
func (s *SessionService) Get(ctx context.Context, id string, query SessionGetQuery, opts ...RequestOption) (*SessionSnapshot, error) {
	var out SessionSnapshot
	path := sessionPath(id) + queryString(query)
	if err := requestJSON(ctx, s.client, "GET", path, nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// HistoryMessage returns one message by absolute history index.
func (s *SessionService) HistoryMessage(ctx context.Context, id string, index uint64, opts ...RequestOption) (*HistoryMessage, error) {
	var out HistoryMessage
	if err := requestJSON(ctx, s.client, "GET", sessionPath(id)+"/history/"+uintString(index), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Turn runs a non-streaming reply/additional/regenerate turn and returns the
// updated session plus execution details.
func (s *SessionService) Turn(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (*TurnResponse, error) {
	var out TurnResponse
	if err := requestJSON(ctx, s.client, "POST", sessionPath(id)+"/turns", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TurnStream runs a streaming turn and returns parsed SSE events.
func (s *SessionService) TurnStream(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (<-chan StreamEvent, error) {
	return streamJSON(ctx, s.client, "POST", sessionPath(id)+"/turns", body, headersFromOptions(opts))
}

// Patch updates non-version session sampling or request settings.
func (s *SessionService) Patch(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (*SessionSnapshot, error) {
	var out SessionSnapshot
	if err := requestJSON(ctx, s.client, "PATCH", sessionPath(id), body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Rewind keeps history through index and drops the rest.
func (s *SessionService) Rewind(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (*SessionSnapshot, error) {
	var out SessionSnapshot
	if err := requestJSON(ctx, s.client, "POST", sessionPath(id)+"/rewind", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Fork creates a new session from history through index.
func (s *SessionService) Fork(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (*SessionSnapshot, error) {
	var out SessionSnapshot
	if err := requestJSON(ctx, s.client, "POST", sessionPath(id)+"/fork", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Edit replaces a history message and drops everything after it.
func (s *SessionService) Edit(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (*SessionSnapshot, error) {
	var out SessionSnapshot
	if err := requestJSON(ctx, s.client, "POST", sessionPath(id)+"/edit", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SessionGetQuery contains optional pagination parameters for GET /v1/sessions/{id}.
type SessionGetQuery struct {
	Limit  *uint64
	Offset *uint64
}

func sessionPath(id string) string {
	return "/sessions/" + id
}

func uintString(value uint64) string {
	return strconv.FormatUint(value, 10)
}

func queryString(query SessionGetQuery) string {
	values := url.Values{}
	if query.Limit != nil {
		values.Set("limit", uintString(*query.Limit))
	}
	if query.Offset != nil {
		values.Set("offset", uintString(*query.Offset))
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}
