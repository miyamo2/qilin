---
title: Streamable HTTP Transport
description: About the Streamable HTTP transport in Qilin.
---

Qilin supports the [Streamable HTTP transport](https://modelcontextprotocol.io/specification/2025-03-26/basic/transports#streamable-http) as defined in the MCP specification. This transport allows your MCP server to communicate over HTTP.

The Streamable HTTP transport is particularly useful when:
- You need to expose your MCP server as a web service
- You want to integrate with web-based clients
- You're building a service that will be accessed by multiple clients simultaneously

## Basic Setup

To use the Streamable HTTP transport, you need to create a new streamable listener and pass it to the `Start` method of your Qilin instance. Here's how to set it up:

```go
q := qilin.New("beer hall")
listener := transport.NewStreamable()
q.Start(qilin.StartWithListener(listener))
```

## Customize the address or `net.Listener`

By default, the Streamable HTTP transport listens on port 3001. You can customize the port or address settings using options when creating the listener.

```go
listener := transport.NewStreamable(transport.StreamableWithAddress("127.0.0.1:0"))
```

```go
netListener, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
    panic(err)
}
listener := transport.NewStreamable(transport.StreamableWithNetListener(netListener))
```

## Authorization

The Streamable HTTP transport supports authorization, allowing you to control access to your MCP server. To implement authorization, you need to create an authorizer that implements the `transport.Authorizer` interface:

```go
var _ transport.Authorizer = &authorizer{}

type authorizer struct{}

func (a *authorizer) Authorize(credential string) error {
    // Implement your authorization logic here.
    // The credential parameter contains the authorization token from the client.
    // Return nil if authorization is successful, or an error if it fails.
    return nil
}
```

Once you've created your authorizer, you can attach it to the Streamable HTTP transport using the `StreamableWithAuthorizer` option:

```go
streamable := transport.NewStreamable(transport.StreamableWithAuthorizer(&authorizer{}))
```

This will ensure that all requests to your MCP server are authorized before being processed. Clients will need to include an appropriate authorization header in their requests.
