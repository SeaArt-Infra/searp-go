package rp

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSessionService_Create(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sessions" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %q", got)
		}
		writeJSON(w, 200, SessionSnapshot{ID: "session-1", Revision: 1})
	})
	defer srv.Close()

	session, err := client.Sessions.Create(context.Background(), JSONMap{"user_id": "u1", "card_id": "c1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.ID != "session-1" {
		t.Fatalf("unexpected session: %+v", session)
	}
}

func TestSessionService_Turn(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sessions/session-1/turns" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["action"] != "reply" {
			t.Fatalf("unexpected body: %v", body)
		}
		writeJSON(w, 200, TurnResponse{
			SessionSnapshot: SessionSnapshot{ID: "session-1", Revision: 2},
		})
	})

	turn, err := client.Sessions.Turn(context.Background(), "session-1", JSONMap{
		"action": "reply",
		"text":   "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if turn.Revision != 2 {
		t.Fatalf("unexpected turn response: %+v", turn)
	}
}

func TestSessionService_TurnStream(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions/session-1/turns" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: token\ndata: Hel\n\nevent: token\ndata: lo\n\nevent: done\ndata: session-1\n\n"))
	})

	events, err := client.Sessions.TurnStream(context.Background(), "session-1", JSONMap{"action": "reply", "text": "hello", "stream": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var tokens string
	done := false
	for event := range events {
		if event.Err != nil {
			t.Fatalf("unexpected stream error: %v", event.Err)
		}
		if event.Done {
			done = true
			continue
		}
		tokens += string(event.Data)
	}
	if !done || tokens != "Hel"+"lo" {
		t.Fatalf("unexpected stream result: done=%v tokens=%q", done, tokens)
	}
}

func TestSessionService_Proactive(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sessions/session-1/proactive" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["idempotency_key"] != "proactive-1" {
			t.Fatalf("unexpected body: %v", body)
		}
		writeJSON(w, 200, map[string]any{
			"id": "session-1", "revision": 3, "text": "Hello again!",
			"operation": map[string]any{"id": "proactive-1"},
		})
	})

	response, err := client.Sessions.Proactive(context.Background(), "session-1", JSONMap{"idempotency_key": "proactive-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Revision != 3 || response.Text != "Hello again!" || response.Operation["id"] != "proactive-1" {
		t.Fatalf("unexpected proactive response: %+v", response)
	}
}

func TestSessionService_ProactiveStream(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions/session-1/proactive" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: operation\ndata: {\"id\":\"proactive-1\"}\n\nevent: token\ndata: Hi\n\nevent: done\ndata: session-1\\nHi\n\n"))
	})

	events, err := client.Sessions.ProactiveStream(context.Background(), "session-1", JSONMap{"idempotency_key": "proactive-1", "stream": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var names []string
	for event := range events {
		if event.Err != nil {
			t.Fatalf("unexpected stream error: %v", event.Err)
		}
		names = append(names, event.Event)
	}
	if got := len(names); got != 3 || names[0] != "operation" || names[1] != "token" || names[2] != "done" {
		t.Fatalf("unexpected events: %v", names)
	}
}

func TestSessionService_ErrorBody(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "session not found", http.StatusNotFound)
	})
	_, err := client.Sessions.Get(context.Background(), "missing", SessionGetQuery{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	seaErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *rp.Error, got %T", err)
	}
	if seaErr.Kind != ErrNotFound || seaErr.Status != http.StatusNotFound {
		t.Fatalf("unexpected error: %+v", seaErr)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
