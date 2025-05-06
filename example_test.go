package qilin_test

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

func Example() {
	q := qilin.New("calc")
	q.Tool("add", (*Req)(nil), func(c qilin.ToolContext) error {
		var req Req
		c.Bind(&req)
		res := Res{
			Result: req.X + req.Y,
		}
		return c.JSON(res)
	})
	q.Start() // listen and serve on stdio
}

func ExampleQilin_Tool() {
	q := qilin.New("calc")
	q.Tool("add", (*Req)(nil), func(c qilin.ToolContext) error {
		var req Req
		c.Bind(&req)
		res := Res{
			Result: req.X + req.Y,
		}
		return c.JSON(res)
	})
	q.Start() // listen and serve on stdio
}

func ExampleToolWithDescription() {
	q := qilin.New("calc")
	q.Tool("add", (*Req)(nil), func(c qilin.ToolContext) error {
		var req Req
		c.Bind(&req)
		res := Res{
			Result: req.X + req.Y,
		}
		return c.JSON(res)
	}, qilin.ToolWithDescription("add y to x"))
	q.Start() // listen and serve on stdio
}

func ExampleToolWithAnnotations() {
	q := qilin.New("calc")
	q.Tool("add", (*Req)(nil), func(c qilin.ToolContext) error {
		var req Req
		c.Bind(&req)
		res := Res{
			Result: req.X + req.Y,
		}
		return c.JSON(res)
	}, qilin.ToolWithAnnotations(qilin.ToolAnnotations{
		Title:           "Calculation of x add y",
		ReadOnlyHint:    true,
		DestructiveHint: false,
		IdempotentHint:  false,
		OpenWorldHint:   false,
	}))
	q.Start() // listen and serve on stdio
}

func ExampleToolWithMiddleware() {
	q := qilin.New("calc")
	q.Tool("add", (*Req)(nil), func(c qilin.ToolContext) error {
		var req Req
		c.Bind(&req)
		res := Res{
			Result: req.X + req.Y,
		}
		return c.JSON(res)
	}, qilin.ToolWithMiddleware(func(next qilin.ToolHandlerFunc) qilin.ToolHandlerFunc {
		return func(c qilin.ToolContext) error {
			// do something before the handler
			return next(c)
		}
	}))
	q.Start() // listen and serve on stdio
}

type Employee struct {
	ID   string `json:"id" jsonschema:"title=ID"`
	Name string `json:"name" jsonschema:"title=Name"`
}

func ExampleQilin_Resource() {
	q := qilin.New("employee_management")
	q.Resource("get_employee", "example://example.com/{id}", func(c qilin.ResourceContext) error {
		c.Param("id")
		res := Employee{
			ID:   c.Param("id"),
			Name: "Bob",
		}
		return c.JSON(res)
	})
	q.Start() // listen and serve on stdio
}

func ExampleResourceWithDescription() {
	q := qilin.New("employee_management")
	q.Resource("get_employee", "example://example.com/{id}", func(c qilin.ResourceContext) error {
		c.Param("id")
		res := Employee{
			ID:   c.Param("id"),
			Name: "Bob",
		}
		return c.JSON(res)
	}, qilin.ResourceWithDescription("Get employee by ID"))
	q.Start() // listen and serve on stdio
}

func ExampleResourceWithMimeType() {
	q := qilin.New("employee_management")
	q.Resource("get_employee", "example://example.com/{id}", func(c qilin.ResourceContext) error {
		c.Param("id")
		res := Employee{
			ID:   c.Param("id"),
			Name: "Bob",
		}
		return c.JSON(res)
	}, qilin.ResourceWithMimeType("application/json"))
	q.Start() // listen and serve on stdio
}
