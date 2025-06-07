# Qilin Copilot Instructions

## Coding Style

Must comply with the Uber Go Style Guide.  
https://github.com/uber-go/guide/blob/master/style.md

## Project Structure

```text
.
├── examples/                               # Example applications
│   └── weather/                            # Calculator example
│       ├── cmd/                            # Entry points
│       │   ├── weather-stdio/main.go       # Stdio transport example
│       │   └── weather-streamable/main.go  # HTTP transport example
│       ├── domain/                         # Domain logic
│       │   ├── model/                      # Domain models
│       │   └── repository/                 # Repository interfaces
│       ├── infrastructure/                 # Infrastructure code
│       │   └── api/weather.go              # API client for weather service
│       └── go.mod                          # Go module file
├── internal/                               # Internal packages
│   └── transport/transport.go              # Internal logic for transport packages
├── transport/
│   ├── example_test.go                     # Example Code for transport
│   ├── stdio.go                            # Stdio transport
│   ├── streamable.go                       # Streamable HTTP transport
│   └── transport.go                        # Shared logic for transport packages
├── context.go                              # ToolContext, ResourceContext, PromptContext etc.
├── context_test.go                         # test for context.go
├── errors.go                               # Error
├── example_test.go                         # Example Code
├── helpers_for_test.go                     # Helper functions for tests
├── qilin.go                                # User-facing Interface
├── session.go                              # Session management
├── subscription.go                         # Subscription management
├── types.go                                # Model Context Protocol types
└── go.mod
```

## Reference

- [Model Context Protocol Specification](https://modelcontextprotocol.io/specification/latest)
- [DeepWiki](https://deepwiki.com/miyamo2/qilin)
- [GoDoc](https://pkg.go.dev/github.com/miyamo2/qilin)