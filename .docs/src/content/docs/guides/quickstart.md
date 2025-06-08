---
title: Quick Start
description: Quick Start guide for Qilin MCP server framework.
---

## Installation

qilin requires Go 1.24 or later. You can install it using the following commands:

```sh
mkdir mymcp && cd mymcp
go mod init mymcp
go get github.com/miyamo2/qilin
```

## Building Simple MCP Server

Create `main.go`

```go
package main

import (
        "fmt"
        "github.com/miyamo2/qilin"
        "maps"
)

type OrderBeerRequest struct {
        BeerName string `json:"beer_name" jsonschema:"title=Beer Name"`
        Quantity int      `json:"quantity" jsonschema:"title=Quantity of Beers"`
}

type OrderBeerResponse struct {
    Amount float64 `json:"amount"`
}

var beers = map[string]string{
        "IPA":     "A hoppy beer with a bitter finish.",
        "Stout": "A dark beer with a rich, roasted flavor.",
        "Lager": "A light, crisp beer with a smooth finish.",
}

func main() {
    q := qilin.New("beer hall", qilin.WithVersion("v0.1.0"))

    q.Resource("menu_list", "resources://beer_list", beerListHandler)

    q.Tool("order_beer", (*OrderBeerRequest)(nil), orderBeerHandler,
        qilin.ToolWithDescription("Order a beer"))

    q.Start()
}

func beerListHandler(c qilin.ResourceContext) error {
    return c.JSON(maps.Keys(beers))
}

func orderBeerHandler(c qilin.ToolContext) error {
    var req OrderBeerRequest
    if err := c.Bind(&req); err != nil {
        return err
    }
    _, ok := beers[req.BeerName]
    if !ok {
        return fmt.Errorf("beer %s not found", req.BeerName)
    }
    var price float64
    switch req.BeerName {
    case "IPA":
        price = 8.0
    case "Stout":
        price = 9.0
    case "Lager":
        price = 7.0
    }
    amount := price * float64(req.Quantity)
    return c.JSON(OrderBeerResponse{Amount: amount})
}
```

Start the mcp server on stdio.

```sh
go run main.go
```

## Debugging

You can use the MCP inspector to debug the server.

```sh
cd <path to your project root>
npx @modelcontextprotocol/inspector go run main.go
```

## Install your MCP server on the Claude Desktop

Edit Claude Desktop's configuration file to include your MCP server.  

- macOS: `~/Library/ApplicationSupport/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
    "mcpServers": {
        "beer_hall": {
            "command": "go",
            "args": ["run", "/path/to/your/main.go"]
        }
    }
}
```
