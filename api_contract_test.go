package rp

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestCardPaginationLocalizationAndTranslations(t *testing.T) {
	ctx := context.Background()
	var requests []string
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		if r.Method == http.MethodPost {
			var body JSONMap
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["expected_latest_version"] != float64(7) {
				t.Fatalf("missing CAS: %v", body)
			}
			if r.URL.Path == "/cards/c1/versions/3/translations" {
				if body["user_id"] != "u1" || body["lang"] != "zh" || body["translations"].(map[string]any)["name"] != "艾达" {
					t.Fatalf("unexpected translation: %v", body)
				}
			}
		}
		if r.URL.Path == "/cards" || r.URL.Path == "/users/u1/cards" {
			writeJSON(w, 200, []JSONMap{})
			return
		}
		writeJSON(w, 200, JSONMap{"ok": true})
	})
	limit, offset, before := uint64(20), uint64(0), uint64(7)
	if _, err := client.Cards.List(ctx, CardQuery{Limit: &limit, Offset: &offset}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Cards.ListByUser(ctx, "u1", CardQuery{Limit: &limit, Offset: &offset}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Cards.ListVersionsWithQuery(ctx, "c1", CardVersionQuery{Limit: &limit, BeforeVersion: &before}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Cards.GetVersionWithQuery(ctx, "c1", 3, CardQuery{Lang: "zh"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Cards.ListTranslations(ctx, "c1", 3); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Cards.SaveTranslations(ctx, "c1", 3, JSONMap{"user_id": "u1", "expected_latest_version": 7, "lang": "zh", "translations": JSONMap{"name": "艾达"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Cards.RestoreVersionWithBody(ctx, "c1", 3, JSONMap{"expected_latest_version": 7, "change_note": "restore"}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET /cards?limit=20&offset=0", "GET /users/u1/cards?limit=20&offset=0",
		"GET /cards/c1/versions?before_version=7&limit=20", "GET /cards/c1/versions/3?lang=zh",
		"GET /cards/c1/versions/3/translations", "POST /cards/c1/versions/3/translations", "POST /cards/c1/versions/3/restore",
	}
	if !reflect.DeepEqual(requests, want) {
		t.Fatalf("requests: %v", requests)
	}
}

func TestMultimodalHistoryAndLatestResponseFields(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/sessions/s1/history/0" {
			_, _ = w.Write([]byte(`{"index":0,"total":2,"role":"user","content":[{"type":"image_url","image_url":{"url":"https://cdn.example/a.png"}}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"s1","history":[{"role":"assistant","content":"hello"},{"role":"user","content":[{"type":"text","text":"what?"},{"type":"image_url","image_url":{"url":"https://cdn.example/a.png","detail":"auto"}}]}],"prompt_redacted":true,"execution":{"prompt":{"user":"Visitor","warnings":["fallback"]},"messages":[],"prompt_tokens":100,"prompt_redacted":true},"history_import":{"seeded":{"messages":2,"dropped":1}}}`))
	})
	out, err := client.Sessions.Turn(context.Background(), "s1", JSONMap{"action": "reply", "message": []ContentPart{{Type: "image_url", ImageURL: &ImageURL{URL: "https://cdn.example/a.png"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if out.History[0].Content != "hello" || len(out.History[1].Content.([]any)) != 2 || !out.PromptRedacted || !out.Execution.PromptRedacted || out.Execution.Prompt.User != "Visitor" || len(out.Execution.Prompt.Warnings) != 1 || len(out.HistoryImport) == 0 {
		t.Fatalf("lost fields: %+v", out)
	}
	message, err := client.Sessions.HistoryMessage(context.Background(), "s1", 0)
	if err != nil || len(message.Content.([]any)) != 1 {
		t.Fatalf("history: %+v, %v", message, err)
	}
}

func TestLiveMetricsAndDebugErrorEvidence(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/live":
			writeJSON(w, 200, JSONMap{"status": "live"})
		case "/metrics":
			_, _ = w.Write([]byte("searp_requests_total 1\n"))
		default:
			writeJSON(w, 422, JSONMap{"error": JSONMap{"code": "invalid", "message": "unknown field"}, "session_id": "s1"})
		}
	})
	if out, err := client.Engine.Live(context.Background()); err != nil || out["status"] != "live" {
		t.Fatalf("live: %v %v", out, err)
	}
	if out, err := client.Engine.Metrics(context.Background()); err != nil || out != "searp_requests_total 1\n" {
		t.Fatalf("metrics: %q %v", out, err)
	}
	_, err := client.Engine.DebugChat(context.Background(), JSONMap{})
	seaErr, ok := err.(*Error)
	if !ok || seaErr.Kind != ErrInvalid || seaErr.Code != "invalid" || seaErr.Message != "unknown field" || seaErr.SessionID != "s1" || seaErr.Payload == nil {
		t.Fatalf("error: %+v", err)
	}
}

func TestCinemaStreamIDsAndWhitespace(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "text/event-stream" || r.Header.Get("Last-Event-ID") != "evt-1" || r.Header.Get("Idempotency-Key") != "round-1" {
			t.Fatalf("headers: %v", r.Header)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("id: evt-2\nevent: cinema\ndata:  leading and trailing  \n\nevent: done\ndata: finished"))
	})
	events, err := client.Cinema.CreateRoundStream(context.Background(), "s1", JSONMap{"content": "go"}, WithHeader("Last-Event-ID", "evt-1"), WithHeader("Idempotency-Key", "round-1"))
	if err != nil {
		t.Fatal(err)
	}
	var seen []StreamEvent
	for event := range events {
		if event.Err != nil {
			t.Fatal(event.Err)
		}
		seen = append(seen, event)
	}
	if len(seen) != 2 || seen[0].ID != "evt-2" || string(seen[0].Data) != " leading and trailing  " || !seen[1].Done || string(seen[1].Data) != "finished" {
		t.Fatalf("events: %+v", seen)
	}
}

func TestCinemaListRoundsReturnsArrayAndPages(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions/s1/cinema/rounds" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		if r.URL.RawQuery != "" && r.URL.RawQuery != "limit=20&offset=0" {
			t.Fatalf("query: %s", r.URL.RawQuery)
		}
		writeJSON(w, 200, []JSONMap{{"round_id": "r1", "cost": "0.01"}})
	})
	ctx := context.Background()
	limit, offset := uint64(20), uint64(0)
	rounds, err := client.Cinema.ListRoundsWithQuery(ctx, "s1", SessionGetQuery{Limit: &limit, Offset: &offset})
	if err != nil || len(rounds) != 1 || rounds[0]["round_id"] != "r1" {
		t.Fatalf("rounds: %v %v", rounds, err)
	}
	if rounds, err := client.Cinema.ListRounds(ctx, "s1"); err != nil || len(rounds) != 1 {
		t.Fatalf("rounds: %v %v", rounds, err)
	}
}

func TestBatchImportUsesExplicitControlAPIAndProjectToken(t *testing.T) {
	srv, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin/v1/cards/import/batch" || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer project-key" || r.Header.Get("X-Request-ID") != "req-1" {
			t.Fatalf("unexpected batch request: %s %s %v", r.Method, r.URL.Path, r.Header)
		}
		var body JSONMap
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body["cards"].([]any)) != 1 {
			t.Fatalf("body: %v", body)
		}
		writeJSON(w, 201, JSONMap{"items": []JSONMap{{"id": "c1"}}})
	})
	client, err := New(&ClientConfig{APIKey: "project-key", APIBaseURL: "https://engine.example/v1", ControlAPIBaseURL: srv.URL + "/admin/v1/"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Cards.ImportBatch(context.Background(), JSONMap{"cards": []JSONMap{{"name": "Ada"}}}, WithHeader("X-Request-ID", "req-1"))
	if err != nil || len(result["items"].([]any)) != 1 {
		t.Fatalf("batch: %v %v", result, err)
	}
	withoutControl, _ := New(nil)
	if _, err := withoutControl.Cards.ImportBatch(context.Background(), JSONMap{}); err == nil {
		t.Fatal("expected missing control URL error")
	}
}
