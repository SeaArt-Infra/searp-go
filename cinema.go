package rp

import (
	"context"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// CinemaService provides the cinema round and image-task endpoints.
type CinemaService struct {
	client *transport.Client
}

// ListRounds lists a session's cinema rounds.
func (c *CinemaService) ListRounds(ctx context.Context, sessionID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "GET", cinemaSessionPath(sessionID)+"/rounds", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateRound creates a cinema round. Set stream=true for SSE.
func (c *CinemaService) CreateRound(ctx context.Context, sessionID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "POST", cinemaSessionPath(sessionID)+"/rounds", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateRoundStream creates a streaming cinema round.
func (c *CinemaService) CreateRoundStream(ctx context.Context, sessionID string, body JSONMap, opts ...RequestOption) (<-chan StreamEvent, error) {
	return streamJSON(ctx, c.client, "POST", cinemaSessionPath(sessionID)+"/rounds", body, headersFromOptions(opts))
}

// GetRound returns one cinema round.
func (c *CinemaService) GetRound(ctx context.Context, sessionID, roundID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "GET", cinemaSessionPath(sessionID)+"/rounds/"+roundID, nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetImageTask returns an owned cinema image task.
func (c *CinemaService) GetImageTask(ctx context.Context, sessionID, taskID string, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "GET", cinemaSessionPath(sessionID)+"/image-tasks/"+taskID, nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GenerateImageTask generates a cinema image for an image task.
func (c *CinemaService) GenerateImageTask(ctx context.Context, taskID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "POST", "/cinema/image-tasks/"+taskID+"/generate", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SaveImageResult records a cinema image result.
func (c *CinemaService) SaveImageResult(ctx context.Context, taskID string, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, c.client, "POST", "/cinema/image-tasks/"+taskID+"/result", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func cinemaSessionPath(sessionID string) string {
	return "/sessions/" + sessionID + "/cinema"
}
