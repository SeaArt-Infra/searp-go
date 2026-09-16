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

	session, err := client.Sessions.CreateExperience(context.Background(), rp.JSONMap{
		"user_id": "visitor-1",
	}, rp.WithHeader("x-request-id", "request-1"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(session.ID)
}
```

`BaseURL` defaults to `http://127.0.0.1:8788`. The SDK derives the API base as
`<BaseURL>/v1` unless the base already ends in `/v1`.

## Services

| Service | Purpose |
| --- | --- |
| `client.Sessions` | Playground, card, and version sessions; history; turns; rewind/fork/edit |
| `client.Operations` | Idempotent paid operations, recovery, and local traces |
| `client.Engine` | Health, capabilities, models, LLM checks, generations, assemble, debug chat |
| `client.Cards` | Role-card CRUD, versions, translations, listing |
| `client.Versions` | Version preview |
| `client.Cinema` | Cinema rounds and image tasks |

## Chat Turn

```go
turn, err := client.Sessions.Turn(ctx, session.ID, rp.JSONMap{
	"action":            "reply",
	"text":              "Hello.",
	"expected_revision": session.Revision,
}, rp.WithHeaders(http.Header{
	"x-infra-project-id": {projectID},
	"x-infra-user-id":    {userID},
	"x-request-id":       {requestID},
}))
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
