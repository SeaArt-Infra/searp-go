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
	Timeout: 300 * time.Second,
})
```

## Session Lifecycle

Create a card or version session with `client.Sessions.Create`:

```go
session, err := client.Sessions.Create(ctx, rp.JSONMap{
	"user_id": "user-123",
	"card_info": rp.JSONMap{
		"name": "Ada", "gender": 2, "lang": "en",
		"introduction": "Port pilot", "greeting": "Welcome to the harbor.",
		"background": "Knows the tides and shipping lanes.",
	},
})
```

If your project has already published an experience version, use
`client.Sessions.CreateExperience` instead:

```go
session, err := client.Sessions.CreateExperience(ctx, rp.JSONMap{
	"user_id": "user-123",
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

For temporary request/sampling configuration, use `Engine.DebugChat`;
`Sessions.Create` rejects nested `request` and sampling fields.

## Group Chat And Suggestions

Use `GroupCards` for story cards and rosters, `GroupSessions.Create` for a
session pinned to a group version, and `GroupSessions.Turn` / `TurnStream`
for generation. Pass `Idempotency-Key` with `WithHeader` for safe retries.
`GroupTurnResponse` exposes the updated session revision, operation and group
record. Members are soft-deleted with a body containing `user_id` and
`expected_latest_version`. Read all group history through `Sessions.Get`.

`Sessions.Inspiration(ctx, id, JSONMap{"lang": "zh"})` generates positive and
neutral player suggestions without modifying history. It is JSON-only and
each retry calls the model again. `Cards.Untranslated` and
`Cards.ForkTranslations` expose translation completeness and reuse.

`Cinema.ListRoundsWithOrder` takes `CinemaRoundQuery{Order: "desc"}` for the
latest rounds; existing pagination-only methods remain compatible.

`Cards.Create` is deprecated with a documented sunset of 2027-01-01.
Configure `ControlAPIBaseURL` and use `Cards.Fork` to copy catalog cards into a
project. See API_REFERENCE.md for differences between import documentation
and registered release handlers.

## Models, Assemble, And Debug Examples

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
