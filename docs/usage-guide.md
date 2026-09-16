# SeaRP Go SDK Usage Guide

The SDK is organized around one `rp.Client` and a set of service fields.
Requests are dynamic JSON maps, while common responses such as
`SessionSnapshot`, `TurnResponse`, and `PublicModel` are typed.

## Client Configuration

```go
client, err := rp.New(&rp.ClientConfig{
	APIKey:     "rp-your-project-token",
	BaseURL:    "https://rp.example.com", // optional
	APIBaseURL: "https://rp.example.com/v1", // optional override
	Headers: http.Header{
		"x-fixed-header": {"fixed-value"},
	},
	Timeout: 120 * time.Second,
})
```

## Session Lifecycle

Create a session from the current published version:

```go
session, err := client.Sessions.CreateExperience(ctx, rp.JSONMap{
	"user_id": "user-123",
})
```

Create a playground or card session with `client.Sessions.Create`:

```go
session, err := client.Sessions.Create(ctx, rp.JSONMap{
	"user_id": "user-123",
	"request": rp.JSONMap{
		"character": rp.JSONMap{"name": "Ada", "gender": 2},
		"style":     1,
		"lang":      "en",
	},
	"max_context":   16000,
	"pinned_memory": "Keep the lamp lit.",
})
```

## Chat And History

- `Sessions.Turn(ctx, id, body, opts...)` runs a JSON turn.
- `Sessions.TurnStream(ctx, id, body, opts...)` runs an SSE turn.
- `Sessions.HistoryMessage(ctx, id, index, opts...)` reads one history element.
- `Sessions.Rewind`, `Sessions.Fork`, `Sessions.Edit`, and `Sessions.Patch`
  mutate session history or settings.

## Idempotent Operations

Prefer the operations API for paid generation:

```go
op, err := client.Operations.Run(ctx, session.ID, rp.JSONMap{
	"idempotency_key":   "reply-001",
	"action":            "reply",
	"text":              "Hello.",
	"expected_revision": session.Revision,
})
```

Reuse the same key for retries; changing input under an existing key returns a
conflict.

## Models, Assemble, And Debug

```go
models, _ := client.Engine.Models(ctx)
generationModels, _ := client.Engine.GenerationModels(ctx)
task, _ := client.Engine.CreateGeneration(ctx, rp.JSONMap{
	"model":  "nano_banana_2_lite",
	"prompt": "a red apple",
})
assembled, _ := client.Engine.Assemble(ctx, rp.JSONMap{
	"request":      rp.JSONMap{"character": rp.JSONMap{"name": "Ada"}},
	"history":      []any{},
	"user_message": "Hello.",
})
```

See `README.md` and the generated `skills/searp-go/SKILL.md` for endpoint-level
examples.
