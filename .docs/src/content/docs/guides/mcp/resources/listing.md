---
title: Listing Resources
description: listing resources in qilin
---

Resource listing is a feature of the Model Context Protocol (MCP) that allows clients to discover what resources are available on your server. When a client connects to your MCP server, it typically requests a list of available resources to understand what data and functionality it can access.

## Default Resource List Handler

By default, Qilin provides a built-in resource list handler that automatically lists all statically registered resources. This default handler works well for most simple applications where all resources are known at startup time.

## Custom Resource List Handlers

If you are using [resource templates](/qilin/guides/mcp/resources/reading/#dynamic-uri-resource-resource-template), you must register a `ResourceListHandler` using the `q.ResourceList` method.
This is necessary because the [default resource list handler](https://pkg.go.dev/github.com/miyamo2/qilin#DefaultResourceListHandler) ignores resource templates.

```go
employees := []Employee{
    {ID: "1", Name: "Alice"},
    {ID: "2", Name: "Bob"},
}

q.ResourceList(myResourceListHandlerfunc)

func myResourceListHandlerfunc(c qilin.ResourceListContext) error {
    err := qilin.DefaultResourceListHandler(c)
    if err != nil {
        return err
    }
    // resolve resource template
    for i, v := range employees {
        uri, err := url.Parse(fmt.Sprintf("example://example.com/%d", i))
        if err != nil {
            return err
        }
        c.SetResource(uri.String(), qilin.Resource{
            URI:         (*qilin.ResourceURI)(uri),
            Name:        v.Name,
            Description: fmt.Sprintf("Employee %d", i),
            MimeType:    "application/json",
        })
    }
    return nil
}
```

In the example code:

1. Defines a sample list of employees that we want to expose as resources
2. Registers custom resource list handler function `myResourceListHandlerfunc` using `q.ResourceList()`
3. Inside the handler function:
     - First call the default handler to include all statically registered resources
     - Iterate through our employee list and create a resource for each employee
     - For each employee, generate a unique URI and set resource metadata
     - Use `c.SetResource()` to add each resource to the list that will be returned to clients

This approach allows you to dynamically generate resources based on your application's data, while still including all statically registered resources.
