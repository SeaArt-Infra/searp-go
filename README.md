# SeaRP Go SDK

Official Go client for the SeaRP engine HTTP API. It covers sessions, chat
turns, idempotent operations, cards, versions, generations, assemble, debug
chat, and cinema routes exposed by the engine `/v1` API.

## Install

```bash
go get github.com/SeaArt-Infra/searp-go
```

Requirements:

- Go 1.22+

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	rp "github.com/SeaArt-Infra/searp-go"
)

func main() {
	client, err := rp.New(&rp.ClientConfig{
		APIKey:  "rp-your-project-token",
		BaseURL: "https://rp.example.com",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	session, err := client.Sessions.Create(ctx, rp.JSONMap{
		"user_id": "visitor-1",
		"request": rp.JSONMap{
			"character": rp.JSONMap{"name": "Ada", "gender": 2},
			"style":     1,
			"lang":      "en",
		},
	}, rp.WithHeader("x-request-id", "request-1"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(session.ID)
}
```

`BaseURL` defaults to `http://127.0.0.1:8788`. The SDK derives the API base as
`<BaseURL>/v1` unless the base already ends in `/v1`.

`Sessions.Create` is the general way to create a playground, card, or version
session. `Sessions.CreateExperience` is only for projects that have already
published an experience version.

## Services

| Service | Purpose |
| --- | --- |
| `client.Sessions` | Playground, card, and version sessions; history; turns; rewind/fork/edit |
| `client.Operations` | Idempotent paid operations, recovery, and local traces |
| `client.Engine` | Health, capabilities, models, LLM checks, generations, assemble, debug chat |
| `client.Cards` | Role-card CRUD, versions, translations, listing |
| `client.Versions` | Version preview |
| `client.Cinema` | Cinema rounds and image tasks |

## Role Card Session

Create a role card with `client.Cards.Create`, then open a session from that
card:

```go
ctx := context.Background()

card, err := client.Cards.Create(ctx, rp.JSONMap{
	"user_id":      "visitor-1",
	"name":         "Ada",
	"gender":       2,
	"introduction": "Port pilot",
	"greeting":     "Welcome to the fog harbor.",
	"background":   "Knows the tides and shipping lanes.",
	"lang":         "en",
})
if err != nil {
	log.Fatal(err)
}

cardID, _ := card["id"].(string)
session, err := client.Sessions.Create(ctx, rp.JSONMap{
	"user_id": "visitor-1",
	"card_id": cardID,
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(session.ID)
```

`card_version` is optional; omit it to pin the latest card version. Use
`client.Cards.List`, `client.Cards.Get`, `client.Cards.Update`, and
`client.Cards.Delete` to maintain cards.

## Chat Turn

```go
ctx := context.Background()

turn, err := client.Sessions.Turn(ctx, session.ID, rp.JSONMap{
	"action":            "reply",
	"text":              "Hello.",
	"expected_revision": session.Revision,
}, rp.WithHeader("x-infra-project-id", projectID),
	rp.WithHeader("x-infra-user-id", userID),
	rp.WithHeader("x-request-id", requestID))
if err != nil {
	log.Fatal(err)
}
fmt.Println(turn.Revision)
```

Use `client.Sessions.TurnStream` for SSE token streaming:

```go
events, err := client.Sessions.TurnStream(ctx, session.ID, rp.JSONMap{
	"action": "reply",
	"text":   "Hello.",
	"stream": true,
})
if err != nil {
	log.Fatal(err)
}
for event := range events {
	if event.Err != nil {
		log.Fatal(event.Err)
	}
	if event.Done {
		break
	}
	if event.Event == "token" {
		fmt.Print(string(event.Data))
	}
}
```

## Error Handling

The engine reports plain-text HTTP errors. Inspect `*rp.Error` at request
boundaries:

```go
_, err := client.Sessions.Get(ctx, "missing", rp.SessionGetQuery{})
if err != nil {
	var seaErr *rp.Error
	if errors.As(err, &seaErr) {
		fmt.Println(seaErr.Kind, seaErr.Status)
	}
}
```

Common error kinds are `rp.ErrAuth`, `rp.ErrInvalid`, `rp.ErrNotFound`,
`rp.ErrConflict`, `rp.ErrQuota`, `rp.ErrTimeout`, `rp.ErrNetwork`, and
`rp.ErrGeneral`.

## Caller Context Headers

Paid generation routes accept SeaInfra attribution headers such as
`x-infra-project-id`, `x-infra-af-id`, `x-infra-session-id`,
`x-infra-user-id`, `x-infra-input-message-id`,
`x-infra-replay-message-id`, `x-task-id`, `x-request-id`, and `x-payload`.
Pass them per request with `rp.WithHeader` or `rp.WithHeaders`; do not bake
request-specific caller identity into long-lived `ClientConfig.Headers`.

## Development

```bash
make check
```

<script
  type="text/plain"
  data-doc-skill
  data-doc-skill-id="searp-go"
  data-doc-skill-label="SeaRP Go SDK"
  data-doc-skill-filename="searp-go-SKILL.md"
  data-doc-skill-version="1"
>
---
name: searp-go
description: Build and troubleshoot SeaRP engine integrations with the searp-go client. Use when creating role-play sessions, running chat turns or streaming replies, using idempotent operations, importing cards, previewing versions, assembling prompts, generating images, or debugging chats from Go.
---

# SeaRP Go SDK

Use `github.com/SeaArt-Infra/searp-go` to call the SeaRP engine `/v1` API from
Go 1.22+. Import it as `rp`. The SDK uses only the standard library.

## Install

```bash
go get github.com/SeaArt-Infra/searp-go
```

## Workflow

1. Create one `rp.Client` with the project bearer token and reuse it.
2. Use `client.Sessions` for sessions and turns; `client.Operations` for paid,
   idempotent generation; `client.Engine` for models, generations, assemble,
   and debug chat; `client.Cards`, `client.Versions`, and `client.Cinema` for
   their routes.
3. Prefer `POST /v1/sessions/{id}/operations` with a stable idempotency key for
   chat generation; reuse the same key for retries.
4. HTTP 200 does not guarantee a generated reply. Check the operation `status`;
   an SSE stream must receive `done` and may also receive `error`.
5. Pass request-specific SeaInfra attribution headers with `rp.WithHeaders`.

## Initialize Client

```go
client, err := rp.New(&rp.ClientConfig{
	APIKey:  "rp-your-project-token",
	BaseURL: "https://rp.example.com",
})
if err != nil {
	log.Fatal(err)
}
```

## Create A Session

```go
session, err := client.Sessions.Create(ctx, rp.JSONMap{
	"user_id": "user-123",
	"request": rp.JSONMap{
		"character": rp.JSONMap{"name": "Ada", "gender": 2},
		"style":     1,
		"lang":      "en",
	},
})
```

Use `client.Sessions.CreateExperience` only when the project has already
published an experience version.

## Run A Reply

```go
op, err := client.Operations.Run(ctx, session.ID, rp.JSONMap{
	"idempotency_key":   "reply-001",
	"action":            "reply",
	"text":              "Hello.",
	"expected_revision": session.Revision,
})
```

For streaming, use `client.Sessions.TurnStream` and stop on `done`.

## Errors

Catch `*rp.Error` and branch on `Kind`. `rp.ErrConflict` usually means a stale
`expected_revision` or an idempotency key reused with different input.

## Route Reference

- `Sessions.Create`, `CreateExperience`, `Get`, `HistoryMessage`, `Turn`,
  `TurnStream`, `Patch`, `Rewind`, `Fork`, `Edit`
- `Operations.Run`, `Get`, `List`, `Recover`, `Traces`
- `Engine.Health`, `Capabilities`, `Models`, `LLMStatus`, `LLMCheck`,
  `GenerationModels`, `CreateGeneration`, `GetGeneration`, `Assemble`,
  `DebugChat`, `DebugChatStream`
- `Cards.List`, `Create`, `Get`, `Update`, `Delete`, `Import`, `SetListing`,
  `ListVersions`, `GetVersion`, `DeleteVersion`, `UpdateTranslation`,
  `RestoreVersion`, `ListByUser`
- `Versions.Preview`
- `Cinema.ListRounds`, `CreateRound`, `CreateRoundStream`, `GetRound`,
  `GetImageTask`, `GenerateImageTask`, `SaveImageResult`
</script>
