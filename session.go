package qilin

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type SessionStore interface {
	// Issue creates a new session and returns its ID.
	Issue(ctx context.Context) (sessionID string, err error)
	// Delete removes the session from the store.
	Delete(ctx context.Context, sessionID string) (err error)
	// Context returns the session-scoped context.
	Context(ctx context.Context, sessionID string) (sessionCtx context.Context, err error)
}

// SessionManager manages session
type SessionManager interface {
	// Start starts a session
	//
	//   ## params
	//
	//   - ctx: context
	//
	//   ## returns
	//
	//   - sessionID: the ID of the started session
	//   - err: an error if the session failed to start
	Start(ctx context.Context) (sessionID string, err error)

	// Context returns a session scoped context
	//
	//   ## params
	//
	//   - ctx: context
	//   - sessionID: session ID
	//
	//   ## returns
	//
	//   - ctx: a context that is scoped to the session
	//   - err: an error if failed to get the session context
	Context(ctx context.Context, sessionID string) (context.Context, error)

	// Discard discards a session
	Discard(ctx context.Context, sessionID string) error
}

var _ SessionStore = (*InMemorySessionStore)(nil)

// InMemorySessionStore is an in-memory implementation of SessionStore
//
// NOTE: It will only work properly if Qilin is running on a single server.
type InMemorySessionStore struct {
	_        struct{}
	sessions sync.Map
}

func (s *InMemorySessionStore) Issue(_ context.Context) (sessionID string, err error) {
	_uuid, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	id := _uuid.String()
	sessionCtx, cancel := context.WithCancel(context.Background())
	session := &stateFullSession{
		ctx:    sessionCtx,
		cancel: cancel,
	}
	s.sessions.LoadOrStore(id, session)
	return id, nil
}

func (s *InMemorySessionStore) Delete(_ context.Context, sessionID string) (err error) {
	v, loaded := s.sessions.LoadAndDelete(sessionID)
	if !loaded {
		return nil
	}
	session := v.(*stateFullSession)
	session.cancel()
	return nil
}

func (s *InMemorySessionStore) Context(
	_ context.Context,
	sessionID string,
) (sessionCtx context.Context, err error) {
	v, ok := s.sessions.Load(sessionID)
	if !ok {
		return nil, fmt.Errorf("session '%s' not found", sessionID)
	}
	session := v.(*stateFullSession)
	//nolint:staticcheck
	//lint:ignore SA1029 Tentative hack to create a simple child context.
	return context.WithValue(session.ctx, struct{}{}, struct{}{}), nil
}

// compatibility check
var _ SessionManager = (*sessionManager)(nil)

// sessionManager is the simplest implementation of SessionManager
type sessionManager struct {
	_ struct{}

	store SessionStore
}

type stateFullSession struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *sessionManager) Start(ctx context.Context) (string, error) {
	return s.store.Issue(ctx)
}

func (s *sessionManager) Context(ctx context.Context, sessionID string) (context.Context, error) {
	return s.store.Context(ctx, sessionID)
}

func (s *sessionManager) Discard(ctx context.Context, sessionID string) error {
	return s.store.Delete(ctx, sessionID)
}

// DefaultSessionManager default session manager
var DefaultSessionManager SessionManager = &sessionManager{
	store: &InMemorySessionStore{},
}
