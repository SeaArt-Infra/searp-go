package rp

import (
	"context"
	"net/http"
	"testing"
)

func TestEngineService_Assemble(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/assemble" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, AssembleResponse{
			PromptTokens: 12,
			Messages:     []Message{{Role: "system", Content: "prompt"}},
		})
	})
	result, err := client.Engine.Assemble(context.Background(), JSONMap{
		"request": JSONMap{"character": JSONMap{"name": "Ada"}},
		"history": []any{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PromptTokens != 12 || len(result.Messages) != 1 {
		t.Fatalf("unexpected assemble result: %+v", result)
	}
}
