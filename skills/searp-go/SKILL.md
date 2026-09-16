---
name: searp-go
description: Build and troubleshoot SeaRP engine and control-plane integrations with the searp-go client. Use when creating role-play sessions, running chat turns or streaming replies, using idempotent operations, importing cards, previewing versions, assembling prompts, generating images, debugging chats, or managing projects through the /admin/v1 gateway from Go.
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
   their routes. Use `client.Admin` for the `/admin/v1` control plane.
3. Prefer `POST /v1/sessions/{id}/operations` with a stable idempotency key for
   chat generation; reuse the same key for retries.
4. HTTP 200 does not guarantee a generated reply. Check the operation `status`;
   an SSE stream must receive `done` and may also receive `error`.
5. Pass request-specific SeaInfra attribution headers with `rp.WithHeaders`.

## Control Plane

```go
health, err := client.Admin.Health(ctx)
whoami, err := client.Admin.Whoami(ctx)
projects, err := client.Admin.ListProjects(ctx)
project, err := client.Admin.CreateProject(ctx, rp.JSONMap{
	"id":    "project-id",
	"token": "project-token-with-at-least-16-characters",
})
```

For control-plane routes without a typed method, use
`client.Admin.Request(ctx, method, path, body, opts...)`. Admin errors carry
their envelope `code` in `*rp.Error.Code`.

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

Create from the current published version:

```go
session, err := client.Sessions.CreateExperience(ctx, rp.JSONMap{
	"user_id": "user-123",
})
```

Create a playground session from a role request:

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

## Run A Reply

```go
op, err := client.Operations.Run(ctx, session.ID, rp.JSONMap{
	"idempotency_key":   "reply-001",
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
```

For streaming, call `client.Sessions.TurnStream` and stop on `done`:

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

## Errors

Catch `*rp.Error` at the request boundary and branch on `Kind` where retry or
user feedback differs. `rp.ErrConflict` usually means a stale
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
- `Admin.Request`, `Health`, `Whoami`, `AgentContract`, `ListProjects`,
  `CreateProject`, `GetProject`, `DeleteProject`, `RotateProjectToken`,
  `GetProjectLive`, `UpdateProjectLive`
- `Admin.Raw`, `ProjectRequest`, `GetGlobalPack`, `UpdateGlobalPack`,
  `GetProjectPack`, `PatchProjectPack`, `DeleteProjectPack`, `ForkProjectPack`
- `Admin.ListCatalog`, `ImportCatalog`, `GetCatalogCard`, `UpdateCatalogCard`,
  `DeleteCatalogCard`, `GetCatalogCardCover`
- `Admin.ListProjectCards`, `GetProjectCard`, `UpdateProjectCard`,
  `DeleteProjectCard`, `SetProjectCardListing`, `ImportProjectCard`,
  `ImportProjectCardsBatch`, `ForkProjectCard`, `ListProjectCardVersions`,
  `GetProjectCardVersion`, `DeleteProjectCardVersion`,
  `RestoreProjectCardVersion`
- `Admin.ListProjectExperiments`, `CreateProjectExperiment`,
  `GetProjectExperiment`, `UpdateProjectExperiment`, `StartProjectExperiment`,
  `PauseProjectExperiment`, `StopProjectExperiment`
- `Admin.GetProjectLLM`, `UpdateProjectLLM`, `DeleteProjectLLM`
- `Admin.ListProjectVersions`, `CreateProjectVersion`, `GetProjectVersion`,
  `DiffProjectVersion`, `PublishProjectVersion`, `GetProjectRelease`,
  `ListProjectReleases`, `RollbackProjectRelease`
- `Admin.ListProjectSystemPrompts`, `CreateProjectSystemPrompt`,
  `SetProjectSystemPromptDefault`, `GetProjectSystemPrompt`,
  `UpdateProjectSystemPrompt`, `ListGlobalSystemPrompts`,
  `CreateGlobalSystemPrompt`, `SetGlobalSystemPromptDefault`,
  `GetGlobalSystemPrompt`
- `Admin.ListProjectUserSessions`, `GetProjectUserSession`,
  `GetProjectIdentityMigration`, `StartProjectIdentityMigration`,
  `PrepareProjectIdentityMigration`, `PurgeProjectIdentityMigration`,
  `AdoptProjectIdentityMigration`, `RevertProjectIdentityMigration`,
  `PreviewProjectIdentityMigration`
