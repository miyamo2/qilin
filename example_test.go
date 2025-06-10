package qilin_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/miyamo2/qilin"
	"github.com/miyamo2/qilin/transport"
)

type Req struct {
	X float64 `json:"x" jsonschema:"title=X"`
	Y float64 `json:"y" jsonschema:"title=Y"`
}

type Res struct {
	Result float64 `json:"result"`
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

type authorizer struct {
	transport.Authorizer
}

func Example_withAuthorization() {
	q := qilin.New("calc")

	// add a tool, resource, or other components here

	// Create streamable transport with an authorizer
	streamable := transport.NewStreamable(
		transport.StreamableWithAuthorizer(&authorizer{}))

	q.Start(qilin.StartWithListener(streamable))
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
}

type Employee struct {
	ID   string `json:"id"   jsonschema:"title=ID"`
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
}

func ExampleQilin_ResourceChangeObserver() {
	q := qilin.New("employee_management")
	q.ResourceChangeObserver("example://example.com/{id}", func(c qilin.ResourceChangeContext) {
		// Handle resource change
		return
	})
}

func ExampleQilin_ResourceList() {
	q := qilin.New("employee_management")
	q.ResourceList(func(c qilin.ResourceListContext) error {
		i := 1
		for _, v := range c.Resources() {
			uri, err := url.Parse(
				strings.Replace((*url.URL)(v.URI).String(), "{id}", fmt.Sprintf("%d", i), 1))
			if err != nil {
				return err
			}
			c.SetResource("example://example.com/1", qilin.Resource{
				URI:         (*qilin.ResourceURI)(uri),
				Name:        v.Name,
				Description: fmt.Sprintf("Employee %d", i),
				MimeType:    "application/json",
			})
			i++
		}
		return nil
	})
}

func ExampleQilin_ResourceListChangeObserver() {
	q := qilin.New("employee_management")
	q.ResourceListChangeObserver(func(c qilin.ResourceListChangeContext) {
		// Handle resource list change
		return
	})
}

func ExampleQilin_Start() {
	q := qilin.New("calc")

	// add a tool, resource, or other components here

	q.Start() // listen and serve on stdio
}

func ExampleQilin_Start_withStdio() {
	q := qilin.New("calc")

	// add a tool, resource, or other components here

	listener := transport.NewStdio(context.Background())
	q.Start(qilin.StartWithListener(listener))
}

func ExampleQilin_Start_withStreamable() {
	q := qilin.New("calc")

	// add a tool, resource, or other components here

	listener := transport.NewStreamable()
	q.Start(qilin.StartWithListener(listener))
}

func ExampleStartWithListener_withStdio() {
	q := qilin.New("calc")

	// add a tool, resource, or other components here

	listener := transport.NewStdio(context.Background())
	q.Start(qilin.StartWithListener(listener))
}

func ExampleStartWithListener_withStreamable() {
	q := qilin.New("calc")

	// add a tool, resource, or other components here

	listener := transport.NewStreamable()
	q.Start(qilin.StartWithListener(listener))
}

func ExampleStartWithContext() {
	q := qilin.New("calc")

	// add a tool, resource, or other components here

	ctx := context.Background()
	q.Start(qilin.StartWithContext(ctx))
}

func ExampleStartWithReadySignal() {
	q := qilin.New("calc")

	// add a tool, resource, or other components here

	ready := make(chan struct{}, 1)
	q.Start(qilin.StartWithReadySignal(ready))

	// Wait for the server to be ready
	<-ready
	fmt.Println("Server is ready")
}
