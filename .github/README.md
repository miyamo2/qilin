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
  <a href="https://deepwiki.com/miyamo2/qilin">
    <img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki">
  </a>
  <a href="https://gitmcp.io/miyamo2/qilin">
    <img src="https://img.shields.io/endpoint?url=https://gitmcp.io/badge/miyamo2/qilin" alt="Ask DeepWiki">
  </a>
</h1>

## 🌟 Highlights

|            &nbsp;             | &nbsp;                                                                                                                 |
|:-----------------------------:|------------------------------------------------------------------------------------------------------------------------|
|   ⚡ **Zero‑config server**    | `qilin.New().Start()` launches an MCP server on **STDIN/STDOUT**                                                       |
| 🤝 **Familiar look and feel** | Handlers inspired by Go's well-known web application framework. Developers familiar with them will feel right at home. |
|     ⏩ **Streamable HTTP**     | Supports Streamable HTTP transport.                                                                                    |

## 🚀 Quick Start

```sh
go get github.com/miyamo2/qilin
```

```go
package main

import (
	"fmt"
	"github.com/miyamo2/qilin"
	"maps"
)

type OrderBeerRequest struct {
	BeerName string `json:"beer_name" jsonschema:"title=Beer Name"`
	Quantity int    `json:"quantity"  jsonschema:"title=Quantity of Beers"`
}

type OrderBeerResponse struct {
	Amount float64 `json:"amount"`
}

var beers = map[string]string{
	"IPA":   "A hoppy beer with a bitter finish.",
	"Stout": "A dark beer with a rich, roasted flavor.",
	"Lager": "A light, crisp beer with a smooth finish.",
}

func main() {
	q := qilin.New("beer hall", qilin.WithVersion("v0.1.0"))

	q.Resource("menu_list",
		"resources://beer_list",
		func(c qilin.ResourceContext) error {
			return c.JSON(maps.Keys(beers))
		})

	q.Tool("order_beer",
		(*OrderBeerRequest)(nil),
		func(c qilin.ToolContext) error {
			var req OrderBeerRequest
			if err := c.Bind(&req); err != nil {
				return err
			}
			_, ok := beers[req.BeerName]
			if !ok {
				return fmt.Errorf("beer %s not found", req.BeerName)
			}
			amount := 8 * req.Quantity // Assume unit cost of all beers is $8.00.
			return c.JSON(OrderBeerResponse{Amount: float64(amount)})
		})
	q.Start() // listen on stdio
}
```

For more detailed usage, please refer to the [Qilin User Guide](https://miyamo2.github.io/qilin/).

## 🛤 Roadmap

### Transports

- [x] Stdio
- [x] Streamable HTTP

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
    - [X] Subscriptions
- [ ] Prompt

## 📜 License

**Qilin** released under the [MIT License](https://github.com/miyamo2/qilin/blob/main/LICENSE)
