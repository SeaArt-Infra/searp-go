package rp

import (
	"context"
	"net/url"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// GroupSessionsService creates group sessions and generates or inspects turns.
// Read history, edit, rewind and fork using Sessions and its session revision.
type GroupSessionsService struct {
	client *transport.Client
}

// Create opens a session pinned to a group card and its member versions.
func (g *GroupSessionsService) Create(ctx context.Context, body JSONMap, opts ...RequestOption) (*SessionSnapshot, error) {
	var out SessionSnapshot
	if err := requestJSON(ctx, g.client, "POST", "/group-sessions", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Turn generates a group turn. Supply Idempotency-Key for safe retries.
func (g *GroupSessionsService) Turn(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (*GroupTurnResponse, error) {
	var out GroupTurnResponse
	if err := requestJSON(ctx, g.client, "POST", sessionPath(id)+"/group/turns", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TurnStream returns token, segment, parser_warning, usage and done/error SSE.
func (g *GroupSessionsService) TurnStream(ctx context.Context, id string, body JSONMap, opts ...RequestOption) (<-chan StreamEvent, error) {
	return streamJSON(ctx, g.client, "POST", sessionPath(id)+"/group/turns", body, headersFromOptions(opts))
}

// GetTurn reads the group record, including segments and parser warnings.
func (g *GroupSessionsService) GetTurn(ctx context.Context, id, turnID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, g.client, "GET", sessionPath(id)+"/group/turns/"+url.PathEscape(turnID), nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Prompt reads a turn's rendered prompt and hashes; requires prompt access.
// An empty turnID asks the server for its default turn.
func (g *GroupSessionsService) Prompt(ctx context.Context, id, turnID string, opts ...RequestOption) (JSONMap, error) {
	path := sessionPath(id) + "/group/prompt"
	if turnID != "" {
		path += "?" + url.Values{"turn_id": {turnID}}.Encode()
	}
	var out JSONMap
	if err := requestJSON(ctx, g.client, "GET", path, nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}
