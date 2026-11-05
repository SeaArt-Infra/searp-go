package rp

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestReleaseGroupCardRosterAndLocalization(t *testing.T) {
	var requests []string
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.EscapedPath())
		if r.Header.Get("X-Request-ID") != "req-group" {
			t.Error("lost caller header")
		}
		switch r.Method {
		case "GET":
			if r.URL.Path == "/group-cards" {
				if r.URL.Query().Get("listed") != "false" || r.URL.Query().Get("offset") != "0" || r.URL.Query().Get("q") != "harbor & stars" {
					t.Errorf("query: %s", r.URL.RawQuery)
				}
				writeJSON(w, 200, []JSONMap{{"id": "g1"}})
			} else if r.URL.Path == "/group-cards/g1/members" || r.URL.Path == "/group-cards/g1/versions" {
				writeJSON(w, 200, []JSONMap{})
			} else {
				if r.URL.Query().Get("lang") != "zh" {
					t.Errorf("lang: %s", r.URL.RawQuery)
				}
				writeJSON(w, 200, JSONMap{"title": "故事"})
			}
		case "DELETE":
			var body JSONMap
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["expected_latest_version"] != float64(3) || body["user_id"] != "owner" {
				t.Errorf("DELETE body: %v", body)
			}
			writeJSON(w, 200, JSONMap{"group_version": 4})
		default:
			writeJSON(w, 200, JSONMap{})
		}
	})
	ctx := context.Background()
	opt := WithHeader("X-Request-ID", "req-group")
	listed, offset := false, uint64(0)
	if out, err := client.GroupCards.List(ctx, GroupCardQuery{Q: "harbor & stars", Listed: &listed, Offset: &offset}, opt); err != nil || len(out) != 1 {
		t.Fatalf("list: %v %v", out, err)
	}
	if _, err := client.GroupCards.Get(ctx, "g1", GroupCardQuery{Lang: "zh"}, opt); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GroupCards.GetVersion(ctx, "g1", 3, GroupCardQuery{Lang: "zh"}, opt); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GroupCards.ListVersions(ctx, "g1", opt); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GroupCards.ListMembers(ctx, "g1", opt); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GroupCards.RemoveMember(ctx, "g1", "tipsy:one/two", JSONMap{"user_id": "owner", "expected_latest_version": 3}, opt); err != nil {
		t.Fatal(err)
	}
	if requests[len(requests)-1] != "DELETE /group-cards/g1/members/tipsy:one%2Ftwo" {
		t.Fatalf("escaped path: %v", requests)
	}
}

func TestReleaseGroupTurnEnvelopeAndSSE(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions/s1/group/turns" || r.Method != "POST" || r.Header.Get("Idempotency-Key") != "group-1" {
			t.Errorf("request: %s %s %v", r.Method, r.URL, r.Header)
		}
		if r.Header.Get("Accept") == "text/event-stream" {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("id: evt-1\nevent: operation\ndata: {\"id\":\"group-1\"}\n\nevent: segment_start\ndata: {\"segment_index\":0}\n\nevent: token\ndata:  Hi \n\nevent: segment_delta\ndata: {\"delta\":\"Hi\"}\n\nevent: segment_end\ndata: {\"speaker_valid\":true}\n\nevent: parser_warning\ndata: {\"code\":\"unknown_speaker\"}\n\nevent: usage\ndata: {\"total_tokens\":9}\n\nevent: done\ndata: s1\ndata:  Hi \n\n"))
			return
		}
		_, _ = w.Write([]byte(`{"id":"s1","revision":4,"usage":null,"execution":{"prompt_redacted":true},"operation":{"id":"group-1"},"turn":{"segments":[{"speaker_id":"tipsy:1","speaker_valid":true}],"parser_warnings":[],"cost":null}}`))
	})
	ctx := context.Background()
	opt := WithHeader("Idempotency-Key", "group-1")
	body := JSONMap{"content": "Hello", "expected_revision": 3, "addressed_participant_ids": []string{"tipsy:1"}}
	out, err := client.GroupSessions.Turn(ctx, "s1", body, opt)
	if err != nil || out.Revision != 4 || out.Usage != nil || out.Operation["id"] != "group-1" || len(out.Turn["segments"].([]any)) != 1 {
		t.Fatalf("envelope: %+v %v", out, err)
	}
	stream, err := client.GroupSessions.TurnStream(ctx, "s1", body, opt)
	if err != nil {
		t.Fatal(err)
	}
	var events []string
	for e := range stream {
		if e.Err != nil {
			t.Fatal(e.Err)
		}
		events = append(events, e.Event)
		if e.ID != "evt-1" {
			t.Errorf("lost event ID: %q", e.ID)
		}
		if e.Done && string(e.Data) != "s1\n Hi " {
			t.Errorf("done data: %q", e.Data)
		}
	}
	if !reflect.DeepEqual(events, []string{"operation", "segment_start", "token", "segment_delta", "segment_end", "parser_warning", "usage", "done"}) {
		t.Fatalf("events: %v", events)
	}
}

func TestReleaseInspirationAndTranslationQueries(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sessions/s1/inspiration":
			var body JSONMap
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body == nil || len(body) != 0 {
				t.Errorf("empty body: %v %v", body, err)
			}
			_, _ = w.Write([]byte(`{"replies":[{"style":"positive","text":"Yes!"},{"style":"neutral","text":"Maybe."}],"model":"m1","usage":null,"provider_id":null,"finish_reason":"stop"}`))
		case "/cards/untranslated":
			if r.URL.Query().Get("include_partial") != "false" || r.URL.Query().Get("locale") != "zh-CN" || r.URL.Query().Get("offset") != "0" {
				t.Errorf("query: %s", r.URL.RawQuery)
			}
			writeJSON(w, 200, JSONMap{"items": []JSONMap{}, "summary": JSONMap{"total_cards": 0}})
		case "/cards/translations/fork":
			writeJSON(w, 200, JSONMap{"items": []JSONMap{}, "errors": []JSONMap{{"code": "missing_source"}}})
		default:
			if r.URL.Query().Get("order") != "desc" {
				t.Errorf("order: %s", r.URL.RawQuery)
			}
			writeJSON(w, 200, []JSONMap{})
		}
	})
	ctx := context.Background()
	out, err := client.Sessions.Inspiration(ctx, "s1", nil)
	if err != nil || len(out.Replies) != 2 || out.Usage != nil || out.Replies[0].Style != "positive" {
		t.Fatalf("inspiration: %+v %v", out, err)
	}
	partial, offset := false, uint64(0)
	if _, err := client.Cards.Untranslated(ctx, UntranslatedCardsQuery{Locale: "zh-CN", IncludePartial: &partial, Offset: &offset}); err != nil {
		t.Fatal(err)
	}
	fork, err := client.Cards.ForkTranslations(ctx, JSONMap{"items": []JSONMap{{"card_id": "c1", "card_version": 3, "locales": []string{"zh"}}}})
	if err != nil || len(fork["errors"].([]any)) != 1 {
		t.Fatalf("fork: %v %v", fork, err)
	}
	if _, err := client.Cinema.ListRoundsWithOrder(ctx, "s1", CinemaRoundQuery{Order: "desc", Offset: &offset}); err != nil {
		t.Fatal(err)
	}
}

func TestReleaseWorldBooksCustomTaskAndTranslationUpdates(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cards/untranslated" {
			if r.URL.Query().Get("card_ids") != "c1,c2" || r.URL.Query().Get("locales") != "zh,ja" || r.URL.Query().Get("missing_only") != "true" {
				t.Errorf("untranslated query: %s", r.URL.RawQuery)
			}
		} else if r.URL.Path == "/sessions/s1/custom-task" {
			if r.Method != "POST" {
				t.Errorf("custom task method: %s", r.Method)
			}
		} else if r.URL.Path == "/cards/c1/translate" {
			if r.Method != "POST" {
				t.Errorf("translate method: %s", r.Method)
			}
		} else if r.URL.Path == "/world-books" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeJSON(w, 200, JSONMap{"books": []JSONMap{}, "items": []JSONMap{}})
	})
	ctx := context.Background()
	missing := true
	if _, err := client.Cards.Untranslated(ctx, UntranslatedCardsQuery{CardIDs: []string{"c1", "c2"}, Locales: []string{"zh", "ja"}, MissingOnly: &missing}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Cards.Translate(ctx, "c1", JSONMap{"lang": "zh"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Sessions.CustomTask(ctx, "s1", JSONMap{"pe": "summarize", "model": "m1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WorldBooks.List(ctx, WorldBookQuery{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WorldBooks.Create(ctx, JSONMap{"title": "Lore"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WorldBooks.Get(ctx, "book/1"); err != nil {
		t.Fatal(err)
	}
	if err := client.WorldBooks.Delete(ctx, "book/1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WorldBooks.CatalogList(ctx, WorldBookQuery{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WorldBooks.CatalogGet(ctx, "book/1"); err != nil {
		t.Fatal(err)
	}
}

func TestReleaseCatalogForkAndGenerationEvidence(t *testing.T) {
	srv, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/v1/projects/project-1/cards/fork" {
			if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer project-key" {
				t.Errorf("fork request: %s %v", r.Method, r.Header)
			}
			var body JSONMap
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if len(body["catalog_ids"].([]any)) != 1 {
				t.Errorf("body: %v", body)
			}
			writeJSON(w, 200, JSONMap{"items": []JSONMap{{"id": "c1"}}, "errors": []JSONMap{{"code": "not_found"}}})
			return
		}
		_, _ = w.Write([]byte(`{"id":"task-1","status":"failed","model":"m1","usage":{"cost":"0.01"},"error":{"code":"upstream_failed"}}`))
	})
	client, err := New(&ClientConfig{APIKey: "project-key", APIBaseURL: srv.URL, ControlAPIBaseURL: srv.URL + "/admin/v1"})
	if err != nil {
		t.Fatal(err)
	}
	out, err := client.Cards.Fork(context.Background(), "project-1", JSONMap{"catalog_ids": []string{"c1"}})
	if err != nil || len(out["errors"].([]any)) != 1 {
		t.Fatalf("fork response: %v %v", out, err)
	}
	task, err := client.Engine.GetGeneration(context.Background(), "task-1")
	if err != nil || task.Usage == nil || len(task.Error) == 0 {
		t.Fatalf("lost evidence: %+v %v", task, err)
	}
	withoutControl, _ := New(nil)
	if _, err := withoutControl.Cards.Fork(context.Background(), "project-1", JSONMap{}); err == nil {
		t.Fatal("missing Control API base must fail locally")
	}
}
