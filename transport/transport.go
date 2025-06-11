package transport

import (
	"golang.org/x/exp/jsonrpc2"
)

const (
	MCPSessionID = "mcp-session-id"
)

// SessionIDHolder provides methods to get and set the session ID
type SessionIDHolder interface {
	// SessionID returns the session ID from the current connection.
	SessionID() string
	// SetSessionID sets the session ID for the current connection.
	SetSessionID(sessionID string)
}

// DefaultStdioFramer returns a default jsonrpc2.Framer for stdio transport.
func DefaultStdioFramer() jsonrpc2.Framer {
	return newStdioFramer()
}

// DefaultStreamableFramer returns a default jsonrpc2.Framer for streamable transport.
func DefaultStreamableFramer() jsonrpc2.Framer {
	return newStreamableFramer()
}

// Authorizer authorizes requests
type Authorizer interface {
	// Authorize checks if the request is authorized.
	// It should return an error if the request is not authorized.
	//
	//   ## params
	//
	//   - credential: Representing the credential to authorize. Provided from the authorization header.
	Authorize(credential string) error
}

// compatibility check
var _ Authorizer = (*noopAuthorizer)(nil)

type noopAuthorizer struct{}

func (n noopAuthorizer) Authorize(_ string) error {
	return nil
}

// DefaultAuthorizer returns a no-op authorizer that always allows requests.
func DefaultAuthorizer() Authorizer {
	return &noopAuthorizer{}
}

type ErrorNotifier interface {
	NoticeError(err error)
}
