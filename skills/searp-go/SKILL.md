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
session, err := client.Sessions.CreateExperience(ctx, rp.JSONMap{
	"user_id": "user-123",
})
```

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
