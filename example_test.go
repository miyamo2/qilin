package qilin_test

import (
	"fmt"
	"github.com/miyamo2/qilin"
)

func Example() {
	type Req struct {
		X float64 `json:"x" jsonschema:"title=X"`
		Y float64 `json:"y" jsonschema:"title=Y"`
	}

	type Res struct {
		Result float64 `json:"result" jsonschema:"title=Result,description=The result of the operation"`
	}

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
	q.Start() // listen and server on stdio
}
