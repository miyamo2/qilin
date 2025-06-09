# Project Guidelines

## Coding Style

Must comply with the Uber Go Style Guide.  
https://github.com/uber-go/guide/blob/master/style.md

## Project Structure

## Project Structure

```text
.
├── .assets/                                # Assets for the project
├── .docs/                                  # User Guide. See Under `.docs` section for details.
├── .github/                                # GitHub related files
│   ├── workflows/                          # GitHub Actions workflows
│   ├── CODE_OF_CONDUCT.md                  # Code of Conduct
│   ├── copilot-instructions.md             # Copilot instructions
│   ├── README.md                           # Issue templates
│   └── renovate.json                       # Renovate configuration
├── .junie/                                 # Junie related files
│   └── guidelines.md                       # Project guidelines
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

### Under `.docs`

directory tree under `.docs`

```text
.
├── src/                                
│   ├── content/                                 # Content for the documentation
│   │   ├── docs/                                # Documentation files
│   │   │   ├── guides/                          # Guides for using Qilin
│   │   │   │   ├── mcp/                         # MCP related guides
│   │   │   │   │   ├── resources/               # Resources related guides
│   │   │   │   │   │   ├── list_changed.md      # Resource List Changed Notification guide
│   │   │   │   │   │   ├── listing.md           # Listing resources guide
│   │   │   │   │   │   ├── reading.md           # Reading resources guide
│   │   │   │   │   │   └── subscribe.md         # Resource Subscribing guide
│   │   │   │   │   ├── sessions/                # Sessions related guides
│   │   │   │   │   │   └── manage.md            # Managing sessions guide
│   │   │   │   │   └── tools/                   # Tools related guides
│   │   │   │   │       └── calling.md           # Calling tools guide
│   │   │   │   └── transports/                  # Transports related guides
│   │   │   │       ├── stdio.md                 # Stdio transport guide
│   │   │   │       └── streamable.md            # Streamable transport guide
│   │   │   ├── index.mdx                        # Index page for documentation
│   │   │   └── introduction.md                  # Introduction to Qilin
│   │   └──  content.config.ts                   # Configuration for the content
│   └── assets/                                  # Assets for the documentation
│       └── logo.webp                            # Logo for the documentation
├── astro.config.mjs                             # Astro configuration file
├── package.json                                 # Package configuration file
├── package-lock.json                            # Package lock file
└── tsconfig.json                                # TypeScript configuration file
```

## Reference

- [Model Context Protocol Specification](https://modelcontextprotocol.io/specification/latest)
- [Qilin User Guide](https://miyamo2.github.io/qilin/)
- [DeepWiki](https://deepwiki.com/miyamo2/qilin)
- [GoDoc](https://pkg.go.dev/github.com/miyamo2/qilin)