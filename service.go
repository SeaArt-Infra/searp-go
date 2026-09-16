package rp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/SeaArt-Infra/searp-go/internal/transport"
)

func requestJSON(ctx context.Context, c *transport.Client, method, path string, body any, headers http.Header, out any) error {
	status, payload, err := c.Request(ctx, method, path, body, headers)
	if err != nil {
		return publicError(err)
	}
	if status >= 400 {
		return parseRawResponse(status, payload)
	}
	if out == nil || len(payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return &Error{Kind: ErrGeneral, Message: "failed to decode response: " + err.Error()}
	}
	return nil
}

func streamJSON(ctx context.Context, c *transport.Client, method, path string, body any, headers http.Header) (<-chan StreamEvent, error) {
	resp, err := c.RequestStream(ctx, method, path, body, headers)
	if err != nil {
		return nil, publicError(err)
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		payload, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, &Error{Kind: ErrGeneral, Message: "failed to read stream error response: " + readErr.Error()}
		}
		return nil, parseRawResponse(resp.StatusCode, payload)
	}

	ch := make(chan StreamEvent, 8)
	go func() {
		defer close(ch)
		transport.StreamResponse(ctx, resp, func(ctx context.Context, event *transport.StreamEvent) bool {
			out := StreamEvent{Event: event.Event}
			if event.Err != nil {
				out.Err = publicError(event.Err)
			}
			if event.Done || event.Event == "done" {
				out.Done = true
			}
			if len(event.Data) > 0 {
				out.Data = RawResponse(event.Data)
			}
			select {
			case <-ctx.Done():
				return false
			case ch <- out:
				return true
			}
		})
	}()
	return ch, nil
}

func publicError(err error) error {
	var transportErr *transport.Error
	if errors.As(err, &transportErr) {
		return newHTTPError(transportErr.Status, transportErr.Message)
	}
	return err
}

func headersFromOptions(opts []RequestOption) http.Header {
	return buildRequestOptions(opts).headers
}
