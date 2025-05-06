<h1 align="center">
  <picture>
      <img height="200" alt="Qilin Logo" src="https://raw.githubusercontent.com/miyamo2/qilin/refs/heads/main/.assets/logo.png">
  </picture>
  <p>Qilin 🌩️🐲🌩️ – Model Context Protocol Framework for Go</p>
  <a href="https://pkg.go.dev/github.com/miyamo2/qilin">
    <img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/miyamo2/qilin.svg" />
  </a>
  <img alt="Go Version" src="https://img.shields.io/github/go-mod/go-version/miyamo2/qilin" />
  <a href="https://goreportcard.com/report/github.com/miyamo2/qilin">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/miyamo2/qilin" />
  </a>
  <a href="https://github.com/miyamo2/qilin/blob/main/LICENSE">
    <img alt="License MIT" src="https://img.shields.io/github/license/miyamo2/qilin?&color=blue" />
  </a>
  <img alt="Status WIP" src="https://img.shields.io/badge/status-WIP-orange" />
</h1>

## ✨ Features

|            &nbsp;             | &nbsp;                                                                                                                 |
|:-----------------------------:|------------------------------------------------------------------------------------------------------------------------|
|   ⚡ **Zero‑config server**    | `qilin.New().Start()` launches an MCP server on **STDIN/STDOUT**                                                       |
| 👀 **Familiar look and feel** | Handlers inspired by Go's well-known web application framework. Developers familiar with them will feel right at home. |

## 🚀 Quick Start

```sh
go get github.com/miyamo2/qilin
```

```go
package main

import (
  "github.com/miyamo2/qilin"
)

type Req struct {
  X float64 `json:"x" jsonschema:"title=X"`
  Y float64 `json:"y" jsonschema:"title=Y"`
}

type Res struct {
  Result float64 `json:"result" jsonschema:"title=Result,description=The result of the operation"`
}

func main() {
  q := qilin.New("calc")

  q.Tool("add", (*Req)(nil), func(c qilin.ToolContext) error {
    var req Req
    c.Bind(&req)
    return c.JSON(Res{Result: req.X + req.Y})
  })

  q.Tool("sub", (*Req)(nil), func(c qilin.ToolContext) error {
    var req Req
    c.Bind(&req)
    return c.JSON(Res{Result: req.X - req.Y})
  }, qilin.ToolWithDescription("subtract y from x"))

  q.Start() // listens & serves on stdio
}
```

## 🛤 Roadmap

### Transports

- [x] Stdio
- [ ] SSE

### Features

- [x] Tool
  - [X] Listing
  - [X] Calling
    - [X] Middleware
- [x] Resource
  - [X] Listing
  - [X] Reading
    - [X] Middleware
  - [X] Templates
  - [X] List Changed Notification
  - [ ] Subscriptions
- [ ] Prompt

## 📜 License

**Qilin** released under the [MIT License](https://github.com/miyamo2/qilin/blob/main/LICENSE)