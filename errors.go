package rp

import "fmt"

const (
	ErrAuth     = "auth"
	ErrInvalid  = "invalid"
	ErrNotFound = "not_found"
	ErrConflict = "conflict"
	ErrQuota    = "quota"
	ErrTimeout  = "timeout"
	ErrNetwork  = "network"
	ErrGeneral  = "general"
)

// Error is the SDK error returned by SeaRP engine API calls.
// The engine returns plain-text error bodies and uses HTTP status codes as its
// public error contract, so Status is the primary machine-readable field.
type Error struct {
	Kind    string
	Code    string
	Message string
	Status  int
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Message == "" {
		return fmt.Sprintf("HTTP %d", e.Status)
	}
	return e.Message
}

func newHTTPError(status int, message string) *Error {
	return newHTTPErrorWithCode(status, "", message)
}

func newHTTPErrorWithCode(status int, code, message string) *Error {
	kind := ErrGeneral
	switch status {
	case 400:
		kind = ErrInvalid
	case 401, 403:
		kind = ErrAuth
	case 404:
		kind = ErrNotFound
	case 409:
		kind = ErrConflict
	case 429:
		kind = ErrQuota
	case 408, 504:
		kind = ErrTimeout
	}
	return &Error{Kind: kind, Code: code, Status: status, Message: message}
}
