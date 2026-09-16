package transport

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
)

// StreamEvent is the transport-level SSE event representation.
type StreamEvent struct {
	Event string
	Data  []byte
	Done  bool
	Err   error
}

// StreamResponse parses server-sent events from resp.Body until EOF, context
// cancellation, or emit returns false. It closes the response body.
func StreamResponse(ctx context.Context, resp *http.Response, emit func(ctx context.Context, event *StreamEvent) bool) {
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	eventName := ""
	dataLines := make([]string, 0, 4)

	flush := func() bool {
		if len(dataLines) == 0 && eventName == "" {
			return true
		}
		event := &StreamEvent{Event: eventName, Data: []byte(strings.Join(dataLines, "\n"))}
		eventName = ""
		dataLines = dataLines[:0]
		if string(event.Data) == "[DONE]" {
			event.Done = true
			event.Data = nil
		}
		if event.Done || len(event.Data) > 0 || event.Event != "" {
			return emit(ctx, event)
		}
		return true
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				flush()
				return
			}
			emit(ctx, &StreamEvent{Err: &Error{Kind: "network", Message: "stream read failed: " + err.Error()}})
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if !flush() {
				return
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "event:"):
			eventName = strings.TrimSpace(line[len("event:"):])
		case strings.HasPrefix(line, "data:"):
			dataLines = append(dataLines, strings.TrimSpace(line[len("data:"):]))
		}
	}
}
