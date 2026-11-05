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
Use `client.Sessions.CustomTask` for read-only caller-supplied PE work,
`client.Cards.Translate` for model-backed card translation, and
`client.WorldBooks` for project/catalog world books.
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
