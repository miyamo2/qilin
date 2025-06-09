---
title: Stdio Transport
description: About the Stdio transport in Qilin.
---

By default, Qilin uses the [Stdio Transport](https://modelcontextprotocol.io/specification/2025-03-26/basic/transports#stdio).
This is the simplest way to run an MCP server.

## Basic Usage

When using Stdio, typically calling `q.Start()` is sufficient, as it's the default transport:

```go
q := qilin.New("beer hall")
q.Start() // Uses Stdio transport by default
```

## Explicit Configuration

If you need to specify the Stdio transport explicitly or provide a custom context, you can do so with the following code:

```go /transport.NewStdio/
q := qilin.New("beer hall")
listener := transport.NewStdio(context.Background())
q.Start(qilin.StartWithListener(listener))
```
