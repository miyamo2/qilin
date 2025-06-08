---
title: Calling Tools
description: calling tools in qilin
---

Calling Tools is a feature of the Model Context Protocol (MCP) that allows clients to interact with server-side functionality. In Qilin, tools can easily be registered via the `Tool` method.

## Registering Tools

To register a tool, you use the `Tool` method on your Qilin instance. This method takes three required parameters:
1. The name of the tool (a string)
2. A pointer to a nil instance of the request type (used for schema generation)
3. A handler function that processes the request and returns a response

```go
q.Tool("tool name", (*RequestSchema)(nil),
    func(c qilin.ToolContext) error {
        // Tool handler logic here
    })
```

## Binding Request Data

You can bind request data using the `c.Bind()` method. This method automatically decodes incoming request data into parameters.

```go
func(c qilin.ToolContext) error {
    var req Req
    if err := c.Bind(&req); err != nil {
        return err
    }
    // Tool handler logic here
}
```

Although there are no restrictions on the parameters, it is generally best to use the provided schema when registering the tool.

## Response Methods

Qilin provides several methods for returning different types of content from your tools. Each method is designed for a specific content type and format.

### Text Content (JSON)

`c.JSON(i any)` - Returns JSON-formatted data

```go
func(c qilin.ToolContext) error {
    var req Req
    if err := c.Bind(&req); err != nil {
        return err
    }
    // Process the request and generate a response
    result := map[string]interface{}{
        "message": "Hello, " + req.Name,
        "timestamp": time.Now().Unix(),
    }
    // Return the result as JSON
    return c.JSON(result)
}
```

### Text Content (String)

`c.String(s string)` - Returns plain text content

```go
func(c qilin.ToolContext) error {
    var req Req
    if err := c.Bind(&req); err != nil {
        return err
    }
    // Process the request and generate a text response
    result := fmt.Sprintf("Hello, %s! The current time is %s", 
        req.Name, time.Now().Format(time.RFC3339))
    // Return the result as a string
    return c.String(result)
}
```

### Image Content

`c.Image(data []byte, mimeType string)` - Returns image data with a specified MIME type

```go
func(c qilin.ToolContext) error {
    var req Req
    if err := c.Bind(&req); err != nil {
        return err
    }
    // Process the request and generate an image
    data, err := generateImage(req.Width, req.Height, req.Format)
    if err != nil {
        return fmt.Errorf("failed to generate image: %w", err)
    }
    // Return the image with appropriate MIME type
    return c.Image(data, "image/png")
}
```

### Audio Content

`c.Audio(data []byte, mimeType string)` - Returns audio data with a specified MIME type

```go
func(c qilin.ToolContext) error {
    var req Req
    if err := c.Bind(&req); err != nil {
        return err
    }
    // Process the request and generate audio data
    data, err := generateAudio(req.Text, req.Voice, req.Format)
    if err != nil {
        return fmt.Errorf("failed to generate audio: %w", err)
    }
    // Return the audio with appropriate MIME type
    return c.Audio(data, "audio/wav")
}
```

### Embedding Resource Content (Text; JSON)

`c.JSONResource(uri *url.URL, i any, mimeType string)` - Returns JSON data as an embedded resource with a URI

```go
func(c qilin.ToolContext) error {
    var req Req
    if err := c.Bind(&req); err != nil {
        return err
    }

    // Process the request and generate a resource URI
    uri, err := url.Parse(fmt.Sprintf("https://example.com/resources/%s", req.ID))
    if err != nil {
        return fmt.Errorf("failed to parse URI: %w", err)
    }

    // Generate the JSON data for the resource
    data := map[string]interface{}{
        "id": req.ID,
        "name": req.Name,
        "created_at": time.Now().Unix(),
    }

    // Return the JSON data as a resource with the URI
    return c.JSONResource(uri, data, "application/json")
}
```

### Embedding Resource Content (Text; String)

`c.StringResource(uri *url.URL, s string, mimeType string)` - Returns plain text data as an embedded resource with a URI

```go
func(c qilin.ToolContext) error {
    var req Req
    if err := c.Bind(&req); err != nil {
        return err
    }

    // Process the request and generate a resource URI
    uri, err := url.Parse(fmt.Sprintf("https://example.com/documents/%s", req.ID))
    if err != nil {
        return fmt.Errorf("failed to parse URI: %w", err)
    }

    // Generate the text content for the resource
    data := fmt.Sprintf("Document: %s\nCreated by: %s\nDate: %s", 
        req.Title, 
        req.Author, 
        time.Now().Format(time.RFC3339))

    // Return the text data as a resource with the URI
    return c.StringResource(uri, data, "text/plain")
}
```

### Embedding Resource Content (Blob)

`c.BinaryResource(uri *url.URL, data []byte, mimeType string)` - Returns binary data as an embedded resource with a URI

```go
func(c qilin.ToolContext) error {
    var req Req
    if err := c.Bind(&req); err != nil {
        return err
    }

    // Process the request and generate a resource URI
    uri, err := url.Parse(fmt.Sprintf("https://example.com/files/%s", req.ID))
    if err != nil {
        return fmt.Errorf("failed to parse URI: %w", err)
    }

    // Generate or retrieve the binary data for the resource
    data, err := getFileData(req.ID, req.Format)
    if err != nil {
        return fmt.Errorf("failed to get file data: %w", err)
    }

    // Return the binary data as a resource with the URI
    return c.BinaryResource(uri, data, "application/octet-stream")
}
```

## Options

You can provide more detailed tools information to clients by specifying options.

### With Description

```go
q.Tool(
    "example_tool",
    (*Req)(nil),
    func(c qilin.ToolContext) error {
        // tool logic here
        return c.JSON(map[string]string{"message": "Tool called successfully"})
    }, qilin.ToolWithDescription("This is an example tool"))
```

### With Annotations

Tool annotations provide additional metadata about your tool that can help clients understand its behavior and display it appropriately. The `ToolWithAnnotations` option allows you to specify the following properties.

#### Title

A user-friendly name for the tool that may be displayed in user interfaces.

```go
q.Tool(
    "fetch_weather_data",
    (*WeatherRequest)(nil),
    weatherHandler,
    qilin.ToolWithAnnotations(qilin.ToolAnnotations{
        Title: "Get Weather Forecast",
    }),
)
```

#### ReadOnlyHint

Indicates that this tool only reads data and doesn't modify any state (safe to call repeatedly).

```go
q.Tool(
    "get_user_profile",
    (*UserProfileRequest)(nil),
    userProfileHandler,
    qilin.ToolWithAnnotations(qilin.ToolAnnotations{
        ReadOnlyHint: true,
    }),
)
```

#### DestructiveHint

Indicates that this tool performs destructive operations that might be irreversible.

```go
q.Tool(
    "delete_account",
    (*DeleteAccountRequest)(nil),
    deleteAccountHandler,
    qilin.ToolWithAnnotations(qilin.ToolAnnotations{
        DestructiveHint: true,
    }),
)
```

#### IdempotentHint

Indicates that calling this tool multiple times with the same parameters will produce the same result.

```go
q.Tool(
    "update_user_preferences",
    (*UserPreferencesRequest)(nil),
    updatePreferencesHandler,
    qilin.ToolWithAnnotations(qilin.ToolAnnotations{
        IdempotentHint: true,
    }),
)
```

#### OpenWorldHint

Indicates that this tool interacts with external systems or resources.

```go
q.Tool(
    "search_web",
    (*WebSearchRequest)(nil),
    webSearchHandler,
    qilin.ToolWithAnnotations(qilin.ToolAnnotations{
        OpenWorldHint: true,
    }),
)
```