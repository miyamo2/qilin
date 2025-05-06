#  Qilin MCP Framework 🌩️🐲🌩️


<img align="center" width="500px" src="./.assets/logo.png" alt="logo" />

**Qilin**(麒麟) is a model context protocol framework written in Go.  

[![Go Reference](https://pkg.go.dev/badge/github.com/miyamo2/qilin.svg)](https://pkg.go.dev/github.com/miyamo2/qilin)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/miyamo2/qilin)](https://img.shields.io/github/go-mod/go-version/miyamo2/qilin)
[![Go Report Card](https://goreportcard.com/badge/github.com/miyamo2/qilin)](https://goreportcard.com/report/github.com/miyamo2/qilin)
[![GitHub License](https://img.shields.io/github/license/miyamo2/qilin?&color=blue)](https://img.shields.io/github/license/miyamo2/qilin?&color=blue)


> [!WARNING]
> 
> **Qilin** still **🚧WIP🚧**

## Getting Started

### Prerequisites

- Go 1.24+

### Installation

```bash
go get github.com/miyamo2/qilin
```
### Usage

```sh
package main

import (
	"fmt"
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
	q.Tool("add", "add y to x", (*Req)(nil), func(c qilin.ToolContext) error {
		var req Req
		c.Bind(&req)
		res := Res{
			Result: req.X + req.Y,
		}
		return c.JSON(res)
	})
	q.Tool("sub", "subtract y from x", (*Req)(nil), func(c qilin.ToolContext) error {
		var req Req
		c.Bind(&req)
		res := Res{
			Result: req.X - req.Y,
		}
		return c.JSON(res)
	})
	q.Tool("mul", "multiply x by y", (*Req)(nil), func(c qilin.ToolContext) error {
		var req Req
		c.Bind(&req)
		res := Res{
			Result: req.X * req.Y,
		}
		return c.JSON(res)
	})
	q.Tool("div", "divide x by y", (*Req)(nil), func(c qilin.ToolContext) error {
		var req Req
		c.Bind(&req)
		if req.Y == 0 {
			return fmt.Errorf("'Y' must not be 0")
		}
		res := Res{
			Result: req.X / req.Y,
		}
		return c.JSON(res)
	})
	q.Start() // listen and serve on stdio
}
```

### Support

#### Transports

- [x] Stdio
- [ ] SSE

#### Features

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