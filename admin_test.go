package rp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew_DefaultAdminBaseURL(t *testing.T) {
	client, err := New(&ClientConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.adminBaseURL != defaultAdminBaseURL {
		t.Fatalf("unexpected adminBaseURL: %s", client.adminBaseURL)
	}
}

func TestNew_DerivesAdminBaseFromBaseURL(t *testing.T) {
	client, err := New(&ClientConfig{
		APIKey:  "test-key",
		BaseURL: "https://engine.example.com/",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.adminBaseURL != "https://engine.example.com/admin/v1" {
		t.Fatalf("unexpected adminBaseURL: %s", client.adminBaseURL)
	}
}

func newAdminTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := New(&ClientConfig{
		APIKey:       "test-key",
		AdminBaseURL: srv.URL,
		Timeout:      5 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed to create admin test client: %v", err)
	}
	return srv, client
}

func TestAdminService_Health(t *testing.T) {
	_, client := newAdminTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/health" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, JSONMap{"ok": true})
	})

	result, err := client.Admin.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["ok"] != true {
		t.Fatalf("unexpected health result: %v", result)
	}
}

func TestAdminService_RequestAndProjectPaths(t *testing.T) {
	_, client := newAdminTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/projects":
			writeJSON(w, 200, JSONMap{"items": []any{}})
		case r.Method == http.MethodGet && r.URL.EscapedPath() == "/projects/a%2Fb":
			writeJSON(w, 200, JSONMap{"id": "a/b"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	if _, err := client.Admin.ListProjects(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	project, err := client.Admin.GetProject(context.Background(), "a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project["id"] != "a/b" {
		t.Fatalf("unexpected project: %v", project)
	}
}

func TestAdminService_ErrorEnvelope(t *testing.T) {
	_, client := newAdminTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 409, JSONMap{"error": JSONMap{"code": "conflict", "message": "revision mismatch"}})
	})

	_, err := client.Admin.UpdateProjectLive(context.Background(), "p1", JSONMap{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	seaErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *rp.Error, got %T", err)
	}
	if seaErr.Status != 409 || seaErr.Code != "conflict" || seaErr.Message != "revision mismatch" {
		t.Fatalf("unexpected error: %+v", seaErr)
	}
}
