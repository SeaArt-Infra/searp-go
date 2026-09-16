package rp

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestOperationService_Run(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sessions/session-1/operations" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["idempotency_key"] != "reply-1" {
			t.Fatalf("unexpected body: %v", body)
		}
		writeJSON(w, 200, JSONMap{"id": "reply-1", "status": "completed"})
	})

	op, err := client.Operations.Run(context.Background(), "session-1", JSONMap{
		"idempotency_key": "reply-1",
		"action":          "reply",
		"text":            "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if op["id"] != "reply-1" {
		t.Fatalf("unexpected operation: %v", op)
	}
}

func TestOperationService_ListQuery(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("session_id"); got != "session-1" {
			t.Fatalf("unexpected session_id: %q", got)
		}
		writeJSON(w, 200, OperationPage{Items: []OperationView{{"id": "op-1"}}})
	})
	page, err := client.Operations.List(context.Background(), OperationQuery{SessionID: "session-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
}
