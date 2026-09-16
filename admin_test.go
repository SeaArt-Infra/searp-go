package rp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

func newAdminTestService(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *adminService) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	service := &adminService{client: &transport.Client{
		APIKey:     "test-key",
		BaseURL:    srv.URL,
		UserAgent:  "searp-go/test",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}}
	if service.client == nil {
		t.Fatal("failed to create admin test service")
	}
	return srv, service
}

func TestAdminService_Health(t *testing.T) {
	_, service := newAdminTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/health" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, JSONMap{"ok": true})
	})

	result, err := service.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["ok"] != true {
		t.Fatalf("unexpected health result: %v", result)
	}
}

func TestAdminService_RequestAndProjectPaths(t *testing.T) {
	_, service := newAdminTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/projects":
			writeJSON(w, 200, JSONMap{"items": []any{}})
		case r.Method == http.MethodGet && r.URL.EscapedPath() == "/projects/a%2Fb":
			writeJSON(w, 200, JSONMap{"id": "a/b"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	if _, err := service.ListProjects(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	project, err := service.GetProject(context.Background(), "a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project["id"] != "a/b" {
		t.Fatalf("unexpected project: %v", project)
	}
}

func TestAdminService_ErrorEnvelope(t *testing.T) {
	_, service := newAdminTestService(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 409, JSONMap{"error": JSONMap{"code": "conflict", "message": "revision mismatch"}})
	})

	_, err := service.UpdateProjectLive(context.Background(), "p1", JSONMap{})
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

func TestAdminService_ProjectAndCatalogPaths(t *testing.T) {
	_, service := newAdminTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.EscapedPath() == "/projects/p1/versions" && r.URL.Query().Get("cursor") == "2":
			writeJSON(w, 200, JSONMap{"items": []any{}})
		case r.Method == http.MethodGet && r.URL.EscapedPath() == "/projects/p1/versions/v1/diff" && r.URL.Query().Get("against") == "v0":
			writeJSON(w, 200, JSONMap{"changes": []any{}})
		case r.Method == http.MethodGet && r.URL.Path == "/catalog/card-1/cover":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("png-bytes"))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.EscapedPath())
		}
	})

	if _, err := service.ListProjectVersions(context.Background(), "p1", map[string]string{"cursor": "2"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := service.DiffProjectVersion(context.Background(), "p1", "v1", "v0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cover, err := service.GetCatalogCardCover(context.Background(), "card-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(cover) != "png-bytes" {
		t.Fatalf("unexpected cover: %q", cover)
	}
}
