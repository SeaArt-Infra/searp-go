package rp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew_DefaultBaseURLs(t *testing.T) {
	client, err := New(&ClientConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.baseURL != defaultBaseURL {
		t.Fatalf("unexpected baseURL: %s", client.baseURL)
	}
	if client.apiBaseURL != defaultBaseURL+"/v1" {
		t.Fatalf("unexpected apiBaseURL: %s", client.apiBaseURL)
	}
}

func TestNew_DerivesAPIBaseFromBaseURL(t *testing.T) {
	client, err := New(&ClientConfig{
		APIKey:  "test-key",
		BaseURL: "https://engine.example.com/",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.baseURL != "https://engine.example.com" {
		t.Fatalf("unexpected baseURL: %s", client.baseURL)
	}
	if client.apiBaseURL != "https://engine.example.com/v1" {
		t.Fatalf("unexpected apiBaseURL: %s", client.apiBaseURL)
	}
}

func TestNew_KeepsExistingV1Path(t *testing.T) {
	client, err := New(&ClientConfig{
		APIKey:  "test-key",
		BaseURL: "https://engine.example.com/v1/",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.apiBaseURL != "https://engine.example.com/v1" {
		t.Fatalf("unexpected apiBaseURL: %s", client.apiBaseURL)
	}
}

func TestNew_InvalidBaseURL(t *testing.T) {
	_, err := New(&ClientConfig{APIKey: "test-key", BaseURL: "://bad"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNew_ClonesHeaders(t *testing.T) {
	headers := http.Header{"X-Infra-Project-Id": {"project-123"}}
	client, err := New(&ClientConfig{APIKey: "test-key", Headers: headers})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	headers.Set("X-Infra-Project-Id", "changed-after-new")
	if got := client.headers.Get("X-Infra-Project-Id"); got != "project-123" {
		t.Fatalf("expected cloned headers, got %q", got)
	}
}

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := New(&ClientConfig{
		APIKey:     "test-key",
		APIBaseURL: srv.URL,
		Timeout:    5 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return srv, client
}
