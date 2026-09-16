package rp

import (
	"encoding/json"
	"fmt"
)

// JSONMap is a convenient representation for the SeaRP API's dynamic JSON
// request and response bodies.
type JSONMap = map[string]any

// RawResponse preserves a successful response body without decoding it.
type RawResponse = json.RawMessage

// StreamEvent is one parsed server-sent event from a streaming endpoint.
type StreamEvent struct {
	Event string
	Data  RawResponse
	Done  bool
	Err   error
}

// Decode unmarshals a raw SeaRP response into the requested type.
func Decode[T any](raw RawResponse) (*T, error) {
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, &Error{Kind: ErrGeneral, Message: "failed to decode response: " + err.Error()}
	}
	return &out, nil
}

// Message is one chat message in history or in the assembled model envelope.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Character is the character card block inside a render request.
type Character struct {
	Name        string `json:"name"`
	Gender      int64  `json:"gender"`
	Personality string `json:"personality"`
	Background  string `json:"background"`
	Greeting    string `json:"greeting"`
	Scenario    string `json:"scenario"`
	DialogStyle any    `json:"dialog_style,omitempty"`
}

// User is the player persona block inside a render request.
type User struct {
	Name    string `json:"name"`
	Lang    string `json:"lang,omitempty"`
	Gender  string `json:"gender"`
	Info    string `json:"info"`
	Like    string `json:"like"`
	Dislike string `json:"dislike"`
}

// Request is the stateless render/playground request bound into a session.
type Request struct {
	Character              Character `json:"character"`
	User                   User      `json:"user"`
	Style                  int64     `json:"style"`
	Lang                   string    `json:"lang"`
	Skeleton               string    `json:"skeleton"`
	MinWords               int64     `json:"min_words"`
	MaxWords               int64     `json:"max_words"`
	ExtraReplyIndex        int64     `json:"extra_reply_index"`
	ExampleText            string    `json:"example_text"`
	ReplyRequirementCustom string    `json:"reply_requirement_custom"`
	StylePromptOverride    string    `json:"style_prompt_override"`
	CharPlaceholder        string    `json:"char_placeholder"`
	UserPlaceholder        string    `json:"user_placeholder"`
	SkipOutput             bool      `json:"skip_output"`
	CardInstructions       *string   `json:"card_instructions,omitempty"`
}

// Rendered is the rendered prompt envelope returned by assemble and snapshot
// responses.
type Rendered struct {
	SkeletonKey             string `json:"skeleton_key"`
	Lang                    string `json:"lang"`
	SystemPrompt            string `json:"system_prompt"`
	OutputInstruction       string `json:"output_instruction"`
	ConversationStylePrompt string `json:"conversation_style_prompt"`
}

// SessionSample is the model sampling snapshot pinned by a session.
type SessionSample struct {
	TopP                 *float64 `json:"top_p,omitempty"`
	ThinkingMode         *bool    `json:"thinking_mode,omitempty"`
	ThinkingBudgetTokens *uint32  `json:"thinking_budget_tokens,omitempty"`
	Model                string   `json:"model"`
	Temperature          float64  `json:"temperature"`
	MaxTokens            uint32   `json:"max_tokens"`
}

// Assignment is one experiment variant pinned on a session.
type Assignment struct {
	ExperimentID string `json:"experiment_id"`
	VariantID    string `json:"variant_id"`
}

// Branch is an archived history replaced by a regenerate.
type Branch struct {
	Revision uint64    `json:"revision"`
	History  []Message `json:"history"`
}

// SessionSnapshot is the JSON session body returned by create, get, and JSON
// turn endpoints.
type SessionSnapshot struct {
	SystemTemplateID string         `json:"system_template_id"`
	ID               string         `json:"id"`
	ProjectID        string         `json:"project_id"`
	UserID           string         `json:"user_id"`
	Revision         uint64         `json:"revision"`
	VersionID        string         `json:"version_id"`
	PEDigest         string         `json:"pe_digest"`
	CardID           string         `json:"card_id,omitempty"`
	CardVersion      *uint64        `json:"card_version,omitempty"`
	Request          Request        `json:"request"`
	History          []Message      `json:"history"`
	HistoryTotal     uint64         `json:"history_total"`
	HistoryOffset    uint64         `json:"history_offset"`
	Branches         []Branch       `json:"branches,omitempty"`
	Pending          string         `json:"pending,omitempty"`
	Prompt           Rendered       `json:"prompt"`
	Messages         []Message      `json:"messages,omitempty"`
	Greeting         string         `json:"greeting"`
	MaxContext       uint64         `json:"max_context"`
	PinnedMemory     string         `json:"pinned_memory,omitempty"`
	PromptTokens     uint64         `json:"prompt_tokens"`
	Recalled         int            `json:"recalled"`
	Assignments      []Assignment   `json:"assignments,omitempty"`
	Sample           *SessionSample `json:"sample,omitempty"`
}

// HistoryMessage is one absolute-indexed element from the session history list.
type HistoryMessage struct {
	Index   uint64 `json:"index"`
	Total   uint64 `json:"total"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AssembleResponse is the response returned by POST /v1/assemble.
type AssembleResponse struct {
	Prompt       Rendered  `json:"prompt"`
	Messages     []Message `json:"messages"`
	PromptTokens uint64    `json:"prompt_tokens"`
}

// TurnResponse is the non-streaming response returned by POST /v1/sessions/{id}/turns.
type TurnResponse struct {
	SessionSnapshot
	Execution AssembleResponse `json:"execution"`
	Usage     *json.RawMessage `json:"usage"`
}

// PublicModel is a chat or multimodal model returned by the engine catalog.
type PublicModel struct {
	Model       string   `json:"model"`
	Provider    string   `json:"provider,omitempty"`
	ModelType   string   `json:"model_type,omitempty"`
	Modalities  []string `json:"modalities,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
}

// Generation is a SeaInfra multimodal generation task.
type Generation struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Model  string `json:"model"`
	Output any    `json:"output,omitempty"`
}

func (e *Error) Unwrap() error { return nil }

func decodeJSON[T any](raw []byte) (*T, error) {
	return Decode[T](RawResponse(raw))
}

func stringValue(body JSONMap, key string) (string, error) {
	value, ok := body[key]
	if !ok {
		return "", nil
	}
	s, ok := value.(string)
	if !ok {
		return "", &Error{Kind: ErrInvalid, Message: fmt.Sprintf("%s must be a string", key)}
	}
	return s, nil
}
