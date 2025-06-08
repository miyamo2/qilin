---
title: Reading Resources
description: reading resources in qilin
---

Reading Resources is a feature of the Model Context Protocol (MCP) that allows clients to access data from your server. In Qilin, resources are identified by URIs and can be either static (fixed) or dynamic (parameterized).

## Registering Resources

To register a resource in Qilin, you use the `Resource` method on your Qilin instance. This method takes three required parameters:
1. The name of the resource (a string)
2. The URI of the resource (a string)
3. A handler function that processes the request and returns the resource data

```go
q.Resource("resource_name", "resource://example.com", func(c qilin.ResourceContext) error {
    // Resource handler logic here
    return nil
})
```

### Static URI Resource

Static resources are defined with a fixed URI and do not accept parameters.

```go
q.Resource(
    "get_employee",
    "example://example.com/1", 
    func(c qilin.ResourceContext) error {
        res := map[string]string{
            "id":      "1", 
            "name":  "Bob",
        }
        return c.JSON(res)
    })
```

In this example:
- `"get_employee"` is the name of the resource
- `"example://example.com/1"` is the fixed URI that clients will use to access this resource

### Dynamic URI Resource (Resource Template)

A Resource Template is a special type of resource that uses a parameterized URI pattern instead of a fixed URI. This allows a single resource handler to serve multiple different resources based on the parameters provided in the URI.

Dynamic resource URIs allow you to define parameterized endpoints with placeholders, such as `{id}`. These placeholders are replaced with actual values when the client requests the resource.

You can access these parameters in your handler by using `c.Param("id")`, as shown below.

```go
q.Resource(
    "get_employee",
    "example://example.com/{id}", 
    func(c qilin.ResourceContext) error {
        res := map[string]string{
            "id":      c.Param("id"),
            "name":  "Bob",
        }
        return c.JSON(res)
    })
```

In this example:
- The URI contains a placeholder `{id}` that can be replaced with any value
- The handler function uses `c.Param("id")` to access the value provided in the URI
- This allows a single resource handler to serve multiple different resources based on the parameter

Additionally, resources registered in this way are treated as resource templates.  
They are omitted from the resource list by default, so you must define [your own Resource List Handler](/qilin/guides/mcp/resources/listing/).

## Content Types

Qilin supports multiple content types for resources, allowing you to return different types of data to clients.

### JSON Content

`c.JSON(i any)` - Return a JSON content.

```go
func(c qilin.ResourceContext) error {
    res := map[string]string{
        "id":      c.Param("id"),
        "name":  "Bob",
    }
    return c.JSON(res)
}
```

### Plain Text Content

`c.String(s string)` - Return plain text content.

```go
func(c qilin.ResourceContext) error {
    return c.String("Hello, World!")
}
```

### Binary Content

`c.Blob(data []byte, mimeType string)` - Return binary data with a specified MIME type.

```go
func(c qilin.ResourceContext) error {
    data, err := os.ReadFile("image.png")
    if err != nil {
        return err
    }
    return c.Blob(data, "image/png")
}
```

## Options

You can provide more detailed resource information to clients by specifying options.

### With Description

Adding a description helps clients understand the purpose of the resource.

```go
q.Resource(
    "get_employee",
    "example://example.com/{id}",
    func(c qilin.ResourceContext) error {
        res := map[string]string{
            "id":      c.Param("id"),
            "name":  "Bob",
        }
        return c.JSON(res)
    }, qilin.ResourceWithDescription("Get an employee by ID"))
```

### With Mime Type

Specifying the MIME Type helps clients understand the format of the resource data.

```go
q.Resource(
    "get_employee",
    "example://example.com/{id}",
    func(c qilin.ResourceContext) error {
        res := map[string]string{
            "id":      c.Param("id"),
            "name":  "Bob",
        }
        return c.JSON(res)
    }, qilin.ResourceWithMimeType("application/json"))
```
