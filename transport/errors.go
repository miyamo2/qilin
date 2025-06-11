package transport

import "errors"

var (
	// ErrSessionNotFound occurs when a session is not found.
	ErrSessionNotFound = errors.New("session not found")

	// ErrMissingSessionID occurs when a session ID is missing in the request.
	ErrMissingSessionID = errors.New("missing session id")
)
