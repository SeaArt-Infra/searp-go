package rp

import (
	"context"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

// EngineService provides the non-session engine endpoints: health,
// capabilities, models, LLM connection checks, generations, assemble, and
// debug chat.
type EngineService struct {
	client *transport.Client
}

// Health returns the engine health snapshot from GET /v1/health.
func (e *EngineService) Health(ctx context.Context) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, e.client, "GET", "/health", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Capabilities returns the engine capability and default-sample snapshot.
func (e *EngineService) Capabilities(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, e.client, "GET", "/capabilities", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Models returns the available chat model catalog.
func (e *EngineService) Models(ctx context.Context, opts ...RequestOption) ([]PublicModel, error) {
	var out []PublicModel
	if err := requestJSON(ctx, e.client, "GET", "/models", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// LLMStatus returns the current project LLM connection state.
func (e *EngineService) LLMStatus(ctx context.Context, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, e.client, "GET", "/llm", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// LLMCheck verifies the current project LLM connection.
func (e *EngineService) LLMCheck(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, e.client, "POST", "/llm/check", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GenerationModels returns the SeaInfra multimodal catalog.
func (e *EngineService) GenerationModels(ctx context.Context, opts ...RequestOption) ([]PublicModel, error) {
	var out []PublicModel
	if err := requestJSON(ctx, e.client, "GET", "/generations/models", nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateGeneration submits a multimodal generation task.
func (e *EngineService) CreateGeneration(ctx context.Context, body JSONMap, opts ...RequestOption) (*Generation, error) {
	var out Generation
	if err := requestJSON(ctx, e.client, "POST", "/generations", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetGeneration polls a multimodal generation task.
func (e *EngineService) GetGeneration(ctx context.Context, id string, opts ...RequestOption) (*Generation, error) {
	var out Generation
	if err := requestJSON(ctx, e.client, "GET", "/generations/"+id, nil, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Assemble renders a prompt envelope without touching Redis.
func (e *EngineService) Assemble(ctx context.Context, body JSONMap, opts ...RequestOption) (*AssembleResponse, error) {
	var out AssembleResponse
	if err := requestJSON(ctx, e.client, "POST", "/assemble", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DebugChat performs a stateless card debug chat. The request body controls
// whether the session/history is persisted; SSE is available by setting
// stream=true or Accept: text/event-stream.
func (e *EngineService) DebugChat(ctx context.Context, body JSONMap, opts ...RequestOption) (JSONMap, error) {
	var out JSONMap
	if err := requestJSON(ctx, e.client, "POST", "/debug/chat", body, headersFromOptions(opts), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DebugChatStream performs a streaming stateless card debug chat.
func (e *EngineService) DebugChatStream(ctx context.Context, body JSONMap, opts ...RequestOption) (<-chan StreamEvent, error) {
	return streamJSON(ctx, e.client, "POST", "/debug/chat", body, headersFromOptions(opts))
}
