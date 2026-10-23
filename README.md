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
		"card_info": rp.JSONMap{
			"name": "Ada", "gender": 2, "lang": "en",
			"introduction": "Port pilot",
			"greeting": "Welcome to the fog harbor.",
			"background": "Knows the tides and shipping lanes.",
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

`Sessions.Create` accepts `card_id`, inline `card_info`, or `version_id`.
Use `Engine.DebugChat` for temporary card/sampling configurations;
`request` and sampling fields are rejected by `Sessions.Create`.
`Sessions.CreateExperience` opens the currently published experience version.

See [API reference](API_REFERENCE.md) for all routes, pagination, translations,
multimodal input, per-turn parameters, and migration notes.

## Release API update

Aligned with SeaRP release API docs `af4e254` on 2026-10-23. The SDK now
covers group story cards and sessions, player inspiration suggestions, missing
translation queries, translation reuse, catalog fork and Cinema order.

```go
suggestions, err := client.Sessions.Inspiration(ctx, session.ID, rp.JSONMap{"lang": "zh"})
if err != nil {
    log.Fatal(err)
}
fmt.Println(suggestions.Replies)
```

```go
group, err := client.GroupSessions.Create(ctx, rp.JSONMap{
    "group_id": groupID, "user_id": "visitor-1", "mode": "group_chat",
})
if err != nil {
    log.Fatal(err)
}
turn, err := client.GroupSessions.Turn(ctx, group.ID, rp.JSONMap{
    "content": "Hello everyone.", "expected_revision": group.Revision,
}, rp.WithHeader("Idempotency-Key", "group-turn-001"))
if err != nil {
    log.Fatal(err)
}
fmt.Println(turn.Revision, turn.Turn["segments"])
```

Use `client.GroupSessions.TurnStream` for group SSE.
Handle segment events, errors and disconnects; only done confirms completion.
Inspiration changes no history and calls the model again on each retry.

`POST /v1/cards` is deprecated and scheduled to sunset on 2027-01-01.
The role-card creation example below remains for existing integrations.
For catalog cards, configure `ControlAPIBaseURL` and call
`client.Cards.Fork` with a project ID and `catalog_ids`; inspect
both `items` and `errors`. See [API reference](API_REFERENCE.md) for route
coverage, import-document differences, group request fields and recovery.

## Services

| Service | Purpose |
| --- | --- |
| `client.Sessions` | Playground, card, and version sessions; turns and proactive messages; history; rewind/fork/edit |
| `client.Operations` | Idempotent paid operations, recovery, and local traces |
| `client.Engine` | Health, capabilities, models, LLM checks, generations, assemble, debug chat |
| `client.Cards` | Role-card CRUD, versions, translations, listing |
| `client.Versions` | Version preview |
| `client.Cinema` | Cinema rounds and image tasks |
| `client.GroupCards` | Group story cards, versions, members and translations |
| `client.GroupSessions` | Group session creation, JSON/SSE turns and prompt records |

## Legacy Role Card Creation

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

Use `client.Sessions.Proactive` for an idempotent character-initiated message
without user input, or `ProactiveStream` for SSE:

```go
proactive, err := client.Sessions.Proactive(ctx, session.ID, rp.JSONMap{
	"idempotency_key": "proactive-001",
	"expected_revision": session.Revision,
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(proactive.Text)
```

## Error Handling

The engine reports plain-text or JSON errors. Inspect `*rp.Error` at request
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
2. Use `client.Sessions` for sessions, turns, and idempotent proactive messages; `client.Operations` for paid,
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
	"card_info": rp.JSONMap{
		"name": "Ada", "gender": 2, "lang": "en",
		"introduction": "Port pilot",
		"greeting": "Welcome to the fog harbor.",
		"background": "Knows the tides and shipping lanes.",
	},
})
```

Use `card_id`, `card_info`, or `version_id` with `Sessions.Create`;
use `Engine.DebugChat` with `card_info` for temporary sampling/configuration.
`Sessions.Create` rejects a nested `request`.
Use `client.Sessions.CreateExperience` for the currently published version.

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

## Release API additions (2026-10-23)

Use `client.GroupCards` for story cards, immutable versions, members,
imports, preview and translations. Use `client.GroupSessions` for creating group
sessions, generating JSON/SSE turns, querying turn records and prompts.
Member removals require a JSON body with `user_id` and
`expected_latest_version`. Group generation takes `content` and supports a
stable `Idempotency-Key` header; preserve the key and original input on retry.
JSON responses include the flattened session revision and group `turn.segments`.
The shared SSE parser also exposes `segment_*` and `parser_warning` events.
Read history through the ordinary session get method.

`client.Sessions.Inspiration` returns positive/neutral player reply
suggestions. It accepts an empty body or per-call language and sampling,
changes no history/revision, and has no SSE or idempotency. Every retry calls
the model again. `client.Cards.Untranslated` queries missing
translations; `ForkTranslations` copies existing catalog translations.
Check per-item errors even when HTTP is 200.

Card creation is deprecated with sunset 2027-01-01; existing create examples
are for compatibility. Use catalog fork with `catalog_ids` and an explicit
Control API base for onboarding published cards. Legacy batch-import is kept
for older deployments. Consult API_REFERENCE.md for import payload differences
between the release documents and registered handlers.
Cinema lists support descending offset order; returned arrays remain sorted
by ascending sequence. Recover abandon always needs
`acknowledge_unknown: true`; use operation revision for recovery and session
revision for generation. Terminal operation evidence defaults to 7 hours.

## Route Reference

- `Sessions.Create`, `CreateExperience`, `Get`, `HistoryMessage`, `Turn`,
  `TurnStream`, `Proactive`, `ProactiveStream`, `Inspiration`, `Patch`, `Rewind`, `Fork`, `Edit`
- `Operations.Run`, `Get`, `List`, `Recover`, `Traces`
- `Engine.Health`, `Live`, `Metrics`, `Capabilities`, `Models`, `LLMStatus`, `LLMCheck`,
  `GenerationModels`, `CreateGeneration`, `GetGeneration`, `Assemble`,
  `DebugChat`, `DebugChatStream`
- `Cards.List`, `Create`, `Get`, `Update`, `Delete`, `Import`, `SetListing`,
  `ListVersions`, `GetVersion`, `DeleteVersion`, `UpdateTranslation`,
  `ListVersionsWithQuery`, `GetVersionWithQuery`, `RestoreVersionWithBody`,
  `ListTranslations`, `SaveTranslations`, `ImportBatch`, `ListByUser`
- `Versions.Preview`
- `Cinema.ListRounds`, `ListRoundsWithQuery`, `CreateRound`, `CreateRoundStream`, `GetRound`,
  `GetImageTask`, `GenerateImageTask`, `SaveImageResult`

- `Cards.Untranslated`, `ForkTranslations`, `Fork`
- `GroupCards`: CRUD, listing, versions, roster, characters, preview, import, translations
- `GroupSessions`: create, turn, turn stream, get turn, prompt
</script>
