---
title: Session Management
description: Managing sessions in Qilin.
---

Session management is a feature of the Model Context Protocol (MCP) Streamable HTTP Transport that allows your server to maintain state across multiple requests from the same client. In Qilin, sessions are managed automatically by default, but you can customize the behavior if needed.

## Default Session Management

By default, Qilin uses an in-memory session store and manager. This works well with applications running on a single server and requires no additional configuration.

```go
q := qilin.New("my_service")
// Sessions are automatically managed
q.Start()
```

When a client connects to your MCP server, a session is automatically created and maintained throughout the client's interaction with your server.

## Custom Session Store

If you need to customize how sessions are stored (for example, to persist sessions to a database), you can implement the `SessionStore` interface and provide it to your Qilin instance:

```go
// compatibility check
var _ qilin.SessionStore = &MyCustomSessionStore{}

type MyCustomSessionStore struct {
    // Your custom fields here
}

// Implement the SessionStore interface methods
func (s *MyCustomSessionStore) Issue(ctx context.Context) (sessionID string, err error) {
    // Your implementation here
    return 
}

func (s *MyCustomSessionStore) Delete(ctx context.Context, sessionID string) (err error) {
    // Your implementation here
    return 
}

func (s *MyCustomSessionStore) Context(ctx context.Context, sessionID string) (sessionCtx context.Context, err error) {
    // Your implementation here
    return 
}
```

Use your custom session store.

```go
q := qilin.New("my_service", qilin.WithSessionStore(&MyCustomSessionStore{}))
q.Start()
```

## Session Lifecycle

Sessions in Qilin follow this lifecycle:

1. **Creation**: When a client connects, a new session is created with a unique ID
2. **Identification**: The session ID is sent to the client and included in subsequent requests
3. **Retrieval**: For each request, the session is retrieved using the session ID
4. **Termination**: When the client disconnects the session is discarded
