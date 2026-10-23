# SeaRP API reference

Synced on 2026-10-23 against the SeaRP release API docs at `af4e254`:
`docs/searp-api.md`, `docs/card-import-api.md` and the card creation migration
guide. Paths below are relative to engine `/v1` unless marked Control API.
Documentation/route differences are recorded below.

## Routes

| Method | Path | Client method |
| --- | --- | --- |
| GET | `/health` | `Engine.Health` |
| GET | `/live` | `Engine.Live` |
| GET | `/metrics` | `Engine.Metrics` |
| GET | `/capabilities` | `Engine.Capabilities` |
| GET | `/models` | `Engine.Models` |
| GET | `/llm` | `Engine.LLMStatus` |
| POST | `/llm/check` | `Engine.LLMCheck` |
| GET | `/generations/models` | `Engine.GenerationModels` |
| POST | `/generations` | `Engine.CreateGeneration` |
| GET | `/generations/{id}` | `Engine.GetGeneration` |
| POST | `/assemble` | `Engine.Assemble` |
| POST | `/debug/chat` | `Engine.DebugChat / DebugChatStream` |
| POST | `/sessions` | `Sessions.Create` |
| POST | `/experience/sessions` | `Sessions.CreateExperience` |
| GET | `/sessions/{id}` | `Sessions.Get` |
| PATCH | `/sessions/{id}` | `Sessions.Patch` |
| GET | `/sessions/{id}/history/{index}` | `Sessions.HistoryMessage` |
| POST | `/sessions/{id}/turns` | `Sessions.Turn / TurnStream` |
| POST | `/sessions/{id}/proactive` | `Sessions.Proactive / ProactiveStream` |
| POST | `/sessions/{id}/rewind` | `Sessions.Rewind` |
| POST | `/sessions/{id}/fork` | `Sessions.Fork` |
| POST | `/sessions/{id}/edit` | `Sessions.Edit` |
| POST | `/sessions/{id}/operations` | `Operations.Run` |
| GET | `/operations` | `Operations.List` |
| GET | `/operations/{id}` | `Operations.Get` |
| POST | `/operations/{id}/recover` | `Operations.Recover` |
| GET | `/traces` | `Operations.Traces` |
| POST | `/versions/{id}/preview` | `Versions.Preview` |
| GET | `/cards` | `Cards.List` |
| POST (deprecated; sunset 2027-01-01) | `/cards` | `Cards.Create` |
| GET | `/cards/{id}` | `Cards.Get` |
| PATCH | `/cards/{id}` | `Cards.Update` |
| DELETE | `/cards/{id}` | `Cards.Delete` |
| POST | `/cards/import` | `Cards.Import` |
| PATCH | `/cards/{id}/listing` | `Cards.SetListing` |
| GET | `/users/{user_id}/cards` | `Cards.ListByUser` |
| GET | `/cards/{id}/versions` | `Cards.ListVersions / ListVersionsWithQuery` |
| GET | `/cards/{id}/versions/{version}` | `Cards.GetVersion / GetVersionWithQuery` |
| DELETE | `/cards/{id}/versions/{version}` | `Cards.DeleteVersion` |
| POST | `/cards/{id}/versions/{version}/restore` | `Cards.RestoreVersionWithBody` |
| PATCH | `/cards/{id}/versions/{version}/translations/{lang}/{field}` | `Cards.UpdateTranslation` |
| GET | `/cards/{id}/versions/{version}/translations` | `Cards.ListTranslations` |
| POST | `/cards/{id}/versions/{version}/translations` | `Cards.SaveTranslations` |
| GET | `/sessions/{id}/cinema/rounds` | `Cinema.ListRounds / ListRoundsWithQuery / ListRoundsWithOrder` |
| POST | `/sessions/{id}/cinema/rounds` | `Cinema.CreateRound / CreateRoundStream` |
| GET | `/sessions/{id}/cinema/rounds/{round_id}` | `Cinema.GetRound` |
| GET | `/sessions/{id}/cinema/image-tasks/{task_id}` | `Cinema.GetImageTask` |
| POST | `/cinema/image-tasks/{task_id}/generate` | `Cinema.GenerateImageTask` |
| POST | `/cinema/image-tasks/{task_id}/result` | `Cinema.SaveImageResult` |
| POST | `/sessions/{id}/inspiration` | `Sessions.Inspiration` |
| GET | `/cards/untranslated` | `Cards.Untranslated` |
| POST | `/cards/translations/fork` | `Cards.ForkTranslations` |
| POST | `/admin/v1/projects/{project_id}/cards/fork` | `Cards.Fork (Control API)` |
| GET | `/group-cards` | `GroupCards.List` |
| POST | `/group-cards` | `GroupCards.Create` |
| GET | `/group-cards/{id}` | `GroupCards.Get` |
| PATCH | `/group-cards/{id}` | `GroupCards.Update` |
| DELETE | `/group-cards/{id}` | `GroupCards.Delete` |
| PATCH | `/group-cards/{id}/listing` | `GroupCards.SetListing` |
| GET | `/group-cards/{id}/versions` | `GroupCards.ListVersions` |
| GET | `/group-cards/{id}/versions/{version}` | `GroupCards.GetVersion` |
| GET | `/group-cards/{id}/members` | `GroupCards.ListMembers` |
| POST | `/group-cards/{id}/members` | `GroupCards.AddMember` |
| PATCH | `/group-cards/{id}/members/{participant_id}` | `GroupCards.UpdateMember` |
| DELETE | `/group-cards/{id}/members/{participant_id}` | `GroupCards.RemoveMember` |
| POST | `/group-cards/{id}/characters` | `GroupCards.ExtractCharacters` |
| POST | `/group-cards/{id}/preview` | `GroupCards.Preview` |
| POST | `/group-cards/import` | `GroupCards.Import` |
| GET | `/group-cards/{id}/versions/{version}/translations` | `GroupCards.ListTranslations` |
| POST | `/group-cards/{id}/versions/{version}/translations` | `GroupCards.SaveTranslations` |
| POST | `/group-sessions` | `GroupSessions.Create` |
| POST | `/sessions/{id}/group/turns` | `GroupSessions.Turn / TurnStream` |
| GET | `/sessions/{id}/group/turns/{turn_id}` | `GroupSessions.GetTurn` |
| GET | `/sessions/{id}/group/prompt` | `GroupSessions.Prompt` |

## Requests and current contract

Request bodies use the server's snake_case JSON keys and are passed through
unchanged. Authentication and request-specific headers come from the client and
request options. Keep project credentials on the caller's server.

- Session creation requires `user_id` and `card_id`, inline `card_info`, or
  `version_id`. Inline cards require nonempty `name`, `introduction`,
  `greeting`, and `background`; `gender` accepts 0/1/2 or the documented
  male/female/other aliases. A nested `request` and sampling fields are rejected.
  Optional `card_version` pins a stored card version. Optional
  `system_template_id` selects a template ID; `user` is the player persona.
- Experience creation accepts `user_id`, optional `user` and
  `system_template_id`, and uses the currently published version.
- Debug chat takes inline `card_info` plus `text` or `message` on the first
  call. It defaults to `persist: true`; continue with the returned
  `session_id` and `expected_revision`. Use `persist: false` with
  `card_info` and `history` for a stateless call. Debug chat accepts sampling,
  memory and per-call prompt parameters; it does not accept a nested `request`.
- Turns take `action` (reply/additional/regenerate), `text` or `message`,
  optional `expected_revision`, sampling, memory, `user`, `lang`,
  `system_prompt_override`, `lastest_system_prompt` (keep this spelling),
  `favorability_score` (-1..100), and `metadata.history`.
- Proactive takes a required `idempotency_key`, optional `expected_revision`,
  sampling, memory and per-call prompt fields, and never accepts user input;
  use `ProactiveStream` for the same SSE operation/token/usage/done events as
  turns.
- Operations take `idempotency_key`, `action`, `text` or `message`,
  `expected_revision` and `metadata` at the top level. Put model,
  temperature, max_tokens, user and request overrides under `live`; put
  max_context and pinned_memory under `memory`. Sampling options
  `top_p`, `thinking_mode`, `thinking_budget_tokens` and prompt parameters
  `lastest_system_prompt`, `favorability_score`, `lang` can also be top-level.
  `system_prompt_override` belongs under `live`. Top-level non-null values
  override matching live values. Operations always return JSON.
- `metadata.history` only seeds an empty session. The response's
  `history_import` records seeded/skipped status; leading assistant messages
  are dropped. On operations this transcript participates in the input digest:
  keep it unchanged when retrying the same key.
- Sampling fields are `model`, `temperature` (0..2), `max_tokens`
  (1..1000000), `top_p` (0 < p <= 1), `thinking_mode` and
  `thinking_budget_tokens`. Thinking support depends on the selected model;
  Claude budgets require enabled thinking and 1024 <= budget < max_tokens.
  Gemini 3 does not accept thinking_mode=false.
- `lang`, per-turn `user` and prompt instructions apply to that call only.
  Sampling and memory updates on non-version sessions are saved on success.
  Version sessions reject fixed configuration overrides; card sessions reject
  replacing their `request`. PATCH accepts `request` only for debug sessions.
- Assemble requires `request`; optional `history`, `user_message`,
  `max_context`, `pinned_memory`, `cinema` and prompt parameters are passed
  through. Version preview requires `text` and optional prompt parameters.
- Rewind/edit/fork accept `expected_revision`. Rewind and edit require
  an absolute history `index`; fork can omit it to copy all retained history.
  Edit takes text `content` and preserves attached images.
- Recover requires the **operation** `expected_revision` and `action`
  (commit/abandon); every abandonment requires
  `acknowledge_unknown: true`. Inspect operation status before recovery.

## Pagination and translations

| Resource | Query | Response / continuation |
| --- | --- | --- |
| Session history | limit, offset | Latest page by default, limit default/max 10; continue with history_offset + history.length, checking revision across pages |
| Project/user cards | lang, limit, offset; project list also ids/card_ids | Array; limit default 50, max 200; explicit offset=0 is preserved |
| Card versions | limit, before_version | {items, next_before_version}; limit default 20, max 100 |
| Card version detail | lang | Localized snapshot inside version metadata |
| Cinema rounds | limit, offset, order (asc/desc) | Array; limit default 10, max 100 |
| Operations | session_id, version_id, status, limit, offset | {items, next_offset}; continue even if filtering leaves an empty page, until next_offset is null |

Retained session history can start after index 0. A cold-history handoff can
return 409; retry the read rather than combining mismatched revisions.

Translation GET returns `{language_code: {field: text}}`. Translation POST
saves owner-provided fields with
`{user_id, expected_latest_version, lang, translations: {name: "...", ...}}`;
it does not run a translation model. The single-field PATCH requires
`{value, expected_latest_version}`. Supported fields are name, background,
dialog_style, reply_settings, introduction, greeting and scenario.

Restore requires `{expected_latest_version, change_note?}`. It copies the
selected snapshot into a new latest version; it does not reset the version
counter. Card updates/restores can use an `Idempotency-Key` request header.
Translation writes do not create content versions.

## Images, streaming and Cinema

For image chat, replace `text` with `message`:
`[{type: "text", text: "What is this?"}, {type: "image_url", image_url: {url: "https://cdn.example/a.png", detail: "auto"}}]`.
Do not send both nonempty text and message. Additional/regenerate omit both.
The server accepts up to 8 images, 1 MiB per inline image URL and 2 MiB inline
total. History `content` is a string or a content-part array, including in the
single-history-message response.

Streaming methods set `Accept: text/event-stream` automatically. Preserve
token whitespace and handle SSE error events and disconnects; receiving HTTP
200 alone is not generation success. Sessions/turns done data is plain text
(session ID followed by reply); debug-chat done data is JSON.

Cinema requires a stable `Idempotency-Key` header. Round bodies take content,
option_index (0 is valid), custom_input and expected_revision, plus optional
model, temperature, max_tokens, top_p, max_context, pinned_memory,
conversation_style, user, lang and image_model. Those overrides apply to the
round. Cinema's outer SSE event name is `cinema`; parse data JSON and inspect
its inner `event` (including message_end and error), rather than waiting for a
chat-style done event. Event IDs are exposed by the SDK. Events are live
transport only; recover final state using operation, Round and image-task
queries after disconnect. A Last-Event-ID header must refer to an existing
operation for the same round; it does not guarantee historical event replay.

Image generate/result bodies require `session_id`. Generate accepts optional
model and only claims pending tasks. Result accepts status, image_url,
watermark_url, material_id and error; status defaults to ready, which requires
image_url. Costs are provider-reported strings or null. The SDK preserves
these values and does not invent cost estimates.

## Card creation, import and visibility

`POST /v1/cards` remains callable but is deprecated, with a documented sunset
of 2027-01-01. Existing create methods remain available for compatibility.
For cards already in the catalog, configure `ControlAPIBaseURL` to the full
`/admin/v1` base and use `Cards.Fork(projectID, {catalog_ids: [...]})`
(with context as the first argument in Go). The 1..100 catalog IDs are forked
independently: inspect both `items` and `errors` even on HTTP 200.

The legacy batch-import method still targets its explicitly configured Control
API base. Current release routes do not register either that legacy batch route
or the engine batch route shown in the main API document. It is retained for
older deployments; use catalog fork for current project onboarding.

Delegated engine imports/listing need short-lived project delegation. Main docs
show an import `{row}`, the migration guide shows `{source, content, publication}`,
and the release handler accepts `{card, overwrite?}`. Import bodies pass through
unchanged; use the payload supported by the target deployment. No automatic
payload conversion or fallback to a second write route is performed.

### Documentation and route differences

The main document lists `POST /v1/cards/{id}/versions`, but release registers
only GET there. Submit versioned edits through `PATCH /v1/cards/{id}` using
`base_version`, `expected_latest_version`, and optional `change_note`.
It also lists `GET /v1/sessions/{id}/history` for groups, but the registered
history list is `GET /v1/sessions/{id}?offset=&limit=`; use the existing session
get method for both single and group history. These unregistered routes are
not exposed as new SDK methods.

Plain-text and JSON HTTP errors are supported, with 400/422 mapped to invalid.
JSON error code/message and the complete response payload are retained; a
debug-chat error can include session_id to query before retrying.
Ordinary project responses may set `prompt_redacted: true`; do not interpret
empty prompt/messages as missing generation. Traces require prompt-inspection
delegation and return 404 for an ordinary project token.

## Go usage and migration

Use `CardQuery{Limit: &limit, Offset: &offset}`,
`ListVersionsWithQuery(ctx, id, CardVersionQuery{Limit: &limit, BeforeVersion: &before})`,
`GetVersionWithQuery(ctx, id, version, CardQuery{Lang: "zh"})` and
`Cinema.ListRoundsWithQuery(ctx, sessionID, SessionGetQuery{Limit: &limit, Offset: &offset})`.
Existing version queries without query arguments still work.

Call `Cards.RestoreVersionWithBody(ctx, id, version, JSONMap{"expected_latest_version": latest})`.
The old bodyless RestoreVersion is deprecated and returns a migration error.

Configure `ClientConfig.ControlAPIBaseURL` with a full base such as
`https://rp.example.com/admin/v1`, then call
`Cards.Fork(ctx, projectID, JSONMap{"catalog_ids": []string{catalogID}})`.

`Cinema.ListRounds` now returns `[]JSONMap` to match the server array.
`Message.Content` and `HistoryMessage.Content` are now `any`: text decodes
as string and content arrays as `[]any`. Use a type switch when reading.
For outgoing images, use `[]ContentPart` with `ImageURL`.
`StreamEvent.ID` retains the SSE ID. Errors expose `Payload` and
`SessionID` as well as Code/Message/Status.

## Inspiration and translation reuse

`Sessions.Inspiration` posts `{}` or per-call `lang`, `model`,
`temperature`, `max_tokens`, `top_p`. It returns two suggestions with
`style=positive/neutral`, `text`, model and provider evidence. It does not modify
history/revision or create an operation. It calls the model on every request;
there is no idempotency or SSE variant. Thinking parameters, persona/card
fields and a body `session_id` are rejected by the server. Missing usage stays
null. Group sessions are unsupported.

`Cards.Untranslated` takes required `locale`, optional `limit`
(default 50, max 200), `offset`, and `include_partial`. Its `summary` covers
only the returned page. Explicit false and offset=0 are preserved.
`Cards.ForkTranslations` takes
`{items: [{card_id, card_version, locales: [...]}]}`; inspect per-item `errors`
and `action=created/skipped`. Existing edited translations are preserved.

## Group cards and sessions

Use `GroupCards` for story CRUD, immutable versions, members, imports,
preview, character inference, and owner translations. Lists, versions and
members return arrays; the version list currently returns at most 50 versions.
Group list filters are `q`, `listed`, `include_deleted`, `limit`, `offset`.
Get and version detail accept `lang`. Member removal is a DELETE with a JSON
body containing `user_id` and `expected_latest_version`; it soft-deletes the
member and preserves historic speaker attribution. Other roster changes pass
the documented body unchanged. Preview does not call a model; character
inference requires prompt-inspection permission.

Use `GroupSessions` to create a session with `group_id`, optional `group_version`,
`user_id`, `user`, `mode`, `output_format`, `active_participant_ids`,
`max_context`, `pinned_memory`. Sessions pin both story and member versions.
Group turns take `content`, optional `action`, `expected_revision`,
`addressed_participant_ids` and per-call sampling/prompt overrides. Supply a
stable `Idempotency-Key` request header for safe retries; without it, resending
can generate another paid turn. JSON responses flatten the session snapshot
and add `execution`, nullable `usage`, `operation` and `turn` (segments,
parser warnings, speaker validation and cost).

Group streaming uses the existing SSE parser: `operation`, `token`, `usage`,
`done`, `error`, plus `segment_start`, `segment_delta`, `segment_end` and
`parser_warning`. Data remains raw text; parse segment data as JSON.
Event IDs and `Last-Event-ID` pass through. A done frame contains the session
ID and canonical text on separate lines. Query the group turn after a disconnect;
handle errors and incomplete streams before treating text as committed.
Prompt lookup accepts `turn_id` and requires prompt-inspection permission.
Read group history and use edit/rewind/fork via the common session service.

## Cinema order and recovery updates

`Cinema.ListRoundsWithOrder` accepts `order=asc/desc`; order determines where offset counts
from, while each returned array remains in ascending sequence order.
Use desc with offset=0 and limit=1 to fetch the latest round.
A completed Cinema round ends with inner `message_end`, not a chat done event.
Failed rounds are not committed; inspect the operation and session before a
new attempt with a new key.

Every explicit recover `abandon` requires `acknowledge_unknown: true`.
Terminal operations are retained for a documented default of 7 hours; unresolved
operations use the session TTL (default 1 day). Idempotency ends when evidence
expires, so persist business results and complete retries within that window.
Use the actual operation `deadline_ms`, not the historical capabilities
`operation_deadline_seconds`, for lease decisions.
